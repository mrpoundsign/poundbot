package discord

import (
	"fmt"
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
}

type guildUserStorer interface {
	GetPlayerIDsByDiscordIDs([]models.PlayerDiscordID) ([]models.PlayerID, error)
	SetGuildUsers(dinfo []models.DiscordInfo, gid string) error
}

// type guildUserListGetter inferace {

// }

type guildHandler struct {
	as guildAccountStorer
	ug guildUserStorer
}

func newGuildCreate(as guildAccountStorer, ug guildUserStorer) func(*discordgo.Session, *discordgo.GuildCreate) {
	gc := guildHandler{as: as, ug: ug}
	return gc.guildCreate
}

func getGuildMembers(gid string, s *discordgo.Session) ([]models.DiscordInfo, error) {
	dinfo := []models.DiscordInfo{}
	uid := ""

	for {
		log.Tracef("Finding from %s", uid)
		g, err := s.GuildMembers(gid, uid, 1000)
		if err != nil {
			log.Printf("Could not read players for gid %s at %s", gid, uid)
			return dinfo, fmt.Errorf("could not read players for gid %s at \"%s\": %w", gid, uid, err)
		}

		log.Tracef("Found %d", len(g))

		if len(g) == 0 {
			log.Tracef("Exiting user list")
			break
		}

		newUsers := make([]models.DiscordInfo, len(g))
		for i, u := range g {
			newUsers[i] = models.DiscordInfo{
				DiscordName: fmt.Sprintf("%s#%s", u.User.Username, u.User.Discriminator),
				Snowflake:   models.PlayerDiscordID(u.User.ID),
			}
			log.Tracef("User: %s - %s:%s", gid, u.User.Username, u.User.ID)
		}

		dinfo = append(dinfo, newUsers...)

		log.Trace("Finding more...")
		time.Sleep(100 * time.Microsecond)
		uid = g[len(g)-1].User.ID

	}

	log.Printf("Found %d players for %s", len(dinfo), gid)

	return dinfo, nil
}

func (gh guildHandler) guildCreate(s *discordgo.Session, gc *discordgo.GuildCreate) {
	gid := gc.ID
	gcLog := log.WithFields(logrus.Fields{"sys": "guildHandler", "gID": gid, "guildName": gc.Name})

	log.Trace("++ Loading Members")
	members, err := getGuildMembers(gid, s)
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

type guildRemover interface {
	Remove(string) error
}

func newGuildDelete(gr guildRemover) func(*discordgo.Session, *discordgo.GuildDelete) {
	return func(s *discordgo.Session, gd *discordgo.GuildDelete) {
		guildDelete(gr, gd.Guild.ID)
	}
}

func guildDelete(gr guildRemover, gID string) {
	gr.Remove(gID)
}

type userFinder interface {
	GetByDiscordID(models.PlayerDiscordID) (models.User, error)
}

type guildMemberAdder interface {
	GetByDiscordGuild(key string) (models.Account, error)
	AddRegisteredPlayerIDs(accountSnowflake string, playerIDs []models.PlayerID) error
}

func newGuildMemberAdd(uf userFinder, gma guildMemberAdder) func(*discordgo.Session, *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, dgma *discordgo.GuildMemberAdd) {
		log.Tracef("%s#%s", dgma.User.Username, dgma.User.Discriminator)
		guildMemberAdd(uf, gma, dgma.GuildID, dgma.Member.User.ID)
	}
}

func guildMemberAdd(uf userFinder, gma guildMemberAdder, gID, uID string) {
	gmaLog := log.WithFields(logrus.Fields{"sys": "guildMemberAdd", "gID": gID, "uID": uID})
	user, err := uf.GetByDiscordID(models.PlayerDiscordID(uID))
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

func newGuildMemberRemove(uf userFinder, gmr guildMemberRemover) func(*discordgo.Session, *discordgo.GuildMemberRemove) {
	return func(s *discordgo.Session, dgmr *discordgo.GuildMemberRemove) {
		guildMemberRemove(uf, gmr, dgmr.GuildID, dgmr.Member.User.ID)
	}
}

func guildMemberRemove(uf userFinder, gmr guildMemberRemover, gID, uID string) {
	gmrLog := log.WithFields(logrus.Fields{"sys": "guildMemberRemove", "gID": gID, "uID": uID})
	user, err := uf.GetByDiscordID(models.PlayerDiscordID(uID))
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
}
