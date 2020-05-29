package account

import "github.com/poundbot/poundbot/pkg/models"

type Service interface {
	All(*[]models.Account) error
	GetByDiscordGuild(snowflake string) (models.Account, error)
	GetByServerKey(serverKey string) (models.Account, error)
	UpsertBase(models.BaseAccount) error
	Remove(guildsnowflakee string) error

	AddServer(snowflake string, server models.AccountServer) error
	UpdateServer(snowflake, oldKey string, server models.AccountServer) error
	RemoveServer(snowflake, serverKey string) error

	AddClan(serverKey string, clan models.Clan) error
	RemoveClan(serverKey, clanTag string) error
	SetClans(serverKey string, clans []models.Clan) error

	SetRegisteredPlayerIDs(accountID string, playerIDsList []models.PlayerID) error
	AddRegisteredPlayerIDs(accountID string, playerIDs []models.PlayerID) error
	RemoveRegisteredPlayerIDs(accountID string, playerIDs []models.PlayerID) error

	RemoveNotInDiscordGuildList(guildIDs []models.BaseAccount) error
	Touch(serverKey string) error
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return service{repo: repo}
}

func (s service) All(accounts *[]models.Account) error {
	return s.repo.All(accounts)
}

func (s service) GetByDiscordGuild(snowflake string) (models.Account, error) {
	return s.repo.GetByDiscordGuild(snowflake)
}

func (s service) GetByServerKey(serverKey string) (models.Account, error) {
	return s.repo.GetByServerKey(serverKey)
}

func (s service) UpsertBase(account models.BaseAccount) error {
	return s.repo.UpsertBase(account)
}

func (s service) Remove(guildsnowflake string) error {
	return s.repo.Remove(guildsnowflake)
}

func (s service) AddServer(snowflake string, server models.AccountServer) error {
	return s.repo.AddServer(snowflake, server)
}

func (s service) UpdateServer(snowflake, oldKey string, server models.AccountServer) error {
	return s.repo.UpdateServer(snowflake, oldKey, server)
}

func (s service) RemoveServer(snowflake, serverKey string) error {
	return s.repo.RemoveServer(snowflake, serverKey)
}

func (s service) AddClan(serverKey string, clan models.Clan) error {
	return s.repo.AddClan(serverKey, clan)
}

func (s service) RemoveClan(serverKey, clanTag string) error {
	return s.repo.RemoveClan(serverKey, clanTag)
}

func (s service) SetClans(serverKey string, clans []models.Clan) error {
	return s.repo.SetClans(serverKey, clans)
}

func (s service) SetRegisteredPlayerIDs(accountID string, playerIDsList []models.PlayerID) error {
	return s.repo.SetRegisteredPlayerIDs(accountID, playerIDsList)
}

func (s service) AddRegisteredPlayerIDs(accountID string, playerIDs []models.PlayerID) error {
	return s.repo.AddRegisteredPlayerIDs(accountID, playerIDs)
}

func (s service) RemoveRegisteredPlayerIDs(accountID string, playerIDs []models.PlayerID) error {
	return s.repo.RemoveRegisteredPlayerIDs(accountID, playerIDs)
}

func (s service) RemoveNotInDiscordGuildList(guildIDs []models.BaseAccount) error {
	return s.repo.RemoveNotInDiscordGuildList(guildIDs)
}

func (s service) Touch(serverKey string) error {
	return s.repo.Touch(serverKey)
}
