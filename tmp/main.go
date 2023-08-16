// package main

// import (
// 	"bufio"
// 	"context"
// 	"fmt"
// 	"os"
// 	"strings"

// 	openai "github.com/sashabaranov/go-openai"
// )

// func main() {
// 	input, err := requestInput()
// 	if err != nil {
// 		fmt.Println("An error occured while reading input.")
// 		return
// 	}
// 	question := "Write title for a youtube video about " + input
// 	key := os.Getenv("OPENAI_KEY")
// 	client := openai.NewClient(key)
// 	resp, err := client.CreateChatCompletion(
// 		context.Background(),
// 		openai.ChatCompletionRequest{
// 			Model: openai.GPT3Dot5Turbo,
// 			Messages: []openai.ChatCompletionMessage{
// 				{
// 					Role:    openai.ChatMessageRoleUser,
// 					Content: question,
// 				},
// 			},
// 		},
// 	)
// 	if err != nil {
// 		fmt.Printf("ChatCompletion error: %v\n", err)
// 		return
// 	}
// 	fmt.Println(resp.Choices[0].Message.Content)
// }

// func requestInput() (string, error) {
// 	fmt.Print("Enter the subject of the video: ")
// 	reader := bufio.NewReader(os.Stdin)
// 	input, err := reader.ReadString('\n')
// 	if err != nil {
// 		return "", err
// 	}
// 	input = strings.TrimSuffix(input, "\n")
// 	return input, nil
// }
