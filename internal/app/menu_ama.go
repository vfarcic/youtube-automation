package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/charmbracelet/huh"
)

const (
	amaActionGenerate = iota
	amaActionApply
	amaActionBack
)

// HandleAMAMenu handles the AMA video enhancement workflow
func (m *MenuHandler) HandleAMAMenu() error {
	var videoID string
	var title string
	var description string
	var tags string
	var timecodes string
	var action int

	// Store transcript and publish date for saving
	var transcript string
	var publishedAt string

	for {
		// Build form with all fields on one screen
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("YouTube Video ID").
					Description("e.g., dQw4w9WgXcQ from https://youtu.be/dQw4w9WgXcQ").
					Placeholder("video ID").
					Value(&videoID),
				huh.NewInput().
					Title("Title").
					Description("Video title (will replace current title)").
					Value(&title),
				huh.NewText().
					Title("Description").
					Description("Video description (replaces content before boilerplate)").
					Lines(5).
					CharLimit(5000).
					Value(&description),
				huh.NewText().
					Title("Tags").
					Description("Comma-separated tags (max 450 characters)").
					Lines(2).
					CharLimit(450).
					Value(&tags),
				huh.NewText().
					Title("Timecodes").
					Description("Timestamped Q&A segments (appended to description)").
					Lines(10).
					CharLimit(5000).
					Value(&timecodes),
				huh.NewSelect[int]().
					Title("Action").
					Options(
						huh.NewOption("Generate with AI", amaActionGenerate),
						huh.NewOption("Apply to YouTube", amaActionApply),
						huh.NewOption("Back", amaActionBack),
					).
					Value(&action),
			),
		)

		err := form.Run()
		if err != nil {
			return fmt.Errorf("failed to run AMA form: %w", err)
		}

		switch action {
		case amaActionGenerate:
			if videoID == "" {
				fmt.Println(m.orangeStyle.Render("Please enter a Video ID first."))
				continue
			}

			fmt.Println(m.normalStyle.Render("Fetching video metadata..."))

			// Fetch video metadata for publish date
			metadata, err := publishing.GetVideoMetadata(videoID)
			if err != nil {
				fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch video metadata: %v", err)))
				continue
			}
			publishedAt = metadata.PublishedAt

			fmt.Println(m.normalStyle.Render("Fetching transcript from YouTube..."))

			transcript, err = publishing.GetTranscript(videoID)
			if err != nil {
				if errors.Is(err, publishing.ErrNoCaptions) {
					fmt.Println(m.errorStyle.Render("No captions available for this video."))
					fmt.Println(m.orangeStyle.Render("Tip: Make sure the video has auto-generated captions enabled in YouTube Studio."))
				} else {
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch transcript: %v", err)))
				}
				continue
			}

			fmt.Println(m.greenStyle.Render("Transcript fetched successfully!"))
			fmt.Println(m.normalStyle.Render("Generating content with AI..."))

			ctx := context.Background()
			content, err := ai.GenerateAMAContent(ctx, transcript)
			if err != nil {
				fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to generate content: %v", err)))
				continue
			}

			// Populate fields with generated content
			title = content.Title
			description = content.Description
			tags = content.Tags
			timecodes = content.Timecodes

			fmt.Println(m.greenStyle.Render("Content generated successfully!"))
			fmt.Println(m.normalStyle.Render("Review and edit the fields above, then select 'Apply to YouTube' when ready."))

		case amaActionApply:
			if videoID == "" {
				fmt.Println(m.orangeStyle.Render("Please enter a Video ID first."))
				continue
			}

			if title == "" && description == "" && tags == "" && timecodes == "" {
				fmt.Println(m.orangeStyle.Render("No content to apply. Generate or enter content first."))
				continue
			}

			fmt.Println(m.normalStyle.Render("Applying changes to YouTube..."))

			err := publishing.UpdateAMAVideo(videoID, title, description, tags, timecodes)
			if err != nil {
				fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to update video: %v", err)))
				continue
			}

			fmt.Println(m.greenStyle.Render("Video updated successfully!"))

			// Save to local files
			fmt.Println(m.normalStyle.Render("Saving to local files..."))
			if saveErr := m.saveAMAFiles(videoID, title, description, tags, timecodes, publishedAt, transcript); saveErr != nil {
				fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Warning: Failed to save local files: %v", saveErr)))
			} else {
				fmt.Println(m.greenStyle.Render("Local files saved to manuscript/ama/"))
			}

			return nil

		case amaActionBack:
			return nil
		}
	}
}

// saveAMAFiles saves AMA video data to YAML and MD files in manuscript/ama/
func (m *MenuHandler) saveAMAFiles(videoID, title, description, tags, timecodes, publishedAt, transcript string) error {
	const amaDir = "manuscript/ama"

	// Create directory if needed
	if err := os.MkdirAll(amaDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Build file name: YYYY-MM-DD-videoID
	datePrefix := extractDateFromISO(publishedAt)
	baseName := fmt.Sprintf("%s-%s", datePrefix, videoID)
	yamlPath := filepath.Join(amaDir, baseName+".yaml")
	mdPath := filepath.Join(amaDir, baseName+".md")

	// Create Video struct
	video := storage.Video{
		Name:        baseName,
		Path:        yamlPath,
		Category:    "ama",
		VideoId:     videoID,
		Title:       title,
		Description: description,
		Tags:        tags,
		Timecodes:   timecodes,
		Date:        publishedAt,
		Gist:        mdPath,
	}

	// Save YAML using existing storage
	yaml := storage.NewYAML("")
	if err := yaml.WriteVideo(video, yamlPath); err != nil {
		return fmt.Errorf("failed to save YAML: %w", err)
	}

	// Save transcript as MD file
	mdContent := fmt.Sprintf("# %s\n\nVideo ID: %s\n\n## Transcript\n\n%s", title, videoID, transcript)
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("failed to save transcript: %w", err)
	}

	return nil
}

// extractDateFromISO extracts YYYY-MM-DD from an ISO 8601 timestamp
func extractDateFromISO(isoDate string) string {
	if isoDate == "" {
		return time.Now().UTC().Format("2006-01-02")
	}

	// Try to parse and extract date
	if t, err := time.Parse(time.RFC3339, isoDate); err == nil {
		return t.Format("2006-01-02")
	}

	// Fallback: try to extract first 10 chars if it matches YYYY-MM-DD pattern
	if len(isoDate) >= 10 {
		dateStr := isoDate[:10]
		if _, err := time.Parse("2006-01-02", dateStr); err == nil {
			return dateStr
		}
	}

	return time.Now().UTC().Format("2006-01-02")
}
