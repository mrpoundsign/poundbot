package user

import "github.com/poundbot/poundbot/pkg/models"

type Service interface {
	GetByPlayerID(models.PlayerID) (models.User, error)
	GetByDiscordID(models.PlayerDiscordID) (models.User, error)
	GetPlayerIDsByDiscordIDs(snowflakes []models.PlayerDiscordID) ([]models.PlayerID, error)
	UpsertPlayer(info UserInfoGetter) error
	RemovePlayerID(models.PlayerDiscordID, models.PlayerID) error
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) GetByPlayerID(PlayerID models.PlayerID) (models.User, error) {
	return s.repo.GetByPlayerID(PlayerID)
}

func (s *service) GetByDiscordID(snowflake models.PlayerDiscordID) (models.User, error) {
	return s.repo.GetByDiscordID(snowflake)
}

func (s *service) GetPlayerIDsByDiscordIDs(snowflakes []models.PlayerDiscordID) ([]models.PlayerID, error) {
	return s.repo.GetPlayerIDsByDiscordIDs(snowflakes)
}

func (s *service) UpsertPlayer(info UserInfoGetter) error {
	return s.repo.UpsertPlayer(info)
}

func (s *service) RemovePlayerID(snowflake models.PlayerDiscordID, playerID models.PlayerID) error {
	return s.repo.RemovePlayerID(snowflake, playerID)
}
