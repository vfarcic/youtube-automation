package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

type Video struct {
	ProjectName string
	ProjectURL  string
	Sponsored   string
	Subject     string
	Date        string
	Code        bool
	Screen      bool
	Head        bool
	Thumbnails  bool
	Title       string
	Description string
	Location    string
}

// ## Thumbnail

// Elements:

// * Logo: TODO:
// * Text: TODO:
// * Screenshots: TODO:

// Ideas:

// * TODO:

// ## Animations

// * Animation: Subscribe (anywhere in the video)
// * Animation: Like (anywhere in the video)
// * Lower third: Viktor Farcic (anywhere in the video)
// * Animation: Join the channel (anywhere in the video)
// * Animation: Sponsor the channel (anywhere in the video)
// * Lower third: TODO: + logo + URL (TODO:) (add to a few places when I mention TODO:)
// * Title roll: TODO:
// * Text: Gist with the commands + an arrow pointing below
// * Thumbnails: (TODO:, TODO:) + text "The link is in the description" + an arrow pointing below
// * Logo: TODO:
// * Section: TODO:
// * Text: TODO:
// * Text: TODO: (big)
// * Plug: TODO: + logo + URL (TODO:) (use their website for animations or screenshots; make it look different from the main video; I'll let you know where to put it once the main video is ready)
// * Diagram: TODO:
// * Header: Cons; Items: TODO:
// * Header: Pros; Items: TODO:
// * Member shoutouts: Thanks a ton to the new members for supporting the channel: TODO: (ping me when you get to this part and I'll send you the latest list)
// * Outro roll

// ## Tasks

// - [ ] Tagline
// - [ ] Other logos
// - [ ] Screenshots
// - [ ] Thumbnail ideas
// - [ ] Animation bullets
// - [ ] Create movie
// - [ ] Title
// - [ ] Description
// - [ ] Diagrams
// - [ ] Sections
// - [ ] Pros
// - [ ] Cons
// - [ ] Members
// - [ ] Proofread
// - [ ] Gist
// - [ ] Sponsor message
// - [ ] Timecodes
// - [ ] Corrections
// - [ ] Tags
// - [ ] Comment
// - [ ] Twitter written
// - [ ] Publish on YouTube
// - [ ] Publish on Twitter
// - [ ] Publish on LinkedIn
// - [ ] Publish on Slack
// - [ ] Publish on Reddit
// - [ ] Hacker News
// - [ ] Publish on TechnologyConversations.com
// - [ ] Add to YT spotlight
// - [ ] Add a comment to the video
// - [ ] Respond to comments
// - [ ] Add to slides
// - [ ] Add to https://gde.advocu.com
// - [ ] Modify repo README.md
// - [ ] Publish on a Twitter space
// - [ ] Convert to Crossplane
// - [ ] Email
// - [ ] Top X?
// - [ ] Add to cncf-demo

// ## Sponsored Message

// TODO:

// ## YouTube Timecodes

// TODO:

// ## Corrections

// N/A

// ## Corrections (Miki)

// N/A

// ## Twitter

// TODO:

// @DevOpsToolkit

// ## Description

// TODO:

// ## Tags

// TODO:

// ## Comment

// TODO:

var redStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("1"))

var greenStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("2"))

func main() {
	// CLI
	Execute()
	video := readYaml(path)
	// Choices
	for {
		choices := []string{
			getChoiceTextFromString("Set project name", video.ProjectName),
			getChoiceTextFromString("Set project URL", video.ProjectURL),
			getChoiceTextFromString("Set sponsorship", video.Sponsored),
			getChoiceTextFromString("Set the subject", video.Subject),
			getChoiceTextFromString("Set publish date", video.Date),
			getChoiceTextFromBool("Wrote code?", video.Code),
			getChoiceTextFromBool("Recorded screen?", video.Screen),
			getChoiceTextFromBool("Recorded talking head?", video.Head),
			getChoiceTextFromBool("Downloaded thumbnails?", video.Thumbnails),
			getChoiceTextFromString("Set files location", video.Location),
			getChoiceTextFromString("Generate title", video.Title),
			getChoiceTextFromString("Write/Modify title", video.Title),
			getChoiceTextFromString("Generate description", video.Description),
			getChoiceTextFromString("Write/Modify description", video.Description),
			"Exit",
		}
		println()
		choice, _ := getChoice(choices, "What would you like to do?")
		err := error(nil)
		switch choice {
		case 0: // Project name
			video.ProjectName, err = getInputFromString("Set project name)", video.ProjectName)
		case 1: // Project URL
			video.ProjectURL, err = getInputFromString("Set project URL", video.ProjectURL)
		case 2: // Sponsored
			video.Sponsored, err = getInputFromString("Sponsorship amount ('-' or 'N/A' if not sponsored)", video.Sponsored)
		case 3: // Subject
			video.Subject, err = getInputFromString("What is the subject of the video?", video.Subject)
		case 4: // Date
			video.Date, err = getInputFromString("What is the publish of the video?", video.Date)
		case 5: // Code
			video.Code = getInputFromBool(video.Code)
		case 6: // Screen
			video.Screen = getInputFromBool(video.Screen)
		case 7: // Head
			video.Head = getInputFromBool(video.Head)
		case 8: // Thumbnails
			video.Thumbnails = getInputFromBool(video.Thumbnails)
		case 9: // Location
			video.Location, err = getInputFromString("Where are files located?", video.Location)
		case 10: // Generate title
			video, err = generateTitle(video)
		case 11: // Modify title
			video, err = modifyTitle(video)
		case 12: // Generate description
			video, err = generateDescription(video)
		case 13: // Modify description
			video, err = modifyDescription(video)
		case 14: // Exit
			return
		}
		if err != nil {
			println(fmt.Sprintf("\n%s", err.Error()))
			continue
		}
		writeYaml(video, path)
	}
}

func getChoiceTextFromString(choice, value string) string {
	valueLength := len(value)
	if valueLength > 100 {
		value = fmt.Sprintf("%s...", value[0:100])
	}
	text := choice
	if value != "" && value != "-" && value != "N/A" {
		text = fmt.Sprintf("%s (%s)", text, value)
	}
	if value == "" {
		return redStyle.Render(text)
	}
	return greenStyle.Render(text)
}

func getChoiceTextFromBool(choice string, value bool) string {
	text := fmt.Sprintf("%s (%t)", choice, value)
	if !value {
		return redStyle.Render(text)
	}
	return greenStyle.Render(text)
}

func generateTitle(video Video) (Video, error) {
	if len(video.Subject) == 0 {
		return video, fmt.Errorf("subject was not specified")
	}
	aiQuestion := "Write up to 75 characters title for a youtube video about " + video.Subject
	titles := askOpenAI(aiQuestion)
	println()
	_, video.Title = getChoice(titles, "Which video title do you prefer?")
	return video, nil
}

func modifyTitle(video Video) (Video, error) {
	if len(video.Subject) == 0 {
		return video, fmt.Errorf("title was not specified")
	}
	println()
	video.Title, _ = getInputFromString("Rewrite the title:", video.Title)
	return video, nil
}

func generateDescription(video Video) (Video, error) {
	if len(video.Title) == 0 {
		return video, fmt.Errorf("title was not generated")
	}
	aiQuestion := "Write a short description for a youtube video about " + video.Title
	descriptions := askOpenAI(aiQuestion)
	println()
	choices := []string{}
	for i := range descriptions {
		println(strconv.Itoa(i) + ": " + descriptions[i])
		choices = append(choices, strconv.Itoa(i))
	}
	println()
	_, descriptionIndex := getChoice(choices, "Which description do you prefer?")
	descriptionIndexInt, _ := strconv.Atoi(descriptionIndex)
	video.Description = descriptions[descriptionIndexInt]
	println(video.Description)
	return video, nil
}

func modifyDescription(video Video) (Video, error) {
	if len(video.Description) == 0 {
		return video, fmt.Errorf("description was not generated")
	}
	println()
	video.Description = getTextArea(video.Description)
	return video, nil
}
