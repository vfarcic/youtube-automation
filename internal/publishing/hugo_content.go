package publishing

import (
	"regexp"
	"strings"
)

// ExtractIntro finds the "## Intro" section in a manuscript and returns the intro
// paragraphs and the remaining body without the intro section.
// If no "## Intro" is found, intro is empty and body is the full content.
func ExtractIntro(manuscript string) (intro, bodyWithoutIntro string) {
	lines := strings.Split(manuscript, "\n")
	introStart := -1
	introEnd := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if introStart == -1 {
			if trimmed == "## Intro" {
				introStart = i
			}
			continue
		}
		// We're inside the intro section; look for the next ## heading
		if strings.HasPrefix(trimmed, "## ") {
			introEnd = i
			break
		}
	}

	if introStart == -1 {
		return "", manuscript
	}

	// If no next heading found, intro goes to end of file
	if introEnd == -1 {
		introEnd = len(lines)
	}

	// Extract intro content (lines between "## Intro" and next heading)
	introLines := lines[introStart+1 : introEnd]
	intro = strings.TrimSpace(strings.Join(introLines, "\n"))

	// Build body without the intro section
	var bodyLines []string
	bodyLines = append(bodyLines, lines[:introStart]...)
	bodyLines = append(bodyLines, lines[introEnd:]...)
	bodyWithoutIntro = strings.Join(bodyLines, "\n")

	// Clean up leading/trailing whitespace in body
	bodyWithoutIntro = strings.TrimSpace(bodyWithoutIntro)

	return intro, bodyWithoutIntro
}

// RemoveTODOAndFIXMELines strips lines where the trimmed text starts with "TODO:" or "FIXME:".
func RemoveTODOAndFIXMELines(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "TODO:") || strings.HasPrefix(trimmed, "FIXME:") {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

var imageRefRegex = regexp.MustCompile(`!\[.*?\]\(([^)]+)\)`)

// ParseImageReferences extracts image filenames from markdown image references.
// It matches patterns like ![alt](filename) and returns the filenames.
func ParseImageReferences(content string) []string {
	matches := imageRefRegex.FindAllStringSubmatch(content, -1)
	var filenames []string
	for _, match := range matches {
		if len(match) >= 2 {
			ref := match[1]
			// Skip URLs (http/https)
			if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
				continue
			}
			filenames = append(filenames, ref)
		}
	}
	return filenames
}

func escapeTomlString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// BuildHugoPost assembles a Hugo blog post with proper frontmatter, intro as excerpt,
// YouTube shortcode, and body content.
func BuildHugoPost(title, date, videoId, intro, body string) string {
	youtubeShortcode := ""
	if videoId != "" {
		youtubeShortcode = "{{< youtube " + videoId + " >}}"
	}

	var sb strings.Builder
	sb.WriteString("\n+++\n")
	sb.WriteString("title = \"" + escapeTomlString(title) + "\"\n")
	sb.WriteString("date = " + date + ":00+00:00\n")
	sb.WriteString("draft = false\n")
	sb.WriteString("+++\n\n")

	if intro != "" {
		sb.WriteString(intro)
		sb.WriteString("\n")
	}

	sb.WriteString("\n<!--more-->\n\n")

	if youtubeShortcode != "" {
		sb.WriteString(youtubeShortcode)
		sb.WriteString("\n\n")
	}

	sb.WriteString(body)
	sb.WriteString("\n")

	return sb.String()
}
