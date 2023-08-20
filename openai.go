package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func askOpenAI(question string, iterations int) []string {
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

func generateDescription(video Video) (Video, error) {
	if len(video.Title) == 0 {
		return video, fmt.Errorf(redStyle.Render("Title was not generated!"))
	}
	aiQuestion := "Write a short description for a youtube video about " + video.Title
	descriptions := askOpenAI(aiQuestion, 5)
	println()
	choices := make(map[int]string)
	for i := range descriptions {
		println(strconv.Itoa(i) + ": " + descriptions[i])
		choices[i] = strconv.Itoa(i)
	}
	println()
	_, descriptionIndex := getChoice(choices, "Which description do you prefer?")
	descriptionIndexInt, _ := strconv.Atoi(descriptionIndex)
	video.Description = descriptions[descriptionIndexInt]
	println(video.Description)
	return video, nil
}

func generateTags(title string) (string, error) {
	if len(title) == 0 {
		return "", fmt.Errorf(redStyle.Render("Title was not generated!"))
	}
	aiQuestion := fmt.Sprintf("Write tags for youtube video about %s. Separate them with comma.", title)
	result := askOpenAI(aiQuestion, 1)
	println(result[0])
	return result[0], nil
}

func generateTitle(video Video) (Video, error) {
	if len(video.Subject) == 0 {
		return video, fmt.Errorf(redStyle.Render("Subject was not specified"))
	}
	aiQuestion := "Write up to 75 characters title for a youtube video about " + video.Subject
	results := askOpenAI(aiQuestion, 5)
	titlesMap := make(map[int]string)
	for index := range results {
		titlesMap[index] = results[index]
	}
	println()
	_, video.Title = getChoice(titlesMap, "Which video title do you prefer?")
	return video, nil
}

func generateTweet(title, videoId string) (string, error) {
	if len(title) == 0 {
		return "", fmt.Errorf(redStyle.Render("Title was not generated!"))
	}
	if len(videoId) == 0 {
		return "", fmt.Errorf(redStyle.Render("Video was NOT uploaded!"))
	}
	aiQuestion := fmt.Sprintf("Write a tweet for a youtube video about %s.", title)
	results := askOpenAI(aiQuestion, 5)
	resultsMap := make(map[int]string)
	for index := range results {
		resultsMap[index] = results[index]
	}
	println()
	_, tweet := getChoice(resultsMap, "Which tweet do you prefer?")
	tweet = fmt.Sprintf("%s\n\nhttps://youtu.be/%s", tweet, videoId)
	return tweet, nil
}
