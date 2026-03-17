package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"
)

// mockEmailService implements EmailService for testing.
type mockEmailService struct {
	sendThumbnailCalled   bool
	sendEditCalled        bool
	sendEditVideo         storage.Video
	sendSponsorsCalled    bool
	sendSponsorsFrom      string
	sendSponsorsTo        string
	sendSponsorsVideoID   string
	sendSponsorsPrice     string
	sendSponsorsTitle     string
	returnErr             error
	sendSponsorsErr       error
}

func (m *mockEmailService) SendThumbnail(from, to string, video storage.Video) error {
	m.sendThumbnailCalled = true
	return m.returnErr
}

func (m *mockEmailService) SendEdit(from, to string, video storage.Video) error {
	m.sendEditCalled = true
	m.sendEditVideo = video
	return m.returnErr
}

func (m *mockEmailService) SendSponsors(from, to string, videoID, sponsorshipPrice, videoTitle string) error {
	m.sendSponsorsCalled = true
	m.sendSponsorsFrom = from
	m.sendSponsorsTo = to
	m.sendSponsorsVideoID = videoID
	m.sendSponsorsPrice = sponsorshipPrice
	m.sendSponsorsTitle = videoTitle
	if m.sendSponsorsErr != nil {
		return m.sendSponsorsErr
	}
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
	if resp.EmailError == "" {
		t.Error("expected emailError to explain why email was not sent")
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

func TestHandleRequestEdit_DoesNotCorruptGistPath(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockEmailService{}
	env.server.SetEmailService(mock, &configuration.SettingsEmail{
		From:   "from@test.com",
		EditTo: "edit@test.com",
	})

	originalGist := "manuscript/devops/test.md"
	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Gist:     originalGist,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-edit/test-video?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Read back the persisted video and verify Gist was not modified
	saved, err := env.server.videoService.GetVideo("test-video", "devops")
	if err != nil {
		t.Fatalf("failed to read back video: %v", err)
	}
	if saved.Gist != originalGist {
		t.Errorf("Gist was corrupted: got %q, want %q", saved.Gist, originalGist)
	}
}

func TestHandleRequestEdit_NoEmailConfigured(t *testing.T) {
	env := setupTestEnv(t)
	// No email service configured

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/request-edit/test-video?category=devops", nil)
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
	if resp.EmailError == "" {
		t.Error("expected emailError to explain why email was not sent")
	}
	if !resp.Video.RequestEdit {
		t.Error("expected video.requestEdit to be true even without email")
	}
}

func TestHandleRequestEdit_AdFile(t *testing.T) {
	tests := []struct {
		name            string
		adFile          string
		createAdFile    bool
		adFileContent   string
		expectEmailSent bool
		checkAdContent  func(t *testing.T, adContent string)
	}{
		{
			name:            "with valid ad file",
			adFile:          "kilo.md",
			createAdFile:    true,
			adFileContent:   "## Kilo Ad\nPlease read this script at the 5 minute mark.",
			expectEmailSent: true,
			checkAdContent: func(t *testing.T, adContent string) {
				if adContent != "## Kilo Ad\nPlease read this script at the 5 minute mark." {
					t.Errorf("expected ad file content, got %q", adContent)
				}
			},
		},
		{
			name:            "with missing ad file",
			adFile:          "nonexistent.md",
			createAdFile:    false,
			expectEmailSent: true,
			checkAdContent: func(t *testing.T, adContent string) {
				if !strings.Contains(adContent, "[Warning:") {
					t.Errorf("expected AdContent to contain warning, got %q", adContent)
				}
			},
		},
		{
			name:            "no ad file for non-sponsored video",
			adFile:          "",
			createAdFile:    false,
			expectEmailSent: true,
			checkAdContent: func(t *testing.T, adContent string) {
				if adContent != "" {
					t.Errorf("expected empty AdContent for non-sponsored video, got %q", adContent)
				}
			},
		},
		{
			name:            "path traversal attempt is sanitized",
			adFile:          "../../../etc/passwd",
			createAdFile:    false,
			expectEmailSent: true,
			checkAdContent: func(t *testing.T, adContent string) {
				// filepath.Base strips directory traversal, so it looks for "passwd" in manuscript/ads/
				if !strings.Contains(adContent, "[Warning:") {
					t.Errorf("expected warning for traversal attempt, got %q", adContent)
				}
				// Ensure the resolved path stayed within manuscript/ads/ (not /etc/passwd)
				if strings.Contains(adContent, "etc/passwd") && !strings.Contains(adContent, "manuscript/ads/passwd") {
					t.Errorf("path traversal should be confined to manuscript/ads/, got %q", adContent)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			mock := &mockEmailService{}
			env.server.SetEmailService(mock, &configuration.SettingsEmail{
				From:   "from@test.com",
				EditTo: "edit@test.com",
			})

			if tt.createAdFile {
				adsDir := filepath.Join(env.tmpDir, "manuscript", "ads")
				if err := os.MkdirAll(adsDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(adsDir, tt.adFile), []byte(tt.adFileContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			video := storage.Video{
				Name:     "test-video",
				Category: "devops",
				Gist:     "test.md",
			}
			if tt.adFile != "" {
				video.Sponsorship = storage.Sponsorship{AdFile: tt.adFile}
			}
			seedVideo(t, env, video)

			req := httptest.NewRequest(http.MethodPost, "/api/actions/request-edit/test-video?category=devops", nil)
			w := httptest.NewRecorder()
			env.server.Router().ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}

			if !mock.sendEditCalled {
				t.Fatal("expected SendEdit to be called")
			}
			tt.checkAdContent(t, mock.sendEditVideo.AdContent)
		})
	}
}

func TestEmailNotConfiguredMessage(t *testing.T) {
	tests := []struct {
		name           string
		svc            EmailService
		settings       *configuration.SettingsEmail
		recipientField string
		wantContains   string
	}{
		{
			name:           "nil service",
			svc:            nil,
			settings:       nil,
			recipientField: "ThumbnailTo",
			wantContains:   "EMAIL_PASSWORD",
		},
		{
			name:           "nil settings",
			svc:            &mockEmailService{},
			settings:       nil,
			recipientField: "ThumbnailTo",
			wantContains:   "email settings are missing",
		},
		{
			name:           "empty from",
			svc:            &mockEmailService{},
			settings:       &configuration.SettingsEmail{},
			recipientField: "ThumbnailTo",
			wantContains:   "EMAIL_FROM",
		},
		{
			name:           "missing recipient",
			svc:            &mockEmailService{},
			settings:       &configuration.SettingsEmail{From: "a@b.com"},
			recipientField: "EditTo",
			wantContains:   "EditTo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := emailNotConfiguredMessage(tt.svc, tt.settings, tt.recipientField)
			if msg == "" {
				t.Fatal("expected non-empty message")
			}
			if !contains(msg, tt.wantContains) {
				t.Errorf("message %q does not contain %q", msg, tt.wantContains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

var errTestEmail = fmt.Errorf("smtp connection failed")

func TestHandleNotifySponsors(t *testing.T) {
	tests := []struct {
		name               string
		url                string
		seedVideo          *storage.Video
		emailService       *mockEmailService
		emailSettings      *configuration.SettingsEmail
		wantStatus         int
		wantAlreadyReq     bool
		wantEmailSent      bool
		wantEmailError     bool
		wantNotified       bool
		wantSponsorsCalled bool
	}{
		{
			name:       "missing category",
			url:        "/api/actions/notify-sponsors/test-video",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "video not found",
			url:        "/api/actions/notify-sponsors/nonexistent?category=devops",
			wantStatus: http.StatusNotFound,
		},
		{
			name: "already notified",
			url:  "/api/actions/notify-sponsors/test-video?category=devops",
			seedVideo: &storage.Video{
				Name: "test-video", Category: "devops", NotifiedSponsors: true,
				Sponsorship: storage.Sponsorship{Amount: "1000", Emails: "sponsor@test.com"},
			},
			wantStatus:     http.StatusOK,
			wantAlreadyReq: true,
			wantNotified:   true,
		},
		{
			name: "empty sponsorship amount",
			url:  "/api/actions/notify-sponsors/test-video?category=devops",
			seedVideo: &storage.Video{
				Name: "test-video", Category: "devops",
				Sponsorship: storage.Sponsorship{Amount: "", Emails: "sponsor@test.com"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "N/A sponsorship amount",
			url:  "/api/actions/notify-sponsors/test-video?category=devops",
			seedVideo: &storage.Video{
				Name: "test-video", Category: "devops",
				Sponsorship: storage.Sponsorship{Amount: "N/A", Emails: "sponsor@test.com"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "dash sponsorship amount",
			url:  "/api/actions/notify-sponsors/test-video?category=devops",
			seedVideo: &storage.Video{
				Name: "test-video", Category: "devops",
				Sponsorship: storage.Sponsorship{Amount: "-", Emails: "sponsor@test.com"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "no sponsor emails",
			url:  "/api/actions/notify-sponsors/test-video?category=devops",
			seedVideo: &storage.Video{
				Name: "test-video", Category: "devops",
				Sponsorship: storage.Sponsorship{Amount: "1000", Emails: ""},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "success",
			url:  "/api/actions/notify-sponsors/test-video?category=devops",
			seedVideo: &storage.Video{
				Name: "test-video", Category: "devops", VideoId: "abc123",
				Sponsorship: storage.Sponsorship{Amount: "1000", Emails: "sponsor@test.com"},
			},
			emailService:       &mockEmailService{},
			emailSettings:      &configuration.SettingsEmail{From: "from@test.com"},
			wantStatus:         http.StatusOK,
			wantEmailSent:      true,
			wantNotified:       true,
			wantSponsorsCalled: true,
		},
		{
			name: "no email configured",
			url:  "/api/actions/notify-sponsors/test-video?category=devops",
			seedVideo: &storage.Video{
				Name: "test-video", Category: "devops",
				Sponsorship: storage.Sponsorship{Amount: "1000", Emails: "sponsor@test.com"},
			},
			wantStatus:     http.StatusOK,
			wantEmailError: true,
		},
		{
			name: "email failure",
			url:  "/api/actions/notify-sponsors/test-video?category=devops",
			seedVideo: &storage.Video{
				Name: "test-video", Category: "devops", VideoId: "abc123",
				Sponsorship: storage.Sponsorship{Amount: "1000", Emails: "sponsor@test.com"},
			},
			emailService:       &mockEmailService{sendSponsorsErr: errTestEmail},
			emailSettings:      &configuration.SettingsEmail{From: "from@test.com"},
			wantStatus:         http.StatusOK,
			wantEmailError:     true,
			wantSponsorsCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)

			if tt.emailService != nil {
				env.server.SetEmailService(tt.emailService, tt.emailSettings)
			}
			if tt.seedVideo != nil {
				seedVideo(t, env, *tt.seedVideo)
			}

			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()
			env.server.Router().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var resp ActionResponse
			json.NewDecoder(w.Body).Decode(&resp)

			if resp.AlreadyRequested != tt.wantAlreadyReq {
				t.Errorf("alreadyRequested = %v, want %v", resp.AlreadyRequested, tt.wantAlreadyReq)
			}
			if resp.EmailSent != tt.wantEmailSent {
				t.Errorf("emailSent = %v, want %v", resp.EmailSent, tt.wantEmailSent)
			}
			if tt.wantEmailError && resp.EmailError == "" {
				t.Error("expected emailError to be set")
			}
			if !tt.wantEmailError && resp.EmailError != "" {
				t.Errorf("unexpected emailError: %s", resp.EmailError)
			}
			if resp.Video.NotifiedSponsors != tt.wantNotified {
				t.Errorf("notifiedSponsors = %v, want %v", resp.Video.NotifiedSponsors, tt.wantNotified)
			}
			if tt.emailService != nil && tt.emailService.sendSponsorsCalled != tt.wantSponsorsCalled {
				t.Errorf("sendSponsorsCalled = %v, want %v", tt.emailService.sendSponsorsCalled, tt.wantSponsorsCalled)
			}
			if tt.wantSponsorsCalled && tt.seedVideo != nil && tt.emailSettings != nil {
				if tt.emailService.sendSponsorsFrom != tt.emailSettings.From {
					t.Errorf("sendSponsors from = %q, want %q", tt.emailService.sendSponsorsFrom, tt.emailSettings.From)
				}
				if tt.emailService.sendSponsorsTo != tt.seedVideo.Sponsorship.Emails {
					t.Errorf("sendSponsors to = %q, want %q", tt.emailService.sendSponsorsTo, tt.seedVideo.Sponsorship.Emails)
				}
				if tt.emailService.sendSponsorsVideoID != tt.seedVideo.VideoId {
					t.Errorf("sendSponsors videoID = %q, want %q", tt.emailService.sendSponsorsVideoID, tt.seedVideo.VideoId)
				}
				if tt.emailService.sendSponsorsPrice != tt.seedVideo.Sponsorship.Amount {
					t.Errorf("sendSponsors price = %q, want %q", tt.emailService.sendSponsorsPrice, tt.seedVideo.Sponsorship.Amount)
				}
				if tt.emailService.sendSponsorsTitle != tt.seedVideo.GetUploadTitle() {
					t.Errorf("sendSponsors title = %q, want %q", tt.emailService.sendSponsorsTitle, tt.seedVideo.GetUploadTitle())
				}
			}
		})
	}
}
