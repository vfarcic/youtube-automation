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

// mockPlaylistItemsListDoer returns canned responses or errors per call.
type mockPlaylistItemsListDoer struct {
	responses  []*youtube.PlaylistItemListResponse
	errs       []error
	numDoCalls int
}

func (m *mockPlaylistItemsListDoer) Do(opts ...googleapi.CallOption) (*youtube.PlaylistItemListResponse, error) {
	idx := m.numDoCalls
	m.numDoCalls++
	if idx < len(m.errs) && m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return &youtube.PlaylistItemListResponse{}, nil
}

// mockPlaylistItemsLister captures arguments and returns the shared doer.
type mockPlaylistItemsLister struct {
	doer                *mockPlaylistItemsListDoer
	capturedParts       [][]string
	capturedPlaylistIDs []string
	capturedPageTokens  []string
	numListCalls        int
}

func (m *mockPlaylistItemsLister) List(_ context.Context, part []string, playlistID, pageToken string) playlistItemsListDoer {
	m.numListCalls++
	m.capturedParts = append(m.capturedParts, part)
	m.capturedPlaylistIDs = append(m.capturedPlaylistIDs, playlistID)
	m.capturedPageTokens = append(m.capturedPageTokens, pageToken)
	return m.doer
}

func playlistItem(videoID, title, publishedAt string) *youtube.PlaylistItem {
	return &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			Title: title,
		},
		ContentDetails: &youtube.PlaylistItemContentDetails{
			VideoId:          videoID,
			VideoPublishedAt: publishedAt,
		},
	}
}

func TestListPlaylistVideosInner(t *testing.T) {
	tests := []struct {
		name             string
		playlistID       string
		responses        []*youtube.PlaylistItemListResponse
		errs             []error
		want             []PlaylistVideo
		wantErr          bool
		wantErrSubstring string
		wantNumListCalls int
		wantPageTokens   []string
	}{
		{
			name:       "happy path single page sorted by publish date desc",
			playlistID: "PL123",
			responses: []*youtube.PlaylistItemListResponse{
				{
					Items: []*youtube.PlaylistItem{
						playlistItem("vidA", "Older", "2026-01-01T10:00:00Z"),
						playlistItem("vidB", "Newest", "2026-03-01T10:00:00Z"),
						playlistItem("vidC", "Middle", "2026-02-01T10:00:00Z"),
					},
				},
			},
			want: []PlaylistVideo{
				{VideoID: "vidB", Title: "Newest", PublishedAt: "2026-03-01T10:00:00Z"},
				{VideoID: "vidC", Title: "Middle", PublishedAt: "2026-02-01T10:00:00Z"},
				{VideoID: "vidA", Title: "Older", PublishedAt: "2026-01-01T10:00:00Z"},
			},
			wantNumListCalls: 1,
			wantPageTokens:   []string{""},
		},
		{
			name:             "empty playlist returns empty slice no error",
			playlistID:       "PLempty",
			responses:        []*youtube.PlaylistItemListResponse{{Items: nil}},
			want:             nil,
			wantNumListCalls: 1,
			wantPageTokens:   []string{""},
		},
		{
			name:       "pagination merges pages and sorts",
			playlistID: "PLpaged",
			responses: []*youtube.PlaylistItemListResponse{
				{
					Items: []*youtube.PlaylistItem{
						playlistItem("v1", "Page1-A", "2026-01-01T10:00:00Z"),
						playlistItem("v2", "Page1-B", "2026-04-01T10:00:00Z"),
					},
					NextPageToken: "tok-2",
				},
				{
					Items: []*youtube.PlaylistItem{
						playlistItem("v3", "Page2-A", "2026-05-01T10:00:00Z"),
						playlistItem("v4", "Page2-B", "2026-02-01T10:00:00Z"),
					},
				},
			},
			want: []PlaylistVideo{
				{VideoID: "v3", Title: "Page2-A", PublishedAt: "2026-05-01T10:00:00Z"},
				{VideoID: "v2", Title: "Page1-B", PublishedAt: "2026-04-01T10:00:00Z"},
				{VideoID: "v4", Title: "Page2-B", PublishedAt: "2026-02-01T10:00:00Z"},
				{VideoID: "v1", Title: "Page1-A", PublishedAt: "2026-01-01T10:00:00Z"},
			},
			wantNumListCalls: 2,
			wantPageTokens:   []string{"", "tok-2"},
		},
		{
			name:             "API error on first page",
			playlistID:       "PLfail",
			errs:             []error{errors.New("quota exceeded")},
			wantErr:          true,
			wantErrSubstring: "failed to list playlist items",
			wantNumListCalls: 1,
			wantPageTokens:   []string{""},
		},
		{
			name:       "API error on second page",
			playlistID: "PLpartial",
			responses: []*youtube.PlaylistItemListResponse{
				{
					Items: []*youtube.PlaylistItem{
						playlistItem("v1", "First", "2026-01-01T10:00:00Z"),
					},
					NextPageToken: "tok-2",
				},
				nil,
			},
			errs:             []error{nil, errors.New("network error")},
			wantErr:          true,
			wantErrSubstring: "failed to list playlist items",
			wantNumListCalls: 2,
			wantPageTokens:   []string{"", "tok-2"},
		},
		{
			name:       "items with nil snippet are skipped",
			playlistID: "PLpartial",
			responses: []*youtube.PlaylistItemListResponse{
				{
					Items: []*youtube.PlaylistItem{
						{ContentDetails: &youtube.PlaylistItemContentDetails{VideoId: "vBad", VideoPublishedAt: "2026-09-01T10:00:00Z"}},
						playlistItem("vGood", "Good", "2026-08-01T10:00:00Z"),
					},
				},
			},
			want: []PlaylistVideo{
				{VideoID: "vGood", Title: "Good", PublishedAt: "2026-08-01T10:00:00Z"},
			},
			wantNumListCalls: 1,
			wantPageTokens:   []string{""},
		},
		{
			name:       "items with nil contentDetails are skipped",
			playlistID: "PLpartial",
			responses: []*youtube.PlaylistItemListResponse{
				{
					Items: []*youtube.PlaylistItem{
						{Snippet: &youtube.PlaylistItemSnippet{Title: "Bad"}},
						playlistItem("vGood", "Good", "2026-08-01T10:00:00Z"),
					},
				},
			},
			want: []PlaylistVideo{
				{VideoID: "vGood", Title: "Good", PublishedAt: "2026-08-01T10:00:00Z"},
			},
			wantNumListCalls: 1,
			wantPageTokens:   []string{""},
		},
		{
			name:       "nil item entries are skipped",
			playlistID: "PLpartial",
			responses: []*youtube.PlaylistItemListResponse{
				{
					Items: []*youtube.PlaylistItem{
						nil,
						playlistItem("vGood", "Good", "2026-08-01T10:00:00Z"),
					},
				},
			},
			want: []PlaylistVideo{
				{VideoID: "vGood", Title: "Good", PublishedAt: "2026-08-01T10:00:00Z"},
			},
			wantNumListCalls: 1,
			wantPageTokens:   []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doer := &mockPlaylistItemsListDoer{
				responses: tt.responses,
				errs:      tt.errs,
			}
			lister := &mockPlaylistItemsLister{doer: doer}

			got, err := listPlaylistVideos(context.Background(), lister, tt.playlistID)

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
			if lister.numListCalls != tt.wantNumListCalls {
				t.Errorf("numListCalls = %d, want %d", lister.numListCalls, tt.wantNumListCalls)
			}
			if !reflect.DeepEqual(lister.capturedPageTokens, tt.wantPageTokens) {
				t.Errorf("page tokens = %v, want %v", lister.capturedPageTokens, tt.wantPageTokens)
			}
			for i, p := range lister.capturedParts {
				if !reflect.DeepEqual(p, []string{"snippet", "contentDetails"}) {
					t.Errorf("call %d requested parts %v, want [snippet contentDetails]", i, p)
				}
			}
			for i, id := range lister.capturedPlaylistIDs {
				if id != tt.playlistID {
					t.Errorf("call %d playlistID = %q, want %q", i, id, tt.playlistID)
				}
			}
		})
	}
}

func TestListPlaylistVideosEmptyID(t *testing.T) {
	got, err := ListPlaylistVideos(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty playlist ID")
	}
	if got != nil {
		t.Errorf("expected nil result, got %v", got)
	}
	const want = "playlist ID cannot be empty"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestListPlaylistVideosListerFactoryError(t *testing.T) {
	original := newPlaylistItemsLister
	t.Cleanup(func() { newPlaylistItemsLister = original })

	wantErr := errors.New("oauth refused")
	newPlaylistItemsLister = func(ctx context.Context) (playlistItemsLister, error) {
		return nil, wantErr
	}

	got, err := ListPlaylistVideos(context.Background(), "PL123")
	if err == nil {
		t.Fatal("expected error from factory")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want wrap of %v", err, wantErr)
	}
	if got != nil {
		t.Errorf("expected nil result, got %v", got)
	}
}

func TestListPlaylistVideosViaFactory(t *testing.T) {
	original := newPlaylistItemsLister
	t.Cleanup(func() { newPlaylistItemsLister = original })

	doer := &mockPlaylistItemsListDoer{
		responses: []*youtube.PlaylistItemListResponse{
			{
				Items: []*youtube.PlaylistItem{
					playlistItem("vidA", "Older", "2026-01-01T10:00:00Z"),
					playlistItem("vidB", "Newer", "2026-02-01T10:00:00Z"),
				},
			},
		},
	}
	lister := &mockPlaylistItemsLister{doer: doer}
	newPlaylistItemsLister = func(ctx context.Context) (playlistItemsLister, error) {
		return lister, nil
	}

	got, err := ListPlaylistVideos(context.Background(), "PL123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []PlaylistVideo{
		{VideoID: "vidB", Title: "Newer", PublishedAt: "2026-02-01T10:00:00Z"},
		{VideoID: "vidA", Title: "Older", PublishedAt: "2026-01-01T10:00:00Z"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestBuildPlaylistItemsLister(t *testing.T) {
	lister, err := buildPlaylistItemsLister(context.Background(), http.DefaultClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lister == nil {
		t.Fatal("expected non-nil lister")
	}
	if _, ok := lister.(*realPlaylistItemsLister); !ok {
		t.Errorf("expected *realPlaylistItemsLister, got %T", lister)
	}
}

func TestRealPlaylistItemsListerList(t *testing.T) {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("failed to create youtube service: %v", err)
	}
	lister := &realPlaylistItemsLister{svc: service.PlaylistItems}

	if doer := lister.List(ctx, []string{"snippet"}, "PL123", ""); doer == nil {
		t.Error("expected non-nil doer when pageToken is empty")
	}

	if doer := lister.List(ctx, []string{"snippet", "contentDetails"}, "PL123", "tok-2"); doer == nil {
		t.Error("expected non-nil doer when pageToken is set")
	}
}

func TestGetVideoDescriptionEmptyID(t *testing.T) {
	got, err := GetVideoDescription(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty video ID")
	}
	if got != "" {
		t.Errorf("expected empty description, got %q", got)
	}
	const want = "video ID cannot be empty"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestGetVideoDescriptionSuccess(t *testing.T) {
	original := fetchVideoMetadata
	t.Cleanup(func() { fetchVideoMetadata = original })

	fetchVideoMetadata = func(_ context.Context, videoID string) (*VideoMetadata, error) {
		if videoID != "vid42" {
			t.Errorf("videoID = %q, want vid42", videoID)
		}
		return &VideoMetadata{Description: "hello world"}, nil
	}

	got, err := GetVideoDescription(context.Background(), "vid42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestGetVideoDescriptionFetchError(t *testing.T) {
	original := fetchVideoMetadata
	t.Cleanup(func() { fetchVideoMetadata = original })

	wantErr := errors.New("api down")
	fetchVideoMetadata = func(_ context.Context, videoID string) (*VideoMetadata, error) {
		return nil, wantErr
	}

	got, err := GetVideoDescription(context.Background(), "vid42")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
	if got != "" {
		t.Errorf("expected empty description, got %q", got)
	}
}

// TestGetVideoDescriptionNilMetadata verifies the defensive guard against a
// misbehaving fetchVideoMetadata override that returns (nil, nil). Without the
// guard, dereferencing metadata.Description would panic.
func TestGetVideoDescriptionNilMetadata(t *testing.T) {
	original := fetchVideoMetadata
	t.Cleanup(func() { fetchVideoMetadata = original })

	fetchVideoMetadata = func(_ context.Context, videoID string) (*VideoMetadata, error) {
		return nil, nil
	}

	got, err := GetVideoDescription(context.Background(), "vid42")
	if err == nil {
		t.Fatal("expected error from nil metadata")
	}
	if !strings.Contains(err.Error(), "no metadata returned") {
		t.Errorf("error = %q, want substring %q", err.Error(), "no metadata returned")
	}
	if got != "" {
		t.Errorf("expected empty description, got %q", got)
	}
}

// TestListPlaylistVideos_NilResponse verifies the defensive guard against a
// (nil response, nil error) from the YouTube SDK. The loop must not panic on
// resp.Items deref.
func TestListPlaylistVideos_NilResponse(t *testing.T) {
	doer := &mockPlaylistItemsListDoer{responses: []*youtube.PlaylistItemListResponse{nil}}
	lister := &mockPlaylistItemsLister{doer: doer}

	got, err := listPlaylistVideos(context.Background(), lister, "PLnil")
	if err == nil {
		t.Fatal("expected error for nil response")
	}
	if !strings.Contains(err.Error(), "nil response") {
		t.Errorf("error = %q, want substring %q", err.Error(), "nil response")
	}
	if got != nil {
		t.Errorf("expected nil result, got %v", got)
	}
}

// TestListPlaylistVideos_EmptyVideoIDSkipped verifies items with an empty
// ContentDetails.VideoId are skipped (rather than producing PlaylistVideo
// records with empty IDs that downstream callers would have to filter).
func TestListPlaylistVideos_EmptyVideoIDSkipped(t *testing.T) {
	doer := &mockPlaylistItemsListDoer{responses: []*youtube.PlaylistItemListResponse{
		{
			Items: []*youtube.PlaylistItem{
				{
					Snippet:        &youtube.PlaylistItemSnippet{Title: "Empty ID"},
					ContentDetails: &youtube.PlaylistItemContentDetails{VideoId: "", VideoPublishedAt: "2026-01-01T00:00:00Z"},
				},
				playlistItem("vGood", "Good", "2026-02-01T00:00:00Z"),
			},
		},
	}}
	lister := &mockPlaylistItemsLister{doer: doer}

	got, err := listPlaylistVideos(context.Background(), lister, "PLmix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []PlaylistVideo{{VideoID: "vGood", Title: "Good", PublishedAt: "2026-02-01T00:00:00Z"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// TestListPlaylistVideos_CtxForwarded verifies the ctx is forwarded through to
// the underlying lister.List call rather than dropped on the floor.
func TestListPlaylistVideos_CtxForwarded(t *testing.T) {
	type ctxKey struct{}
	parent := context.WithValue(context.Background(), ctxKey{}, "marker")

	doer := &mockPlaylistItemsListDoer{responses: []*youtube.PlaylistItemListResponse{
		{Items: []*youtube.PlaylistItem{playlistItem("v1", "t1", "2026-01-01T00:00:00Z")}},
	}}
	lister := &ctxCapturingLister{inner: &mockPlaylistItemsLister{doer: doer}}

	if _, err := listPlaylistVideos(parent, lister, "PL"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lister.captured) != 1 {
		t.Fatalf("expected 1 captured ctx, got %d", len(lister.captured))
	}
	if v, _ := lister.captured[0].Value(ctxKey{}).(string); v != "marker" {
		t.Errorf("ctx value = %q, want marker (ctx not forwarded)", v)
	}
}

// ctxCapturingLister wraps mockPlaylistItemsLister to capture each ctx
// passed into List, since the inner mock discards it.
type ctxCapturingLister struct {
	inner    *mockPlaylistItemsLister
	captured []context.Context
}

func (c *ctxCapturingLister) List(ctx context.Context, part []string, playlistID, pageToken string) playlistItemsListDoer {
	c.captured = append(c.captured, ctx)
	return c.inner.List(ctx, part, playlistID, pageToken)
}

// TestListPlaylistVideos_PageCap verifies the maxPlaylistPages bound triggers
// an error instead of looping forever when the API keeps returning a non-empty
// NextPageToken (a malformed-server safety net, not a normal-flow path).
func TestListPlaylistVideos_PageCap(t *testing.T) {
	responses := make([]*youtube.PlaylistItemListResponse, maxPlaylistPages+5)
	for i := range responses {
		responses[i] = &youtube.PlaylistItemListResponse{
			Items:         []*youtube.PlaylistItem{playlistItem("v", "t", "2026-01-01T00:00:00Z")},
			NextPageToken: "always-more",
		}
	}
	doer := &mockPlaylistItemsListDoer{responses: responses}
	lister := &mockPlaylistItemsLister{doer: doer}

	got, err := listPlaylistVideos(context.Background(), lister, "PLrunaway")
	if err == nil {
		t.Fatal("expected page-cap error, got nil")
	}
	if got != nil {
		t.Errorf("expected nil result on cap, got %d videos", len(got))
	}
	if !strings.Contains(err.Error(), "page cap") {
		t.Errorf("error = %q, want substring %q", err.Error(), "page cap")
	}
	if lister.numListCalls != maxPlaylistPages {
		t.Errorf("called lister %d times, want %d", lister.numListCalls, maxPlaylistPages)
	}
}
