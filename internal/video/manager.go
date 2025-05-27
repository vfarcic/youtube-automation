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

// CalculateDefinePhaseCompletion calculates the completed and total tasks for the Definition phase.
func (m *Manager) CalculateDefinePhaseCompletion(video storage.Video) (completed int, total int) {
	fieldsToCount := []interface{}{
		video.Title,
		video.Description,
		video.Highlight,
		video.Tags,
		video.DescriptionTags,
		video.Tweet,
		video.Animations,
		video.RequestThumbnail, // This is a bool
		video.Gist,             // Added Gist back
	}
	total = len(fieldsToCount)

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
