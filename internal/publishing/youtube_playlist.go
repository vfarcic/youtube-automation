package publishing

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// PlaylistVideo holds the minimal metadata for a single video in a YouTube playlist.
type PlaylistVideo struct {
	VideoID     string
	Title       string
	PublishedAt string // ISO 8601 (e.g., "2025-12-06T15:00:00Z")
}

// playlistItemsListDoer wraps the chained .Do() of a PlaylistItems.List call.
type playlistItemsListDoer interface {
	Do(opts ...googleapi.CallOption) (*youtube.PlaylistItemListResponse, error)
}

// playlistItemsLister abstracts service.PlaylistItems for testing. The ctx is
// forwarded to .Context(ctx) on the underlying call so callers can cancel
// pagination via the standard context plumbing.
type playlistItemsLister interface {
	List(ctx context.Context, part []string, playlistID, pageToken string) playlistItemsListDoer
}

// realPlaylistItemsLister adapts *youtube.PlaylistItemsService to playlistItemsLister.
type realPlaylistItemsLister struct {
	svc *youtube.PlaylistItemsService
}

func (r *realPlaylistItemsLister) List(ctx context.Context, part []string, playlistID, pageToken string) playlistItemsListDoer {
	call := r.svc.List(part).PlaylistId(playlistID).MaxResults(50).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	return call
}

// buildPlaylistItemsLister constructs a lister from an authenticated *http.Client.
// Split out so tests can exercise service construction without a real OAuth flow.
func buildPlaylistItemsLister(ctx context.Context, client *http.Client) (playlistItemsLister, error) {
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}
	return &realPlaylistItemsLister{svc: service.PlaylistItems}, nil
}

// newPlaylistItemsLister constructs an authenticated lister for the
// PlaylistItems endpoint. Tests may override this to inject a mock. The ctx
// is propagated through OAuth and service construction so cancellation
// terminates startup work as well as the eventual API calls.
var newPlaylistItemsLister = func(ctx context.Context) (playlistItemsLister, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("OAuth failed: %w", err)
	}
	return buildPlaylistItemsLister(ctx, client)
}

// maxPlaylistPages caps pagination so a pathological playlist (or buggy API
// response that returns NextPageToken indefinitely) cannot trap the scheduler
// in an unbounded loop. At maxResults=50 this is 2,500 videos — well above
// any realistic playlist size.
const maxPlaylistPages = 50

// ListPlaylistVideos fetches all videos in the given playlist, sorted by
// publish date descending (most recent first). Pagination is capped at
// maxPlaylistPages to prevent unbounded loops on malformed API responses;
// the supplied ctx is forwarded to every API call so callers can cancel.
func ListPlaylistVideos(ctx context.Context, playlistID string) ([]PlaylistVideo, error) {
	if playlistID == "" {
		return nil, fmt.Errorf("playlist ID cannot be empty")
	}
	lister, err := newPlaylistItemsLister(ctx)
	if err != nil {
		return nil, err
	}
	return listPlaylistVideos(ctx, lister, playlistID)
}

// listPlaylistVideos is the testable inner implementation. It paginates through
// all results, enforces maxPlaylistPages, and returns them sorted by
// PublishedAt descending.
func listPlaylistVideos(ctx context.Context, lister playlistItemsLister, playlistID string) ([]PlaylistVideo, error) {
	var result []PlaylistVideo
	pageToken := ""

	for page := 0; ; page++ {
		if page >= maxPlaylistPages {
			return nil, fmt.Errorf("playlist %s exceeded page cap of %d (possible malformed pagination response)", playlistID, maxPlaylistPages)
		}

		resp, err := lister.List(ctx, []string{"snippet", "contentDetails"}, playlistID, pageToken).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list playlist items: %w", err)
		}
		if resp == nil {
			return nil, fmt.Errorf("playlist %s returned nil response on page %d", playlistID, page)
		}

		for _, item := range resp.Items {
			if item == nil || item.Snippet == nil || item.ContentDetails == nil {
				slog.Warn("skipping playlist item with missing metadata", "playlistID", playlistID, "page", page)
				continue
			}
			if item.ContentDetails.VideoId == "" {
				slog.Warn("skipping playlist item with empty video ID", "playlistID", playlistID, "page", page, "title", item.Snippet.Title)
				continue
			}
			result = append(result, PlaylistVideo{
				VideoID:     item.ContentDetails.VideoId,
				Title:       item.Snippet.Title,
				PublishedAt: item.ContentDetails.VideoPublishedAt,
			})
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].PublishedAt > result[j].PublishedAt
	})

	return result, nil
}

// fetchVideoMetadata is a hook over GetVideoMetadata so tests can mock it.
// The ctx is forwarded to the underlying YouTube API call so callers can
// cancel pagination or video-level fetches via the standard context plumbing.
var fetchVideoMetadata = func(ctx context.Context, videoID string) (*VideoMetadata, error) {
	return GetVideoMetadata(ctx, videoID)
}

// GetVideoDescription returns the current description of a YouTube video.
// Used by the AMA scheduler to detect the timecodes-marker for idempotency.
// Delegates to GetVideoMetadata (in youtube_update.go) to avoid duplicating
// the videos.list call. The ctx is forwarded so callers can cancel the
// underlying API work.
func GetVideoDescription(ctx context.Context, videoID string) (string, error) {
	if videoID == "" {
		return "", fmt.Errorf("video ID cannot be empty")
	}
	metadata, err := fetchVideoMetadata(ctx, videoID)
	if err != nil {
		return "", err
	}
	if metadata == nil {
		return "", fmt.Errorf("no metadata returned for video %s", videoID)
	}
	return metadata.Description, nil
}
