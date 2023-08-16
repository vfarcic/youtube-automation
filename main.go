package main

import "strconv"

// import "strconv"

func main() {
	// Title
	question := getInput("What is the subject of the video?", "")
	aiQuestion := "Write up to 75 characters title for a youtube video about " + question
	titles := askOpenAI(aiQuestion)
	println()
	title := getChoice(titles)
	println()
	title = getInput("Rewrite the title:", title)

	// Description
	aiQuestion = "Write a short description for a youtube video about " + title
	descriptions := askOpenAI(aiQuestion)
	println()
	choices := []string{}
	for i := range descriptions {
		println(strconv.Itoa(i) + ": " + descriptions[i])
		choices = append(choices, strconv.Itoa(i))
	}
	println()
	descriptionIndex := getChoice(choices)
	descriptionIndexInt, _ := strconv.Atoi(descriptionIndex)
	description := descriptions[descriptionIndexInt]
	println(description)
	description = getTextArea(description)
	println()

	// Output
	println("Selected title: " + title)
	println()
	println("Selected description:\n" + description)
	println()
}
