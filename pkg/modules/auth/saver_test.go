package auth

import (
	"errors"
	"testing"

	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/pkg/modules/user"
)

type testUserUpserter struct{}

func (t testUserUpserter) UpsertPlayer(uig user.UserInfoGetter) error {
	if uig.GetPlayerID() == "error:upsert" {
		return errors.New("error")
	}
	return nil
}

type testDiscordAuthRemover struct{}

func (t testDiscordAuthRemover) Remove(playerID models.PlayerID) error {
	if playerID == "error:remove" {
		return errors.New("error")
	}
	return nil
}

func TestAuthSaver_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		with *models.DiscordAuth
		want *models.DiscordAuth
	}{
		{
			name: "With nothing",
		},
		{
			name: "With AuthSuccess",
			with: &models.DiscordAuth{PlayerID: "game:1001"},
			want: &models.DiscordAuth{PlayerID: "game:1001"},
		},
		{
			name: "With Upsert failure",
			with: &models.DiscordAuth{PlayerID: "error:upsert"},
			want: &models.DiscordAuth{PlayerID: "game:1001"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan interface{})
			defer func() {
				for len(done) > 0 {
					<-done
				}
				close(done)
			}()

			ch := make(chan models.DiscordAuth)
			go func() {
				defer func() { done <- nil }()
				defer close(ch)
				if tt.with != nil {
					ch <- *tt.with
				}
			}()

			server := saver{
				das:         testDiscordAuthRemover{},
				us:          testUserUpserter{},
				authSuccess: ch,
				done:        done,
			}

			server.Run()
		})
	}
}
