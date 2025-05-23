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
func (a *App) Run() {
	for {
		a.menuHandler.ChooseIndex()
	}
}

// Import types from workflow package for compatibility
type Directory = workflow.Directory
type DirectorySelector = workflow.DirectorySelector
type confirmer = workflow.Confirmer

// defaultConfirmer is the default implementation of confirmer using utils.ConfirmAction
type defaultConfirmer struct{}

func (dc defaultConfirmer) Confirm(message string) bool {
	return utils.ConfirmAction(message)
}
