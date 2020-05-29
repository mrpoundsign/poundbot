package playerauth

import "github.com/poundbot/poundbot/pkg/models"

type Service interface {
	GetByDiscordID(models.PlayerDiscordID) (models.DiscordAuth, error)
	GetByDiscordName(discordName string) (models.DiscordAuth, error)
	Upsert(models.DiscordAuth) error
	Remove(models.PlayerID) error
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) GetByDiscordID(snowflake models.PlayerDiscordID) (models.DiscordAuth, error) {
	return s.repo.GetByDiscordID(snowflake)
}

func (s *service) GetByDiscordName(discordName string) (models.DiscordAuth, error) {
	return s.repo.GetByDiscordName(discordName)
}

func (s *service) Upsert(auth models.DiscordAuth) error {
	return s.repo.Upsert(auth)
}

func (s *service) Remove(playerID models.PlayerID) error {
	return s.repo.Remove(playerID)
}
