package auth

import (
	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/pkg/modules/user"
	"github.com/sirupsen/logrus"
)

type discordAuthRemover interface {
	Remove(models.PlayerID) error
}

type userIDGetter interface {
	GetPlayerID() models.PlayerID
	GetDiscordID() models.PlayerDiscordID
}

type userUpserter interface {
	UpsertPlayer(user.UserInfoGetter) error
}

// An AuthSaver saves Discord -> Rust user authentications
type saver struct {
	das         discordAuthRemover
	us          userUpserter
	authSuccess <-chan models.DiscordAuth
	done        <-chan interface{}
}

// NewAuthSaver creates a new AuthSaver
func NewSaver(da discordAuthRemover, u userUpserter, as <-chan models.DiscordAuth, done <-chan interface{}) *saver {
	return &saver{
		das:         da,
		us:          u,
		authSuccess: as,
		done:        done,
	}
}

// Run updates users sent in through the AuthSuccess channel
func (a *saver) Run() {
	defer log.Warn("Auth Server Stopped.")

	log.Info("Starting Auth Server")

	for {
		select {
		case da, more := <-a.authSuccess:
			if !more {
				continue
			}
			rLog := log.WithFields(logrus.Fields{
				"gID":       da.GuildSnowflake,
				"pID":       da.PlayerID,
				"discordID": da.Snowflake,
				"name":      da.DiscordName,
			})

			rLog.WithField("pin", da.Pin).Info("auth success")
			if err := a.us.UpsertPlayer(da); err != nil {
				rLog.WithError(err).Error("storage error saving player")
				if da.Ack != nil {
					rLog.Trace("sending auth failure ACK")
					da.Ack(false)
				}
				continue
			}

			if err := a.das.Remove(da.GetPlayerID()); err != nil {
				log.WithError(err).Error("storage error removing DiscordAuth")
				if da.Ack != nil {
					rLog.Trace("sending auth failure ACK")
					da.Ack(false)
				}
				continue
			}

			if da.Ack != nil {
				rLog.Trace("sending auth success ACK")
				da.Ack(true)
			}
		case <-a.done:
			return
		}
	}
}
