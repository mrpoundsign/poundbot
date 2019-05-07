package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/globalsign/mgo/bson"
)

type Server struct {
	Name         string
	Key          string
	Address      string
	Clans        []Clan
	ChatChanID   string
	ServerChanID string
	RaidDelay    string
	Timestamp    `bson:",inline"`
}

func (s Server) UsersClan(playerIDs []string) (bool, Clan) {
	for _, clan := range s.Clans {
		for _, member := range clan.Members {
			for _, id := range playerIDs {
				if member == id {
					return true, clan
				}
			}
		}
	}
	return false, Clan{}
}

type BaseAccount struct {
	GuildSnowflake         string
	OwnerSnowflake         string
	CommandPrefix          string
	AdminSnowflakes        []string `bson:",omitempty"`
	AuthenticatedPlayerIDs []string
}

type Account struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	BaseAccount `bson:",inline" json:",inline"`
	Servers     []Server `bson:",omitempty"`
	Timestamp   `bson:",inline" json:",inline"`
	Disabled    bool
}

// ServerFromKey finds a server for a given key. Errors if not found.
func (a Account) ServerFromKey(apiKey string) (Server, error) {
	for i := range a.Servers {
		if a.Servers[i].Key == apiKey {
			return a.Servers[i], nil
		}
	}
	return Server{}, errors.New("server not found")
}

// GetCommandPrefix the Discord command prefix. Defaults to "!pb"
func (a Account) GetCommandPrefix() string {
	if a.CommandPrefix == "" {
		return "!pb"
	}
	return a.CommandPrefix
}

// GetAdminIDs returns the Discord IDs considered "admins"
func (a Account) GetAdminIDs() []string {
	return append(a.AdminSnowflakes, a.OwnerSnowflake)
}

// GetRegisteredPlayerIDs returns a list of player IDs
// for the game requested. These ids are stripped of their
// prefix (e.g. "rust:1001" would be "1001")
func (a Account) GetRegisteredPlayerIDs(game string) []string {
	ids := []string{}
	gamePrefix := game + ":"
	for _, id := range a.AuthenticatedPlayerIDs {
		if strings.HasPrefix(id, gamePrefix) {
			ids = append(ids, id[len(gamePrefix):])
		}
	}
	return ids
}

// Clan is a clan from the game
type Clan struct {
	Tag        string
	OwnerID    string
	Members    []string `bson:",omitempty"`
	Moderators []string `bson:",omitempty"`
}

// SetGame adds game name to all IDs
func (c *Clan) SetGame(game string) {
	c.OwnerID = fmt.Sprintf("%s:%s", game, c.OwnerID)
	for i := range c.Members {
		c.Members[i] = fmt.Sprintf("%s:%s", game, c.Members[i])
	}

	for i := range c.Moderators {
		c.Moderators[i] = fmt.Sprintf("%s:%s", game, c.Moderators[i])
	}
}
