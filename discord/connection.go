package discord

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func connLog() *logrus.Entry {
	return log.WithField("sys", "CONN")
}

// Disconnected is a handler for the Disconnected discord call
func disconnected(status chan<- bool) func(s *discordgo.Session, event *discordgo.Disconnect) {
	return func(s *discordgo.Session, event *discordgo.Disconnect) {
		status <- false
		connLog().Warn("Disconnected!")
	}
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(status chan<- bool) func(s *discordgo.Session, event *discordgo.Ready) {
	return func(s *discordgo.Session, event *discordgo.Ready) {
		connLog().Info("Connection Ready")

		status <- true
		connLog().Trace("Connection Ready Sent")
	}
}

type sessionOpener interface {
	Open() error
}

func connect(sess sessionOpener, status chan<- bool) {
	connLog().Info("Connecting")

	for {
		err := sess.Open()
		if err != nil {
			connLog().WithError(err).Warn("Error connecting; Attempting reconnect...")
			time.Sleep(1 * time.Second)
			continue
		}

		status <- true

		connLog().Info(
			"Connected",
		)
		return

	}
}
