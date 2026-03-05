package api

import (
	"errors"
	"net/http"

	"devopstoolkit/youtube-automation/internal/aspect"

	"github.com/go-chi/chi/v5"
)

// CompletionCriteriaResponse wraps a single field's completion criteria.
type CompletionCriteriaResponse struct {
	AspectKey          string `json:"aspectKey"`
	FieldKey           string `json:"fieldKey"`
	CompletionCriteria string `json:"completionCriteria"`
}

// handleGetAspects returns all aspects with full field metadata.
func (s *Server) handleGetAspects(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, s.aspectService.GetAspects())
}

// handleGetAspectsOverview returns lightweight aspect summaries without fields.
func (s *Server) handleGetAspectsOverview(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, s.aspectService.GetAspectsOverview())
}

// handleGetAspectFields returns fields for a specific aspect.
func (s *Server) handleGetAspectFields(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	fields, err := s.aspectService.GetAspectFields(key)
	if err != nil {
		if errors.Is(err, aspect.ErrAspectNotFound) {
			respondError(w, http.StatusNotFound, "aspect not found", key)
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get aspect fields", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, fields)
}

// handleGetFieldCompletion returns the completion criteria for a specific field.
func (s *Server) handleGetFieldCompletion(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	field := chi.URLParam(r, "field")

	criteria := s.aspectService.GetFieldCompletionCriteria(key, field)

	respondJSON(w, http.StatusOK, CompletionCriteriaResponse{
		AspectKey:          key,
		FieldKey:           field,
		CompletionCriteria: criteria,
	})
}
