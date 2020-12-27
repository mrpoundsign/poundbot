package migrations

import (
	migrate "github.com/eminetto/mongo-migrate"
	"github.com/globalsign/mgo"
)

func init() {
	migrate.Register(func(db *mgo.Database) error { //Up
		return db.C("users").DropIndexName("playerids_1")
	}, func(db *mgo.Database) error { //Down
		return nil
	})
}
