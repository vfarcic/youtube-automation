package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Hugo struct{}

func (r *Hugo) Post(gist, title, date string) (string, error) {
	if gist == "N/A" {
		return "", nil
	}
	post, err := r.getPost(gist, title, date)
	if err != nil {
		return "", err
	}
	return r.hugoFromMarkdown(gist, title, post)
}

func (r *Hugo) hugoFromMarkdown(filePath, title, post string) (string, error) {
	// Convert the manuscript path to a content path
	relPath, err := filepath.Rel(filepath.Join(settings.Hugo.Path, "manuscript"), filepath.Dir(filePath))
	if err != nil {
		// If we can't make a relative path, try to extract the category from the path structure
		relPath = filepath.Base(filepath.Dir(filePath))
	}

	// Use filepath.Join for proper path construction
	categoryDir := filepath.Join(settings.Hugo.Path, "content", relPath)

	// Sanitize the title for use as a directory name
	postDir := title
	postDir = strings.ReplaceAll(postDir, " ", "-")
	postDir = strings.ReplaceAll(postDir, "(", "")
	postDir = strings.ReplaceAll(postDir, ")", "")
	postDir = strings.ReplaceAll(postDir, ":", "")
	postDir = strings.ReplaceAll(postDir, "&", "")
	postDir = strings.ReplaceAll(postDir, "/", "-")
	postDir = strings.ReplaceAll(postDir, "'", "")
	postDir = strings.ReplaceAll(postDir, "!", "")
	postDir = strings.ToLower(postDir)

	// Create the full directory path using filepath.Join
	fullDir := filepath.Join(categoryDir, postDir)
	if err := os.MkdirAll(fullDir, os.FileMode(0755)); err != nil {
		return "", err
	}

	// Create the output file path using filepath.Join
	hugoPath := filepath.Join(fullDir, "_index.md")
	if err := os.WriteFile(hugoPath, []byte(post), 0644); err != nil {
		return "", err
	}
	return hugoPath, nil
}

func (r *Hugo) getPost(filePath, title, date string) (string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err // Return error instead of log.Fatal for better testability
	}
	content := fmt.Sprintf(`
+++
title = '%s'
date = %s:00+00:00
draft = false
+++

FIXME:

<!--more-->

{{< youtube FIXME: >}}

%s
`, title, date, string(contentBytes))
	return content, nil
}
