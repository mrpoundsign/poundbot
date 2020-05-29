package gameapi

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/storage"
)

const upgradeURL = "https://umod.org/plugins/pound-bot"

type discordHandler interface {
	RaidNotify(models.RaiAlertWithMessageChannel)
	AuthDiscord(models.DiscordAuth)
	SendChatMessage(models.ChatMessage)
	SendGameMessage(models.GameMessage, time.Duration) error
	ServerChannels(models.ServerChannelsRequest)
	SetRole(models.RoleSet, time.Duration) error
}

// ServerConfig contains the base Server configuration
type ServerConfig struct {
	BindAddr string
	Port     int
	Storage  storage.Storage
}

type ServerChannels struct {
	AuthSuccess <-chan models.DiscordAuth
	ChatQueue   storage.ChatQueueStore
}

// A Server runs the HTTP server, notification channels, and DB writing.
type Server struct {
	http.Server
	sc              *ServerConfig
	channels        ServerChannels
	shutdownRequest chan struct{}
	dh              discordHandler
}

// NewServer creates a Server
func NewServer(sc *ServerConfig, dh discordHandler, channels ServerChannels) *Server {
	s := Server{
		Server: http.Server{
			Addr: fmt.Sprintf("%s:%d", sc.BindAddr, sc.Port),
		},
		sc:       sc,
		dh:       dh,
		channels: channels,
	}

	sa := newServerAuth(sc.Storage.Accounts())
	r := mux.NewRouter()

	// Handles all /api requests, and sets the server auth handler
	api := r.PathPrefix("/api").Subrouter()
	api.Use(sa.handle)
	api.Use(requestUUID)

	s.routes(api)

	s.Handler = r

	s.shutdownRequest = make(chan struct{})

	return &s
}

// Start starts the HTTP server, raid alerter, and Discord auth manager
func (s *Server) Start() error {
	// Start the AuthSaver
	go func() {
		var newConn = s.sc.Storage.Copy()
		defer newConn.Close()

		var as = newAuthSaver(newConn.DiscordAuths(), newConn.Users(), s.channels.AuthSuccess, s.shutdownRequest)
		as.Run()
	}()

	// Start the RaidAlerter
	go func() {
		var newConn = s.sc.Storage.Copy()
		defer newConn.Close()

		var ra = newRaidAlerter(newConn.RaidAlerts(), s.dh, s.shutdownRequest)
		ra.Run()
	}()

	go func() {
		log.Printf("Starting HTTP Server on %s:%d", s.sc.BindAddr, s.sc.Port)
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.WithError(err).Warn("HTTP server died with error\n")
		} else {
			log.Print("HTTP server graceful shutdown")
		}
	}()

	return nil
}

// Stop stops the http server
func (s *Server) Stop() {
	log.Warn("Shutting down HTTP server ...")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() { //Create shutdown context with 10 second timeout
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//shutdown the server
		err := s.Shutdown(ctx)
		if err != nil {
			log.WithError(err).Warn("Shutdown request error")
		}
	}()
	s.shutdownRequest <- struct{}{} // AuthSaver
	s.shutdownRequest <- struct{}{} // RaidAlerter
	wg.Wait()
}
