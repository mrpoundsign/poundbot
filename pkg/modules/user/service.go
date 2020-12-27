package user

import "github.com/poundbot/poundbot/pkg/models"

// type Service interface {
// 	GetByPlayerID(models.PlayerID) (models.User, error)
// 	GetByDiscordID(models.PlayerDiscordID) (models.User, error)
// 	GetPlayerIDsByDiscordIDs(snowflakes []models.PlayerDiscordID) ([]models.PlayerID, error)
// 	UpsertPlayer(info UserInfoGetter) error
// 	RemovePlayerID(models.PlayerDiscordID, models.PlayerID) error
// }

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByPlayerID(PlayerID models.PlayerID) (models.User, error) {
	return s.repo.GetByPlayerID(PlayerID)
}

func (s *Service) GetByDiscordID(snowflake models.PlayerDiscordID) (models.User, error) {
	return s.repo.GetByDiscordID(snowflake)
}

func (s *Service) GetPlayerIDsByDiscordIDs(snowflakes []models.PlayerDiscordID) ([]models.PlayerID, error) {
	return s.repo.GetPlayerIDsByDiscordIDs(snowflakes)
}

func (s *Service) UpsertPlayer(info UserInfoGetter) error {
	return s.repo.UpsertPlayer(info)
}

func (s *Service) RemovePlayerID(snowflake models.PlayerDiscordID, playerID models.PlayerID) error {
	return s.repo.RemovePlayerID(snowflake, playerID)
}

func (s *Service) SetGuildUsers(dinfo []models.DiscordInfo, gid string) error {
	return s.repo.SetGuildUsers(dinfo, gid)
}
