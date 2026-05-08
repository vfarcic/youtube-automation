package publishing

import (
	"context"
	"fmt"
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

// playlistItemsLister abstracts service.PlaylistItems for testing.
type playlistItemsLister interface {
	List(part []string, playlistID, pageToken string) playlistItemsListDoer
}

// realPlaylistItemsLister adapts *youtube.PlaylistItemsService to playlistItemsLister.
type realPlaylistItemsLister struct {
	svc *youtube.PlaylistItemsService
}

func (r *realPlaylistItemsLister) List(part []string, playlistID, pageToken string) playlistItemsListDoer {
	call := r.svc.List(part).PlaylistId(playlistID).MaxResults(50)
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
// PlaylistItems endpoint. Tests may override this to inject a mock.
var newPlaylistItemsLister = func() (playlistItemsLister, error) {
	ctx := context.Background()
	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("OAuth failed: %w", err)
	}
	return buildPlaylistItemsLister(ctx, client)
}

// ListPlaylistVideos fetches all videos in the given playlist, sorted by
// publish date descending (most recent first).
func ListPlaylistVideos(playlistID string) ([]PlaylistVideo, error) {
	if playlistID == "" {
		return nil, fmt.Errorf("playlist ID cannot be empty")
	}
	lister, err := newPlaylistItemsLister()
	if err != nil {
		return nil, err
	}
	return listPlaylistVideos(lister, playlistID)
}

// listPlaylistVideos is the testable inner implementation. It paginates through
// all results and returns them sorted by PublishedAt descending.
func listPlaylistVideos(lister playlistItemsLister, playlistID string) ([]PlaylistVideo, error) {
	var result []PlaylistVideo
	pageToken := ""

	for {
		resp, err := lister.List([]string{"snippet", "contentDetails"}, playlistID, pageToken).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list playlist items: %w", err)
		}

		for _, item := range resp.Items {
			if item == nil || item.Snippet == nil || item.ContentDetails == nil {
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
var fetchVideoMetadata = GetVideoMetadata

// GetVideoDescription returns the current description of a YouTube video.
// Used by the AMA scheduler to detect the timecodes-marker for idempotency.
// Delegates to GetVideoMetadata (in youtube_update.go) to avoid duplicating
// the videos.list call.
func GetVideoDescription(videoID string) (string, error) {
	metadata, err := fetchVideoMetadata(videoID)
	if err != nil {
		return "", err
	}
	return metadata.Description, nil
}
