package gameapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

type muxFuncHandler interface {
	// HandleFunc registers a new route with a matcher for the URL path.
	// See Route.Path() and Route.HandlerFunc().
	HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route
}

func (s *Server) routes(router muxFuncHandler) {
	initEntityDeath(router, "/entity_death", s.sc.Storage.RaidAlerts())
	initDiscordAuth(router, "/discord_auth", s.sc.Storage.DiscordAuths(), s.sc.Storage.Users(), s.dh)
	initChat(router, "/chat", s.channels.ChatQueue)
	initMessages(router, "/messages", s.dh)
	initClans(router, "/clans", s.sc.Storage.Accounts(), s.sc.Storage.Users())
	initRoles(router, "/roles", s.dh)
	initPlayers(router, "/players")
}
