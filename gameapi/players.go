package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/poundbot/poundbot/pkg/models"
)

type playerIDs []string

type registeredPlayers struct{}

func initPlayers(api muxFuncHandler, path string) {
	rp := registeredPlayers{}
	api.HandleFunc(fmt.Sprintf("%s/registered", path), rp.handle).Methods(http.MethodGet)
}

func (p *registeredPlayers) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())
	hLog := logWithRequest(r.RequestURI, sc)

	if err != nil {
		hLog.WithError(err).Info("Can't find server")
		handleError(w, models.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		b, err := json.Marshal(sc.account.GetRegisteredPlayerIDs(sc.game))
		if err != nil {
			hLog.WithError(err).Info("Could not marshal JSON")
			return
		}

		setJSONContentType(w, http.StatusOK)
		w.Write(b)
	}
}
