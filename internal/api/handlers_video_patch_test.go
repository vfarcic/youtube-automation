package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- PATCH tests ------------------------------------------------------------

func TestHandlePatchVideoAspect_StringField(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	body := `{"projectName":"NewProject"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?category=devops&aspect=initial-details", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result VideoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "NewProject", result.ProjectName)
}

func TestHandlePatchVideoAspect_BoolField(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	body := `{"code":true}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?category=devops&aspect=work-progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result VideoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.True(t, result.Code)
}

func TestHandlePatchVideoAspect_NestedField(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	body := `{"sponsorship.amount":"5000"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?category=devops&aspect=initial-details", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result VideoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "5000", result.Sponsorship.Amount)
}

func TestHandlePatchVideoAspect_MissingCategory(t *testing.T) {
	env := setupTestEnv(t)

	body := `{"projectName":"X"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?aspect=initial-details", strings.NewReader(body))
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlePatchVideoAspect_MissingAspect(t *testing.T) {
	env := setupTestEnv(t)

	body := `{"projectName":"X"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?category=devops", strings.NewReader(body))
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlePatchVideoAspect_InvalidAspect(t *testing.T) {
	env := setupTestEnv(t)

	body := `{"projectName":"X"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?category=devops&aspect=bogus", strings.NewReader(body))
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlePatchVideoAspect_UnknownFieldForAspect(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	// "code" belongs to work-progress, not initial-details
	body := `{"code":true}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?category=devops&aspect=initial-details", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlePatchVideoAspect_VideoNotFound(t *testing.T) {
	env := setupTestEnv(t)

	body := `{"projectName":"X"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/no-such-video?category=devops&aspect=initial-details", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandlePatchVideoAspect_InvalidJSON(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?category=devops&aspect=initial-details", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlePatchVideoAspect_PreservesUnchangedFields(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{
		Name:        "test-vid",
		Category:    "devops",
		ProjectName: "OldProject",
		ProjectURL:  "https://example.com",
	})

	// Only update projectName, projectURL should remain
	body := `{"projectName":"NewProject"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/videos/test-vid?category=devops&aspect=initial-details", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result VideoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "NewProject", result.ProjectName)
	assert.Equal(t, "https://example.com", result.ProjectURL)
}

// --- Progress tests ---------------------------------------------------------

func TestHandleGetVideoProgress(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	req := httptest.NewRequest(http.MethodGet, "/api/videos/test-vid/progress?category=devops", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result OverallProgressResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, 7, len(result.Aspects))
	assert.GreaterOrEqual(t, result.Overall.Total, 0)
}

func TestHandleGetVideoAspectProgress(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	req := httptest.NewRequest(http.MethodGet, "/api/videos/test-vid/progress/initial-details?category=devops", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result AspectProgressInfo
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "initial-details", result.AspectKey)
	assert.Equal(t, "Initial Details", result.Title)
	assert.GreaterOrEqual(t, result.Total, 0)
}

func TestHandleGetVideoAspectProgress_InvalidAspect(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	req := httptest.NewRequest(http.MethodGet, "/api/videos/test-vid/progress/bogus?category=devops", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleGetVideoProgress_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/videos/no-such/progress?category=devops", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- Manuscript tests -------------------------------------------------------

func TestHandleGetVideoManuscript(t *testing.T) {
	env := setupTestEnv(t)

	mdPath := filepath.Join(env.tmpDir, "manuscript", "devops", "test-vid.md")

	seedVideo(t, env, storage.Video{
		Name:     "test-vid",
		Category: "devops",
		Gist:     mdPath,
	})

	// Write manuscript AFTER seedVideo so it doesn't get overwritten
	require.NoError(t, os.WriteFile(mdPath, []byte("# Hello World\nSome content"), 0644))

	req := httptest.NewRequest(http.MethodGet, "/api/videos/test-vid/manuscript?category=devops", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result ManuscriptResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Contains(t, result.Content, "Hello World")
}

func TestHandleGetVideoManuscript_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/videos/no-such/manuscript?category=devops", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- Animations tests -------------------------------------------------------

func TestHandleGetVideoAnimations(t *testing.T) {
	env := setupTestEnv(t)

	mdPath := filepath.Join(env.tmpDir, "manuscript", "devops", "test-vid.md")

	seedVideo(t, env, storage.Video{
		Name:     "test-vid",
		Category: "devops",
		Gist:     mdPath,
	})

	// Write manuscript AFTER seedVideo so it doesn't get overwritten
	mdContent := "## Intro\nSome intro\nTODO: Add logo animation\n## Section One\nContent here\nTODO: Show diagram\n## Destroy\n"
	require.NoError(t, os.WriteFile(mdPath, []byte(mdContent), 0644))

	req := httptest.NewRequest(http.MethodGet, "/api/videos/test-vid/animations?category=devops", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result AnimationsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Contains(t, result.Animations, "Add logo animation")
	assert.Contains(t, result.Animations, "Show diagram")
	assert.Contains(t, result.Animations, "Section: Section One")
	assert.Contains(t, result.Sections, "Section: Section One")
}

func TestHandleGetVideoAnimations_NoGist(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{Name: "test-vid", Category: "devops"})

	req := httptest.NewRequest(http.MethodGet, "/api/videos/test-vid/animations?category=devops", nil)
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
