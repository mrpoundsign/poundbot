package user

import "github.com/poundbot/poundbot/pkg/models"

type UserInfoGetter interface {
	GetPlayerID() models.PlayerID
	GetDiscordID() models.PlayerDiscordID
}
