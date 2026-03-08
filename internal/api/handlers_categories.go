package api

import (
	"net/http"
)

// handleGetCategories returns the available video categories.
func (s *Server) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.videoService.GetCategories()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get categories", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, categories)
}
