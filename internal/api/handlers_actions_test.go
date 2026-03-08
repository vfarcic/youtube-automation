package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"
)

// mockEmailService implements EmailService for testing.
type mockEmailService struct {
	sendThumbnailCalled bool
	sendEditCalled      bool
	returnErr           error
}

func (m *mockEmailService) SendThumbnail(from, to string, video storage.Video) error {
	m.sendThumbnailCalled = true
	return m.returnErr
}

func (m *mockEmailService) SendEdit(from, to string, video storage.Video) error {
	m.sendEditCalled = true
	return m.returnErr
}

func TestHandleRequestThumbnail_MissingCategory(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-thumbnail/test-video", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleRequestThumbnail_VideoNotFound(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-thumbnail/nonexistent?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleRequestThumbnail_AlreadyRequested(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{
		Name:             "test-video",
		Category:         "devops",
		RequestThumbnail: true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-thumbnail/test-video?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ActionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.AlreadyRequested {
		t.Error("expected alreadyRequested to be true")
	}
}

func TestHandleRequestThumbnail_Success(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockEmailService{}
	env.server.SetEmailService(mock, &configuration.SettingsEmail{
		From:        "from@test.com",
		ThumbnailTo: "thumb@test.com",
	})

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-thumbnail/test-video?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ActionResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.AlreadyRequested {
		t.Error("expected alreadyRequested to be false")
	}
	if !resp.EmailSent {
		t.Error("expected emailSent to be true")
	}
	if !mock.sendThumbnailCalled {
		t.Error("expected SendThumbnail to be called")
	}
	if !resp.Video.RequestThumbnail {
		t.Error("expected video.requestThumbnail to be true")
	}
}

func TestHandleRequestThumbnail_NoEmailConfigured(t *testing.T) {
	env := setupTestEnv(t)
	// No email service configured

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-thumbnail/test-video?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ActionResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.EmailSent {
		t.Error("expected emailSent to be false when email not configured")
	}
	if resp.Video.RequestThumbnail != true {
		t.Error("expected video.requestThumbnail to be true even without email")
	}
}

func TestHandleRequestThumbnail_EmailFailure(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockEmailService{returnErr: errTestEmail}
	env.server.SetEmailService(mock, &configuration.SettingsEmail{
		From:        "from@test.com",
		ThumbnailTo: "thumb@test.com",
	})

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-thumbnail/test-video?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ActionResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.EmailSent {
		t.Error("expected emailSent to be false on email failure")
	}
	if resp.EmailError == "" {
		t.Error("expected emailError to be set")
	}
	if resp.Video.RequestThumbnail != true {
		t.Error("expected video.requestThumbnail to be true even when email fails")
	}
}

func TestHandleRequestEdit_MissingCategory(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-edit/test-video", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleRequestEdit_VideoNotFound(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-edit/nonexistent?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleRequestEdit_AlreadyRequested(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{
		Name:        "test-video",
		Category:    "devops",
		RequestEdit: true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-edit/test-video?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ActionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.AlreadyRequested {
		t.Error("expected alreadyRequested to be true")
	}
}

func TestHandleRequestEdit_Success(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockEmailService{}
	env.server.SetEmailService(mock, &configuration.SettingsEmail{
		From:   "from@test.com",
		EditTo: "edit@test.com",
	})

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Gist:     "test.md",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-edit/test-video?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ActionResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.AlreadyRequested {
		t.Error("expected alreadyRequested to be false")
	}
	if !resp.EmailSent {
		t.Error("expected emailSent to be true")
	}
	if !mock.sendEditCalled {
		t.Error("expected SendEdit to be called")
	}
	if !resp.Video.RequestEdit {
		t.Error("expected video.requestEdit to be true")
	}
}

var errTestEmail = fmt.Errorf("smtp connection failed")
