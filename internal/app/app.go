package app

import (
	// "bytes" // Unused, removing
	"errors"
	"fmt"
	"strings"

	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/ui"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow" // For workflow.Directory, workflow.DirectorySelector
	"devopstoolkit/youtube-automation/pkg/utils"
	// "github.com/charmbracelet/lipgloss" // Unused in this file, removing
)

// Confirmer defines an interface for confirming actions.
// This is local to the app package.
type Confirmer interface {
	Confirm(prompt string) bool
}

// App represents the main application
type App struct {
	menuHandler *MenuHandler
	config      *configuration.Settings
	uiRenderer  *ui.Renderer
}

// simpleConfirmer implements the local app.Confirmer interface
type simpleConfirmer struct{}

func (sc *simpleConfirmer) Confirm(prompt string) bool {
	var response string
	// TODO: This direct fmt.Scanln might interact poorly with huh forms if huh is active.
	// Consider if huh provides its own confirmation mechanism that should be used when available.
	// For now, this matches a basic CLI confirmation.
	fmt.Printf("%s [y/N]: ", prompt)
	_, err := fmt.Scanln(&response) // Capture error from Scanln
	if err != nil {
		// If Scanln encounters EOF (e.g., piped input ends) or other errors,
		// it might be appropriate to return false and the error, or just false.
		// For a simple CLI, returning false on error is safer.
		return false
	}
	response = strings.TrimSpace(response)
	if response == "" {
		return false // Default to 'N' when user just presses Enter
	}
	return strings.ToLower(response) == "y"
}

// New creates a new application instance
func New() *App {
	cfg := configuration.GlobalSettings

	// Initialize ui.Renderer
	uiRenderer := &ui.Renderer{} // Simple instantiation

	fsOps := filesystem.NewOperations()
	videoManager := video.NewManager(fsOps.GetFilePath)
	videoService := service.NewVideoService("index.yaml", fsOps, videoManager)
	aspectService := aspect.NewService()

	confirmerInstance := &simpleConfirmer{}

	// Instantiate MenuHandler directly here, as NewMenuHandler doesn't exist
	mh := &MenuHandler{
		confirmer:         confirmerInstance,
		uiRenderer:        uiRenderer,
		videoManager:      videoManager,
		filesystem:        fsOps,
		videoService:      videoService,
		aspectService:     aspectService,
		greenStyle:        ui.GreenStyle,
		orangeStyle:       ui.OrangeStyle,
		redStyle:          ui.RedStyle,
		farFutureStyle:    ui.FarFutureStyle,
		confirmationStyle: ui.ConfirmationStyle,
		errorStyle:        ui.ErrorStyle,
		settings:          cfg, // Initialize the settings field
	}
	mh.dirSelector = mh // MenuHandler implements DirectorySelector

	return &App{
		config:      &cfg,
		menuHandler: mh,
		uiRenderer:  uiRenderer,
	}
}

// Run starts the main application loop
func (a *App) Run() error {
	for {
		if err := a.menuHandler.ChooseIndex(); err != nil {
			// If ChooseIndex returns an error, check if it's the signal to exit.
			if errors.Is(err, ErrExitApplication) { // Use errors.Is for checking sentinel errors
				return nil // Graceful exit from the application
			}
			// Otherwise, it's an actual error that should be propagated.
			return err
		}
		// If ChooseIndex returns nil, it means a sub-menu returned, so the main menu loop continues.
	}
	// The loop is infinite and only exits via a return statement above.
}

// Import types from workflow package for compatibility
type Directory = workflow.Directory
type DirectorySelector = workflow.DirectorySelector

// defaultConfirmer is the default implementation of confirmer using utils.ConfirmAction
type defaultConfirmer struct{}

func (dc defaultConfirmer) Confirm(message string) bool {
	return utils.ConfirmAction(message, nil)
}
