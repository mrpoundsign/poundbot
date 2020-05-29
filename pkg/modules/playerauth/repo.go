package playerauth

import "github.com/poundbot/poundbot/pkg/models"

type Repo interface {
	GetByDiscordName(discordName string) (models.DiscordAuth, error)
	GetByDiscordID(models.PlayerDiscordID) (models.DiscordAuth, error)
	Upsert(models.DiscordAuth) error
	Remove(models.PlayerID) error
}
