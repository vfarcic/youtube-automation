package filesystem

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

// SanitizeName applies the same sanitization logic used for filenames
// This converts to lowercase, replaces spaces with hyphens, and removes problematic characters
func (o *Operations) SanitizeName(name string) string {
	// First, convert to lower case and replace spaces, similar to category
	sanitizedName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	// Then apply more general sanitization for problematic characters
	sanitizedName = sanitizeFileName(sanitizedName)
	return sanitizedName
}

// GetFilePath generates the full file path for a given category, name, and extension
// Note: This method expects the name to already be sanitized (lowercase, spaces replaced with hyphens)
// as this is now handled at the service layer for consistency
func (o *Operations) GetFilePath(category, name, extension string) string {
	// Convert category to lowercase and replace spaces with hyphens
	sanitizedCategory := strings.ReplaceAll(strings.ToLower(category), " ", "-")

	// Name is expected to already be sanitized at the service level
	// No additional sanitization needed here

	return filepath.Join("manuscript", sanitizedCategory, name+"."+extension)
}

// GetAnimations extracts animation cues and section titles from the specified markdown file.
// It processes the file line by line:
//   - Lines starting with "TODO:" are considered animation cues; the text after "TODO:" (trimmed) is added to the animations list.
//   - Lines starting with "## " are considered section headers, unless they are "## Intro", "## Setup", or "## Destroy".
//     The text after "## " (trimmed), prefixed with "Section: ", is added to both the animations and sections lists.
//
// It returns a slice of animation strings, a slice of section title strings, and any error encountered.
// Note: This function is moved from the old repository.Repo struct.
func (o *Operations) GetAnimations(filePath string) (animations, sections []string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, "\u00a0", " ") // Non-breaking space (U+00A0)
		if strings.HasPrefix(line, "TODO:") {
			line = strings.ReplaceAll(line, "TODO:", "")
			line = strings.TrimSpace(line)
			animations = append(animations, line)
		} else if strings.HasPrefix(line, "## ") {
			containsAny := false
			for _, value := range []string{"## Intro", "## Introduction", "## Setup", "## Destroy"} {
				if line == value {
					containsAny = true
					break
				}
			}
			if !containsAny {
				line = strings.Replace(line, "## ", "", 1)
				line = strings.TrimSpace(line)
				line = fmt.Sprintf("Section: %s", line)
				animations = append(animations, line)
				sections = append(sections, line)
			}
		}
	}
	if errScan := scanner.Err(); errScan != nil {
		return nil, nil, fmt.Errorf("error scanning file %s: %w", filePath, errScan)
	}

	if animations == nil {
		animations = []string{}
	}
	if sections == nil {
		sections = []string{}
	}
	return animations, sections, nil
}
