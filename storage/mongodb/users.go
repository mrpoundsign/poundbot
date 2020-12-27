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

// A Users implements db.UsersStore
type Users struct {
	collection *mgo.Collection
}

// Get implements db.UsersStore.Get
func (u Users) GetByPlayerID(gameUserID models.PlayerID) (models.User, error) {
	var user models.User
	err := u.collection.Find(bson.M{userPlayerIDsField: gameUserID}).One(&user)
	return user, err
}

func (u Users) GetByDiscordID(snowflake models.PlayerDiscordID) (models.User, error) {
	var user models.User
	err := u.collection.Find(bson.M{userSnowflakeField: snowflake}).One(&user)
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
func (u Users) SetGuildUsers(dinfo []models.DiscordInfo, gid string) error {
	pids := make([]models.PlayerDiscordID, len(dinfo))
	for i, d := range dinfo {
		pids[i] = d.Snowflake
		_, err := u.collection.Upsert(
			bson.M{userSnowflakeField: d.Snowflake},
			bson.M{
				"$setOnInsert": bson.M{
					userSnowflakeField:   d.Snowflake,
					userDiscordNameField: d.DiscordName,
					userPlayerIDsField:   []string{},
					"createdat":          time.Now().UTC(),
					"updatedat":          time.Now().UTC(),
				},
				"$addToSet": bson.M{userGuildsField: gid},
			},
		)

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
