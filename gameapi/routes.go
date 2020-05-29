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
	initEntityDeath(router, "/entity_death", s.sc.RAS)
	initDiscordAuth(router, "/discord_auth", s.sc.PAS, s.sc.US, s.dh)
	initChat(router, "/chat", s.channels.ChatQueue)
	initMessages(router, "/messages", s.dh)
	initClans(router, "/clans", s.sc.AS)
	initRoles(router, "/roles", s.dh)
	initPlayers(router, "/players")
}
