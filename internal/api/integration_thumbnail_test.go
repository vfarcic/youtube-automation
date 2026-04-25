package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/thumbnail"
)

// trackingDriveService records upload calls for verification.
type trackingDriveService struct {
	mockDriveService
	uploadCalls []driveUploadCall
}

type driveUploadCall struct {
	Filename string
	MimeType string
	FolderID string
	Data     []byte
}

func (t *trackingDriveService) UploadFile(_ context.Context, filename string, content io.Reader, mimeType string, folderID string) (string, error) {
	data, _ := io.ReadAll(content)
	t.uploadCalls = append(t.uploadCalls, driveUploadCall{
		Filename: filename,
		MimeType: mimeType,
		FolderID: folderID,
		Data:     data,
	})
	return t.returnFileID, t.returnErr
}

// TestIntegration_ThumbnailGeneration_FullFlow exercises the complete end-to-end
// thumbnail generation flow through the API layer with mocked providers:
//
//  1. POST /api/ai/tagline-and-illustrations/{category}/{name} → get tagline & illustration suggestions
//  1b. POST /api/videos/{videoName}/thumbnail-config → save selection
//  2. POST /api/thumbnails/generate → generate thumbnails with mock providers
//  3. GET /api/thumbnails/generated/{id} → download generated image, verify bytes
//  4. POST /api/thumbnails/generated/{id}/select → select thumbnail, verify Drive
//     upload, verify image removed from store, verify ThumbnailVariant saved
func TestIntegration_ThumbnailGeneration_FullFlow(t *testing.T) {
	// --- Setup ---
	env := setupTestEnv(t)

	// Mock AI service for tagline & illustration suggestions
	aiMock := &mockAIService{
		taglineAndIllustrations: &ai.TaglineAndIllustrationsResult{
			Taglines:      []string{"Secure Everything", "Lock It Down", "Zero Trust"},
			Illustrations: []string{"Fortress protecting servers", "Shield around clusters", "Cracked lock being fixed"},
		},
	}
	env.server.aiService = aiMock

	// Mock image generators (two providers, like Gemini + GPT Image)
	// Use PNG magic bytes so DetectContentType recognizes them
	geminiImageData := []byte("\x89PNG\r\n\x1a\n-gemini-thumbnail-data-")
	gptImageData := []byte("\x89PNG\r\n\x1a\n-gpt-image-thumbnail-data-")
	geminiGen := &mockImageGenerator{name: "gemini", data: geminiImageData}
	gptGen := &mockImageGenerator{name: "gpt-image", data: gptImageData}

	// Image store for generated thumbnails
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)

	// Create a photo directory with a test creator photo
	photoDir := filepath.Join(env.tmpDir, "photos")
	if err := os.MkdirAll(photoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(photoDir, "creator.png"), []byte("creator-photo"), 0644); err != nil {
		t.Fatal(err)
	}

	env.server.SetThumbnailGeneration(
		[]thumbnail.ImageGenerator{geminiGen, gptGen},
		store,
		photoDir,
	)

	// Mock Drive service with screenshots for photo loading
	driveMock := &trackingDriveService{
		mockDriveService: mockDriveService{
			returnFileID: "drive-file-abc123",
			listFiles: []gdrive.DriveFileInfo{
				{ID: "screenshot-01", Name: "screenshot-01.png", MimeType: "image/png"},
			},
			fileContents: map[string][]byte{
				"screenshot-01": []byte("fake-creator-photo"),
			},
		},
	}
	env.server.SetDriveService(driveMock, "root-folder-id")

	// Seed a video with manuscript
	seedVideoWithManuscript(t, env, "my-video", "devops",
		"# My Video\n\nThis is a manuscript about Kubernetes security best practices.")

	router := env.server.Router()

	// ====================================================
	// Step 1: Suggest tagline & illustrations from manuscript
	// ====================================================
	t.Run("Step1_SuggestTaglineAndIllustrations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/ai/tagline-and-illustrations/devops/my-video", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("POST /api/ai/tagline-and-illustrations: expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp AITaglineAndIllustrationsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Taglines) != 3 {
			t.Errorf("expected 3 tagline suggestions, got %d", len(resp.Taglines))
		}
		if len(resp.Illustrations) != 3 {
			t.Errorf("expected 3 illustration suggestions, got %d", len(resp.Illustrations))
		}
	})

	// ====================================================
	// Step 1b: Save tagline & illustration selection
	// ====================================================
	t.Run("Step1b_SaveThumbnailConfig", func(t *testing.T) {
		body := `{"category": "devops", "tagline": "Secure Everything", "illustration": "Fortress protecting servers"}`
		req := httptest.NewRequest(http.MethodPost, "/api/videos/my-video/thumbnail-config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("POST /api/videos/my-video/thumbnail-config: expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// ====================================================
	// Step 2: Generate thumbnails (reads stored tagline + illustration)
	// ====================================================
	var generatedIDs []ThumbnailGenerateMeta

	t.Run("Step2_GenerateThumbnails", func(t *testing.T) {
		body := `{
			"category": "devops",
			"name": "my-video"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("POST /api/thumbnails/generate: expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp ThumbnailGenerateResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode generate response: %v", err)
		}

		// 2 providers x 2 styles (with/without illustration) = 4 thumbnails
		if len(resp.Thumbnails) != 4 {
			t.Fatalf("expected 4 thumbnails (2 providers x 2 styles), got %d", len(resp.Thumbnails))
		}

		// Verify each thumbnail has an ID, provider, and style
		providers := map[string]int{}
		styles := map[string]int{}
		for _, m := range resp.Thumbnails {
			if m.ID == "" {
				t.Error("expected non-empty ID for generated thumbnail")
			}
			providers[m.Provider]++
			styles[m.Style]++
		}

		if providers["gemini"] != 2 {
			t.Errorf("expected 2 gemini thumbnails, got %d", providers["gemini"])
		}
		if providers["gpt-image"] != 2 {
			t.Errorf("expected 2 gpt-image thumbnails, got %d", providers["gpt-image"])
		}
		if styles["with illustration"] != 2 {
			t.Errorf("expected 2 'with illustration' thumbnails, got %d", styles["with illustration"])
		}
		if styles["without illustration"] != 2 {
			t.Errorf("expected 2 'without illustration' thumbnails, got %d", styles["without illustration"])
		}

		// No errors expected since both mock providers succeed
		if len(resp.Errors) != 0 {
			t.Errorf("expected no errors, got %v", resp.Errors)
		}

		// Verify all 4 images are in the store
		if store.Len() != 4 {
			t.Errorf("expected 4 images in store, got %d", store.Len())
		}

		generatedIDs = resp.Thumbnails
	})

	if len(generatedIDs) == 0 {
		t.Fatal("no thumbnails generated; cannot continue integration test")
	}

	// ====================================================
	// Step 3: Download a generated thumbnail and verify bytes
	// ====================================================
	// Find a gemini thumbnail to download
	var geminiThumb ThumbnailGenerateMeta
	for _, m := range generatedIDs {
		if m.Provider == "gemini" {
			geminiThumb = m
			break
		}
	}

	t.Run("Step3_DownloadGeneratedThumbnail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/thumbnails/generated/"+geminiThumb.ID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("GET /api/thumbnails/generated/{id}: expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify Content-Type is image/png (due to PNG magic bytes)
		ct := w.Header().Get("Content-Type")
		if ct != "image/png" {
			t.Errorf("expected Content-Type image/png, got %q", ct)
		}

		// Verify the image bytes match what the mock provider returned
		if !bytes.Equal(w.Body.Bytes(), geminiImageData) {
			t.Errorf("downloaded image bytes don't match expected data")
		}
	})

	// ====================================================
	// Step 4: Select thumbnail → upload to Drive → verify state
	// ====================================================
	t.Run("Step4_SelectThumbnail", func(t *testing.T) {
		body := `{
			"category": "devops",
			"name": "my-video",
			"variantIndex": 0
		}`
		req := httptest.NewRequest(http.MethodPost,
			"/api/thumbnails/generated/"+geminiThumb.ID+"/select",
			strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("POST /api/thumbnails/generated/{id}/select: expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode select response: %v", err)
		}

		// Verify Drive file ID returned
		if resp["driveFileId"] != "drive-file-abc123" {
			t.Errorf("expected driveFileId 'drive-file-abc123', got %v", resp["driveFileId"])
		}

		// Verify variant index in response
		if vi, ok := resp["variantIndex"].(float64); !ok || int(vi) != 0 {
			t.Errorf("expected variantIndex 0, got %v", resp["variantIndex"])
		}

		// Verify Drive upload was called
		if len(driveMock.uploadCalls) != 1 {
			t.Fatalf("expected 1 Drive upload call, got %d", len(driveMock.uploadCalls))
		}
		upload := driveMock.uploadCalls[0]

		// Verify uploaded image data matches the mock provider output
		if !bytes.Equal(upload.Data, geminiImageData) {
			t.Error("uploaded image data doesn't match expected provider output")
		}

		// Verify filename follows the expected pattern
		if upload.Filename != "thumbnail-0-generated.png" {
			t.Errorf("expected filename 'thumbnail-0-generated.png', got %q", upload.Filename)
		}

		// Verify folder was a subfolder of root
		if upload.FolderID != "root-folder-id-subfolder" {
			t.Errorf("expected folder 'root-folder-id-subfolder', got %q", upload.FolderID)
		}

		// Verify image was removed from store after selection
		if _, found := store.Get(geminiThumb.ID); found {
			t.Error("expected selected image to be removed from store")
		}

		// Verify remaining images are still in store (3 left out of 4)
		if store.Len() != 3 {
			t.Errorf("expected 3 images remaining in store, got %d", store.Len())
		}

		// Verify ThumbnailVariant was saved on the video
		updatedVideo, err := env.server.videoService.GetVideo("my-video", "devops")
		if err != nil {
			t.Fatalf("failed to get video after select: %v", err)
		}

		if len(updatedVideo.ThumbnailVariants) != 1 {
			t.Fatalf("expected 1 thumbnail variant, got %d", len(updatedVideo.ThumbnailVariants))
		}

		variant := updatedVideo.ThumbnailVariants[0]
		if variant.DriveFileID != "drive-file-abc123" {
			t.Errorf("expected DriveFileID 'drive-file-abc123', got %q", variant.DriveFileID)
		}
		if variant.Index != 1 {
			t.Errorf("expected variant Index 1, got %d", variant.Index)
		}
	})
}

// TestIntegration_ThumbnailGeneration_PartialProviderFailure verifies that
// when one provider fails, the other still produces results.
func TestIntegration_ThumbnailGeneration_PartialProviderFailure(t *testing.T) {
	env := setupTestEnv(t)

	goodImageData := []byte("\x89PNG\r\n\x1a\n-good-data-")
	goodGen := &mockImageGenerator{name: "gemini", data: goodImageData}
	badGen := &mockImageGenerator{name: "gpt-image", err: io.ErrUnexpectedEOF}

	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{goodGen, badGen}, store, "")

	seedVideo(t, env, storage.Video{Name: "test", Category: "devops", Tagline: "Hello World"})
	env.server.SetDriveService(mockDriveWithScreenshots(), "root-folder")

	body := `{"category":"devops","name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for partial failure, got %d: %s", w.Code, w.Body.String())
	}

	var resp ThumbnailGenerateResponse
	json.NewDecoder(w.Body).Decode(&resp)

	// Good provider produces 2 thumbnails
	if len(resp.Thumbnails) != 2 {
		t.Errorf("expected 2 thumbnails from successful provider, got %d", len(resp.Thumbnails))
	}
	for _, m := range resp.Thumbnails {
		if m.Provider != "gemini" {
			t.Errorf("expected provider 'gemini', got %q", m.Provider)
		}
	}

	// Should report errors from the failing provider
	if len(resp.Errors) == 0 {
		t.Error("expected errors from failing provider")
	}

	// The good thumbnails should be downloadable
	if len(resp.Thumbnails) > 0 {
		dlReq := httptest.NewRequest(http.MethodGet, "/api/thumbnails/generated/"+resp.Thumbnails[0].ID, nil)
		dlW := httptest.NewRecorder()
		env.server.Router().ServeHTTP(dlW, dlReq)

		if dlW.Code != http.StatusOK {
			t.Errorf("expected 200 downloading good thumbnail, got %d", dlW.Code)
		}
		if !bytes.Equal(dlW.Body.Bytes(), goodImageData) {
			t.Error("downloaded image bytes don't match expected")
		}
	}
}

// TestIntegration_ThumbnailGeneration_MultipleSelections verifies that
// a user can select multiple thumbnails (for different variant indices).
func TestIntegration_ThumbnailGeneration_MultipleSelections(t *testing.T) {
	env := setupTestEnv(t)

	imageData := []byte("\x89PNG\r\n\x1a\n-test-image-")
	gen := &mockImageGenerator{name: "test-provider", data: imageData}
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")

	driveMock := &trackingDriveService{
		mockDriveService: mockDriveService{
			returnFileID: "drive-multi",
			listFiles: []gdrive.DriveFileInfo{
				{ID: "screenshot-01", Name: "screenshot-01.png", MimeType: "image/png"},
			},
			fileContents: map[string][]byte{
				"screenshot-01": []byte("fake-photo"),
			},
		},
	}
	env.server.SetDriveService(driveMock, "root-folder")

	seedVideo(t, env, storage.Video{
		Name:     "multi-video",
		Category: "devops",
		Tagline:  "Test",
	})

	// Generate thumbnails
	body := `{"category":"devops","name":"multi-video"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("generate: expected 200, got %d", w.Code)
	}

	var genResp ThumbnailGenerateResponse
	json.NewDecoder(w.Body).Decode(&genResp)

	if len(genResp.Thumbnails) < 2 {
		t.Fatalf("need at least 2 thumbnails, got %d", len(genResp.Thumbnails))
	}

	// Select first thumbnail as variant 0
	selBody1 := `{"category":"devops","name":"multi-video","variantIndex":0}`
	selReq1 := httptest.NewRequest(http.MethodPost,
		"/api/thumbnails/generated/"+genResp.Thumbnails[0].ID+"/select",
		strings.NewReader(selBody1))
	selReq1.Header.Set("Content-Type", "application/json")
	selW1 := httptest.NewRecorder()
	env.server.Router().ServeHTTP(selW1, selReq1)

	if selW1.Code != http.StatusOK {
		t.Fatalf("select variant 0: expected 200, got %d: %s", selW1.Code, selW1.Body.String())
	}

	// Select second thumbnail as variant 1 (auto-create)
	selBody2 := `{"category":"devops","name":"multi-video","variantIndex":1}`
	selReq2 := httptest.NewRequest(http.MethodPost,
		"/api/thumbnails/generated/"+genResp.Thumbnails[1].ID+"/select",
		strings.NewReader(selBody2))
	selReq2.Header.Set("Content-Type", "application/json")
	selW2 := httptest.NewRecorder()
	env.server.Router().ServeHTTP(selW2, selReq2)

	if selW2.Code != http.StatusOK {
		t.Fatalf("select variant 1: expected 200, got %d: %s", selW2.Code, selW2.Body.String())
	}

	// Verify video has 2 variants
	v, err := env.server.videoService.GetVideo("multi-video", "devops")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.ThumbnailVariants) != 2 {
		t.Errorf("expected 2 variants, got %d", len(v.ThumbnailVariants))
	}
	for i, variant := range v.ThumbnailVariants {
		if variant.DriveFileID != "drive-multi" {
			t.Errorf("variant %d: expected DriveFileID 'drive-multi', got %q", i, variant.DriveFileID)
		}
	}

	// Verify both images were removed from store
	if _, found := store.Get(genResp.Thumbnails[0].ID); found {
		t.Error("first selected image should be removed from store")
	}
	if _, found := store.Get(genResp.Thumbnails[1].ID); found {
		t.Error("second selected image should be removed from store")
	}

	// Verify Drive was called twice
	if len(driveMock.uploadCalls) != 2 {
		t.Errorf("expected 2 Drive upload calls, got %d", len(driveMock.uploadCalls))
	}
}

// TestIntegration_ThumbnailGeneration_ConcurrentProviders verifies that
// multiple providers execute concurrently by checking all results arrive.
func TestIntegration_ThumbnailGeneration_ConcurrentProviders(t *testing.T) {
	env := setupTestEnv(t)

	// Create 3 providers to verify concurrent execution
	providers := []thumbnail.ImageGenerator{
		&mockImageGenerator{name: "provider-a", data: []byte("\x89PNG\r\n\x1a\n-A-")},
		&mockImageGenerator{name: "provider-b", data: []byte("\x89PNG\r\n\x1a\n-B-")},
		&mockImageGenerator{name: "provider-c", data: []byte("\x89PNG\r\n\x1a\n-C-")},
	}

	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	env.server.SetThumbnailGeneration(providers, store, "")

	seedVideo(t, env, storage.Video{Name: "test", Category: "devops", Tagline: "Concurrent Test"})
	env.server.SetDriveService(mockDriveWithScreenshots(), "root-folder")

	body := `{"category":"devops","name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ThumbnailGenerateResponse
	json.NewDecoder(w.Body).Decode(&resp)

	// 3 providers x 2 styles = 6 thumbnails
	if len(resp.Thumbnails) != 6 {
		t.Errorf("expected 6 thumbnails (3 providers x 2 styles), got %d", len(resp.Thumbnails))
	}

	// Verify all 3 providers represented
	providerCounts := map[string]int{}
	for _, m := range resp.Thumbnails {
		providerCounts[m.Provider]++
	}
	for _, name := range []string{"provider-a", "provider-b", "provider-c"} {
		if providerCounts[name] != 2 {
			t.Errorf("expected 2 thumbnails from %s, got %d", name, providerCounts[name])
		}
	}
}

// TestIntegration_ExistingManualUploadStillWorks verifies that the manual
// "Upload to Drive" flow (existing functionality) is not broken by thumbnail
// generation being configured on the same server.
func TestIntegration_ExistingManualUploadStillWorks(t *testing.T) {
	env := setupTestEnv(t)

	// Configure both thumbnail generation AND drive service
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "test", data: []byte("data")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")

	driveMock := &mockDriveService{returnFileID: "manual-drive-id"}
	env.server.SetDriveService(driveMock, "root-folder")

	seedVideo(t, env, storage.Video{
		Name:     "manual-test",
		Category: "devops",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Path: "/local/thumb.png"},
		},
	})

	// The manual upload endpoint POST /api/drive/upload/thumbnail/{videoName}
	// should still work. We test that the endpoint is reachable and returns
	// the expected error for missing multipart body (proving the handler is active).
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/manual-test", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	// The handler expects multipart form data; sending none should give 400, not 404/501.
	// This proves the route is still registered and functional.
	if w.Code == http.StatusNotFound || w.Code == http.StatusNotImplemented {
		t.Errorf("manual upload endpoint should still be active, got %d", w.Code)
	}
}
