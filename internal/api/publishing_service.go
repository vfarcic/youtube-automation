package api

import (
	"context"

	"devopstoolkit/youtube-automation/internal/platform/bluesky"
	"devopstoolkit/youtube-automation/internal/publishing"
	slackpkg "devopstoolkit/youtube-automation/internal/slack"
	"devopstoolkit/youtube-automation/internal/storage"
)

// PublishingService abstracts publishing operations for testability.
type PublishingService interface {
	UploadVideo(ctx context.Context, video *storage.Video) (string, error)
	UploadThumbnail(ctx context.Context, videoID, thumbnailPath string) error
	UploadShort(ctx context.Context, filePath string, short storage.Short, mainVideoID string) (string, error)
	CreateHugoPost(ctx context.Context, video *storage.Video, opts *publishing.HugoPostOptions) (string, error)
	GetTranscript(ctx context.Context, videoID string) (string, error)
	GetVideoMetadata(ctx context.Context, videoID string) (*publishing.VideoMetadata, error)
	PostBlueSky(ctx context.Context, text, videoID, thumbnailPath string) error
	PostSlack(ctx context.Context, video *storage.Video, videoPath string) error
	UpdateAMAVideo(ctx context.Context, videoID, title, description, tags, timecodes string) error
}

// DefaultPublishingService delegates to existing publishing functions.
type DefaultPublishingService struct {
	blueskyConfig bluesky.Config
	hugo          *publishing.Hugo
	slackService  *slackpkg.SlackService
}

// NewDefaultPublishingService creates a new DefaultPublishingService.
func NewDefaultPublishingService(bsCfg bluesky.Config, hugo *publishing.Hugo, slackSvc *slackpkg.SlackService) *DefaultPublishingService {
	return &DefaultPublishingService{
		blueskyConfig: bsCfg,
		hugo:          hugo,
		slackService:  slackSvc,
	}
}

func (d *DefaultPublishingService) UploadVideo(_ context.Context, video *storage.Video) (string, error) {
	return publishing.UploadVideo(video)
}

func (d *DefaultPublishingService) UploadThumbnail(_ context.Context, videoID, thumbnailPath string) error {
	return publishing.UploadThumbnail(videoID, thumbnailPath)
}

func (d *DefaultPublishingService) UploadShort(_ context.Context, filePath string, short storage.Short, mainVideoID string) (string, error) {
	return publishing.UploadShort(filePath, short, mainVideoID)
}

func (d *DefaultPublishingService) CreateHugoPost(_ context.Context, video *storage.Video, opts *publishing.HugoPostOptions) (string, error) {
	if d.hugo == nil {
		return "", nil
	}
	return d.hugo.Post(video, opts)
}

func (d *DefaultPublishingService) GetTranscript(_ context.Context, videoID string) (string, error) {
	return publishing.GetTranscript(videoID)
}

func (d *DefaultPublishingService) GetVideoMetadata(_ context.Context, videoID string) (*publishing.VideoMetadata, error) {
	return publishing.GetVideoMetadata(videoID)
}

func (d *DefaultPublishingService) PostBlueSky(_ context.Context, text, videoID, thumbnailPath string) error {
	return bluesky.SendPost(d.blueskyConfig, text, videoID, thumbnailPath)
}

func (d *DefaultPublishingService) PostSlack(_ context.Context, video *storage.Video, videoPath string) error {
	if d.slackService == nil {
		return nil
	}
	return d.slackService.PostVideo(video, videoPath)
}

func (d *DefaultPublishingService) UpdateAMAVideo(_ context.Context, videoID, title, description, tags, timecodes string) error {
	return publishing.UpdateAMAVideo(videoID, title, description, tags, timecodes)
}
