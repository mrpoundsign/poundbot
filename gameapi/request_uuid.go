package gameapi

import (
	"net/http"

	"context"

	"github.com/gofrs/uuid"
	"github.com/poundbot/poundbot/types"
)

func requestUUID(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestUUID := r.Header.Get("X-Request-ID")
			if len(requestUUID) == 0 {
				rUUID, err := uuid.NewV4()
				if err != nil {
					handleError(w, types.RESTError{
						StatusCode: http.StatusInternalServerError,
						Error:      "could not create UUID",
					})
					return
				}
				requestUUID = rUUID.String()
			}

			ctx := context.WithValue(r.Context(), contextKeyRequestUUID, requestUUID)
			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", requestUUID)

			next.ServeHTTP(w, r)
		},
	)
}
