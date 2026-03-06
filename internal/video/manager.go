package video

import (
	"fmt"

	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/workflow"
)

// ProgressCalculator computes aspect-based progress from aspect mappings
type ProgressCalculator interface {
	CalculateAspectProgress(aspectKey string, video storage.Video) (int, int)
}

// Manager handles video phase determination and lifecycle operations
type Manager struct {
	filePathFunc   func(category, name, extension string) string
	progressCalc   ProgressCalculator
}

// NewManager creates a new video manager with the provided file path function and progress calculator
func NewManager(filePathFunc func(string, string, string) string, progressCalc ProgressCalculator) *Manager {
	return &Manager{
		filePathFunc: filePathFunc,
		progressCalc: progressCalc,
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

	// Phase 5: Upload
	publishCompleted, publishTotal := m.CalculatePublishingProgress(video)

	// Phase 6: Post-Publish Details
	postPublishCompleted, postPublishTotal := m.CalculatePostPublishProgress(video)

	// Sum all phases
	totalCompleted := initCompleted + workCompleted + defineCompleted + editCompleted + publishCompleted + postPublishCompleted
	totalTasks := initTotal + workTotal + defineTotal + editTotal + publishTotal + postPublishTotal

	return totalCompleted, totalTasks
}

// CalculateDefinePhaseCompletion calculates the completed and total tasks for the Definition phase.
func (m *Manager) CalculateDefinePhaseCompletion(video storage.Video) (int, int) {
	return m.progressCalc.CalculateAspectProgress("definition", video)
}

// CalculateInitialDetailsProgress calculates Initial Details phase progress on-the-fly
func (m *Manager) CalculateInitialDetailsProgress(video storage.Video) (int, int) {
	return m.progressCalc.CalculateAspectProgress("initial-details", video)
}

// CalculateWorkProgressProgress calculates Work Progress phase progress on-the-fly
func (m *Manager) CalculateWorkProgressProgress(video storage.Video) (int, int) {
	return m.progressCalc.CalculateAspectProgress("work-progress", video)
}

// CalculatePostProductionProgress calculates Post-Production phase progress on-the-fly
func (m *Manager) CalculatePostProductionProgress(video storage.Video) (int, int) {
	return m.progressCalc.CalculateAspectProgress("post-production", video)
}

// CalculatePublishingProgress calculates Publishing phase progress on-the-fly
func (m *Manager) CalculatePublishingProgress(video storage.Video) (int, int) {
	return m.progressCalc.CalculateAspectProgress("publishing", video)
}

// CalculatePostPublishProgress calculates Post-Publish phase progress on-the-fly
func (m *Manager) CalculatePostPublishProgress(video storage.Video) (int, int) {
	return m.progressCalc.CalculateAspectProgress("post-publish", video)
}

// CalculateAnalysisProgress calculates Analysis phase progress on-the-fly
func (m *Manager) CalculateAnalysisProgress(video storage.Video) (int, int) {
	return m.progressCalc.CalculateAspectProgress("analysis", video)
}

// CalculateDubbingProgress calculates Dubbing phase progress on-the-fly
// Tracks progress for Spanish dubbing (MVP): 1 long-form video + N shorts + 1 translation + 1 upload
func (m *Manager) CalculateDubbingProgress(video storage.Video) (int, int) {
	// Total = 1 (long-form) + number of shorts + 1 (translation) + 1 (upload)
	total := 1 + len(video.Shorts) + 1 + 1
	completed := 0

	// If no dubbing map exists, return 0/total
	if video.Dubbing == nil {
		return 0, total
	}

	// Check long-form video (key = "es")
	if esInfo, ok := video.Dubbing["es"]; ok && esInfo.DubbingStatus == "dubbed" {
		completed++
	}

	// Check each short (key = "es:shortN")
	for i := range video.Shorts {
		shortKey := fmt.Sprintf("es:short%d", i+1)
		if shortInfo, ok := video.Dubbing[shortKey]; ok && shortInfo.DubbingStatus == "dubbed" {
			completed++
		}
	}

	// Check translation (complete when TranslatedTitle is set for long-form)
	if esInfo, ok := video.Dubbing["es"]; ok && esInfo.Title != "" {
		completed++
	}

	// Check upload (complete when all dubbed items have been uploaded)
	// Long-form must be uploaded, plus all dubbed shorts
	allUploaded := false
	if esInfo, ok := video.Dubbing["es"]; ok && esInfo.DubbingStatus == "dubbed" && esInfo.UploadedVideoID != "" {
		allUploaded = true
		// Check all shorts that were dubbed are also uploaded
		for i := range video.Shorts {
			shortKey := fmt.Sprintf("es:short%d", i+1)
			if shortInfo, ok := video.Dubbing[shortKey]; ok && shortInfo.DubbingStatus == "dubbed" {
				if shortInfo.UploadedVideoID == "" {
					allUploaded = false
					break
				}
			}
		}
	}
	if allUploaded {
		completed++
	}

	return completed, total
}

