package gameapi

import (
	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/storage"
	"github.com/sirupsen/logrus"
)

type discordAuthRemover interface {
	Remove(storage.UserInfoGetter) error
}

type userIDGetter interface {
	GetPlayerID() string
	GetDiscordID() string
}

type userUpserter interface {
	UpsertPlayer(storage.UserInfoGetter) error
}

// An AuthSaver saves Discord -> Rust user authentications
type AuthSaver struct {
	das         discordAuthRemover
	us          userUpserter
	authSuccess <-chan models.DiscordAuth
	done        <-chan struct{}
}

// NewAuthSaver creates a new AuthSaver
func newAuthSaver(da discordAuthRemover, u userUpserter, as <-chan models.DiscordAuth, done <-chan struct{}) *AuthSaver {
	return &AuthSaver{
		das:         da,
		us:          u,
		authSuccess: as,
		done:        done,
	}
}

// Run updates users sent in through the AuthSuccess channel
func (a *AuthSaver) Run() {
	rLog := log.WithField("sys", "AUTH")
	defer rLog.Warn("AuthServer Stopped.")
	rLog.Info("Starting AuthServer")
	for {
		select {
		case as, more := <-a.authSuccess:
			if !more {
				continue
			}
			rLog = rLog.WithFields(logrus.Fields{
				"gID":       as.GuildSnowflake,
				"pID":       as.PlayerID,
				"discordID": as.Snowflake,
				"name":      as.DiscordName,
			})
			rLog.WithField("pin", as.Pin).Info("auth success")
			if err := a.us.UpsertPlayer(as); err != nil {
				rLog.WithError(err).Error("storage error saving player")
				if as.Ack != nil {
					rLog.Trace("sending auth failure ACK")
					as.Ack(false)
				}
				continue
			}
			if err := a.das.Remove(as); err != nil {
				log.WithError(err).Error("storage error removing DiscordAuth")
				if as.Ack != nil {
					rLog.Trace("sending auth failure ACK")
					as.Ack(false)
				}
				continue
			}

			if as.Ack != nil {
				rLog.Trace("sending auth success ACK")
				as.Ack(true)
			}
		case <-a.done:
			return
		}
	}
}
