package mongodb

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/pkg/modules/user"
)

const userDiscordNameField = "discordname"
const userPlayerIDsField = "playerids"
const userSnowflakeField = "snowflake"
const userGuildsField = "guildids"

type discordUpsertUser struct {
	DiscordInfo models.DiscordInfo `bson:",inline"`
	Timestamp   models.Timestamp   `bson:",inline"`
}

func newDiscordUpsertUser(di models.DiscordInfo) *discordUpsertUser {
	return &discordUpsertUser{
		DiscordInfo: di,
		Timestamp:   *models.NewTimestamp(),
	}
}

// A Users implements db.UsersStore
type Users struct {
	collection *mgo.Collection
}

// GetByPlayerID get a user by their playerid
func (u Users) GetByPlayerID(gameUserID models.PlayerID) (models.User, error) {
	var user models.User
	err := u.collection.Find(bson.M{userPlayerIDsField: gameUserID}).One(&user)
	return user, err
}

// GetByDiscordID gets a user by their discord snowflake
func (u Users) GetByDiscordID(snowflake models.PlayerDiscordID) (models.User, error) {
	var user models.User
	err := u.collection.Find(bson.M{userSnowflakeField: snowflake}).One(&user)
	return user, err
}

// GetByDiscordName gets a user by their discord name
func (u Users) GetByDiscordName(name string) (models.User, error) {
	var user models.User
	err := u.collection.Find(bson.M{userDiscordNameField: name}).One(&user)
	return user, err
}

func (u Users) GetPlayerIDsByDiscordIDs(snowflakes []models.PlayerDiscordID) ([]models.PlayerID, error) {
	var playerIDs []models.PlayerID
	err := u.collection.Find(bson.M{userSnowflakeField: bson.M{"$in": snowflakes}}).
		Distinct(userPlayerIDsField, &playerIDs)
	return playerIDs, err
}

func (u Users) UpsertPlayer(info user.UserInfoGetter) error {
	_, err := u.collection.Upsert(
		bson.M{userSnowflakeField: info.GetDiscordID()},
		bson.M{
			"$setOnInsert": bson.M{
				userSnowflakeField: info.GetDiscordID(),
				"createdat":        time.Now().UTC(),
			},
			"$set":      bson.M{"updatedat": time.Now().UTC()},
			"$addToSet": bson.M{userPlayerIDsField: info.GetPlayerID()},
		},
	)

	return err
}

func (u Users) RemovePlayerID(snowflake models.PlayerDiscordID, playerID models.PlayerID) error {
	if playerID == "all" {
		err := u.collection.Remove(
			bson.M{userSnowflakeField: snowflake},
		)
		return err
	}

	err := u.collection.Update(
		bson.M{userSnowflakeField: snowflake},
		bson.M{
			"$set": bson.M{
				"updatedat": time.Now().UTC(),
			},
			"$pull": bson.M{userPlayerIDsField: playerID},
		},
	)
	return err
}

// SetGuildUsers updates users discord info for the given guild
func (u Users) SetGuildUsers(dis []models.DiscordInfo, gid string) error {
	pids := make([]models.PlayerDiscordID, len(dis))
	for i, di := range dis {
		pids[i] = di.Snowflake
		err := u.AddGuildUser(di, gid)

		if err != nil {
			return fmt.Errorf("setguildplayers: could not update user: %w", err)
		}
	}

	_, err := u.collection.UpdateAll(
		bson.M{
			userSnowflakeField: bson.M{"$nin": pids},
			userGuildsField:    gid,
		},
		bson.M{"$pull": bson.M{userGuildsField: gid}},
	)

	if err != nil {
		return fmt.Errorf("setguildplayers: could remove users: %w", err)
	}

	return nil
}

func (u Users) AddGuildUser(di models.DiscordInfo, gid string) error {
	_, err := u.collection.Upsert(
		bson.M{userSnowflakeField: di.Snowflake},
		bson.M{
			"$setOnInsert": *newDiscordUpsertUser(di),
			"$addToSet":    bson.M{userGuildsField: gid},
		},
	)

	return err
}

func (u Users) RemoveGuildUser(di models.DiscordInfo, gid string) error {
	err := u.collection.Update(
		bson.M{userSnowflakeField: di.Snowflake},
		bson.M{
			"$pull": bson.M{userGuildsField: gid},
		},
	)

	return err
}

func (u Users) RemoveGuild(gid string) error {
	log.Printf("removing guild %s", gid)
	err := u.collection.Update(
		bson.M{userGuildsField: gid},
		bson.M{
			"$pull": bson.M{userGuildsField: gid},
		},
	)

	if err != nil {
		return err
	}

	info, err := u.collection.RemoveAll(
		bson.M{userGuildsField: []string{}},
	)

	if err != nil {
		return err
	}

	log.Printf("Removed %d of %d users for guild %s", info.Removed, info.Matched, gid)

	return nil
}
