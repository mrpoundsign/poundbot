package discord

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/poundbot/poundbot/types"
	"github.com/sirupsen/logrus"
)

type gameDiscordMessageSender interface {
	sendChannelMessage(userID, channelID, message string) error
	sendChannelEmbed(userID, channelID, message string, color int) error
}

type guildFinder func(string) (*discordgo.Guild, error)

// gameMessageHandler handles the messages interface from games
func gameMessageHandler(userID string, m types.GameMessage, gf guildFinder, ms gameDiscordMessageSender) {
	defer close(m.ErrorResponse)

	mhLog := log.WithFields(logrus.Fields{
		"cmd":   "gameMessageHandler",
		"gID":   m.Snowflake,
		"cName": m.ChannelName,
	})

	sendErrorResponse := func(errorCh chan<- error, err error) {
		select {
		case errorCh <- err:
		case <-time.After(time.Second / 2):
			mhLog.WithError(err).Error("no response sending message error to channel")
		}
	}

	channelID := ""

	if len(m.Snowflake) == 0 {
		sendErrorResponse(m.ErrorResponse, fmt.Errorf("no server defined"))
		mhLog.Error("no guild id provided with channel name")
		return
	}
	guild, err := gf(m.Snowflake)
	if err != nil {
		sendErrorResponse(m.ErrorResponse, fmt.Errorf("server not found"))
		mhLog.WithError(err).Error("Could not get guild from session")
		return
	}

	for _, gChan := range guild.Channels {
		mhLog.WithField("guildChan", gChan.Name).Trace("checking for channel match")
		if gChan.Type == discordgo.ChannelTypeGuildText && (gChan.Name == m.ChannelName || gChan.ID == m.ChannelName) {
			channelID = gChan.ID
			break
		}
	}

	if len(channelID) == 0 {
		sendErrorResponse(m.ErrorResponse, errors.New("channel not found"))
		mhLog.Info("could not find channel")
		return
	}

	var message string
	for i := range m.MessageParts {
		switch m.MessageParts[i].Escape {
		case true:
			message = message + escapeDiscordString(m.MessageParts[i].Content)
		case false:
			message = message + m.MessageParts[i].Content
		}
	}

	switch m.Type {
	case types.GameMessageTypePlain:
		err = ms.sendChannelMessage(userID, channelID, message)
	case types.GameMessageTypeEmbed:
		err = ms.sendChannelEmbed(userID, channelID, message, m.EmbedStyle.ColorInt())
	}
	if err != nil {
		m.ErrorResponse <- errors.New("could not send to channel")
		mhLog.WithError(err).Error("Error sending chat to channel")
		return
	}

}

// gameChatHandler handles game chat messages
func gameChatHandler(userID string, cm types.ChatMessage, gf guildFinder, ms gameDiscordMessageSender) {
	ccLog := log.WithFields(logrus.Fields{
		"cmd":   "gameChatHandler",
		"pID":   cm.PlayerID,
		"cID":   cm.ChannelID,
		"name":  cm.DisplayName,
		"dName": cm.DiscordName,
	})
	var clan = ""
	if cm.ClanTag != "" {
		clan = fmt.Sprintf("[%s] ", cm.ClanTag)
	}

	err := ms.sendChannelMessage(
		userID,
		cm.ChannelID,
		fmt.Sprintf("☢️ @%s **%s%s**: %s",
			iclock().Now().UTC().Format("01-02 15:04 MST"),
			clan, escapeDiscordString(cm.DisplayName), escapeDiscordString(cm.Message)),
	)

	if err != nil {
		ccLog.WithError(err).Error("Error sending chat to channel.")
	}
}
