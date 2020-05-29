package mongodb

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/pkg/models"
)

// A DiscordAuths implements db.DiscordAuthsStore
type DiscordAuths struct {
	collection *mgo.Collection
}

func (d DiscordAuths) GetByDiscordName(discordName string) (models.DiscordAuth, error) {
	var da models.DiscordAuth
	err := d.collection.Find(bson.M{"discordname": discordName}).One(&da)
	return da, err
}

func (d DiscordAuths) GetByDiscordID(snowflake models.PlayerDiscordID) (models.DiscordAuth, error) {
	var da models.DiscordAuth
	err := d.collection.Find(bson.M{"snowflake": snowflake}).One(&da)
	if err != nil {
		return models.DiscordAuth{}, fmt.Errorf("mongodb could not find snowflake %s (%s)", snowflake, err)
	}
	return da, nil
}

// Remove implements db.DiscordAuthsStore.Remove
func (d DiscordAuths) Remove(playerID models.PlayerID) error {
	return d.collection.Remove(bson.M{"playerid": playerID})
}

// Upsert implements db.DiscordAuthsStore.Upsert
func (d DiscordAuths) Upsert(da models.DiscordAuth) error {
	_, err := d.collection.Upsert(
		bson.M{"playerid": da.PlayerID},
		da,
	)
	return err
}
