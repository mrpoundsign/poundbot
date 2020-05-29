package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/pkg/models"
)

type clanService interface {
	AddClan(serverKey string, clan models.Clan) error
	RemoveClan(serverKey, clanTag string) error
	SetClans(serverKey string, clans []models.Clan) error
}

type serverClan struct {
	Tag        string
	ClanTag    string
	Owner      string
	OwnerID    models.PlayerID
	Members    []models.PlayerID
	Moderators []models.PlayerID
}

func (s serverClan) ToClan() models.Clan {
	c := models.Clan{}
	c.Members = s.Members
	c.Moderators = s.Moderators
	if s.Owner != "" {
		// RustIO Clan
		c.OwnerID = models.PlayerID(s.Owner)
		c.Tag = s.Tag
		return c
	}

	c.OwnerID = s.OwnerID
	c.Tag = s.ClanTag
	return c
}

type clans struct {
	cs clanService
}

func initClans(api muxFuncHandler, path string, cs clanService) {
	c := clans{cs: cs}

	api.HandleFunc(path, c.rootHandler).
		Methods(http.MethodPut)

	api.HandleFunc(fmt.Sprintf("%s/{tag}", path), c.clanHandler).
		Methods(http.MethodDelete, http.MethodPut)
}

// Handle manages clans sync HTTP requests from the Rust server
// These requests are a complete refresh of all clans
func (c *clans) rootHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())
	rhLog := logWithRequest(r.RequestURI, sc)

	if err != nil {
		rhLog.WithError(err).Info("Can't find server")
		handleError(w, models.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	rhLog.Info("Updating all clans")

	decoder := json.NewDecoder(r.Body)

	var sClans []serverClan
	err = decoder.Decode(&sClans)
	if err != nil {
		rhLog.WithError(err).Warn("Could not decode clans")
		handleError(w, models.RESTError{StatusCode: http.StatusBadRequest, Error: "Could not decode clans"})
		return
	}

	var clans = make([]models.Clan, len(sClans))

	for i := range sClans {
		clans[i] = sClans[i].ToClan()
		clans[i].SetGame(sc.game)
	}

	err = c.cs.SetClans(sc.serverKey, clans)
	if err != nil {
		rhLog.WithError(err).Error("Error updating clans")
		handleError(w, models.RESTError{StatusCode: http.StatusInternalServerError, Error: "Could not set clans"})
	}
}

// Handle manages individual clan REST requests form the Rust server
func (c *clans) clanHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())
	chLog := logWithRequest(r.RequestURI, sc)

	if err != nil {
		chLog.WithError(err).Info("Can't find server")
		handleError(w, models.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	vars := mux.Vars(r)
	tag := vars["tag"]

	switch r.Method {
	case http.MethodDelete:
		chLog.Infof("Removing clan \"%s\"", tag)
		err := c.cs.RemoveClan(sc.serverKey, tag)
		if err != nil {
			handleError(w, models.RESTError{
				Error:      "Could not remove clan",
				StatusCode: http.StatusInternalServerError,
			})
			chLog.WithError(err).Errorf("Error removing clan \"%s\"", tag)
		}
		return
	case http.MethodPut:
		chLog.Infof("Updating clan \"%s\"", tag)
		decoder := json.NewDecoder(r.Body)
		var sClan serverClan
		err := decoder.Decode(&sClan)
		if err != nil {
			chLog.WithError(err).Errorf("Error decoding clan \"%s\"", tag)
			handleError(w, models.RESTError{
				Error:      "Could not decode clan data",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		clan := sClan.ToClan()

		clan.SetGame(sc.game)

		err = c.cs.AddClan(sc.serverKey, clan)
		if err != nil {
			handleError(w, models.RESTError{
				Error:      "Could not add clan",
				StatusCode: http.StatusInternalServerError,
			})
			chLog.WithError(err).Errorf("Error adding clan \"%s\"", tag)
		}
	}
}
