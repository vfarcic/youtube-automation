package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/manuscript"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/charmbracelet/huh"
)

// HandleAnalyzeShorts analyzes a video's manuscript for YouTube Shorts candidates,
// lets the user select which ones to keep, and inserts TODO markers into the manuscript.
// Returns the selected shorts with markers inserted in the manuscript file.
func (m *MenuHandler) HandleAnalyzeShorts(video *storage.Video) ([]storage.Short, error) {
	// Check if manuscript path exists
	if video.Gist == "" {
		return nil, fmt.Errorf("manuscript path (Gist) is not set for this video")
	}

	// Read manuscript content
	manuscriptContent, err := os.ReadFile(video.Gist)
	if err != nil {
		return nil, fmt.Errorf("failed to read manuscript from %s: %w", video.Gist, err)
	}

	if strings.TrimSpace(string(manuscriptContent)) == "" {
		return nil, fmt.Errorf("manuscript file is empty: %s", video.Gist)
	}

	fmt.Println(m.normalStyle.Render("Analyzing manuscript for YouTube Shorts candidates..."))
	fmt.Println(m.normalStyle.Render("This may take a moment."))

	// Call AI to analyze manuscript
	ctx := context.Background()
	candidates, err := ai.AnalyzeShortsFromManuscript(ctx, string(manuscriptContent))
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	fmt.Println(m.greenStyle.Render(fmt.Sprintf("✓ Found %d Short candidates", len(candidates))))
	fmt.Println("")

	// Display candidates and let user select
	selectedShorts, err := m.displayAndSelectShortCandidates(candidates)
	if err != nil {
		return nil, err
	}

	// Insert TODO markers into manuscript for selected shorts
	if len(selectedShorts) > 0 {
		fmt.Println(m.normalStyle.Render("Inserting TODO markers into manuscript..."))
		if markerErr := manuscript.InsertShortMarkers(video.Gist, selectedShorts); markerErr != nil {
			// Check if it's a partial success (some markers inserted but not all found)
			if strings.Contains(markerErr.Error(), "markers inserted") {
				fmt.Println(m.orangeStyle.Render(fmt.Sprintf("⚠️  %s", markerErr.Error())))
			} else {
				fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to insert markers: %v", markerErr)))
				// Continue anyway - shorts are still selected, just without markers
			}
		} else {
			fmt.Println(m.greenStyle.Render(fmt.Sprintf("✓ TODO markers inserted in %s", video.Gist)))
		}
	}

	return selectedShorts, nil
}

// displayAndSelectShortCandidates shows candidates and returns user-selected shorts
func (m *MenuHandler) displayAndSelectShortCandidates(candidates []ai.ShortCandidate) ([]storage.Short, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates to display")
	}

	// Display all candidates with details
	fmt.Println(m.normalStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(m.normalStyle.Render("📹 YouTube Shorts Candidates"))
	fmt.Println(m.normalStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println("")

	for i, c := range candidates {
		wordCount := ai.CountWords(c.Text)
		fmt.Println(m.greenStyle.Render(fmt.Sprintf("%d. %s", i+1, c.Title)))
		fmt.Println(m.normalStyle.Render(fmt.Sprintf("   ID: %s | Words: %d", c.ID, wordCount)))
		fmt.Println(m.normalStyle.Render(fmt.Sprintf("   Why: %s", c.Rationale)))
		fmt.Println("")
		// Show truncated text preview
		textPreview := c.Text
		if len(textPreview) > 200 {
			textPreview = textPreview[:200] + "..."
		}
		fmt.Println(m.normalStyle.Render(fmt.Sprintf("   \"%s\"", textPreview)))
		fmt.Println("")
		fmt.Println(m.normalStyle.Render("   ──────────────────────────────────────────────────────"))
		fmt.Println("")
	}

	// Build options for multi-select
	options := make([]huh.Option[string], len(candidates))
	for i, c := range candidates {
		label := fmt.Sprintf("%s (%d words)", c.Title, ai.CountWords(c.Text))
		options[i] = huh.NewOption(label, c.ID)
	}

	// Multi-select form
	var selectedIDs []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select which candidates to keep as Shorts").
				Description("Use space to select, enter to confirm").
				Options(options...).
				Value(&selectedIDs),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selectedIDs) == 0 {
		fmt.Println(m.orangeStyle.Render("No candidates selected."))
		return nil, nil
	}

	// Convert selected candidates to storage.Short structs
	selectedShorts := make([]storage.Short, 0, len(selectedIDs))
	for _, id := range selectedIDs {
		for _, c := range candidates {
			if c.ID == id {
				selectedShorts = append(selectedShorts, storage.Short{
					ID:    c.ID,
					Title: c.Title,
					Text:  c.Text,
					// ScheduledDate and YouTubeID will be set later
				})
				break
			}
		}
	}

	fmt.Println(m.greenStyle.Render(fmt.Sprintf("✓ Selected %d Shorts", len(selectedShorts))))

	return selectedShorts, nil
}
