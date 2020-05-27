package gameapi

import (
	"testing"

	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/storage/mocks"
)

func TestAuthSaver_Run(t *testing.T) {
	t.Parallel()

	var mockU *mocks.UsersStore
	var mockDA *mocks.DiscordAuthsStore
	done := make(chan struct{})
	defer close(done)

	tests := []struct {
		name string
		a    func(<-chan models.DiscordAuth) *AuthSaver
		with *models.DiscordAuth
		want *models.DiscordAuth
	}{
		{
			name: "With nothing",
			a: func(ch <-chan models.DiscordAuth) *AuthSaver {
				return newAuthSaver(mockDA, mockU, ch, done)
			},
		},
		{
			name: "With AuthSuccess",

			a: func(ch <-chan models.DiscordAuth) *AuthSaver {
				result := models.DiscordAuth{PlayerID: "game:1001"}
				mockU = &mocks.UsersStore{}
				mockU.On("UpsertPlayer", result).Return(nil)

				mockDA = &mocks.DiscordAuthsStore{}
				mockDA.On("Remove", result).Return(nil)

				return newAuthSaver(mockDA, mockU, ch, done)
			},
			with: &models.DiscordAuth{PlayerID: "game:1001"},
			want: &models.DiscordAuth{PlayerID: "game:1001"},
		},
	}
	for _, tt := range tests {
		mockU = nil
		mockDA = nil
		for len(done) > 0 {
			<-done
		}

		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan models.DiscordAuth)
			var server = tt.a(ch)

			go func() {
				defer func() { done <- struct{}{} }()
				defer close(ch)
				if tt.with != nil {
					ch <- *tt.with
				}
			}()

			server.Run()
			if mockU != nil {
				mockU.AssertExpectations(t)
			}
			if mockDA != nil {
				mockDA.AssertExpectations(t)
			}
		})
	}
}
