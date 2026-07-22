package api

import (
	"log/slog"
	"net/http"
)

// handleGetCategories returns the available video categories.
//
// Categories are derived from the manuscript directories in the data repo, so
// a category added externally (a new directory pushed to the git remote) only
// appears once the local clone has pulled it. Best-effort throttled pull
// refreshes the clone here; failures are logged but never block the response,
// which is served from the local working copy regardless.
func (s *Server) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	if s.gitSync != nil {
		if err := s.gitSync.PullIfStale(pullOnReadThrottle); err != nil {
			slog.Warn("categories: git pull failed, serving local copy", "err", err)
		}
	}

	categories, err := s.videoService.GetCategories()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get categories", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, categories)
}
