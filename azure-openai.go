package main

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/charmbracelet/huh/spinner"
	"golang.org/x/net/context"
)

type AzureOpenAI struct {
	messages []azopenai.ChatRequestMessageClassification
}

func NewAIChat(systemMessage string) *AzureOpenAI {
	return &AzureOpenAI{
		messages: []azopenai.ChatRequestMessageClassification{
			&azopenai.ChatRequestSystemMessage{Content: to.Ptr(systemMessage)},
		},
	}
}

func NewAIChatYouTube() *AzureOpenAI {
	return NewAIChat("You are helping with YouTube videos.")
}

func (a *AzureOpenAI) Chat(newMessage string) (map[int32]string, error) {
	var resp azopenai.GetChatCompletionsResponse
	var err error
	azureOpenAIKey := os.Getenv("AZURE_OPENAI_KEY")
	modelDeploymentID := os.Getenv("AZURE_OPENAI_DEPLOYMENT")
	if len(modelDeploymentID) == 0 {
		modelDeploymentID = "gpt-4-1106-preview"
	}
	azureOpenAIEndpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	if azureOpenAIKey == "" || modelDeploymentID == "" || azureOpenAIEndpoint == "" {
		return nil, fmt.Errorf("Skipping example, environment variables missing")
	}
	keyCredential := azcore.NewKeyCredential(azureOpenAIKey)
	client, err := azopenai.NewClientWithKeyCredential(azureOpenAIEndpoint, keyCredential, nil)
	if err != nil {
		return nil, err
	}
	a.messages = append(a.messages, &azopenai.ChatRequestUserMessage{Content: azopenai.NewChatRequestUserMessageContent(newMessage)})
	action := func() {
		resp, err = client.GetChatCompletions(context.TODO(), azopenai.ChatCompletionsOptions{
			Messages:       a.messages,
			DeploymentName: &modelDeploymentID,
		}, nil)
	}
	spinner.New().Title("Contemplating how to destroy the world while trying to answer your question...").Action(action).Run()
	if err != nil {
		return nil, err
	}
	out := make(map[int32]string)
	for _, choice := range resp.Choices {
		out[*choice.Index] = *choice.Message.Content
		a.messages = append(a.messages, &azopenai.ChatRequestAssistantMessage{Content: choice.Message.Content})
	}
	return out, nil
}

func (a *AzureOpenAI) Close() {
	a.messages = nil
}
