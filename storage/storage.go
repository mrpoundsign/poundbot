package storage

import (
	"time"

	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/pkg/modules/account"
	"github.com/poundbot/poundbot/pkg/modules/playerauth"
	"github.com/poundbot/poundbot/pkg/modules/user"
)

// type UserInfoGetter interface {
// 	GetPlayerID() string
// 	GetDiscordID() string
// }

type ChatQueueStore interface {
	GetGameServerMessage(serverKey, tag string, timeout time.Duration) (message models.ChatMessage, success bool)
	InsertMessage(message models.ChatMessage) error
}

type MessageLocksStore interface {
	Obtain(mID, mType string) bool
}

// UsersStore is for accessing the user store.
//
// Get gets a user from store.
//
// UpsertBase updates or creates a user in the store
//
// RemoveClan removes a clan tag from all users e.g. when a clan is removed.
//
// RemoveClansNotIn is used for removing all clan tags not in the provided
// list from all users in the data store.
//
// SetClanIn sets the clan tag on all users who have the provided steam IDs.
// type UsersStore interface {
// 	GetByPlayerID(PlayerID string) (models.User, error)
// 	GetByDiscordID(snowflake string) (models.User, error)
// 	GetPlayerIDsByDiscordIDs(snowflakes []string) ([]string, error)
// 	UpsertPlayer(info user.UserInfoGetter) error
// 	RemovePlayerID(snowflake, playerID string) error
// }

// DiscordAuthsStore is for accessing the discord -> user authentications
// in the store.
//
// Upsert created or updates a discord auth
//
// Remove removes a discord auth
// type DiscordAuthsStore interface {
// 	GetByDiscordName(discordName string) (models.DiscordAuth, error)
// 	GetByDiscordID(snowflake string) (models.DiscordAuth, error)
// 	Upsert(models.DiscordAuth) error
// 	Remove(playerauth.UserInfoGetter) error
// }

// RaidAlertsStore is for accessing raid information. The raid information
// comes in as models.EntityDeath and comes out as models.RaidAlert
//
// GetReady gets raid alerts that are ready to alert
//
// AddInfo adds or updated raid information to a raid alert
//
// Remove deletes a raid alert
type RaidAlertsStore interface {
	GetReady() ([]models.RaidAlert, error)
	AddInfo(alertIn, validUntil time.Duration, ed models.EntityDeath) error
	Remove(models.RaidAlert) error
	IncrementNotifyCount(models.RaidAlert) error
	SetMessageID(models.RaidAlert, string) error
}

// AccountsStore is for accounts storage
// type AccountsStore interface {
// 	All(*[]models.Account) error
// 	GetByDiscordGuild(snowflake string) (models.Account, error)
// 	GetByServerKey(serverKey string) (models.Account, error)
// 	UpsertBase(models.BaseAccount) error
// 	Remove(snowflake string) error

// 	AddServer(snowflake string, server models.AccountServer) error
// 	UpdateServer(snowflake, oldKey string, server models.AccountServer) error
// 	RemoveServer(snowflake, serverKey string) error

// 	AddClan(serverKey string, clan models.Clan) error
// 	RemoveClan(serverKey, clanTag string) error
// 	SetClans(serverKey string, clans []models.Clan) error

// 	SetRegisteredPlayerIDs(accountID string, playerIDsList []string) error
// 	AddRegisteredPlayerIDs(accountID string, playerIDs []string) error
// 	RemoveRegisteredPlayerIDs(accountID string, playerIDs []string) error

// 	RemoveNotInDiscordGuildList(guildIDs []models.BaseAccount) error
// 	Touch(serverKey string) error
// }

// Storage is a complete implementation of the data store for users,
// clans, discord auth requests, raid alerts, and chats.
//
// Copy creates a new DB connection. Should always close the connection when
// you're done with it.
//
// Close closes the session
//
// Init creates indexes, and should always be called when Poundbot
// first starts
type Storage interface {
	Copy() Storage
	Close()
	Init()
	Accounts() account.Repo
	Users() user.Repo
	DiscordAuths() playerauth.Repo
	RaidAlerts() RaidAlertsStore
	ChatQueue() ChatQueueStore
}
