package user

import "github.com/poundbot/poundbot/pkg/models"

type Repo interface {
	GetByPlayerID(models.PlayerID) (models.User, error)
	GetByDiscordID(models.PlayerDiscordID) (models.User, error)
	GetByDiscordName(name string) (models.User, error)
	GetPlayerIDsByDiscordIDs([]models.PlayerDiscordID) ([]models.PlayerID, error)
	UpsertPlayer(UserInfoGetter) error
	RemovePlayerID(models.PlayerDiscordID, models.PlayerID) error
	SetGuildUsers(dinfos []models.DiscordInfo, gid string) error
	AddGuildUser(di models.DiscordInfo, gid string) error
	RemoveGuildUser(di models.DiscordInfo, gid string) error
	RemoveGuild(gid string) error
}
