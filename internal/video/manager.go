package video

import (
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/workflow"
	"strings"
)

// Manager handles video phase determination and lifecycle operations
type Manager struct {
	filePathFunc func(category, name, extension string) string
}

// NewManager creates a new video manager with the provided file path function
func NewManager(filePathFunc func(string, string, string) string) *Manager {
	return &Manager{
		filePathFunc: filePathFunc,
	}
}

// GetVideoPhase determines the current phase of a video based on its state
func (m *Manager) GetVideoPhase(vi storage.VideoIndex) int {
	yaml := storage.YAML{}
	video, err := yaml.GetVideo(m.filePathFunc(vi.Category, vi.Name, "yaml"))
	if err != nil {
		// Log the error and return a default phase or handle appropriately
		// For now, returning PhaseIdeas
		// Consider a more robust error handling or logging strategy here
		println("Error getting video for phase determination:", err.Error()) // Basic logging
		return workflow.PhaseIdeas
	}

	return CalculateVideoPhase(video)
}

// CalculateVideoPhase determines the current phase of a video based on its state
// This is a pure function that doesn't require file I/O, making it suitable for use
// in contexts where the video data is already loaded (like API handlers)
func CalculateVideoPhase(video storage.Video) int {
	if len(video.Sponsorship.Blocked) > 0 { // Check for sponsorship block first
		return workflow.PhaseSponsoredBlocked
	} else if video.Delayed { // Then check for delayed
		return workflow.PhaseDelayed
	} else if len(video.Repo) > 0 { // Assuming video.Repo is populated when published
		return workflow.PhasePublished
	} else if len(video.UploadVideo) > 0 && len(video.Tweet) > 0 { // Assuming these indicate pending publish
		return workflow.PhasePublishPending
	} else if video.RequestEdit {
		return workflow.PhaseEditRequested
	} else if video.Code && video.Screen && video.Head && video.Diagrams { // Assuming these are key for material done
		return workflow.PhaseMaterialDone
	} else if len(video.Date) > 0 { // Date implies started
		return workflow.PhaseStarted
	} else {
		return workflow.PhaseIdeas
	}
}

// ===============================================
// SHARED PROGRESS CALCULATION FUNCTIONS
// ===============================================
// These functions provide a single source of truth for progress calculations
// used by both CLI and API to ensure consistency

// CalculateOverallProgress calculates the combined progress across all video phases
// This function is used by both CLI and API to ensure consistent calculations
func (m *Manager) CalculateOverallProgress(video storage.Video) (int, int) {
	// Phase 1: Initial Details
	initCompleted, initTotal := m.CalculateInitialDetailsProgress(video)

	// Phase 2: Work Progress
	workCompleted, workTotal := m.CalculateWorkProgressProgress(video)

	// Phase 3: Definition - special calculation
	defineCompleted, defineTotal := m.CalculateDefinePhaseCompletion(video)

	// Phase 4: Post-Production
	editCompleted, editTotal := m.CalculatePostProductionProgress(video)

	// Phase 5: Publishing Details
	publishCompleted, publishTotal := m.CalculatePublishingProgress(video)

	// Phase 6: Post-Publish Details
	postPublishCompleted, postPublishTotal := m.CalculatePostPublishProgress(video)

	// Sum all phases
	totalCompleted := initCompleted + workCompleted + defineCompleted + editCompleted + publishCompleted + postPublishCompleted
	totalTasks := initTotal + workTotal + defineTotal + editTotal + publishTotal + postPublishTotal

	return totalCompleted, totalTasks
}

// CalculateDefinePhaseCompletion calculates the completed and total tasks for the Definition phase.
func (m *Manager) CalculateDefinePhaseCompletion(video storage.Video) (completed int, total int) {
	// Check Titles array - at least one title with non-empty text is required
	titleComplete := false
	if len(video.Titles) > 0 {
		for _, t := range video.Titles {
			if len(strings.TrimSpace(t.Text)) > 0 && strings.TrimSpace(t.Text) != "-" {
				titleComplete = true
				break
			}
		}
	}

	fieldsToCount := []interface{}{
		video.Description,
		video.Tags,
		video.DescriptionTags,
		video.Tweet,
		video.Animations,
		video.RequestThumbnail, // This is a bool
		// Gist removed - it belongs to InitialDetails phase, not Definition phase
	}
	total = len(fieldsToCount) + 1 // +1 for title check

	// Count title completion
	if titleComplete {
		completed++
	}

	// Count other fields
	for _, field := range fieldsToCount {
		switch v := field.(type) {
		case string:
			if len(strings.TrimSpace(v)) > 0 && strings.TrimSpace(v) != "-" {
				completed++
			}
		case bool:
			if v { // For RequestThumbnail, true means complete
				completed++
			}
		}
	}
	return completed, total
}

// CalculateInitialDetailsProgress calculates Initial Details phase progress on-the-fly
func (m *Manager) CalculateInitialDetailsProgress(video storage.Video) (int, int) {
	var completedCount, totalCount int

	// General fields
	generalFields := []interface{}{
		video.ProjectName,
		video.ProjectURL,
		video.Gist,
		video.Date,
	}
	c, t := m.countCompletedTasks(generalFields)
	completedCount += c
	totalCount += t

	// Sponsorship.Amount
	totalCount++
	if len(video.Sponsorship.Amount) > 0 {
		completedCount++
	}

	// Sponsorship.Name
	totalCount++
	if len(video.Sponsorship.Name) > 0 {
		completedCount++
	}

	// Sponsorship.URL
	totalCount++
	if len(video.Sponsorship.URL) > 0 {
		completedCount++
	}

	// Special conditions (3 additional tasks)
	totalCount += 3

	// Condition 1: Sponsorship Emails
	if len(video.Sponsorship.Amount) == 0 || video.Sponsorship.Amount == "N/A" || video.Sponsorship.Amount == "-" || len(video.Sponsorship.Emails) > 0 {
		completedCount++
	}

	// Condition 2: Sponsorship Blocked
	if len(video.Sponsorship.Blocked) == 0 {
		completedCount++
	}

	// Condition 3: Delayed
	if !video.Delayed {
		completedCount++
	}

	return completedCount, totalCount
}

// CalculateWorkProgressProgress calculates Work Progress phase progress on-the-fly
func (m *Manager) CalculateWorkProgressProgress(video storage.Video) (int, int) {
	fields := []interface{}{
		video.Code,
		video.Head,
		video.Screen,
		video.RelatedVideos,
		video.Thumbnails,
		video.Diagrams,
		video.Screenshots,
		video.Location,
		video.Tagline,
		video.TaglineIdeas,
		video.OtherLogos,
	}
	return m.countCompletedTasks(fields)
}

// CalculatePostProductionProgress calculates Post-Production phase progress on-the-fly
func (m *Manager) CalculatePostProductionProgress(video storage.Video) (int, int) {
	fields := []interface{}{
		video.Thumbnail,
		video.Members,
		video.RequestEdit,
		video.Movie,
		video.Slides,
	}
	completed, total := m.countCompletedTasks(fields)

	// Special handling for Timecodes
	total++
	if video.Timecodes != "" && !m.containsString(video.Timecodes, "FIXME:") {
		completed++
	}

	return completed, total
}

// CalculatePublishingProgress calculates Publishing phase progress on-the-fly
func (m *Manager) CalculatePublishingProgress(video storage.Video) (int, int) {
	fields := []interface{}{
		video.UploadVideo,
		video.HugoPath,
	}
	return m.countCompletedTasks(fields)
}

// CalculatePostPublishProgress calculates Post-Publish phase progress on-the-fly
func (m *Manager) CalculatePostPublishProgress(video storage.Video) (int, int) {
	fields := []interface{}{
		video.DOTPosted,
		video.BlueSkyPosted,
		video.LinkedInPosted,
		video.SlackPosted,
		video.YouTubeHighlight,
		video.YouTubeComment,
		video.YouTubeCommentReply,
		video.GDE,
		video.Repo,
	}
	completed, total := m.countCompletedTasks(fields)

	// Special logic for NotifiedSponsors
	total++
	if video.NotifiedSponsors || len(video.Sponsorship.Amount) == 0 || video.Sponsorship.Amount == "N/A" || video.Sponsorship.Amount == "-" {
		completed++
	}

	return completed, total
}

// CalculateAnalysisProgress calculates Analysis phase progress on-the-fly
func (m *Manager) CalculateAnalysisProgress(video storage.Video) (int, int) {
	// If no titles exist, return 0/0 (nothing to track)
	if len(video.Titles) == 0 {
		return 0, 0
	}

	completed := 0
	total := len(video.Titles)

	// Count titles that have share percentages filled (Share > 0)
	for _, title := range video.Titles {
		if title.Share > 0 {
			completed++
		}
	}

	return completed, total
}

// countCompletedTasks counts completed tasks based on field values
func (m *Manager) countCompletedTasks(fields []interface{}) (completed int, total int) {
	for _, field := range fields {
		switch v := field.(type) {
		case string:
			if len(v) > 0 && v != "-" {
				completed++
			}
		case bool:
			if v {
				completed++
			}
		}
		total++
	}
	return completed, total
}

// containsString checks if a string contains a substring
func (m *Manager) containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
