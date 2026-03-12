package publishing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const latestPostsHeader = "# Latest Posts"
const entrySeparator = "---"

// AddHomepageEntry adds a new entry to the Hugo home page (content/_index.md)
// after the "# Latest Posts" header. The entry uses the established format with
// thumbnail, title link, intro text, and "Full article >>" link.
func AddHomepageEntry(basePath, category, slug, title, intro string) error {
	indexPath := filepath.Join(basePath, "content", "_index.md")

	content, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("reading home page: %w", err)
	}

	postPath := "/" + category + "/" + slug
	entry := buildHomepageEntry(postPath, title, intro)

	lines := strings.Split(string(content), "\n")
	var result []string
	inserted := false
	skipNext := false

	for i, line := range lines {
		if skipNext {
			skipNext = false
			continue
		}
		result = append(result, line)
		if !inserted && strings.TrimSpace(line) == latestPostsHeader {
			// Insert blank line + entry after the header
			result = append(result, "")
			result = append(result, entry)
			inserted = true
			// Skip an existing blank line after header to avoid duplication
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" {
				skipNext = true
			}
		}
	}

	if !inserted {
		return fmt.Errorf("could not find %q header in %s", latestPostsHeader, indexPath)
	}

	return os.WriteFile(indexPath, []byte(strings.Join(result, "\n")), 0644)
}

// TrimHomepageEntries keeps only the first maxEntries entries in the home page.
// Entries are delimited by "---" separators.
func TrimHomepageEntries(basePath string, maxEntries int) error {
	indexPath := filepath.Join(basePath, "content", "_index.md")

	content, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("reading home page: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the "# Latest Posts" header
	headerIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == latestPostsHeader {
			headerIdx = i
			break
		}
	}
	if headerIdx == -1 {
		return nil // No header found, nothing to trim
	}

	// Everything before and including the header (plus following blank line)
	preambleEnd := headerIdx + 1
	if preambleEnd < len(lines) && strings.TrimSpace(lines[preambleEnd]) == "" {
		preambleEnd++
	}
	preamble := lines[:preambleEnd]
	rest := lines[preambleEnd:]

	// Split the rest into entries by "---" separator
	var entries []string
	var current []string
	for _, line := range rest {
		if strings.TrimSpace(line) == entrySeparator {
			if len(current) > 0 {
				entries = append(entries, strings.Join(current, "\n"))
			}
			current = nil
			continue
		}
		current = append(current, line)
	}
	// Don't add trailing content as an entry (it's likely empty lines or footer)
	trailing := ""
	if len(current) > 0 {
		trailing = strings.Join(current, "\n")
	}

	// Trim to maxEntries
	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	// Rebuild the file
	var result []string
	result = append(result, preamble...)
	for _, entry := range entries {
		result = append(result, entry)
		result = append(result, entrySeparator)
		result = append(result, "")
	}
	if trailing != "" && strings.TrimSpace(trailing) != "" {
		result = append(result, trailing)
	}

	return os.WriteFile(indexPath, []byte(strings.Join(result, "\n")), 0644)
}

func buildHomepageEntry(postPath, title, intro string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<a href="%s"><img src="%s/thumbnail.jpg" style="width:50%%; float:right; padding: 10px"></a>`, postPath, postPath))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("## [%s](%s)", title, postPath))
	sb.WriteString("\n\n")
	if intro != "" {
		sb.WriteString(intro)
		sb.WriteString("\n\n")
	}
	sb.WriteString(fmt.Sprintf("**[Full article >>](%s)**", postPath))
	sb.WriteString("\n\n")
	sb.WriteString(entrySeparator)
	sb.WriteString("\n")
	return sb.String()
}
