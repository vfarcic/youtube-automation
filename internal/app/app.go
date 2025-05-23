package app

import (
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/ui"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"
	"devopstoolkit/youtube-automation/pkg/utils"
)

// App represents the main application
type App struct {
	menuHandler *MenuHandler
}

// New creates a new application instance
func New() *App {
	fs := filesystem.NewOperations()

	menuHandler := &MenuHandler{
		confirmer:    &defaultConfirmer{},
		uiRenderer:   ui.NewRenderer(),
		filesystem:   fs,
		videoManager: video.NewManager(fs.GetFilePath),
	}

	// Set up directory functions
	menuHandler.getDirsFunc = menuHandler.doGetAvailableDirectories
	menuHandler.dirSelector = menuHandler // MenuHandler implements DirectorySelector

	return &App{
		menuHandler: menuHandler,
	}
}

// Run starts the main application loop
func (a *App) Run() error {
	for {
		if err := a.menuHandler.ChooseIndex(); err != nil {
			// If ChooseIndex returns an error, it means either a real error occurred
			// or the user chose to exit (which we've mapped to return nil from ChooseIndex).
			// If it's a non-nil error, we propagate it up.
			// If it's nil, it means a graceful exit from the menu, so we break the loop.
			if err.Error() == "user chose to exit" { // A bit fragile, could use a custom error type
				return nil // Graceful exit
			}
			return err // Propagate actual errors
		}
	}
	// return nil // Unreachable due to infinite loop unless break/return inside
}

// Import types from workflow package for compatibility
type Directory = workflow.Directory
type DirectorySelector = workflow.DirectorySelector
type Confirmer = workflow.Confirmer

// defaultConfirmer is the default implementation of confirmer using utils.ConfirmAction
type defaultConfirmer struct{}

func (dc defaultConfirmer) Confirm(message string) bool {
	return utils.ConfirmAction(message)
}
