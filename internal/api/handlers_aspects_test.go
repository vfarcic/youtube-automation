package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devopstoolkit/youtube-automation/internal/aspect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetAspects(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/aspects/", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result aspect.AspectMetadata
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, 7, len(result.Aspects))

	// Verify aspects are ordered
	for i := 1; i < len(result.Aspects); i++ {
		assert.LessOrEqual(t, result.Aspects[i-1].Order, result.Aspects[i].Order)
	}
}

func TestHandleGetAspectsOverview(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/aspects/overview", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result aspect.AspectOverview
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, 7, len(result.Aspects))

	// Each summary should have a non-zero field count
	for _, a := range result.Aspects {
		assert.NotEmpty(t, a.Key)
		assert.Greater(t, a.FieldCount, 0)
	}
}

func TestHandleGetAspectFields(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("valid aspect key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/aspects/initial-details/fields", nil)
		rec := httptest.NewRecorder()
		env.server.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result aspect.AspectFields
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
		assert.Equal(t, "initial-details", result.AspectKey)
		assert.NotEmpty(t, result.Fields)
	})

	t.Run("invalid aspect key returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/aspects/nonexistent/fields", nil)
		rec := httptest.NewRecorder()
		env.server.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandleGetFieldCompletion(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/aspects/initial-details/fields/gist/completion", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result CompletionCriteriaResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "initial-details", result.AspectKey)
	assert.Equal(t, "gist", result.FieldKey)
	assert.NotEmpty(t, result.CompletionCriteria)
}
