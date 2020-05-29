package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blang/semver"
	"github.com/poundbot/poundbot/pkg/models"
)

type raidAlertAdder interface {
	AddInfo(alertIn, validUntil time.Duration, ed models.EntityDeath) error
}

type entityDeath struct {
	raa        raidAlertAdder
	minVersion semver.Version
}

func initEntityDeath(api muxFuncHandler, path string, raa raidAlertAdder) {
	ed := entityDeath{raa: raa, minVersion: semver.Version{Major: 1}}
	api.HandleFunc(path, ed.handle)
}

// handle manages incoming Rust entity death notices and sends them
// to the RaidAlertsStore and RaidAlerts channel
func (e *entityDeath) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())

	edLog := logWithRequest(r.RequestURI, sc)

	if err != nil {
		edLog.WithError(err).Info("Can't find server")
		handleError(w, models.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var ed models.EntityDeath
	err = decoder.Decode(&ed)
	if err != nil {
		edLog.WithError(err).Error("Invalid JSON")
		handleError(w, models.RESTError{
			Error:      "Invalid request",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	for i := range ed.OwnerIDs {
		ed.OwnerIDs[i] = fmt.Sprintf("%s:%s", sc.game, ed.OwnerIDs[i])
	}

	if len(ed.ServerName) == 0 {
		ed.ServerName = sc.server.Name
	}

	if len(ed.ServerKey) == 8 {
		ed.ServerKey = sc.server.Key
	}

	alertAt := 10 * time.Second
	validUntil := 15 * time.Minute

	sAlertAt, err := time.ParseDuration(sc.server.RaidDelay)
	if err == nil {
		alertAt = sAlertAt
	}

	sValidUntil, err := time.ParseDuration(sc.server.RaidCooldown)
	if err == nil {
		validUntil = sValidUntil
	}

	e.raa.AddInfo(alertAt, validUntil, ed)
}
