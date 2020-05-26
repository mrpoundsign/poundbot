package gameapi

import (
	"net/http"
	"strings"

	"context"

	"github.com/blang/semver"
	"github.com/poundbot/poundbot/types"
)

type serverAuthenticator interface {
	GetByServerKey(serverKey string) (types.Account, error)
	Touch(serverKey string) error
}

func newServerAuth(as serverAuthenticator) *serverAuth {
	return &serverAuth{
		as:         as,
		minVersion: semver.Version{Major: 2},
	}
}

type serverAuth struct {
	as         serverAuthenticator
	minVersion semver.Version
}

func (sa serverAuth) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			version, err := semver.Make(r.Header.Get("X-PoundBotConnector-Version"))
			if err != nil {
				handleError(w, types.RESTError{
					StatusCode: http.StatusBadRequest,
					Error:      "PoundBot must be updated. Please download the latest version at" + upgradeURL,
				})
				return
			}
			if version.LT(sa.minVersion) {
				handleError(w, types.RESTError{
					StatusCode: http.StatusBadRequest,
					Error:      "PoundBot must be updated. Please download the latest version at" + upgradeURL,
				})
				return
			}

			s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				handleError(w, types.RESTError{
					StatusCode: http.StatusUnauthorized,
					Error:      "Authorization header is incorrect.",
				})
				return
			}

			game := r.Header.Get("X-PoundBot-Game")

			if len(game) == 0 {
				handleError(w, types.RESTError{
					StatusCode: http.StatusBadRequest,
					Error:      "Missing X-PoundBot-Game header.",
				})
				return
			}

			account, err := sa.as.GetByServerKey(s[1])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			err = sa.as.Touch(s[1])
			if err != nil {
				handleError(w, types.RESTError{
					StatusCode: http.StatusInternalServerError,
					Error:      "Error updating server account",
				})
				log.Printf("Error updating %s (touch)", account.ID)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyServerKey, s[1])
			ctx = context.WithValue(ctx, contextKeyAccount, account)
			ctx = context.WithValue(ctx, contextKeyGame, game)

			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
