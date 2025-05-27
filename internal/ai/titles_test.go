package ai

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestSuggestTitles(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video title generation. It needs multiple suggestions."
	validAIConfig := AITitleGeneratorConfig{
		Endpoint:       "https://fake-titles-endpoint.openai.azure.com/",
		DeploymentName: "fake-titles-deployment",
		APIKey:         "fake-titles-api-key",
		APIVersion:     "2023-07-01-preview",
	}

	originalNewLLMTitles := newLLMClientFuncForTitles
	defer func() { newLLMClientFuncForTitles = originalNewLLMTitles }()

	tests := []struct {
		name              string
		mockLLMResponse   string
		mockLLMError      error
		aiConfig          AITitleGeneratorConfig
		manuscript        string
		wantTitles        []string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:            "Successful titles suggestion - direct array",
			mockLLMResponse: `["AI Title 1", "Catchy Title 2"]`,
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTitles:      []string{"AI Title 1", "Catchy Title 2"},
			wantErr:         false,
		},
		{
			name:            "Successful titles suggestion - with JSON code fence", // Assuming titles.go also strips fences
			mockLLMResponse: "```json\n[\"Fenced Title A\"]\n```",
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTitles:      []string{"Fenced Title A"},
			wantErr:         false,
		},
		{
			name:              "AI returns empty response for titles",
			mockLLMResponse:   "",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for titles",
		},
		{
			name:              "AI returns malformed JSON for titles",
			mockLLMResponse:   `["Unterminated title array]`,
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "failed to parse JSON response",
		},
		{
			name:              "AI generation error for titles",
			mockLLMError:      fmt.Errorf("AI titles service unavailable"),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI titles service unavailable",
		},
		{
			name:              "Incomplete AI config for titles - no endpoint",
			aiConfig:          AITitleGeneratorConfig{DeploymentName: "d", APIKey: "k", APIVersion: "v"},
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "Azure OpenAI Endpoint is not configured",
		},
		{
			name:              "LLM creation fails for titles",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			mockLLMError:      fmt.Errorf("mock llm creation failed for titles"),
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "mock llm creation failed for titles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLLM{ // mockLLM is from highlights_test.go (same package)
				ResponseContent: tt.mockLLMResponse,
				ErrorToReturn:   tt.mockLLMError,
			}

			newLLMClientFuncForTitles = func(options ...openai.Option) (llms.Model, error) {
				if strings.Contains(tt.name, "LLM creation fails") {
					return nil, tt.mockLLMError
				}
				return mock, nil
			}

			gotTitles, err := SuggestTitles(ctx, tt.manuscript, tt.aiConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestTitles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				// Check if the underlying error is what we expect, even if wrapped by the retry logic
				if tt.mockLLMError != nil { // If we are testing an LLM error scenario
					if !strings.Contains(err.Error(), tt.mockLLMError.Error()) {
						t.Errorf("SuggestTitles() error = %q, want to contain underlying error %q", err.Error(), tt.mockLLMError.Error())
					}
				} else if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					// For non-LLM errors (e.g., config errors, empty/malformed JSON response after successful LLM call)
					t.Errorf("SuggestTitles() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("SuggestTitles() unexpected error = %v", err)
					return
				}
			}

			if !reflect.DeepEqual(gotTitles, tt.wantTitles) {
				t.Errorf("SuggestTitles() gotTitles = %v, want %v", gotTitles, tt.wantTitles)
			}

			// Check LLM call count
			if !tt.wantErr && tt.aiConfig.Endpoint != "" && tt.aiConfig.APIKey != "" && !strings.Contains(tt.name, "LLM creation fails") {
				if mock.CallCount == 0 {
					t.Errorf("Expected LLM to be called for SuggestTitles, but CallCount is 0")
				}
			}
		})
	}
}

func TestGetAIConfig(t *testing.T) {
	// Backup and restore original global settings and env vars
	originalSettings := configuration.GlobalSettings
	originalAIKeyEnv := os.Getenv("AI_KEY")

	defer func() {
		configuration.GlobalSettings = originalSettings
		os.Setenv("AI_KEY", originalAIKeyEnv)
	}()

	tests := []struct {
		name              string
		setupFunc         func() // For setting up GlobalSettings and Env vars
		wantConfig        AITitleGeneratorConfig
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name: "All settings from GlobalSettings",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Key:        "global_key",
						Endpoint:   "global_endpoint",
						Deployment: "global_deployment",
						APIVersion: "global_version",
					},
				}
				os.Unsetenv("AI_KEY")
			},
			wantConfig: AITitleGeneratorConfig{
				APIKey:         "global_key",
				Endpoint:       "global_endpoint",
				DeploymentName: "global_deployment",
				APIVersion:     "global_version",
			},
			wantErr: false,
		},
		{
			name: "AI_KEY from env overrides GlobalSettings.AI.Key",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Key:        "global_key_ignored",
						Endpoint:   "endpoint",
						Deployment: "deployment",
						APIVersion: "version",
					},
				}
				os.Setenv("AI_KEY", "env_key")
			},
			wantConfig: AITitleGeneratorConfig{
				APIKey:         "env_key",
				Endpoint:       "endpoint",
				DeploymentName: "deployment",
				APIVersion:     "version",
			},
			wantErr: false,
		},
		{
			name: "AI_KEY from env when GlobalSettings.AI.Key is empty",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Key:        "",
						Endpoint:   "endpoint",
						Deployment: "deployment",
						APIVersion: "version",
					},
				}
				os.Setenv("AI_KEY", "env_key_used")
			},
			wantConfig: AITitleGeneratorConfig{
				APIKey:         "env_key_used",
				Endpoint:       "endpoint",
				DeploymentName: "deployment",
				APIVersion:     "version",
			},
			wantErr: false,
		},
		{
			name: "Error if no AI_KEY in env and GlobalSettings.AI.Key is empty",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Key:        "",
						Endpoint:   "endpoint",
						Deployment: "deployment",
						APIVersion: "version",
					},
				}
				os.Unsetenv("AI_KEY")
			},
			wantErr:           true,
			expectedErrSubstr: "AI_KEY environment variable not set and no key in settings",
		},
		{
			name: "Error if GlobalSettings.AI.Endpoint is empty",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Key:        "some_key",
						Endpoint:   "",
						Deployment: "deployment",
						APIVersion: "version",
					},
				}
				os.Unsetenv("AI_KEY")
			},
			wantErr:           true,
			expectedErrSubstr: "AI endpoint or deployment not configured in settings.yaml",
		},
		{
			name: "Error if GlobalSettings.AI.Deployment is empty",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Key:        "some_key",
						Endpoint:   "endpoint",
						Deployment: "",
						APIVersion: "version",
					},
				}
				os.Unsetenv("AI_KEY")
			},
			wantErr:           true,
			expectedErrSubstr: "AI endpoint or deployment not configured in settings.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			gotConfig, err := GetAIConfig()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetAIConfig() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("GetAIConfig() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GetAIConfig() unexpected error = %v", err)
					return
				}
				if !reflect.DeepEqual(gotConfig, tt.wantConfig) {
					t.Errorf("GetAIConfig() gotConfig = %v, want %v", gotConfig, tt.wantConfig)
				}
			}
		})
	}
}
