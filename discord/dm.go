package discord

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/poundbot/poundbot/messages"
	"github.com/poundbot/poundbot/pkg/models"

	"github.com/sirupsen/logrus"
)

type dmUserStorage interface {
	GetByDiscordID(models.PlayerDiscordID) (models.User, error)
	RemovePlayerID(models.PlayerDiscordID, models.PlayerID) error
}

type dmAuthStorage interface {
	AddRegisteredPlayerIDs(accountID string, playerIDs []models.PlayerID) error
}

type dmDiscordAccountStorage interface {
	GetByDiscordID(models.PlayerDiscordID) (models.DiscordAuth, error)
}

type dm struct {
	us       dmUserStorage
	as       dmAuthStorage
	das      dmDiscordAccountStorage
	authChan chan<- models.DiscordAuth
}

func (i dm) process(m discordgo.MessageCreate) string {
	pLog := log.WithFields(logrus.Fields{"sys": "dm.process()"})
	message := strings.TrimSpace(m.Content)
	authorID := models.PlayerDiscordID(m.Author.ID)

	isPIN, err := regexp.MatchString("\\A[0-9]+\\z", message)
	if err != nil {
		pLog.Error(err)
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "InternalError",
				Other: "Internal error. Please try again.",
			}})
	}

	if isPIN {
		return i.validatePIN(message, authorID)
	}

	parts := strings.Fields(message)

	if len(parts) == 0 {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "InstructInvalidCommand",
				Other: "Invalid command. See `help`",
			}})
	}

	switch parts[0] {
	case localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandStatus",
			Other: "status",
		},
	}):
		return i.status(authorID)
	case localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandHelp",
			Other: "help",
		},
	}):
		return i.help(authorID)
	case localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandUnregister",
			Other: "unregister",
		},
	}):
		return i.unregister(authorID, parts[1:])
	}

	return localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructInvalidCommand",
			Other: "Invalid command. See `help`",
		}})
}

func (i dm) status(authorID models.PlayerDiscordID) string {
	u, err := i.us.GetByDiscordID(authorID)
	if err != nil {
		return "You are not registered anywhere."
	}
	return fmt.Sprintf("Your registered IDs are: %s", strings.Join(u.PlayerIDsAsStrings(), ", "))
}

func (i dm) unregister(authorID models.PlayerDiscordID, parts []string) string {
	u, err := i.us.GetByDiscordID(authorID)
	if err != nil {
		return "You are not registered anywhere."
	}

	if len(parts) == 0 {
		return "Usage: `unregister <game>`"
	}

	if parts[0] == "all" {
		u.PlayerIDs = []models.PlayerID{}
		i.us.RemovePlayerID(u.Snowflake, "all")
		return "You have been removed from all games."
	}

	game := parts[0]
	for _, pID := range u.PlayerIDs {
		if strings.HasPrefix(string(pID), fmt.Sprintf("%s:", game)) {
			err := i.us.RemovePlayerID(u.Snowflake, pID)
			if err != nil {
				// return "Could not remove ID, try again."
				return fmt.Sprintf("Could not remove ID, try again. %s", err)
			}
			return fmt.Sprintf("%s removed", pID)
		}
	}

	return fmt.Sprintf("Could not find an ID for game %s.\n%s", game, i.status(authorID))
}

func (i dm) help(authorID models.PlayerDiscordID) string {
	return messages.DMHelpText()
}

func (i dm) validatePIN(pin string, authorID models.PlayerDiscordID) string {
	vpLog := log.WithFields(logrus.Fields{"sys": "dm.validatePIN()"})
	da, err := i.das.GetByDiscordID(authorID)
	if err != nil {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PINNotRequested",
				Other: "ERROR: PIN is not required at this time. Check `status` or `help`.",
			}})
	}

	vpLog = vpLog.WithFields(logrus.Fields{"playerid": da.PlayerID, "guildid": da.GuildSnowflake})

	if !(pinString(da.Pin) == pin) {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PINInvalid",
				Other: "Invalid PIN. Please try again.",
			}})
	}

	authResult := make(chan string)
	da.Ack = func(authenticated bool) {
		defer close(authResult)
		if authenticated {
			authResult <- localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "PINAuthenticated",
					Other: "You have authenticated!",
				}})
			err = i.as.AddRegisteredPlayerIDs(da.GuildSnowflake, []models.PlayerID{da.PlayerID})
			if err != nil {
				vpLog.Error(err)
			}
			return
		}

		authResult <- localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PINInternalError",
				Other: "Internal error. Please try again.",
			}})
	}
	i.authChan <- da
	return <-authResult
}
