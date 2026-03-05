package api

import (
	"net/http"
	"time"
)

// healthResponse is the payload for GET /health.
type healthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// handleHealth returns a simple health-check response.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, healthResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	})
}
