package main

import (
	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/charmbracelet/huh/spinner"
	"golang.org/x/net/context"
)

type AzureOpenAI struct {
	messages   []azopenai.ChatRequestMessageClassification
	key        string
	endpoint   string
	deployment string
}

func NewAIChat(systemMessage, endpoint, key, deployment string) *AzureOpenAI {
	return &AzureOpenAI{
		messages: []azopenai.ChatRequestMessageClassification{
			&azopenai.ChatRequestSystemMessage{Content: to.Ptr(systemMessage)},
		},
		key:        key,
		endpoint:   endpoint,
		deployment: deployment,
	}
}

func NewAIChatYouTube(endpoint, key, deployment string) *AzureOpenAI {
	return NewAIChat("You are helping with YouTube videos.", endpoint, key, deployment)
}

func (a *AzureOpenAI) Chat(newMessage string) (map[int32]string, error) {
	var resp azopenai.GetChatCompletionsResponse
	keyCredential := azcore.NewKeyCredential(a.key)
	client, err := azopenai.NewClientWithKeyCredential(
		a.endpoint,
		keyCredential,
		nil,
	)
	if err != nil {
		return nil, err
	}
	a.messages = append(
		a.messages,
		&azopenai.ChatRequestUserMessage{
			Content: azopenai.NewChatRequestUserMessageContent(newMessage),
		},
	)
	action := func() {
		resp, err = client.GetChatCompletions(
			context.TODO(),
			azopenai.ChatCompletionsOptions{
				Messages:       a.messages,
				DeploymentName: &a.deployment,
			}, nil)
	}
	spinner.New().
		Title("Contemplating how to destroy the world while trying to answer your question...").
		Action(action).
		Run()
	if err != nil {
		return nil, err
	}
	out := make(map[int32]string)
	for _, choice := range resp.Choices {
		out[*choice.Index] = *choice.Message.Content
		a.messages = append(
			a.messages,
			&azopenai.ChatRequestAssistantMessage{
				Content: choice.Message.Content,
			},
		)
	}
	return out, nil
}

func (a *AzureOpenAI) Close() {
	a.messages = nil
}
