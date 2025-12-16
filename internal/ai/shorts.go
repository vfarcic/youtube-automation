package ai

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"devopstoolkit/youtube-automation/internal/configuration"
)

//go:embed templates/shorts.md
var shortsTemplate string

// ShortCandidate represents a potential YouTube Short identified by AI analysis.
// The Rationale field is only used during selection and is not persisted.
type ShortCandidate struct {
	ID        string `json:"id"`        // Unique identifier (short1, short2, etc.)
	Title     string `json:"title"`     // Catchy title for the Short
	Text      string `json:"text"`      // Exact text segment from manuscript
	Rationale string `json:"rationale"` // Why this makes a good Short (display only)
}

// shortsTemplateData holds the data for the shorts analysis template
type shortsTemplateData struct {
	ManuscriptContent string
	MaxWords          int
	CandidateCount    int
}

// AnalyzeShortsFromManuscript analyzes a manuscript and returns Short candidates.
// It uses AI to identify self-contained, high-impact segments suitable for YouTube Shorts.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - manuscriptContent: The full manuscript text to analyze
//
// Returns:
//   - []ShortCandidate: Ordered list of candidates (best first)
//   - error: If analysis fails
func AnalyzeShortsFromManuscript(ctx context.Context, manuscriptContent string) ([]ShortCandidate, error) {
	if strings.TrimSpace(manuscriptContent) == "" {
		return nil, fmt.Errorf("manuscript content is empty")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Get configuration
	maxWords := configuration.GlobalSettings.Shorts.MaxWords
	candidateCount := configuration.GlobalSettings.Shorts.CandidateCount

	// Parse and execute template
	tmpl, err := template.New("shorts").Parse(shortsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse shorts template: %w", err)
	}

	data := shortsTemplateData{
		ManuscriptContent: manuscriptContent,
		MaxWords:          maxWords,
		CandidateCount:    candidateCount,
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute shorts template: %w", err)
	}

	// Generate content using the provider
	// Use higher token limit to accommodate multiple candidates with full text
	responseContent, err := provider.GenerateContent(ctx, promptBuf.String(), 4096)
	if err != nil {
		return nil, fmt.Errorf("AI shorts analysis failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return nil, fmt.Errorf("AI returned an empty response for shorts analysis")
	}

	// Parse JSON response
	var candidates []ShortCandidate
	if err := ParseJSONResponse(responseContent, &candidates); err != nil {
		return nil, fmt.Errorf("failed to parse shorts candidates: %w", err)
	}

	// Validate candidates
	if err := validateShortCandidates(candidates, maxWords); err != nil {
		return nil, err
	}

	return candidates, nil
}

// validateShortCandidates checks that candidates meet requirements
func validateShortCandidates(candidates []ShortCandidate, maxWords int) error {
	if len(candidates) == 0 {
		return fmt.Errorf("AI returned no short candidates")
	}

	for i, c := range candidates {
		if c.ID == "" {
			return fmt.Errorf("candidate %d has empty ID", i+1)
		}
		if c.Title == "" {
			return fmt.Errorf("candidate %d (%s) has empty title", i+1, c.ID)
		}
		if c.Text == "" {
			return fmt.Errorf("candidate %d (%s) has empty text", i+1, c.ID)
		}

		// Check word count
		wordCount := len(strings.Fields(c.Text))
		if wordCount > maxWords {
			return fmt.Errorf("candidate %s exceeds word limit: %d words (max %d)", c.ID, wordCount, maxWords)
		}
	}

	return nil
}

// CountWords returns the word count of a text segment
func CountWords(text string) int {
	return len(strings.Fields(text))
}
