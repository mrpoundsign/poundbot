package gameapi

import (
	"encoding/json"
	"net/http"

	"github.com/poundbot/poundbot/types"
)

// handleError is a generic JSON HTTP error response
func handleError(w http.ResponseWriter, restError types.RESTError) error {
	setJSONContentType(w, restError.StatusCode)
	return json.NewEncoder(w).Encode(restError)
}

func methodNotAllowed(w http.ResponseWriter) {
	handleError(w, types.RESTError{
		StatusCode: http.StatusMethodNotAllowed,
		Error:      "Method not allowed",
	})
}

func setJSONContentType(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	w.Header().Set("content-type", "application/json; cahrset=utf-8")
}
