package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

type AzureOpenAI struct{}

func (a *AzureOpenAI) Chat(newMessage string) (map[int32]string, error) {
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

	// This is a conversation in progress.
	// NOTE: all messages, regardless of role, count against token usage for this API.
	// messages := []azopenai.ChatMessage{
	messages := []azopenai.ChatRequestMessageClassification{
		// You set the tone and rules of the conversation with a prompt as the system role.
		// {Role: to.Ptr(azopenai.ChatRoleSystem), Content: to.Ptr("You are a helpful assistant.")},
		&azopenai.ChatRequestSystemMessage{Content: to.Ptr("You are a helpful assistant.")},

		// // The user asks a question
		// {Role: to.Ptr(azopenai.ChatRoleUser), Content: to.Ptr("Does Azure OpenAI support customer managed keys?")},
		&azopenai.ChatRequestUserMessage{Content: azopenai.NewChatRequestUserMessageContent(newMessage)},

		// // The reply would come back from the Azure OpenAI model. You'd add it to the conversation so we can maintain context.
		// {Role: to.Ptr(azopenai.ChatRoleAssistant), Content: to.Ptr("Yes, customer managed keys are supported by Azure OpenAI")},

		// // The user answers the question based on the latest reply.
		// {Role: to.Ptr(azopenai.ChatRoleUser), Content: to.Ptr("Do other Azure AI services support this too?")},

		// from here you'd keep iterating, sending responses back from the chat completions API
	}

	resp, err := client.GetChatCompletions(context.TODO(), azopenai.ChatCompletionsOptions{
		Messages:       messages,
		DeploymentName: &modelDeploymentID,
	}, nil)

	if err != nil {
		return nil, err
	}
	out := make(map[int32]string)
	for _, choice := range resp.Choices {
		out[*choice.Index] = *choice.Message.Content
	}
	return out, nil
}
