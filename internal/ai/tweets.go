package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// newLLMClientFuncForTweets is a variable to allow mocking in tests.
var newLLMClientFuncForTweets = func(options ...openai.Option) (llms.Model, error) {
	return openai.New(options...)
}

// isAIConfigComplete is likely defined in titles.go or another file in this package.
// Removing the re-declaration from here.
// func isAIConfigComplete(config AITitleGeneratorConfig) bool {
// 	return config.Endpoint != "" &&
// 		config.DeploymentName != "" &&
// 		config.APIKey != "" &&
// 		config.APIVersion != ""
// }

// SuggestTweets generates 5 tweet suggestions based on the provided manuscript.
// Each tweet should be a maximum of 280 characters.
func SuggestTweets(ctx context.Context, manuscript string, config AITitleGeneratorConfig) ([]string, error) {
	if strings.TrimSpace(manuscript) == "" {
		return nil, errors.New("manuscript content is empty, cannot suggest tweets")
	}
	if !isAIConfigComplete(config) { // This will use the one from titles.go or similar
		return nil, errors.New("AI configuration is not fully set for tweets")
	}

	llm, err := newLLMClientFuncForTweets(
		openai.WithAPIType(openai.APITypeAzure),
		openai.WithToken(config.APIKey),
		openai.WithBaseURL(config.Endpoint),
		openai.WithModel(config.DeploymentName),
		openai.WithAPIVersion(config.APIVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LangChainGo client for tweets: %w", err)
	}

	// Simplified prompt string construction
	prompt := "Given the following manuscript, suggest exactly 5 relevant and engaging tweets to promote a video about it. "
	prompt += "Each tweet MUST be a maximum of 280 characters. "
	prompt += "Each tweet SHOULD be decorated with 1-3 relevant emojis. "
	prompt += "Each tweet SHOULD include 2-3 relevant hashtags. "
	prompt += "The placeholder `[YOUTUBE]` (which will be replaced with the video URL) MUST be included and appear on its own line (or two lines if needed for spacing) at the end of the tweet. "
	prompt += "Return the suggestions as a JSON array of strings. "
	prompt += "Do not add any other text, explanation, or formatting outside the JSON array. "
	prompt += "Example JSON response: [\"Excited to share my new video on X! 🤩 #X #NewVideo\n\n[YOUTUBE]\", \"Learn all about Y in my latest upload! 🚀 #Y #Tutorial\n\n[YOUTUBE]\"]\n\nManuscript:\n"
	prompt += manuscript

	var responseContent string
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		llmResponse, errGen := llms.GenerateFromSinglePrompt(ctx, llm, prompt, llms.WithTemperature(0.7))
		if errGen != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("error generating tweets after %d attempts: %w", maxRetries, errGen)
			}
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		responseContent = strings.TrimSpace(llmResponse)
		break
	}

	if responseContent == "" {
		return nil, errors.New("AI returned an empty response for tweets")
	}

	if strings.HasPrefix(responseContent, "```json") && strings.HasSuffix(responseContent, "```") {
		responseContent = strings.TrimPrefix(responseContent, "```json")
		responseContent = strings.TrimSuffix(responseContent, "```")
		responseContent = strings.TrimSpace(responseContent)
	} else if strings.HasPrefix(responseContent, "```") && strings.HasSuffix(responseContent, "```") {
		responseContent = strings.TrimPrefix(responseContent, "```")
		responseContent = strings.TrimSuffix(responseContent, "```")
		responseContent = strings.TrimSpace(responseContent)
	}

	var tweets []string
	err = json.Unmarshal([]byte(responseContent), &tweets)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response for tweets as JSON: %w. Response: %s", err, responseContent)
	}

	if len(tweets) == 0 {
		return nil, errors.New("AI returned an empty list of tweets")
	}

	for _, tweet := range tweets {
		if len(tweet) > 280 {
			return nil, fmt.Errorf("AI returned a tweet exceeding 280 characters. Tweet: '%s'", tweet)
		}
	}

	return tweets, nil
}

// Removed unused newChatMessage helper function
