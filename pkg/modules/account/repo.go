package account

import "github.com/poundbot/poundbot/pkg/models"

type Repo interface {
	All(*[]models.Account) error
	GetByDiscordGuild(snowflake string) (models.Account, error)
	GetByServerKey(serverKey string) (models.Account, error)
	UpsertBase(models.BaseAccount) error
	Remove(guildsnowflake string) error

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
