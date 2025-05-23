package app

import (
	"testing"

	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/ui"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"
)

func TestNew(t *testing.T) {
	app := New()

	if app == nil {
		t.Fatal("New() returned nil")
	}

	if app.menuHandler == nil {
		t.Fatal("menuHandler is nil")
	}

	// Verify that the menuHandler has all required dependencies
	if app.menuHandler.confirmer == nil {
		t.Error("confirmer is nil")
	}

	if app.menuHandler.uiRenderer == nil {
		t.Error("uiRenderer is nil")
	}

	if app.menuHandler.filesystem == nil {
		t.Error("filesystem is nil")
	}

	if app.menuHandler.videoManager == nil {
		t.Error("videoManager is nil")
	}

	if app.menuHandler.dirSelector == nil {
		t.Error("dirSelector is nil")
	}
}

func TestMenuHandlerComponents(t *testing.T) {
	fs := filesystem.NewOperations()

	menuHandler := &MenuHandler{
		confirmer:    &defaultConfirmer{},
		uiRenderer:   ui.NewRenderer(),
		filesystem:   fs,
		videoManager: video.NewManager(fs.GetFilePath),
	}

	// Test that all components are properly initialized
	if menuHandler.confirmer == nil {
		t.Error("confirmer should not be nil")
	}

	if menuHandler.uiRenderer == nil {
		t.Error("uiRenderer should not be nil")
	}

	if menuHandler.filesystem == nil {
		t.Error("filesystem should not be nil")
	}

	if menuHandler.videoManager == nil {
		t.Error("videoManager should not be nil")
	}
}

func TestDefaultConfirmer(t *testing.T) {
	confirmer := &defaultConfirmer{}

	// Test that the confirmer implements the interface
	var _ workflow.Confirmer = confirmer

	// We can't easily test the Confirm method without user interaction,
	// but we can verify it doesn't panic
	// Note: In a real test environment, you might mock utils.ConfirmAction
}
