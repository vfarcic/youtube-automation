package publishing

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// mockVideoListDoer returns a canned response or error from .Do().
type mockVideoListDoer struct {
	resp *youtube.VideoListResponse
	err  error
}

func (m *mockVideoListDoer) Do(opts ...googleapi.CallOption) (*youtube.VideoListResponse, error) {
	return m.resp, m.err
}

// mockVideosClient captures arguments and returns the canned doers.
// It reuses mockVideoUpdateDoer from youtube_test.go for Update calls.
type mockVideosClient struct {
	listDoer   *mockVideoListDoer
	updateDoer *mockVideoUpdateDoer

	listParts   []string
	listVideoID string
	listCalls   int
	listCtxs    []context.Context
	updateParts []string
	updateVideo *youtube.Video
	updateCalls int
	updateCtxs  []context.Context
}

func (m *mockVideosClient) List(ctx context.Context, part []string, videoID string) videoListDoer {
	m.listCalls++
	m.listParts = part
	m.listVideoID = videoID
	m.listCtxs = append(m.listCtxs, ctx)
	return m.listDoer
}

func (m *mockVideosClient) Update(ctx context.Context, part []string, video *youtube.Video) videoUpdateDoer {
	m.updateCalls++
	m.updateParts = part
	m.updateVideo = video
	m.updateCtxs = append(m.updateCtxs, ctx)
	if m.updateDoer == nil {
		return &mockVideoUpdateDoer{}
	}
	return m.updateDoer
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple comma-separated tags",
			input:    "kubernetes,docker,devops",
			expected: []string{"kubernetes", "docker", "devops"},
		},
		{
			name:     "tags with spaces",
			input:    " kubernetes , docker , devops ",
			expected: []string{"kubernetes", "docker", "devops"},
		},
		{
			name:     "single tag",
			input:    "kubernetes",
			expected: []string{"kubernetes"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "tags with empty entries",
			input:    "kubernetes,,docker,,devops",
			expected: []string{"kubernetes", "docker", "devops"},
		},
		{
			name:     "only commas",
			input:    ",,,",
			expected: []string{},
		},
		{
			name:     "tags with special characters",
			input:    "CI/CD,GitOps,cloud-native",
			expected: []string{"CI/CD", "GitOps", "cloud-native"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseTags(%q) returned %d tags, expected %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("parseTags(%q)[%d] = %q, expected %q", tt.input, i, tag, tt.expected[i])
				}
			}
		})
	}
}

func TestExtractBoilerplate(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "description with boilerplate",
			description: "This is my video description.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nConsider joining the channel: https://youtube.com/...\n\n▬▬▬▬▬▬ 🔗 Additional Info 🔗 ▬▬▬▬▬▬\nMore info here",
			expected:    "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nConsider joining the channel: https://youtube.com/...\n\n▬▬▬▬▬▬ 🔗 Additional Info 🔗 ▬▬▬▬▬▬\nMore info here",
		},
		{
			name:        "description without boilerplate",
			description: "This is just a simple description without any boilerplate.",
			expected:    "",
		},
		{
			name:        "empty description",
			description: "",
			expected:    "",
		},
		{
			name:        "description with boilerplate and timecodes",
			description: "My description.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate content\n\n▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬\n00:00 Intro\n02:00 Question 1",
			expected:    "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate content",
		},
		{
			name:        "boilerplate at start",
			description: "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nOnly boilerplate",
			expected:    "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nOnly boilerplate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBoilerplate(tt.description)
			if result != tt.expected {
				t.Errorf("extractBoilerplate() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestBuildAMADescription(t *testing.T) {
	tests := []struct {
		name               string
		newDescription     string
		currentDescription string
		timecodes          string
		shouldContain      []string
		shouldNotContain   []string
	}{
		{
			name:               "new description with boilerplate and timecodes",
			newDescription:     "This is my new AMA description about Kubernetes.",
			currentDescription: "Old description.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nConsider joining: https://youtube.com/...",
			timecodes:          "00:00 Intro\n02:30 Question about GitOps",
			shouldContain: []string{
				"This is my new AMA description about Kubernetes.",
				"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬",
				"Consider joining: https://youtube.com/...",
				"▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬",
				"00:00 Intro",
				"02:30 Question about GitOps",
			},
			shouldNotContain: []string{
				"Old description.",
			},
		},
		{
			name:               "only new description, no current boilerplate",
			newDescription:     "Brand new description.",
			currentDescription: "",
			timecodes:          "00:00 Start\n01:00 Topic 1",
			shouldContain: []string{
				"Brand new description.",
				"▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬",
				"00:00 Start",
			},
			shouldNotContain: []string{},
		},
		{
			name:               "empty new description preserves boilerplate",
			newDescription:     "",
			currentDescription: "Some text.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate here",
			timecodes:          "00:00 Intro",
			shouldContain: []string{
				"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬",
				"Boilerplate here",
				"▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬",
				"00:00 Intro",
			},
			shouldNotContain: []string{
				"Some text.",
			},
		},
		{
			name:               "no timecodes provided",
			newDescription:     "My description.",
			currentDescription: "Old.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate",
			timecodes:          "",
			shouldContain: []string{
				"My description.",
				"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬",
				"Boilerplate",
			},
			shouldNotContain: []string{
				"▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬",
			},
		},
		{
			name:               "replaces existing timecodes",
			newDescription:     "New content.",
			currentDescription: "Old.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate\n\n▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬\n00:00 Old intro",
			timecodes:          "00:00 New intro\n05:00 New question",
			shouldContain: []string{
				"New content.",
				"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬",
				"Boilerplate",
				"00:00 New intro",
				"05:00 New question",
			},
			shouldNotContain: []string{
				"Old.",
				"00:00 Old intro",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAMADescription(tt.newDescription, tt.currentDescription, tt.timecodes)

			for _, s := range tt.shouldContain {
				if !contains(result, s) {
					t.Errorf("buildAMADescription() should contain %q, but got:\n%s", s, result)
				}
			}

			for _, s := range tt.shouldNotContain {
				if contains(result, s) {
					t.Errorf("buildAMADescription() should NOT contain %q, but got:\n%s", s, result)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestBuildAMADescriptionOrder(t *testing.T) {
	newDesc := "New description here."
	currentDesc := "Old.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate content"
	timecodes := "00:00 Intro"

	result := buildAMADescription(newDesc, currentDesc, timecodes)

	// Verify order: new description comes before boilerplate, timecodes come last
	newDescIdx := indexOf(result, "New description here.")
	boilerplateIdx := indexOf(result, "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬")
	timecodesIdx := indexOf(result, "▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬")

	if newDescIdx == -1 || boilerplateIdx == -1 || timecodesIdx == -1 {
		t.Fatalf("Missing expected content in result:\n%s", result)
	}

	if newDescIdx >= boilerplateIdx {
		t.Errorf("New description should come before boilerplate. newDesc at %d, boilerplate at %d", newDescIdx, boilerplateIdx)
	}

	if boilerplateIdx >= timecodesIdx {
		t.Errorf("Boilerplate should come before timecodes. boilerplate at %d, timecodes at %d", boilerplateIdx, timecodesIdx)
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestGetVideoMetadataEmptyID(t *testing.T) {
	_, err := GetVideoMetadata(context.Background(), "")
	if err == nil {
		t.Error("GetVideoMetadata should return error for empty video ID")
	}
	expectedErr := "video ID cannot be empty"
	if err.Error() != expectedErr {
		t.Errorf("GetVideoMetadata error = %q, expected %q", err.Error(), expectedErr)
	}
}

func TestUpdateAMAVideoEmptyID(t *testing.T) {
	err := UpdateAMAVideo(context.Background(), "", "title", "desc", "tags", "timecodes")
	if err == nil {
		t.Error("UpdateAMAVideo should return error for empty video ID")
	}
	expectedErr := "video ID cannot be empty"
	if err.Error() != expectedErr {
		t.Errorf("UpdateAMAVideo error = %q, expected %q", err.Error(), expectedErr)
	}
}

func TestGetVideoMetadataInner(t *testing.T) {
	tests := []struct {
		name             string
		videoID          string
		listResp         *youtube.VideoListResponse
		listErr          error
		want             *VideoMetadata
		wantErr          bool
		wantErrSubstring string
	}{
		{
			name:    "happy path returns metadata",
			videoID: "vid42",
			listResp: &youtube.VideoListResponse{
				Items: []*youtube.Video{
					{
						Snippet: &youtube.VideoSnippet{
							Title:       "Live AMA",
							Description: "Some description",
							Tags:        []string{"k8s", "devops"},
							PublishedAt: "2026-01-01T10:00:00Z",
						},
					},
				},
			},
			want: &VideoMetadata{
				Title:       "Live AMA",
				Description: "Some description",
				Tags:        []string{"k8s", "devops"},
				PublishedAt: "2026-01-01T10:00:00Z",
			},
		},
		{
			name:             "list API error wraps message",
			videoID:          "vid42",
			listErr:          errors.New("quota exceeded"),
			wantErr:          true,
			wantErrSubstring: "failed to fetch video metadata",
		},
		{
			name:             "video not found returns error",
			videoID:          "missing",
			listResp:         &youtube.VideoListResponse{Items: nil},
			wantErr:          true,
			wantErrSubstring: "video not found: missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockVideosClient{
				listDoer: &mockVideoListDoer{resp: tt.listResp, err: tt.listErr},
			}
			got, err := getVideoMetadata(context.Background(), client, tt.videoID)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrSubstring != "" && !strings.Contains(err.Error(), tt.wantErrSubstring) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrSubstring)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
			if client.listCalls != 1 {
				t.Errorf("listCalls = %d, want 1", client.listCalls)
			}
			if !reflect.DeepEqual(client.listParts, []string{"snippet"}) {
				t.Errorf("listParts = %v, want [snippet]", client.listParts)
			}
			if client.listVideoID != tt.videoID {
				t.Errorf("listVideoID = %q, want %q", client.listVideoID, tt.videoID)
			}
		})
	}
}

func TestGetVideoMetadataFactoryError(t *testing.T) {
	original := newVideosClient
	t.Cleanup(func() { newVideosClient = original })

	wantErr := errors.New("oauth refused")
	newVideosClient = func() (videosClient, error) {
		return nil, wantErr
	}

	got, err := GetVideoMetadata(context.Background(), "vid42")
	if err == nil {
		t.Fatal("expected error from factory")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want wrap of %v", err, wantErr)
	}
	if got != nil {
		t.Errorf("expected nil metadata, got %+v", got)
	}
}

func TestGetVideoMetadataViaFactory(t *testing.T) {
	original := newVideosClient
	t.Cleanup(func() { newVideosClient = original })

	client := &mockVideosClient{
		listDoer: &mockVideoListDoer{resp: &youtube.VideoListResponse{
			Items: []*youtube.Video{
				{
					Snippet: &youtube.VideoSnippet{
						Title:       "From factory",
						Description: "desc",
						PublishedAt: "2026-02-01T00:00:00Z",
					},
				},
			},
		}},
	}
	newVideosClient = func() (videosClient, error) { return client, nil }

	got, err := GetVideoMetadata(context.Background(), "vid42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Title != "From factory" {
		t.Errorf("got %+v, want title=From factory", got)
	}
}

func TestUpdateAMAVideoInner(t *testing.T) {
	currentDescription := "Live stream description.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nLegacy boilerplate."
	currentVideo := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       "Original Title",
			Description: currentDescription,
			Tags:        []string{"old"},
			CategoryId:  "27",
		},
	}

	tests := []struct {
		name             string
		videoID          string
		title            string
		description      string
		tags             string
		timecodes        string
		listResp         *youtube.VideoListResponse
		listErr          error
		updateErr        error
		wantErr          bool
		wantErrSubstring string
		wantTitle        string
		wantTags         []string
		wantContains     []string
	}{
		{
			name:        "happy path updates title, description, tags, timecodes",
			videoID:     "vid42",
			title:       "New AMA Title",
			description: "Fresh AMA description.",
			tags:        "ama,kubernetes",
			timecodes:   "00:00 Intro\n02:00 Question",
			listResp:    &youtube.VideoListResponse{Items: []*youtube.Video{currentVideo}},
			wantTitle:   "New AMA Title",
			wantTags:    []string{"ama", "kubernetes"},
			wantContains: []string{
				"Fresh AMA description.",
				"Legacy boilerplate.",
				TimecodesHeader,
				"00:00 Intro",
			},
		},
		{
			name:        "empty title preserves original",
			videoID:     "vid42",
			title:       "",
			description: "desc",
			tags:        "tag",
			timecodes:   "",
			listResp:    &youtube.VideoListResponse{Items: []*youtube.Video{currentVideo}},
			wantTitle:   "Original Title",
			wantTags:    []string{"tag"},
		},
		{
			name:        "empty tags preserves original tags",
			videoID:     "vid42",
			title:       "T",
			description: "d",
			tags:        "",
			timecodes:   "",
			listResp:    &youtube.VideoListResponse{Items: []*youtube.Video{currentVideo}},
			wantTitle:   "T",
			wantTags:    []string{"old"},
		},
		{
			name:             "list API error",
			videoID:          "vid42",
			listErr:          errors.New("network down"),
			wantErr:          true,
			wantErrSubstring: "failed to fetch video",
		},
		{
			name:             "video not found",
			videoID:          "missing",
			listResp:         &youtube.VideoListResponse{Items: nil},
			wantErr:          true,
			wantErrSubstring: "video not found: missing",
		},
		{
			name:             "update API error",
			videoID:          "vid42",
			title:            "T",
			description:      "d",
			tags:             "tag",
			listResp:         &youtube.VideoListResponse{Items: []*youtube.Video{currentVideo}},
			updateErr:        errors.New("forbidden"),
			wantErr:          true,
			wantErrSubstring: "failed to update video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockVideosClient{
				listDoer:   &mockVideoListDoer{resp: tt.listResp, err: tt.listErr},
				updateDoer: &mockVideoUpdateDoer{ShouldFail: tt.updateErr != nil, ResponseError: tt.updateErr},
			}

			err := updateAMAVideo(context.Background(), client, tt.videoID, tt.title, tt.description, tt.tags, tt.timecodes)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrSubstring != "" && !strings.Contains(err.Error(), tt.wantErrSubstring) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrSubstring)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client.updateCalls != 1 {
				t.Errorf("updateCalls = %d, want 1", client.updateCalls)
			}
			if client.updateVideo == nil || client.updateVideo.Snippet == nil {
				t.Fatal("update video snippet missing")
			}
			if client.updateVideo.Id != tt.videoID {
				t.Errorf("update video Id = %q, want %q", client.updateVideo.Id, tt.videoID)
			}
			if client.updateVideo.Snippet.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", client.updateVideo.Snippet.Title, tt.wantTitle)
			}
			if !reflect.DeepEqual(client.updateVideo.Snippet.Tags, tt.wantTags) {
				t.Errorf("tags = %v, want %v", client.updateVideo.Snippet.Tags, tt.wantTags)
			}
			if client.updateVideo.Snippet.CategoryId != "27" {
				t.Errorf("categoryId = %q, want preserved 27", client.updateVideo.Snippet.CategoryId)
			}
			for _, s := range tt.wantContains {
				if !strings.Contains(client.updateVideo.Snippet.Description, s) {
					t.Errorf("description missing %q; got:\n%s", s, client.updateVideo.Snippet.Description)
				}
			}
		})
	}
}

func TestUpdateAMAVideoFactoryError(t *testing.T) {
	original := newVideosClient
	t.Cleanup(func() { newVideosClient = original })

	wantErr := errors.New("oauth refused")
	newVideosClient = func() (videosClient, error) { return nil, wantErr }

	err := UpdateAMAVideo(context.Background(), "vid42", "t", "d", "tag", "")
	if err == nil {
		t.Fatal("expected error from factory")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want wrap of %v", err, wantErr)
	}
}

func TestUpdateAMAVideoViaFactory(t *testing.T) {
	original := newVideosClient
	t.Cleanup(func() { newVideosClient = original })

	client := &mockVideosClient{
		listDoer: &mockVideoListDoer{resp: &youtube.VideoListResponse{
			Items: []*youtube.Video{
				{Snippet: &youtube.VideoSnippet{Title: "old", Description: "", CategoryId: "27"}},
			},
		}},
	}
	newVideosClient = func() (videosClient, error) { return client, nil }

	if err := UpdateAMAVideo(context.Background(), "vid42", "new", "desc", "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.updateCalls != 1 {
		t.Errorf("updateCalls = %d, want 1", client.updateCalls)
	}
	if client.updateVideo.Snippet.Title != "new" {
		t.Errorf("title = %q, want new", client.updateVideo.Snippet.Title)
	}
}

func TestBuildVideosClient(t *testing.T) {
	client, err := buildVideosClient(context.Background(), http.DefaultClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if _, ok := client.(*realVideosClient); !ok {
		t.Errorf("expected *realVideosClient, got %T", client)
	}
}

func TestRealVideosClientListAndUpdate(t *testing.T) {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("failed to create youtube service: %v", err)
	}
	client := &realVideosClient{svc: service.Videos}

	if doer := client.List(ctx, []string{"snippet"}, "vid42"); doer == nil {
		t.Error("expected non-nil List doer")
	}
	if doer := client.Update(ctx, []string{"snippet"}, &youtube.Video{Id: "vid42"}); doer == nil {
		t.Error("expected non-nil Update doer")
	}
}

// TestGetVideoMetadataCtxCanceled verifies that a pre-canceled ctx short-
// circuits getVideoMetadata before the API call, so the YouTube HTTP request
// is never issued. This is the fast-path that stops cancellation from being
// defeated at the API call boundary.
func TestGetVideoMetadataCtxCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &mockVideosClient{
		listDoer: &mockVideoListDoer{resp: &youtube.VideoListResponse{
			Items: []*youtube.Video{{Snippet: &youtube.VideoSnippet{Title: "should not reach"}}},
		}},
	}

	got, err := getVideoMetadata(ctx, client, "vid42")
	if err == nil {
		t.Fatal("expected error from canceled ctx")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want wrap of context.Canceled", err)
	}
	if got != nil {
		t.Errorf("expected nil metadata on cancel, got %+v", got)
	}
	if client.listCalls != 0 {
		t.Errorf("expected 0 List calls on canceled ctx, got %d", client.listCalls)
	}
}

// TestUpdateAMAVideoCtxCanceled verifies that a pre-canceled ctx short-circuits
// updateAMAVideo before any API calls, so neither the videos.list nor the
// videos.update HTTP request is issued.
func TestUpdateAMAVideoCtxCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &mockVideosClient{
		listDoer: &mockVideoListDoer{resp: &youtube.VideoListResponse{
			Items: []*youtube.Video{{Snippet: &youtube.VideoSnippet{Title: "should not reach", CategoryId: "27"}}},
		}},
	}

	err := updateAMAVideo(ctx, client, "vid42", "t", "d", "tag", "")
	if err == nil {
		t.Fatal("expected error from canceled ctx")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want wrap of context.Canceled", err)
	}
	if client.listCalls != 0 {
		t.Errorf("expected 0 List calls on canceled ctx, got %d", client.listCalls)
	}
	if client.updateCalls != 0 {
		t.Errorf("expected 0 Update calls on canceled ctx, got %d", client.updateCalls)
	}
}

// TestGetVideoMetadataForwardsCtx verifies the ctx is passed through to the
// underlying client.List call rather than dropped on the floor.
func TestGetVideoMetadataForwardsCtx(t *testing.T) {
	type ctxKey struct{}
	parent := context.WithValue(context.Background(), ctxKey{}, "marker")

	client := &mockVideosClient{
		listDoer: &mockVideoListDoer{resp: &youtube.VideoListResponse{
			Items: []*youtube.Video{{Snippet: &youtube.VideoSnippet{Title: "ok"}}},
		}},
	}

	if _, err := getVideoMetadata(parent, client, "vid42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.listCtxs) != 1 {
		t.Fatalf("expected 1 captured ctx, got %d", len(client.listCtxs))
	}
	if v, _ := client.listCtxs[0].Value(ctxKey{}).(string); v != "marker" {
		t.Errorf("ctx value = %q, want marker (ctx not forwarded)", v)
	}
}

// TestUpdateAMAVideoForwardsCtx verifies the ctx is passed through to both the
// videos.list and videos.update calls rather than dropped on the floor.
func TestUpdateAMAVideoForwardsCtx(t *testing.T) {
	type ctxKey struct{}
	parent := context.WithValue(context.Background(), ctxKey{}, "marker")

	client := &mockVideosClient{
		listDoer: &mockVideoListDoer{resp: &youtube.VideoListResponse{
			Items: []*youtube.Video{{Snippet: &youtube.VideoSnippet{Title: "old", CategoryId: "27"}}},
		}},
	}

	if err := updateAMAVideo(parent, client, "vid42", "new", "desc", "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.listCtxs) != 1 {
		t.Fatalf("expected 1 captured List ctx, got %d", len(client.listCtxs))
	}
	if v, _ := client.listCtxs[0].Value(ctxKey{}).(string); v != "marker" {
		t.Errorf("List ctx value = %q, want marker (ctx not forwarded)", v)
	}
	if len(client.updateCtxs) != 1 {
		t.Fatalf("expected 1 captured Update ctx, got %d", len(client.updateCtxs))
	}
	if v, _ := client.updateCtxs[0].Value(ctxKey{}).(string); v != "marker" {
		t.Errorf("Update ctx value = %q, want marker (ctx not forwarded)", v)
	}
}
