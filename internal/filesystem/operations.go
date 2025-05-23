package filesystem

import (
	"fmt"
	"regexp"
	"strings"
)

const baseDir = "manuscript/"

// sanitizeFileName removes or replaces characters that are typically invalid in file names.
func sanitizeFileName(name string) string {
	// Replace known problematic characters with a hyphen or remove them.
	// This is a basic example; a more robust solution might involve a whitelist
	// or more comprehensive blacklist based on OS-specific rules.
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-") // For Windows paths if ever relevant
	name = strings.ReplaceAll(name, "?", "")
	name = strings.ReplaceAll(name, "*", "")
	name = strings.ReplaceAll(name, "<", "")
	name = strings.ReplaceAll(name, ">", "")
	name = strings.ReplaceAll(name, "|", "")
	name = strings.ReplaceAll(name, "\"", "")

	// Consolidate multiple hyphens that might result from replacements
	re := regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")
	return name
}

// Operations handles file and directory path operations
type Operations struct{}

// NewOperations creates a new filesystem operations handler
func NewOperations() *Operations {
	return &Operations{}
}

// GetDirPath generates the directory path for a given category
func (o *Operations) GetDirPath(category string) string {
	// Category name is already sanitized by replacing spaces with hyphens and lowercasing
	sanitizedCategory := strings.ReplaceAll(strings.ToLower(category), " ", "-")
	return fmt.Sprintf("%s%s", baseDir, sanitizedCategory)
}

// GetFilePath generates the full file path for a given category, name, and extension
func (o *Operations) GetFilePath(category, name, extension string) string {
	dirPath := o.GetDirPath(category) // category is sanitized in GetDirPath

	// Sanitize the name part
	// First, convert to lower case and replace spaces, similar to category
	sanitizedName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	// Then apply more general sanitization for problematic characters
	sanitizedName = sanitizeFileName(sanitizedName)

	filePath := fmt.Sprintf("%s/%s.%s", dirPath, sanitizedName, extension)
	// Further sanitization of the whole path is generally not needed if components are clean,
	// but one final pass on the name part after construction can be an option.
	// However, the current sanitizeFileName acts on the name component before path assembly.
	return filePath
}
