package workflow

import (
	"bytes"
)

// Directory represents a selectable directory option.
// Name is for display, Path is the actual file system path.
// Used by getAvailableDirectories and selectTargetDirectory.
type Directory struct {
	Name string
	Path string
}

// DirectorySelector defines an interface for selecting a directory.
// This allows for mocking in tests.
type DirectorySelector interface {
	SelectDirectory(input *bytes.Buffer) (Directory, error)
}

// Confirmer defines an interface for confirming actions.
// This allows for mocking in tests.
type Confirmer interface {
	Confirm(message string) bool
}