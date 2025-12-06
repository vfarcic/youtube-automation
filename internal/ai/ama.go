package ai

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/ama-title.md
var amaTitleTemplate string

//go:embed templates/ama-timecodes.md
var amaTimecodesTemplate string

// AMAContent holds all generated content for an AMA video
type AMAContent struct {
	Title       string
	Timecodes   string
	Description string
	Tags        string
}

// amaTemplateData holds the data for AMA templates
type amaTemplateData struct {
	Transcript string
}

// GenerateAMATitle generates a title for an AMA video based on the transcript.
// It returns a single title string summarizing the main topics discussed.
func GenerateAMATitle(ctx context.Context, transcript string) (string, error) {
	if strings.TrimSpace(transcript) == "" {
		return "", fmt.Errorf("transcript is empty, cannot generate title")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	tmpl, err := template.New("ama-title").Parse(amaTitleTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse AMA title template: %w", err)
	}

	data := amaTemplateData{
		Transcript: transcript,
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute AMA title template: %w", err)
	}

	responseContent, err := provider.GenerateContent(ctx, promptBuf.String(), 100)
	if err != nil {
		return "", fmt.Errorf("AI AMA title generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("AI returned an empty response for AMA title")
	}

	// Clean up the response - remove quotes if present
	result := strings.TrimSpace(responseContent)
	result = strings.Trim(result, "\"'")

	return result, nil
}

// GenerateAMATimecodes generates timestamped Q&A segments from an AMA transcript.
// The first entry is always "00:00 Intro (skip to first question)" for the intro music/animation.
func GenerateAMATimecodes(ctx context.Context, transcript string) (string, error) {
	if strings.TrimSpace(transcript) == "" {
		return "", fmt.Errorf("transcript is empty, cannot generate timecodes")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	tmpl, err := template.New("ama-timecodes").Parse(amaTimecodesTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse AMA timecodes template: %w", err)
	}

	data := amaTemplateData{
		Transcript: transcript,
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute AMA timecodes template: %w", err)
	}

	responseContent, err := provider.GenerateContent(ctx, promptBuf.String(), 1500)
	if err != nil {
		return "", fmt.Errorf("AI AMA timecodes generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("AI returned an empty response for AMA timecodes")
	}

	return strings.TrimSpace(responseContent), nil
}

// GenerateAMADescription generates a description for an AMA video based on the transcript.
// It summarizes the key topics and questions discussed.
func GenerateAMADescription(ctx context.Context, transcript string) (string, error) {
	if strings.TrimSpace(transcript) == "" {
		return "", fmt.Errorf("transcript is empty, cannot generate description")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	prompt := fmt.Sprintf(
		`Based on the following AMA (Ask Me Anything) livestream transcript, generate a concise and engaging video description.

IMPORTANT: The host's name is Viktor (with a K), not Victor.

REQUIREMENTS:
- One or two paragraphs long
- Summarize the main topics and questions discussed
- Highlight key technologies, tools, or concepts mentioned
- Use clear, engaging language
- Do not include links, hashtags, or calls to action

Return only the description text, with no additional formatting or commentary.

TRANSCRIPT:
%s

DESCRIPTION:`,
		transcript,
	)

	responseContent, err := provider.GenerateContent(ctx, prompt, 400)
	if err != nil {
		return "", fmt.Errorf("AI AMA description generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("AI returned an empty response for AMA description")
	}

	return strings.TrimSpace(responseContent), nil
}

// GenerateAMATags generates tags for an AMA video based on the transcript.
// Tags are comma-separated and limited to 450 characters total.
func GenerateAMATags(ctx context.Context, transcript string) (string, error) {
	if strings.TrimSpace(transcript) == "" {
		return "", fmt.Errorf("transcript is empty, cannot generate tags")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	prompt := fmt.Sprintf(
		`Based on the following AMA (Ask Me Anything) livestream transcript, generate a comma-separated list of relevant tags.

IMPORTANT: The host's name is Viktor (with a K), not Victor.

REQUIREMENTS:
- The total length of the comma-separated string MUST NOT exceed 450 characters
- Focus on specific technologies, tools, concepts, and terms discussed
- Include "AMA", "Q&A", and "livestream" as base tags
- Include both specific terms (e.g., "Kubernetes", "ArgoCD") and broader categories (e.g., "DevOps", "cloud native")
- Order tags by relevance (most relevant first)

Return ONLY the comma-separated string of tags, without any additional explanation, preamble, or markdown formatting.

TRANSCRIPT:
%s

TAGS:`,
		transcript,
	)

	responseContent, err := provider.GenerateContent(ctx, prompt, 200)
	if err != nil {
		return "", fmt.Errorf("AI AMA tag generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("AI returned an empty response for AMA tags")
	}

	// Ensure the response does not exceed the character limit
	if len(responseContent) > 450 {
		if idx := strings.LastIndex(responseContent[:450], ","); idx != -1 {
			responseContent = responseContent[:idx]
		} else {
			responseContent = responseContent[:450]
		}
	}

	return strings.TrimSpace(responseContent), nil
}

// GenerateAMAContent generates all content (title, timecodes, description, tags) for an AMA video.
// This is a convenience function that calls all individual generation functions.
func GenerateAMAContent(ctx context.Context, transcript string) (*AMAContent, error) {
	if strings.TrimSpace(transcript) == "" {
		return nil, fmt.Errorf("transcript is empty, cannot generate content")
	}

	content := &AMAContent{}
	var err error

	content.Title, err = GenerateAMATitle(ctx, transcript)
	if err != nil {
		return nil, fmt.Errorf("failed to generate title: %w", err)
	}

	content.Timecodes, err = GenerateAMATimecodes(ctx, transcript)
	if err != nil {
		return nil, fmt.Errorf("failed to generate timecodes: %w", err)
	}

	content.Description, err = GenerateAMADescription(ctx, transcript)
	if err != nil {
		return nil, fmt.Errorf("failed to generate description: %w", err)
	}

	content.Tags, err = GenerateAMATags(ctx, transcript)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tags: %w", err)
	}

	return content, nil
}
