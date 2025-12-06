package app

import (
	"errors"
	"fmt"

	"devopstoolkit/youtube-automation/internal/publishing"

	"github.com/charmbracelet/huh"
)

// HandleAMAMenu displays the Ask Me Anything submenu
func (m *MenuHandler) HandleAMAMenu() error {
	var selectedOption int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Ask Me Anything").
				Options(
					huh.NewOption("Fetch Transcript", 0),
					huh.NewOption("Back", actionReturn),
				).
				Value(&selectedOption),
		),
	)

	err := form.Run()
	if err != nil {
		return fmt.Errorf("failed to run AMA menu form: %w", err)
	}

	switch selectedOption {
	case 0:
		return m.HandleAMAFetchTranscript()
	case actionReturn:
		return nil
	}

	return nil
}

// HandleAMAFetchTranscript prompts for video ID and fetches the transcript
func (m *MenuHandler) HandleAMAFetchTranscript() error {
	var videoID string

	// Prompt for video ID
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("YouTube Video ID").
				Description("Enter the video ID (e.g., dQw4w9WgXcQ from https://youtu.be/dQw4w9WgXcQ)").
				Placeholder("video ID").
				Value(&videoID),
		),
	)

	err := form.Run()
	if err != nil {
		return fmt.Errorf("failed to run video ID form: %w", err)
	}

	if videoID == "" {
		fmt.Println(m.orangeStyle.Render("No video ID provided."))
		return nil
	}

	// Fetch transcript
	fmt.Println(m.normalStyle.Render("Fetching transcript from YouTube..."))

	transcript, err := publishing.GetTranscript(videoID)
	if err != nil {
		if errors.Is(err, publishing.ErrNoCaptions) {
			fmt.Println(m.errorStyle.Render("No captions available for this video."))
			fmt.Println(m.orangeStyle.Render("Tip: Make sure the video has auto-generated captions enabled in YouTube Studio."))
			return nil
		}
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch transcript: %v", err)))
		return nil
	}

	// Display success and transcript preview
	fmt.Println(m.greenStyle.Render("Transcript fetched successfully!"))
	fmt.Println("")
	fmt.Println(m.normalStyle.Render("--------------------------------------------"))
	fmt.Println(m.normalStyle.Render("Transcript Preview (first 2000 chars)"))
	fmt.Println(m.normalStyle.Render("--------------------------------------------"))
	fmt.Println("")

	// Show preview (first 2000 chars)
	preview := transcript
	if len(preview) > 2000 {
		preview = preview[:2000] + "\n... (truncated)"
	}
	fmt.Println(preview)

	fmt.Println("")
	fmt.Println(m.normalStyle.Render("--------------------------------------------"))
	fmt.Println(m.greenStyle.Render(fmt.Sprintf("Total transcript length: %d characters", len(transcript))))

	return nil
}
