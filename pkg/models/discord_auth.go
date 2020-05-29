package models

type DiscordAuth struct {
	GuildSnowflake string
	PlayerID       PlayerID
	DiscordInfo    `bson:",inline"`
	Pin            int
	Ack            func(bool) `bson:"-" json:"-"`
}

func (d DiscordAuth) GetPlayerID() PlayerID {
	return d.PlayerID
}

func (d DiscordAuth) GetDiscordID() PlayerDiscordID {
	return d.Snowflake
}
