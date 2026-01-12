package app

import (
	"errors"

	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/ui"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/charmbracelet/lipgloss"
)

// ErrExitApplication is a sentinel error to signal application termination.
var ErrExitApplication = errors.New("user requested application exit")

// MenuHandler handles the main menu and navigation logic
type MenuHandler struct {
	confirmer         Confirmer
	dirSelector       DirectorySelector
	uiRenderer        *ui.Renderer
	videoManager      *video.Manager
	filesystem        *filesystem.Operations
	videoService      *service.VideoService
	aspectService     *aspect.Service
	greenStyle        lipgloss.Style
	orangeStyle       lipgloss.Style
	redStyle          lipgloss.Style
	farFutureStyle    lipgloss.Style
	confirmationStyle lipgloss.Style
	errorStyle        lipgloss.Style
	normalStyle       lipgloss.Style
	settings          configuration.Settings
}

// Constants for menu indices
const indexCreateVideo = 0
const indexListVideos = 1
const indexAnalyze = 2
const indexAMA = 3

const (
	actionEdit = iota
	actionDelete
	actionMoveFiles
	actionArchive
)
const actionReturn = 99

// Constants for edit phases
const (
	editPhaseInitial = iota
	editPhaseWork
	editPhaseDefinition
	editPhasePostProduction
	editPhasePublishing // "Upload" phase
	editPhaseDubbing    // "Dubbing" phase
	editPhasePostPublish
	editPhaseAnalysis
	// actionReturn can be reused for returning from this sub-menu
)

