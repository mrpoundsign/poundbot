// +build integration

package mongodb

import (
	"testing"

	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/storage/mongodb/mongotest"
	"github.com/stretchr/testify/assert"
)

var baseUser = models.BaseUser{
	GamesInfo: models.GamesInfo{PlayerIDs: []models.PlayerID{"pid1"}},
	DiscordInfo: models.DiscordInfo{
		Snowflake: "did1",
	},
}

func NewUsers(t *testing.T) (*Users, *mongotest.Collection) {
	coll, err := mongotest.NewCollection(usersCollection)
	if err != nil {
		t.Fatal(err)
	}
	return &Users{collection: coll.C}, coll
}

func TestUsers_GetByPlayerID(t *testing.T) {
	t.Parallel()

	type args struct {
		gameUserID models.PlayerID
	}
	tests := []struct {
		name    string
		args    args
		want    *models.User
		wantErr bool
	}{
		{
			name: "found",
			args: args{gameUserID: "pid1"},
			want: &models.User{BaseUser: baseUser},
		},
		{
			name:    "not found",
			args:    args{gameUserID: "notfound"},
			want:    &models.User{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, coll := NewUsers(t)
			defer coll.Close()

			users.collection.Insert(baseUser)

			got, err := users.GetByPlayerID(tt.args.gameUserID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Users.GetByPlayerID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, &got) {
				t.Errorf("Users.GetByPlayerID() = %v, want %v", got, tt.want)
			}
		})
	}
}

type player struct {
	id  models.PlayerID
	did models.PlayerDiscordID
}

func (p player) GetPlayerID() models.PlayerID {
	return p.id
}

func (p player) GetDiscordID() models.PlayerDiscordID {
	return p.did
}

func TestUsers_UpsertPlayer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		player    player
		wantCount int
		wantErr   bool
	}{
		{
			name:      "insert",
			player:    player{id: "pid2", did: "did2"},
			wantCount: 2,
		},
		{
			name:      "upsert",
			player:    player{id: "pid2", did: "did1"},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, coll := NewUsers(t)
			defer coll.Close()

			err := users.collection.Insert(baseUser)
			if err != nil {
				t.Fatal(err)
			}

			err = users.UpsertPlayer(tt.player)
			if err != nil {
				t.Fatal(err)
			}

			count, err := users.collection.Count()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantCount, count)
		})
	}
}
