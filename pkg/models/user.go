package models

import "fmt"

type PlayerID string

func (pid PlayerID) String() string {
	return string(pid)
}

func (pid PlayerID) withGame(game string) string {
	return fmt.Sprintf("%s:%s", game, pid)
}

func (pid PlayerID) PlayerGameID(game string) PlayerID {
	return PlayerID(pid.withGame(game))
}

type PlayerDiscordID string

func (pdid PlayerDiscordID) String() string {
	return string(pdid)
}

type DiscordInfo struct {
	DiscordName string
	Snowflake   PlayerDiscordID
}

// GamesInfo steam id translator between server and DB
// also used as a selector on the DB
type GamesInfo struct {
	PlayerName string     `bson:",omitempty"`
	PlayerIDs  []PlayerID `bson:",omitempty"`
}

func (g GamesInfo) PlayerIDsAsStrings() []string {
	s := make([]string, len(g.PlayerIDs))
	for i, id := range g.PlayerIDs {
		s[i] = string(id)
	}
	return s
}

// BaseUser core user information for upserts
type BaseUser struct {
	GamesInfo   `bson:",inline" json:",inline"`
	DiscordInfo `bson:",inline" json:",inline"`
}

// User full user model
type User struct {
	BaseUser  `bson:",inline" json:",inline"`
	Timestamp `bson:",inline" json:",inline"`
}
