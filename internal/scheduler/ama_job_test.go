package scheduler

import (
	"context"
	"errors"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/publishing"
)

// --- mocks ---

type mockPlaylistLister struct {
	calls    int
	playlist string
	videos   []publishing.PlaylistVideo
	err      error
}

func (m *mockPlaylistLister) ListPlaylistVideos(_ context.Context, playlistID string) ([]publishing.PlaylistVideo, error) {
	m.calls++
	m.playlist = playlistID
	return m.videos, m.err
}

type mockDescriptionReader struct {
	calls   int
	videoID string
	desc    string
	err     error
}

func (m *mockDescriptionReader) GetVideoDescription(_ context.Context, videoID string) (string, error) {
	m.calls++
	m.videoID = videoID
	return m.desc, m.err
}

type mockTranscriptFetcher struct {
	calls      int
	videoID    string
	transcript string
	err        error
}

func (m *mockTranscriptFetcher) GetTranscript(_ context.Context, videoID string) (string, error) {
	m.calls++
	m.videoID = videoID
	return m.transcript, m.err
}

type mockAIContentGenerator struct {
	calls      int
	transcript string
	content    *ai.AMAContent
	err        error
}

func (m *mockAIContentGenerator) GenerateAMAContent(_ context.Context, transcript string) (*ai.AMAContent, error) {
	m.calls++
	m.transcript = transcript
	return m.content, m.err
}

type mockVideoUpdater struct {
	calls    int
	videoID  string
	title    string
	desc     string
	tags     string
	timecode string
	err      error
}

func (m *mockVideoUpdater) UpdateAMAVideo(_ context.Context, videoID, title, description, tags, timecodes string) error {
	m.calls++
	m.videoID = videoID
	m.title = title
	m.desc = description
	m.tags = tags
	m.timecode = timecodes
	return m.err
}

// --- tests ---

const testPlaylistID = "PLtest"

func newAMAVideo(id, publishedAt string) publishing.PlaylistVideo {
	return publishing.PlaylistVideo{VideoID: id, Title: "AMA " + id, PublishedAt: publishedAt}
}

func TestAMAJob_Run(t *testing.T) {
	processedDescription := "Some intro\n\n" + publishing.TimecodesHeader + "\n00:00 hello"
	unprocessedDescription := "Just some plain text, no marker."

	listErr := errors.New("playlist API down")
	descErr := errors.New("video metadata fetch failed")
	transcriptErr := errors.New("transcript not ready")
	aiErr := errors.New("AI provider error")
	updateErr := errors.New("youtube update rejected")

	tests := []struct {
		name string

		// dependency setup
		playlistVideos []publishing.PlaylistVideo
		playlistErr    error
		descResult     string
		descErr        error
		transcript     string
		transcriptErr  error
		aiContent      *ai.AMAContent
		aiErr          error
		updateErr      error

		// expectations
		wantOutcome      Outcome
		wantVideoID      string
		wantVideoURL     string
		wantErrSubstring string

		// call expectations (-1 = don't care, otherwise exact)
		wantDescCalls       int
		wantTranscriptCalls int
		wantAICalls         int
		wantUpdateCalls     int
	}{
		{
			name:                "skipped when marker present",
			playlistVideos:      []publishing.PlaylistVideo{newAMAVideo("vid-recent", "2026-05-08T10:00:00Z"), newAMAVideo("vid-older", "2026-05-01T10:00:00Z")},
			descResult:          processedDescription,
			wantOutcome:         OutcomeSkipped,
			wantVideoID:         "vid-recent",
			wantVideoURL:        "https://www.youtube.com/watch?v=vid-recent",
			wantDescCalls:       1,
			wantTranscriptCalls: 0,
			wantAICalls:         0,
			wantUpdateCalls:     0,
		},
		{
			name:                "processed end-to-end happy path",
			playlistVideos:      []publishing.PlaylistVideo{newAMAVideo("vid-new", "2026-05-08T10:00:00Z")},
			descResult:          unprocessedDescription,
			transcript:          "the transcript",
			aiContent:           &ai.AMAContent{Title: "T", Description: "D", Tags: "a,b", Timecodes: "00:00 intro"},
			wantOutcome:         OutcomeProcessed,
			wantVideoID:         "vid-new",
			wantVideoURL:        "https://www.youtube.com/watch?v=vid-new",
			wantDescCalls:       1,
			wantTranscriptCalls: 1,
			wantAICalls:         1,
			wantUpdateCalls:     1,
		},
		{
			name:                "failed when transcript not yet ready",
			playlistVideos:      []publishing.PlaylistVideo{newAMAVideo("vid-new", "2026-05-08T10:00:00Z")},
			descResult:          unprocessedDescription,
			transcriptErr:       transcriptErr,
			wantOutcome:         OutcomeFailed,
			wantVideoID:         "vid-new",
			wantVideoURL:        "https://www.youtube.com/watch?v=vid-new",
			wantErrSubstring:    "transcript not ready",
			wantDescCalls:       1,
			wantTranscriptCalls: 1,
			wantAICalls:         0,
			wantUpdateCalls:     0,
		},
		{
			name:                "failed when AI content generation errors",
			playlistVideos:      []publishing.PlaylistVideo{newAMAVideo("vid-new", "2026-05-08T10:00:00Z")},
			descResult:          unprocessedDescription,
			transcript:          "transcript",
			aiErr:               aiErr,
			wantOutcome:         OutcomeFailed,
			wantVideoID:         "vid-new",
			wantVideoURL:        "https://www.youtube.com/watch?v=vid-new",
			wantErrSubstring:    "AI provider error",
			wantDescCalls:       1,
			wantTranscriptCalls: 1,
			wantAICalls:         1,
			wantUpdateCalls:     0,
		},
		{
			name:                "failed when AI returns nil content with no error",
			playlistVideos:      []publishing.PlaylistVideo{newAMAVideo("vid-new", "2026-05-08T10:00:00Z")},
			descResult:          unprocessedDescription,
			transcript:          "transcript",
			aiContent:           nil,
			wantOutcome:         OutcomeFailed,
			wantVideoID:         "vid-new",
			wantVideoURL:        "https://www.youtube.com/watch?v=vid-new",
			wantErrSubstring:    "no content",
			wantDescCalls:       1,
			wantTranscriptCalls: 1,
			wantAICalls:         1,
			wantUpdateCalls:     0,
		},
		{
			name:                "failed when YouTube update rejects",
			playlistVideos:      []publishing.PlaylistVideo{newAMAVideo("vid-new", "2026-05-08T10:00:00Z")},
			descResult:          unprocessedDescription,
			transcript:          "transcript",
			aiContent:           &ai.AMAContent{Title: "T", Description: "D", Tags: "x", Timecodes: "00:00 a"},
			updateErr:           updateErr,
			wantOutcome:         OutcomeFailed,
			wantVideoID:         "vid-new",
			wantVideoURL:        "https://www.youtube.com/watch?v=vid-new",
			wantErrSubstring:    "youtube update rejected",
			wantDescCalls:       1,
			wantTranscriptCalls: 1,
			wantAICalls:         1,
			wantUpdateCalls:     1,
		},
		{
			name:                "scheduler-error when playlist list fails",
			playlistErr:         listErr,
			wantOutcome:         OutcomeSchedulerError,
			wantErrSubstring:    "playlist API down",
			wantDescCalls:       0,
			wantTranscriptCalls: 0,
			wantAICalls:         0,
			wantUpdateCalls:     0,
		},
		{
			name:                "scheduler-error when playlist is empty",
			playlistVideos:      nil,
			wantOutcome:         OutcomeSchedulerError,
			wantErrSubstring:    "is empty",
			wantDescCalls:       0,
			wantTranscriptCalls: 0,
			wantAICalls:         0,
			wantUpdateCalls:     0,
		},
		{
			name:                "scheduler-error when description fetch fails",
			playlistVideos:      []publishing.PlaylistVideo{newAMAVideo("vid-new", "2026-05-08T10:00:00Z")},
			descErr:             descErr,
			wantOutcome:         OutcomeSchedulerError,
			wantVideoID:         "vid-new",
			wantVideoURL:        "https://www.youtube.com/watch?v=vid-new",
			wantErrSubstring:    "video metadata fetch failed",
			wantDescCalls:       1,
			wantTranscriptCalls: 0,
			wantAICalls:         0,
			wantUpdateCalls:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lister := &mockPlaylistLister{videos: tt.playlistVideos, err: tt.playlistErr}
			descReader := &mockDescriptionReader{desc: tt.descResult, err: tt.descErr}
			transFetcher := &mockTranscriptFetcher{transcript: tt.transcript, err: tt.transcriptErr}
			aiGen := &mockAIContentGenerator{content: tt.aiContent, err: tt.aiErr}
			updater := &mockVideoUpdater{err: tt.updateErr}

			job := &AMAJob{
				PlaylistID:         testPlaylistID,
				PlaylistLister:     lister,
				DescriptionReader:  descReader,
				TranscriptFetcher:  transFetcher,
				AIContentGenerator: aiGen,
				VideoUpdater:       updater,
			}

			got := job.Run(context.Background())

			if got.Outcome != tt.wantOutcome {
				t.Fatalf("Outcome = %q, want %q (err=%v)", got.Outcome, tt.wantOutcome, got.Err)
			}
			if got.VideoID != tt.wantVideoID {
				t.Errorf("VideoID = %q, want %q", got.VideoID, tt.wantVideoID)
			}
			if got.VideoURL != tt.wantVideoURL {
				t.Errorf("VideoURL = %q, want %q", got.VideoURL, tt.wantVideoURL)
			}
			if tt.wantErrSubstring == "" {
				if got.Err != nil {
					t.Errorf("Err = %v, want nil", got.Err)
				}
			} else {
				if got.Err == nil {
					t.Fatalf("Err = nil, want error containing %q", tt.wantErrSubstring)
				}
				if !strings.Contains(got.Err.Error(), tt.wantErrSubstring) {
					t.Errorf("Err = %q, want substring %q", got.Err.Error(), tt.wantErrSubstring)
				}
			}

			if lister.calls != 1 {
				t.Errorf("PlaylistLister called %d times, want 1", lister.calls)
			}
			if lister.calls == 1 && lister.playlist != testPlaylistID {
				t.Errorf("PlaylistLister got playlistID %q, want %q", lister.playlist, testPlaylistID)
			}
			if descReader.calls != tt.wantDescCalls {
				t.Errorf("DescriptionReader calls = %d, want %d", descReader.calls, tt.wantDescCalls)
			}
			if transFetcher.calls != tt.wantTranscriptCalls {
				t.Errorf("TranscriptFetcher calls = %d, want %d", transFetcher.calls, tt.wantTranscriptCalls)
			}
			if aiGen.calls != tt.wantAICalls {
				t.Errorf("AIContentGenerator calls = %d, want %d", aiGen.calls, tt.wantAICalls)
			}
			if updater.calls != tt.wantUpdateCalls {
				t.Errorf("VideoUpdater calls = %d, want %d", updater.calls, tt.wantUpdateCalls)
			}
		})
	}
}

// TestAMAJob_Run_PicksMostRecent verifies the job operates on the first
// playlist entry — Milestone 1's lister returns videos sorted descending by
// publish date, so [0] is "most recent".
func TestAMAJob_Run_PicksMostRecent(t *testing.T) {
	videos := []publishing.PlaylistVideo{
		newAMAVideo("vid-most-recent", "2026-05-08T10:00:00Z"),
		newAMAVideo("vid-second", "2026-05-01T10:00:00Z"),
		newAMAVideo("vid-oldest", "2026-04-24T10:00:00Z"),
	}
	lister := &mockPlaylistLister{videos: videos}
	descReader := &mockDescriptionReader{desc: "marker absent"}
	transFetcher := &mockTranscriptFetcher{transcript: "t"}
	aiGen := &mockAIContentGenerator{content: &ai.AMAContent{Title: "x"}}
	updater := &mockVideoUpdater{}

	job := &AMAJob{
		PlaylistID:         testPlaylistID,
		PlaylistLister:     lister,
		DescriptionReader:  descReader,
		TranscriptFetcher:  transFetcher,
		AIContentGenerator: aiGen,
		VideoUpdater:       updater,
	}

	got := job.Run(context.Background())

	if got.Outcome != OutcomeProcessed {
		t.Fatalf("Outcome = %q, want processed (err=%v)", got.Outcome, got.Err)
	}
	if descReader.videoID != "vid-most-recent" {
		t.Errorf("DescriptionReader saw videoID %q, want vid-most-recent", descReader.videoID)
	}
	if transFetcher.videoID != "vid-most-recent" {
		t.Errorf("TranscriptFetcher saw videoID %q, want vid-most-recent", transFetcher.videoID)
	}
	if updater.videoID != "vid-most-recent" {
		t.Errorf("VideoUpdater saw videoID %q, want vid-most-recent", updater.videoID)
	}
}

// TestAMAJob_Run_PassesGeneratedContentToUpdater verifies the AI output is
// forwarded verbatim to UpdateAMAVideo (matching the manual /api/ama/apply
// behaviour the PRD says must remain equivalent).
func TestAMAJob_Run_PassesGeneratedContentToUpdater(t *testing.T) {
	content := &ai.AMAContent{
		Title:       "Generated Title",
		Description: "Generated description body",
		Tags:        "tag1,tag2,tag3",
		Timecodes:   "00:00 Intro\n01:23 Topic",
	}

	lister := &mockPlaylistLister{videos: []publishing.PlaylistVideo{newAMAVideo("vid-1", "2026-05-08T10:00:00Z")}}
	descReader := &mockDescriptionReader{desc: "no marker here"}
	transFetcher := &mockTranscriptFetcher{transcript: "transcript text"}
	aiGen := &mockAIContentGenerator{content: content}
	updater := &mockVideoUpdater{}

	job := &AMAJob{
		PlaylistID:         testPlaylistID,
		PlaylistLister:     lister,
		DescriptionReader:  descReader,
		TranscriptFetcher:  transFetcher,
		AIContentGenerator: aiGen,
		VideoUpdater:       updater,
	}

	if got := job.Run(context.Background()); got.Outcome != OutcomeProcessed {
		t.Fatalf("Outcome = %q, want processed (err=%v)", got.Outcome, got.Err)
	}

	if aiGen.transcript != "transcript text" {
		t.Errorf("AI got transcript %q, want %q", aiGen.transcript, "transcript text")
	}
	if updater.videoID != "vid-1" || updater.title != content.Title || updater.desc != content.Description || updater.tags != content.Tags || updater.timecode != content.Timecodes {
		t.Errorf("Updater args = (%q, %q, %q, %q, %q), want (vid-1, %q, %q, %q, %q)",
			updater.videoID, updater.title, updater.desc, updater.tags, updater.timecode,
			content.Title, content.Description, content.Tags, content.Timecodes)
	}
}
