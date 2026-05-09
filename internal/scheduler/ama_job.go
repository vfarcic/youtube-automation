// Package scheduler contains the in-app cron job that automates AMA processing.
//
// The AMA job orchestrates: list playlist videos -> read latest video's
// description -> check for the timecodes marker -> if absent, fetch
// transcript, generate AI content, and apply to YouTube. It returns a typed
// result that downstream code (email notifier, scheduler loop) can act on
// without re-deriving outcome from error inspection.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/publishing"
)

// Outcome enumerates the categorical results of an AMA job run.
//
// The four values map 1:1 to the notification rules in PRD #386:
//   - skipped:         marker present, no email
//   - processed:       marker absent, applied successfully
//   - failed:          marker absent, processing pipeline failed
//   - scheduler-error: pre-decision failure (playlist or description unreadable)
type Outcome string

const (
	OutcomeSkipped        Outcome = "skipped"
	OutcomeProcessed      Outcome = "processed"
	OutcomeFailed         Outcome = "failed"
	OutcomeSchedulerError Outcome = "scheduler-error"
)

// AMAJobResult is the typed result of one AMA job execution.
//
// VideoID and VideoURL are populated whenever the job advanced far enough to
// know which video it was operating on (i.e. anything past the playlist list
// step). Err holds the underlying failure for failed and scheduler-error
// outcomes; it is nil for skipped and processed.
type AMAJobResult struct {
	Outcome  Outcome
	VideoID  string
	VideoURL string
	Err      error
}

// PlaylistLister lists videos in a YouTube playlist, sorted with the most
// recently published video first.
type PlaylistLister interface {
	ListPlaylistVideos(playlistID string) ([]publishing.PlaylistVideo, error)
}

// DescriptionReader returns the current description of a YouTube video. The
// scheduler uses this to detect the idempotency marker.
type DescriptionReader interface {
	GetVideoDescription(videoID string) (string, error)
}

// TranscriptFetcher returns the transcript for a YouTube video.
type TranscriptFetcher interface {
	GetTranscript(ctx context.Context, videoID string) (string, error)
}

// AIContentGenerator generates AMA title/description/tags/timecodes from a
// transcript.
type AIContentGenerator interface {
	GenerateAMAContent(ctx context.Context, transcript string) (*ai.AMAContent, error)
}

// VideoUpdater applies generated AMA content to a YouTube video.
type VideoUpdater interface {
	UpdateAMAVideo(ctx context.Context, videoID, title, description, tags, timecodes string) error
}

// AMAJob bundles the dependencies and configuration needed to run a single
// AMA processing pass. All collaborators are interfaces so tests can supply
// fakes and so the production wiring can compose existing services.
type AMAJob struct {
	PlaylistID         string
	PlaylistLister     PlaylistLister
	DescriptionReader  DescriptionReader
	TranscriptFetcher  TranscriptFetcher
	AIContentGenerator AIContentGenerator
	VideoUpdater       VideoUpdater
}

// Run executes one AMA processing pass and returns a typed result describing
// what happened. It never panics and never returns an error directly; all
// failures are encoded in the result so callers can route by outcome.
func (j *AMAJob) Run(ctx context.Context) AMAJobResult {
	videos, err := j.PlaylistLister.ListPlaylistVideos(j.PlaylistID)
	if err != nil {
		return AMAJobResult{
			Outcome: OutcomeSchedulerError,
			Err:     fmt.Errorf("list playlist videos: %w", err),
		}
	}
	if len(videos) == 0 {
		return AMAJobResult{
			Outcome: OutcomeSchedulerError,
			Err:     fmt.Errorf("playlist %q is empty", j.PlaylistID),
		}
	}

	latest := videos[0]
	videoID := latest.VideoID
	url := videoURL(videoID)

	description, err := j.DescriptionReader.GetVideoDescription(videoID)
	if err != nil {
		return AMAJobResult{
			Outcome:  OutcomeSchedulerError,
			VideoID:  videoID,
			VideoURL: url,
			Err:      fmt.Errorf("read description for video %s: %w", videoID, err),
		}
	}

	if strings.Contains(description, publishing.TimecodesHeader) {
		return AMAJobResult{
			Outcome:  OutcomeSkipped,
			VideoID:  videoID,
			VideoURL: url,
		}
	}

	transcript, err := j.TranscriptFetcher.GetTranscript(ctx, videoID)
	if err != nil {
		return AMAJobResult{
			Outcome:  OutcomeFailed,
			VideoID:  videoID,
			VideoURL: url,
			Err:      fmt.Errorf("fetch transcript: %w", err),
		}
	}

	content, err := j.AIContentGenerator.GenerateAMAContent(ctx, transcript)
	if err != nil {
		return AMAJobResult{
			Outcome:  OutcomeFailed,
			VideoID:  videoID,
			VideoURL: url,
			Err:      fmt.Errorf("generate AMA content: %w", err),
		}
	}
	if content == nil {
		return AMAJobResult{
			Outcome:  OutcomeFailed,
			VideoID:  videoID,
			VideoURL: url,
			Err:      errors.New("generate AMA content: AI returned no content"),
		}
	}

	if err := j.VideoUpdater.UpdateAMAVideo(ctx, videoID, content.Title, content.Description, content.Tags, content.Timecodes); err != nil {
		return AMAJobResult{
			Outcome:  OutcomeFailed,
			VideoID:  videoID,
			VideoURL: url,
			Err:      fmt.Errorf("apply AMA content to YouTube: %w", err),
		}
	}

	return AMAJobResult{
		Outcome:  OutcomeProcessed,
		VideoID:  videoID,
		VideoURL: url,
	}
}

func videoURL(videoID string) string {
	if videoID == "" {
		return ""
	}
	return "https://www.youtube.com/watch?v=" + videoID
}
