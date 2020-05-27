package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/pkg/models"
)

type discordAuthenticator interface {
	AuthDiscord(models.DiscordAuth)
}

type daAuthUpserter interface {
	Upsert(models.DiscordAuth) error
}

type daUserGetter interface {
	GetByPlayerID(string) (models.User, error)
}

type discordAuth struct {
	dau daAuthUpserter
	us  daUserGetter
	da  discordAuthenticator
}

type discordAuthRequest struct {
	models.DiscordAuth
}

func initDiscordAuth(api *mux.Router, path string, dau daAuthUpserter, us daUserGetter, dah discordAuthenticator) {
	da := discordAuth{dau: dau, us: us, da: dah}
	api.HandleFunc(path, da.createDiscordAuth).Methods("PUT")
	api.HandleFunc(fmt.Sprintf("%s/check/{player_id}", path), da.checkPlayer).Methods("GET")
}

// createDiscordAuth takes Discord verification requests from the Rust server
// and sends them to the DiscordAuthsStore and DiscordAuth channel
func (da *discordAuth) createDiscordAuth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())

	params := mux.Vars(r)
	hLog := logWithRequest(r.RequestURI, sc).WithField("pID", params["player_id"])

	if err != nil {
		hLog.WithError(err).Info("Can't find server")
		handleError(w, models.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var dAuth discordAuthRequest

	err = decoder.Decode(&dAuth)
	if err != nil {
		hLog.WithError(err).Info("Bad request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dAuth.PlayerID = fmt.Sprintf("%s:%s", sc.game, dAuth.PlayerID)

	user, err := da.us.GetByPlayerID(dAuth.PlayerID)
	if err == nil {
		handleError(w, models.RESTError{
			StatusCode: http.StatusConflict,
			Error:      fmt.Sprintf("%s is already registered.", user.DiscordName),
		})
		return
	}

	dAuth.GuildSnowflake = sc.account.GuildSnowflake

	err = da.dau.Upsert(dAuth.DiscordAuth)
	if err != nil {
		log.Println(err.Error())
		return
	}
	da.da.AuthDiscord(dAuth.DiscordAuth)
}

func (da *discordAuth) checkPlayer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())

	params := mux.Vars(r)
	cpLog := logWithRequest(r.RequestURI, sc).WithField("pID", params["player_id"])

	if err != nil {
		cpLog.WithError(err).Info("Can't find server")
		handleError(w, models.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	cpLog.Trace("Checking player")
	_, err = da.us.GetByPlayerID(fmt.Sprintf("%s:%s", sc.game, params["player_id"]))
	if err != nil {
		cpLog.Trace("Player not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cpLog.Trace("Player found")
}
