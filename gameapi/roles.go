package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/pkg/models"
)

type discordRoleSetter interface {
	SetRole(models.RoleSet, time.Duration) error
}

type roles struct {
	drs     discordRoleSetter
	timeout time.Duration
}

func initRoles(api *mux.Router, path string, drs discordRoleSetter) {
	r := roles{drs: drs, timeout: 10 * time.Second}

	api.HandleFunc(fmt.Sprintf("%s/{role_name}", path), r.roleHandler).
		Methods(http.MethodPut)
}

func (rs roles) roleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	role := vars["role_name"]

	sc, err := getServerContext(r.Context())
	rhLog := logWithRequest(r.RequestURI, sc)

	if err != nil {
		rhLog.Info("Could not find server")
		handleError(w, models.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var roleSet models.RoleSet

	err = decoder.Decode(&roleSet)
	if err != nil {
		rhLog.WithError(err).Error("Invalid JSON")
		if err := handleError(w, models.RESTError{
			Error:      "Invalid request",
			StatusCode: http.StatusBadRequest,
		}); err != nil {
			rhLog.WithError(err).Error("http response failed to write")
		}
		return
	}

	roleSet.Role = role
	roleSet.GuildID = sc.account.GuildSnowflake
	roleSet.SetGame(sc.game)

	if err := rs.drs.SetRole(roleSet, rs.timeout); err != nil {
		rhLog.Error("timed out sending message to channel")
		if err := handleError(w, models.RESTError{
			Error:      "internal error sending message to discord handler",
			StatusCode: http.StatusInternalServerError,
		}); err != nil {
			rhLog.WithError(err).Error("http response failed to write")
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
