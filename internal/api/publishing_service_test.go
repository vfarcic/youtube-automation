package api

import (
	"context"
	"testing"

	"devopstoolkit/youtube-automation/internal/platform/bluesky"
	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/storage"
)

func TestDefaultPublishingService_PostSlack_NotConfigured(t *testing.T) {
	svc := NewDefaultPublishingService(bluesky.Config{}, nil, nil)

	video := &storage.Video{
		Name:     "test-video",
		Category: "devops",
		VideoId:  "yt-123",
	}

	err := svc.PostSlack(context.Background(), video, "path/to/video.yaml")
	if err == nil {
		t.Fatal("expected error when Slack is not configured, got nil")
	}

	expectedMsg := "Slack is not configured (SLACK_API_TOKEN environment variable required)"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDefaultPublishingService_CreateHugoPost_NotConfigured(t *testing.T) {
	svc := NewDefaultPublishingService(bluesky.Config{}, nil, nil)

	video := &storage.Video{
		Name:     "test-video",
		Category: "devops",
		Titles:   []storage.TitleVariant{{Index: 1, Text: "Test Video"}},
	}

	_, err := svc.CreateHugoPost(context.Background(), video, &publishing.HugoPostOptions{})
	if err == nil {
		t.Fatal("expected error when Hugo is not configured, got nil")
	}

	expectedMsg := "Hugo is not configured (check hugo settings in settings.yaml)"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}
