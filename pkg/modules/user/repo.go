package user

import "github.com/poundbot/poundbot/pkg/models"

type Repo interface {
	GetByPlayerID(models.PlayerID) (models.User, error)
	GetByDiscordID(models.PlayerDiscordID) (models.User, error)
	GetPlayerIDsByDiscordIDs([]models.PlayerDiscordID) ([]models.PlayerID, error)
	UpsertPlayer(UserInfoGetter) error
	RemovePlayerID(models.PlayerDiscordID, models.PlayerID) error
	SetGuildUsers(dinfo []models.DiscordInfo, gid string) error
}
