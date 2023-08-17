package main

import (
	"fmt"
	"strconv"
)

type Video struct {
	Subject     string
	Date        string
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

// - [ ] Date
// - [ ] Sponsored
// - [ ] Code
// - [ ] Record screen
// - [ ] Record face
// - [ ] Download thumbnails
// - [ ] Material uploaded
// - [ ] Product name
// - [ ] Product URL
// - [ ] Other logos
// - [ ] Tagline
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

func main() {
	// CLI
	Execute()
	video := readYaml(path)
	// Choices
	for {
		descriptionLength := len(video.Description)
		if descriptionLength > 100 {
			descriptionLength = 100
		}
		description := ""
		if len(video.Description) > 0 {
			description = fmt.Sprintf("%s...", video.Description[0:descriptionLength])
		}
		choices := []string{
			fmt.Sprintf("Pick a subject (%s)", video.Subject),
			fmt.Sprintf("Select publish date (%s)", video.Date),
			"Generate title",
			fmt.Sprintf("Modify title (%s)", video.Title),
			"Generate description",
			fmt.Sprintf("Modify description (%s)", description),
			fmt.Sprintf("Set files location (%s)", video.Location),
			"Exit",
		}
		println()
		choice, _ := getChoice(choices, "What would you like to do?")
		err := error(nil)
		switch choice {
		case 0: // Subject
			video.Subject = getInput("What is the subject of the video?", video.Subject)
		case 1: // Date
			video.Date = getInput("What is the publish of the video?", video.Date)
		case 2: // Generate title
			video, err = generateTitle(video)
		case 3: // Modify title
			video, err = modifyTitle(video)
		case 4: // Generate description
			video, err = generateDescription(video)
		case 5: // Modify description
			video, err = modifyDescription(video)
		case 6: // Location
			video.Location = getInput("Where are files located?", video.Location)
		case 7: // Exit
			return
		}
		if err != nil {
			println(fmt.Sprintf("\n%s", err.Error()))
			continue
		}
		writeYaml(video, path)
	}
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
	video.Title = getInput("Rewrite the title:", video.Title)
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
