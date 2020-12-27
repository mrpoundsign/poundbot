package discord

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/pkg/modules/account"
	"github.com/poundbot/poundbot/pkg/modules/playerauth"
	"github.com/poundbot/poundbot/storage"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

var iclock = pbclock.Clock

type ChatQueueStore interface {
	GetGameServerMessage(serverKey, tag string, timeout time.Duration) (message models.ChatMessage, success bool)
	InsertMessage(message models.ChatMessage) error
}

type userService interface {
	GetByPlayerID(models.PlayerID) (models.User, error)
	GetByDiscordID(models.PlayerDiscordID) (models.User, error)
	GetPlayerIDsByDiscordIDs(snowflakes []models.PlayerDiscordID) ([]models.PlayerID, error)
	RemovePlayerID(models.PlayerDiscordID, models.PlayerID) error
}

type Runner struct {
	session         *discordgo.Session
	cqs             ChatQueueStore
	as              account.Service
	mls             storage.MessageLocksStore
	authsRepo       playerauth.Service
	us              userService
	token           string
	status          chan bool
	chatChan        chan models.ChatMessage
	raidAlertChan   chan models.RaidAlertWithMessageChannel
	gameMessageChan chan models.GameMessage
	authChan        chan models.DiscordAuth
	AuthSuccess     chan models.DiscordAuth
	channelsRequest chan models.ServerChannelsRequest
	roleSetChan     chan models.RoleSet
	shutdown        bool
}

func NewRunner(token string, as account.Service, ar playerauth.Service,
	us userService, mls storage.MessageLocksStore, cqs ChatQueueStore) *Runner {
	return &Runner{
		cqs:             cqs,
		mls:             mls,
		as:              as,
		authsRepo:       ar,
		us:              us,
		token:           token,
		chatChan:        make(chan models.ChatMessage),
		authChan:        make(chan models.DiscordAuth),
		AuthSuccess:     make(chan models.DiscordAuth),
		raidAlertChan:   make(chan models.RaidAlertWithMessageChannel),
		gameMessageChan: make(chan models.GameMessage),
		channelsRequest: make(chan models.ServerChannelsRequest),
		roleSetChan:     make(chan models.RoleSet),
	}
}

// Start starts the runner
func (r *Runner) Start() error {
	session, err := discordgo.New("Bot " + r.token)
	if err != nil {
		return err
	}

	r.session = session
	r.session.Identify.Intents = discordgo.MakeIntent(
		discordgo.IntentsAllWithoutPrivileged |
			discordgo.IntentsGuildMembers,
	)
	r.session.State.TrackMembers = true
	r.session.State.TrackChannels = true
	r.session.AddHandler(r.messageCreate)
	r.session.AddHandler(ready(r.status))
	r.session.AddHandler(disconnected(r.status))
	r.session.AddHandler(r.resumed)
	r.session.AddHandler(newGuildCreate(r.as, r.us))
	r.session.AddHandler(newGuildDelete(r.as))
	r.session.AddHandler(newGuildMemberAdd(r.us, r.as))
	r.session.AddHandler(newGuildMemberRemove(r.us, r.as))

	r.session.AddHandler(
		func(s *discordgo.Session, e *discordgo.Event) {
			log.Tracef("Event: %s", e.Type)
		},
	)

	r.status = make(chan bool)

	go r.runner()

	connect(r.session, r.status)

	return nil
}

func (r Runner) RaidNotify(ra models.RaidAlertWithMessageChannel) {
	r.raidAlertChan <- ra
}

func (r Runner) AuthDiscord(da models.DiscordAuth) {
	r.authChan <- da
}

func (r Runner) SendChatMessage(cm models.ChatMessage) {
	r.chatChan <- cm
}

// SendGameMessage sends a message from the game to a discord channel
func (r Runner) SendGameMessage(gm models.GameMessage, timeout time.Duration) error {
	select {
	case r.gameMessageChan <- gm:
		return nil
	case <-time.After(timeout):
		return errors.New("no response from discord handler")
	}
}

// ServerChannels sends a request to get the visible chnnels for a discord guild
func (r Runner) ServerChannels(scr models.ServerChannelsRequest) {
	r.channelsRequest <- scr
}

func (r Runner) SetRole(rs models.RoleSet, timeout time.Duration) error {
	// sending message
	select {
	case r.roleSetChan <- rs:
		return nil
	case <-time.After(timeout):
		return errors.New("no response from discord handler")
	}
}

// Stop stops the runner
func (r *Runner) Stop() {
	defer r.session.Close()
	log.WithFields(logrus.Fields{"sys": "RUNNER"}).Info(
		"Disconnecting...",
	)

	r.shutdown = true
}

func (r *Runner) runner() {
	rLog := log.WithFields(logrus.Fields{"sys": "RUNNER"})
	defer rLog.Warn("Runner exited")

	connectedState := false

	for {
		if connectedState {
			rLog.Info("Waiting for messages.")
		Reading:
			for {
				select {
				case connectedState = <-r.status:
					if !connectedState {
						rLog.Warn("Received disconnected message")
						if r.shutdown {
							return
						}
						break Reading
					}

					rLog.Info("Received unexpected connected message")
				case raidAlert := <-r.raidAlertChan:
					raLog := rLog.WithFields(logrus.Fields{"chan": "RAID", "pID": raidAlert.PlayerID})
					raLog.Trace("Got raid alert")
					go func() {
						defer close(raidAlert.MessageIDChannel)
						raUser, err := r.us.GetByPlayerID(raidAlert.PlayerID)
						if err != nil {
							raLog.WithError(err).Error("Player not found trying to send raid alert")
							return
						}

						user, err := r.session.User(raUser.Snowflake.String())
						if err != nil {
							raLog.WithField("uID", raUser.Snowflake).WithError(err).Error(
								"Discord user not found trying to send raid alert",
							)
							return
						}

						id, err := r.sendPrivateMessage(models.PlayerDiscordID(user.ID), raidAlert.MessageID, raidAlert.String())
						if err != nil {
							raLog.WithError(err).Error("could not create private channel to send to user")
							return
						}

						raLog.Tracef("setting message ID to %s", id)

						raidAlert.MessageIDChannel <- id

					}()
				case da := <-r.authChan:
					go r.discordAuthHandler(da)
				case m := <-r.gameMessageChan:
					go gameMessageHandler(r.session.State.User.ID, m, r.session.State.Guild, r)
				case cm := <-r.chatChan:
					go gameChatHandler(r.session.State.User.ID, cm, r.session.State.Guild, r)
				case cr := <-r.channelsRequest:
					go sendChannelList(r.session.State.User.ID, cr.GuildID, cr.ResponseChan, r.session.State)
				case rs := <-r.roleSetChan:
					go rolesSetHandler(r.session.State.User.ID, rs, r.session.State, r.us, r.session)
				}
			}
		}
	Connecting:
		for {
			rLog.Info("Waiting for connected state change...")
			connectedState = <-r.status
			rLog.WithField("sys", "CONN").Infof("Received connection state: %v", connectedState)
			if connectedState {

				rLog.Trace("setting discord status")
				if err := r.session.UpdateStatus(0, localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "DiscordStatus",
						Other: "!pb help",
					}}),
				); err != nil {
					log.WithError(err).Error("failed to update bot status")
				}

				log.Tracef("Found %d guilds", len(r.session.State.Guilds))
				guilds := make([]models.BaseAccount, len(r.session.State.Guilds))
				for i, guild := range r.session.State.Guilds {
					guilds[i] = models.BaseAccount{GuildSnowflake: guild.ID, OwnerSnowflake: models.PlayerDiscordID(guild.OwnerID)}
				}
				if err := r.as.RemoveNotInDiscordGuildList(guilds); err != nil {
					log.WithError(err).Error("could not sync discord guilds")
				}

				break Connecting
			}
			rLog.WithField("sys", "CONN").Info("Received disconnected message")
		}
	}

}

func (r *Runner) discordAuthHandler(da models.DiscordAuth) {
	dLog := log.WithFields(logrus.Fields{
		"chan": "DAUTH",
		"gID":  da.GuildSnowflake,
		"name": da.DiscordInfo.DiscordName,
		"uID":  da.Snowflake,
	})
	dLog.Trace("Got discord auth")
	dUser, err := r.getUserByName(da.GuildSnowflake, da.DiscordInfo.DiscordName)
	if err != nil {
		dLog.WithError(err).Error("Discord user not found")
		err = r.authsRepo.Remove(da.GetPlayerID())
		if err != nil {
			dLog.WithError(err).Error("Error removing discord auth for PlayerID from the database.")
		}
		return
	}

	da.Snowflake = models.PlayerDiscordID(dUser.ID)

	err = r.authsRepo.Upsert(da)
	if err != nil {
		dLog.WithError(err).Error("Error upserting PlayerID ito the database")
		return
	}

	_, err = r.sendPrivateMessage(da.Snowflake, "",
		localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "UserPINPrompt",
				Other: "Enter the PIN provided in-game to validate your account.\nOnce you are validated, you will begin receiving raid alerts!",
			},
		}),
	)

	if err != nil {
		dLog.WithError(err).Error("Could not send PIN request to user")
	}
}

func (r *Runner) resumed(s *discordgo.Session, event *discordgo.Resumed) {
	log.WithField("sys", "CONN").Info("Resumed connection")
	r.status <- true
}

// Returns nil user if they don't exist; Returns error if there was a communications error
func (r *Runner) getUserByName(guildID, name string) (discordgo.User, error) {
	guild, err := r.session.State.Guild(guildID)
	if err != nil {
		return discordgo.User{}, fmt.Errorf("guild %s not found searching for user %s", guildID, name)
	}

	for _, user := range guild.Members {
		log.Tracef("Checking %s: %s", user.User.String(), name)
		if strings.ToLower(user.User.String()) == strings.ToLower(name) {
			return *user.User, nil
		}
	}

	return discordgo.User{}, fmt.Errorf("discord user not found %s for %s", name, guildID)
}
