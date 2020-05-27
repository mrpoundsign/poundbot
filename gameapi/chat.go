package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blang/semver"
	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/pkg/models"
)

var iclock = pbclock.Clock

type chatQueue interface {
	GetGameServerMessage(sk, tag string, to time.Duration) (models.ChatMessage, bool)
}

type discordChat struct {
	ClanTag     string
	DisplayName string
	Message     string
}

func newDiscordChat(cm models.ChatMessage) discordChat {
	return discordChat{
		ClanTag:     cm.ClanTag,
		DisplayName: cm.DisplayName,
		Message:     cm.Message,
	}
}

// A Chat is for handling discord <-> rust chat
type chat struct {
	cqs        chatQueue
	timeout    time.Duration
	minVersion semver.Version
}

// initChat initializes a chat handler and returns it
func initChat(api *mux.Router, path string, cq chatQueue) {
	c := chat{
		cqs:        cq,
		timeout:    10 * time.Second,
		minVersion: semver.Version{Major: 1, Patch: 3},
	}

	api.HandleFunc(path, c.handle).Methods(http.MethodGet, http.MethodPost)
}

// handle manages Discord to GameServer chat requests
//
// HTTP GET requests wait for messages or disconnect with http.StatusNoContent
// after timeout seconds.
func (c *chat) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())
	if err != nil {
		log.Info(fmt.Sprintf("[%s](%s:%s) Can't find server: %s", sc.requestUUID, sc.account.ID.Hex(), sc.serverKey, err.Error()))
		handleError(w, models.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusForbidden,
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		m, found := c.cqs.GetGameServerMessage(sc.serverKey, "chat", c.timeout)
		if !found {
			setJSONContentType(w, http.StatusNoContent)
			return
		}

		b, err := json.Marshal(newDiscordChat(m))
		if err != nil {
			log.Printf("[%s] %s", sc.requestUUID, err.Error())
			return
		}

		setJSONContentType(w, http.StatusOK)

		w.Write(b)
	}
}
