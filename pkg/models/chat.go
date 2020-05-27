package models

import "github.com/globalsign/mgo/bson"

type ChatMessage struct {
	ID           bson.ObjectId `bson:"_id,omitempty"`
	DiscordInfo  `bson:",inline" json:"-"`
	ChannelID    string `json:"-"`
	ClanTag      string
	DisplayName  string
	Message      string
	PlayerID     string
	ServerKey    string `json:"-"`
	SentToServer bool   `json:"-"`
	Tag          string
}
