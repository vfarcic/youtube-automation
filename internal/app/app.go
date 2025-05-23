package app

import (
	// "bytes" // Unused, removing
	"errors"
	"fmt"
	"strings"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
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
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

// New creates a new application instance
func New() *App {
	cfg := configuration.GlobalSettings

	// Initialize ui.Renderer
	uiRenderer := &ui.Renderer{} // Simple instantiation

	fsOps := filesystem.NewOperations()
	videoManager := video.NewManager(fsOps.GetFilePath)

	// Define getDirsFunc and dirSelector (simplified, assuming default behavior for now)
	// In a real scenario, these might come from config or be more complex.
	getDirsFunc := func() ([]Directory, error) {
		// Placeholder: This would typically scan configured directories
		// For now, return an empty list or a predefined one if necessary for basic operation.
		// The actual implementation is in MenuHandler.doGetAvailableDirectories
		// This func is a dependency passed to MenuHandler, so MenuHandler should have its own way to get dirs.
		// Let's assume MenuHandler's internal doGetAvailableDirectories is sufficient and we pass nil or a dummy here
		// if NewMenuHandler doesn't strictly need it for *initialization* itself.
		// However, the MenuHandler constructor in app.go *does* take it.
		// For now, we'll pass a simple implementation that relies on the MenuHandler's internal method later.
		// This is a bit circular if MenuHandler.doGetAvailableDirectories isn't static.
		// Let's assume for the constructor it's okay to be nil if the MenuHandler populates it or uses an internal one.
		// Revisiting doGetAvailableDirectories logic: it's a method on MenuHandler, not a static func.
		// So, MenuHandler needs to be instantiated first, then its method can be used if needed for this func.
		// This suggests that getDirsFunc might not be needed by NewMenuHandler directly if it's for runtime use.

		// Let's create a dummy function for now, as MenuHandler.doGetAvailableDirectories handles the real logic.
		// The MenuHandler's SelectDirectory method will use its internal doGetAvailableDirectories.
		return []Directory{}, nil
	}

	confirmerInstance := &simpleConfirmer{}

	// Instantiate MenuHandler directly here, as NewMenuHandler doesn't exist
	mh := &MenuHandler{
		confirmer:         confirmerInstance,
		getDirsFunc:       getDirsFunc,
		uiRenderer:        uiRenderer,
		videoManager:      videoManager,
		filesystem:        fsOps,
		greenStyle:        ui.GreenStyle,
		orangeStyle:       ui.OrangeStyle,
		redStyle:          ui.RedStyle,
		farFutureStyle:    ui.FarFutureStyle,
		confirmationStyle: ui.ConfirmationStyle,
		errorStyle:        ui.ErrorStyle,
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
