package api

import (
	"net/http"
	"sort"

	"devopstoolkit/youtube-automation/internal/workflow"
)

// PhaseInfo describes a single lifecycle phase.
type PhaseInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// handleGetPhases returns the list of phases with video counts.
func (s *Server) handleGetPhases(w http.ResponseWriter, r *http.Request) {
	counts, err := s.videoService.GetVideoPhases()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get phases", err.Error())
		return
	}

	phases := make([]PhaseInfo, 0, len(workflow.PhaseNames))
	for id, name := range workflow.PhaseNames {
		phases = append(phases, PhaseInfo{
			ID:    id,
			Name:  name,
			Count: counts[id],
		})
	}

	sort.Slice(phases, func(i, j int) bool {
		return phases[i].ID < phases[j].ID
	})

	respondJSON(w, http.StatusOK, phases)
}
