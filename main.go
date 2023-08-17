package main

import (
	"fmt"
	"strconv"
)

// import "strconv"

func main() {
	// CLI
	Execute()
	video := readYaml(path)
	// Choices
	for {
		description := video.Description
		descriptionLength := len(description)
		if descriptionLength > 100 {
			descriptionLength = 100
		}
		choices := []string{
			fmt.Sprintf("Pick a subject (%s)", video.Subject),
			"Generate title",
			fmt.Sprintf("Modify title (%s)", video.Title),
			"Generate description",
			fmt.Sprintf("Modify description (%s)", fmt.Sprintf("%s...", video.Description[0:descriptionLength])),
			"Exit",
		}
		println()
		choice, _ := getChoice(choices, "What would you like to do?")
		err := error(nil)
		switch choice {
		case 0: // Subject
			video.Subject = getInput("What is the subject of the video?", "")
		case 1: // Generate title
			video, err = generateTitle(video)
		case 2: // Modify title
			video, err = modifyTitle(video)
		case 3: // Generate description
			video, err = generateDescription(video)
		case 4: // Modify description
			video, err = modifyDescription(video)
		case 5: // Exit
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
