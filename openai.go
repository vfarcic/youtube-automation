package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAI struct{}

func (o *OpenAI) Ask(question string, iterations int) []string {
	key := os.Getenv("OPENAI_KEY")
	client := openai.NewClient(key)
	responses := make([]string, 0)
	for i := 1; i <= iterations; i++ {
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT3Dot5Turbo,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: question,
					},
				},
			},
		)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return []string{}
		}
		respString := strings.ReplaceAll(resp.Choices[0].Message.Content, "\"", "")
		responses = append(responses, respString)
	}
	return responses
}

func (o *OpenAI) GenerateDescription(video Video) (Video, error) {
	if len(video.Title) == 0 {
		return video, fmt.Errorf(redStyle.Render("Title was not generated!"))
	}
	aiQuestion := "Write a short description of up to 300 characters for a youtube video about " + video.Title
	descriptions := o.Ask(aiQuestion, 5)
	choices := make(map[int]Task)
	for i := range descriptions {
		confirmationMessage += strconv.Itoa(i) + ": " + descriptions[i] + "\n"
		choices[i] = Task{Title: strconv.Itoa(i)}
	}
	_, descriptionIndex := getChoice(choices, "Which description do you prefer?")
	descriptionIndexInt, _ := strconv.Atoi(descriptionIndex)
	video.Description = descriptions[descriptionIndexInt]
	confirmationMessage = video.Description
	return video, nil
}

func (o *OpenAI) GenerateTags(title string) (string, error) {
	if len(title) == 0 {
		return "", fmt.Errorf(redStyle.Render("Title was not generated!"))
	}
	aiQuestion := fmt.Sprintf("Write tags for youtube video about %s. Separate them with comma.", title)
	result := o.Ask(aiQuestion, 1)
	confirmationMessage = result[0]
	return result[0], nil
}

func (o *OpenAI) GenerateTitle(video Video) (Video, error) {
	if len(video.Subject) == 0 {
		return video, fmt.Errorf(redStyle.Render("Subject was not specified"))
	}
	aiQuestion := "Write up to 75 characters title for a youtube video about " + video.Subject
	results := o.Ask(aiQuestion, 5)
	titlesMap := make(map[int]Task)
	for index := range results {
		titlesMap[index] = Task{Title: results[index]}
	}
	_, video.Title = getChoice(titlesMap, "Which video title do you prefer?")
	return video, nil
}

func (o *OpenAI) GenerateTweet(title, videoId string) (string, error) {
	if len(title) == 0 {
		return "", fmt.Errorf(redStyle.Render("Title was not generated!"))
	}
	if len(videoId) == 0 {
		return "", fmt.Errorf(redStyle.Render("Video was NOT uploaded!"))
	}
	aiQuestion := fmt.Sprintf("Write a tweet for a youtube video about %s.", title)
	results := o.Ask(aiQuestion, 5)
	resultsMap := make(map[int]Task)
	for index := range results {
		resultsMap[index] = Task{Title: results[index]}
	}
	_, tweet := getChoice(resultsMap, titleStyle.Render("Which tweet do you prefer?"))
	tweet = fmt.Sprintf("%s\n\n%s", tweet, getYouTubeURL(videoId))
	return tweet, nil
}
