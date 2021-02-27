package discord

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/pkg/models"
	"github.com/sirupsen/logrus"
)

type guildAccountStorer interface {
	UpsertBase(models.BaseAccount) error
	SetRegisteredPlayerIDs(ServerID string, IDs []models.PlayerID) error
	GetByDiscordGuild(string) (models.Account, error)
	Remove(string) error
}

type guildUserStorer interface {
	GetPlayerIDsByDiscordIDs([]models.PlayerDiscordID) ([]models.PlayerID, error)
	SetGuildUsers(dinfo []models.DiscordInfo, gid string) error
	AddGuildUser(di models.DiscordInfo, gid string) error
	RemoveGuildUser(di models.DiscordInfo, gid string) error
	RemoveGuild(gid string) error
}

// type guildUserListGetter inferace {

// }

type guildHandler struct {
	as           guildAccountStorer
	ug           guildUserStorer
	s            sync.Mutex
	listLock     sync.RWMutex
	loadedGuilds []string
}

func newGuildHandler(as guildAccountStorer, ug guildUserStorer) *guildHandler {
	gc := guildHandler{
		as:           as,
		ug:           ug,
		loadedGuilds: []string{},
	}
	return &gc
}

func (gh *guildHandler) registerDiscordHooks(s *discordgo.Session) {
	s.AddHandler(gh.guildCreate)
	s.AddHandler(gh.guildDelete)
	s.AddHandler(gh.ready)
}

func (gh *guildHandler) ready(s *discordgo.Session, _ *discordgo.Ready) {
	gh.listLock.Lock()
	defer gh.listLock.Unlock()

	gh.loadedGuilds = []string{}
}

func (gh *guildHandler) guildMembersLoaded(gID string) bool {
	gh.listLock.RLock()
	defer gh.listLock.RUnlock()

	for _, gid := range gh.loadedGuilds {
		if gid == gID {
			return true
		}
	}

	return false
}

func (gh *guildHandler) setGuildMembersLoaded(gID string) {
	gh.listLock.Lock()
	defer gh.listLock.Unlock()

	gh.loadedGuilds = append(gh.loadedGuilds, gID)
}

func (gh *guildHandler) getGuildMembers(gID string, s *discordgo.Session) ([]models.DiscordInfo, error) {
	gh.s.Lock()
	defer func() {
		time.Sleep(500 * time.Millisecond)
		gh.s.Unlock()
	}()

	dinfo := []models.DiscordInfo{}
	uid := ""

	for {
		log.Tracef("Finding from \"%s\" for gID %s", uid, gID)
		g, err := s.GuildMembers(gID, uid, 1000)
		if err != nil {
			log.Printf("Could not read members for gID %s at %s", gID, uid)
			return dinfo, fmt.Errorf("could not read members for gID %s at \"%s\": %w", gID, uid, err)
		}

		log.Tracef("Found %d", len(g))

		if len(g) == 0 {
			log.Printf("Found %d members for %s", len(dinfo), gID)
			log.Tracef("Exiting user list")
			return dinfo, nil
		}

		newUsers := make([]models.DiscordInfo, len(g))
		for i, u := range g {
			newUsers[i] = models.DiscordInfo{
				DiscordName: fmt.Sprintf("%s#%s", u.User.Username, u.User.Discriminator),
				Snowflake:   models.PlayerDiscordID(u.User.ID),
			}
			log.Tracef("User: %s - %s:%s", gID, u.User.Username, u.User.ID)
		}

		dinfo = append(dinfo, newUsers...)

		log.Trace("Finding more...")
		time.Sleep(200 * time.Millisecond)
		uid = g[len(g)-1].User.ID

	}
}

func (gh *guildHandler) guildCreate(s *discordgo.Session, gc *discordgo.GuildCreate) {
	gid := gc.ID
	gcLog := log.WithFields(logrus.Fields{"sys": "guildHandler", "gID": gid, "guildName": gc.Name})

	if gh.guildMembersLoaded(gid) {
		gcLog.Trace("Skipping guild member load, already loaded")
	}

	gh.setGuildMembersLoaded(gid)

	log.Trace("++ Loading Members")
	members, err := gh.getGuildMembers(gid, s)
	if err != nil {
		log.WithError(err).Errorf("Could not get members for guild %s", gid)
		return
	}

	userIDs := make([]models.PlayerDiscordID, len(members))
	for i, member := range members {
		log.Tracef("    Member: %s:%s", member.DiscordName, member.Snowflake)
		userIDs[i] = models.PlayerDiscordID(member.Snowflake)
	}
	log.Trace("   Loading Members Done")

	account, err := gh.as.GetByDiscordGuild(gc.ID)
	if err != nil {
		if err != mgo.ErrNotFound {
			// Some other storage error
			log.WithError(err).Error("Error loading account")
			return
		}
		account.BaseAccount = models.BaseAccount{GuildSnowflake: gc.ID, OwnerSnowflake: models.PlayerDiscordID(gc.OwnerID)}
	} else {
		account.OwnerSnowflake = models.PlayerDiscordID(gc.OwnerID)
	}

	err = gh.as.UpsertBase(account.BaseAccount)
	if err != nil {
		gcLog.WithError(err).Error("Error upserting account")
		return
	}

	playerIDs, err := gh.ug.GetPlayerIDsByDiscordIDs(userIDs)
	if err != nil {
		gcLog.WithError(err).Error("Error getting playerIDs")
		return
	}

	gcLog = gcLog.WithField("playerIDs", playerIDs)

	gcLog.Trace("Adding players")

	err = gh.ug.SetGuildUsers(members, account.GuildSnowflake)
	if err != nil {
		gcLog.WithError(err).Error("Error setting guild users")
	}

	err = gh.as.SetRegisteredPlayerIDs(account.GuildSnowflake, playerIDs)
	if err != nil {
		gcLog.WithError(err).Error("Error setting playerIDs")
	}
}

func (gh *guildHandler) guildDelete(s *discordgo.Session, gd *discordgo.GuildDelete) {
	log.Printf("Removing guild %s", gd.ID)
	gh.as.Remove(gd.ID)
	gh.ug.RemoveGuild(gd.ID)
}

type userStorer interface {
	GetByDiscordID(models.PlayerDiscordID) (models.User, error)
	AddGuildUser(di models.DiscordInfo, gid string) error
	RemoveGuildUser(di models.DiscordInfo, gid string) error
}

type guildMemberAdder interface {
	GetByDiscordGuild(key string) (models.Account, error)
	AddRegisteredPlayerIDs(accountSnowflake string, playerIDs []models.PlayerID) error
}

func newGuildMemberAdd(us userStorer, gma guildMemberAdder) func(*discordgo.Session, *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, dgma *discordgo.GuildMemberAdd) {
		log.Printf("Adding user %s#%s(%s) to gID %s", dgma.User.Username, dgma.User.Discriminator, dgma.User.ID, dgma.GuildID)
		di := models.DiscordInfo{
			DiscordName: fmt.Sprintf("%s#%s", dgma.User.Username, dgma.User.Discriminator),
			Snowflake:   models.PlayerDiscordID(dgma.User.ID),
		}
		err := us.AddGuildUser(di, dgma.GuildID)
		if err != nil {
			log.WithError(err).Error("Storage error: Could not add user %s:%s to %s", di.DiscordName, di.Snowflake, dgma.GuildID)
		}
		guildMemberAdd(us, gma, dgma.GuildID, dgma.User.ID)
	}
}

func guildMemberAdd(us userStorer, gma guildMemberAdder, gID, uID string) {
	gmaLog := log.WithFields(logrus.Fields{"sys": "guildMemberAdd", "gID": gID, "uID": uID})
	user, err := us.GetByDiscordID(models.PlayerDiscordID(uID))

	if err != nil {
		gmaLog.WithError(err).Trace("Error finding user")
		return
	}

	_, err = gma.GetByDiscordGuild(gID)
	if err != nil {
		if err != mgo.ErrNotFound {
			gmaLog.WithError(err).Trace("Could not get account for guild")
		}
		return
	}

	err = gma.AddRegisteredPlayerIDs(gID, user.PlayerIDs)
	if err != nil {
		gmaLog.WithError(err).Error("Storage error: Could not add player IDs to account")
	}
}

type guildMemberRemover interface {
	GetByDiscordGuild(key string) (models.Account, error)
	RemoveRegisteredPlayerIDs(accountSnowflake string, pids []models.PlayerID) error
}

func newGuildMemberRemove(us userStorer, gmr guildMemberRemover) func(*discordgo.Session, *discordgo.GuildMemberRemove) {
	return func(s *discordgo.Session, dgmr *discordgo.GuildMemberRemove) {
		guildMemberRemove(us, gmr, dgmr.GuildID, dgmr.User.ID)
	}
}

func guildMemberRemove(us userStorer, gmr guildMemberRemover, gID, uID string) {
	log.Printf("Removeing user %s(%s) from gID %s", uID, gID)
	gmrLog := log.WithFields(logrus.Fields{"sys": "guildMemberRemove", "gID": gID, "uID": uID})
	user, err := us.GetByDiscordID(models.PlayerDiscordID(uID))
	if err != nil {
		gmrLog.WithError(err).Trace("Error finding user")
		return
	}
	account, err := gmr.GetByDiscordGuild(gID)
	if err != nil {
		if err != mgo.ErrNotFound {
			gmrLog.WithError(err).Trace("Could not get account for guild ID")
		}
		return
	}
	err = gmr.RemoveRegisteredPlayerIDs(account.GuildSnowflake, user.PlayerIDs)
	if err != nil {
		gmrLog.WithError(err).Error("Could not remove player IDs to account")
	}

	err = us.RemoveGuildUser(user.DiscordInfo, gID)
	if err != nil {
		gmrLog.WithError(err).Error("Storage error: Could not remove guild snowflake from account")
	}
}
