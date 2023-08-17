package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func askOpenAI(question string) []string {
	key := os.Getenv("OPENAI_KEY")
	client := openai.NewClient(key)
	responses := make([]string, 0)
	for i := 1; i <= 5; i++ {
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
