package models

import "fmt"

type RoleSet struct {
	GuildID   string `json:"-"`
	Role      string
	PlayerIDs []PlayerID
}

func (gs *RoleSet) SetGame(game string) {
	for i := range gs.PlayerIDs {
		gs.PlayerIDs[i] = PlayerID(fmt.Sprintf("%s:%s", game, gs.PlayerIDs[i]))
	}
}
