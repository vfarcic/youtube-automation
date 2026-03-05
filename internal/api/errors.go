package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard error payload returned by the API.
type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// respondError writes a standardized JSON error response.
func respondError(w http.ResponseWriter, status int, errMsg, detail string) {
	respondJSON(w, status, ErrorResponse{
		Error:  errMsg,
		Detail: detail,
	})
}
