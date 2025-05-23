package filesystem

import (
	"fmt"
	"strings"
)

// Operations handles file and directory path operations
type Operations struct{}

// NewOperations creates a new filesystem operations handler
func NewOperations() *Operations {
	return &Operations{}
}

// GetDirPath generates the directory path for a given category
func (o *Operations) GetDirPath(category string) string {
	return fmt.Sprintf("manuscript/%s", strings.ReplaceAll(strings.ToLower(category), " ", "-"))
}

// GetFilePath generates the full file path for a given category, name, and extension
func (o *Operations) GetFilePath(category, name, extension string) string {
	dirPath := o.GetDirPath(category)
	filePath := fmt.Sprintf("%s/%s.%s", dirPath, strings.ToLower(name), extension)
	filePath = strings.ReplaceAll(filePath, " ", "-")
	filePath = strings.ReplaceAll(filePath, "?", "")
	return filePath
}