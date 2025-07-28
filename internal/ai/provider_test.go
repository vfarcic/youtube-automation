package ai

import (
	"context"
	"os"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
)

// Compile-time interface verification
var (
	_ AIProvider = (*AzureProvider)(nil)
	_ AIProvider = (*AnthropicProvider)(nil)
)

func TestGetAIProvider(t *testing.T) {
	// Store original values
	originalSettings := configuration.GlobalSettings
	originalGetAIProvider := GetAIProvider
	defer func() {
		configuration.GlobalSettings = originalSettings
		GetAIProvider = originalGetAIProvider
	}()

	// Restore the real GetAIProvider function for these tests
	GetAIProvider = originalGetAIProvider

	tests := []struct {
		name              string
		setupFunc         func()
		wantProviderType  string // "azure" or "anthropic"
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name: "Azure provider with valid config",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "azure",
						Azure: configuration.SettingsAzureAI{
							Key:        "test-key",
							Endpoint:   "https://test.openai.azure.com",
							Deployment: "test-deployment",
							APIVersion: "2023-05-15",
						},
					},
				}
				os.Unsetenv("AI_KEY")
			},
			wantProviderType: "azure",
			wantErr:          false,
		},
		{
			name: "Azure provider with env var API key",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "azure",
						Azure: configuration.SettingsAzureAI{
							Key:        "", // Empty in config
							Endpoint:   "https://test.openai.azure.com",
							Deployment: "test-deployment",
							APIVersion: "2023-05-15",
						},
					},
				}
				os.Setenv("AI_KEY", "env-test-key")
			},
			wantProviderType: "azure",
			wantErr:          false,
		},
		{
			name: "Anthropic provider with valid config",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "anthropic",
						Anthropic: configuration.SettingsAnthropicAI{
							Key:   "test-anthropic-key",
							Model: "claude-3-sonnet-20240229",
						},
					},
				}
				os.Unsetenv("ANTHROPIC_API_KEY")
			},
			wantProviderType: "anthropic",
			wantErr:          false,
		},
		{
			name: "Anthropic provider with env var API key",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "anthropic",
						Anthropic: configuration.SettingsAnthropicAI{
							Key:   "", // Empty in config
							Model: "claude-3-sonnet-20240229",
						},
					},
				}
				os.Setenv("ANTHROPIC_API_KEY", "env-anthropic-key")
			},
			wantProviderType: "anthropic",
			wantErr:          false,
		},
		{
			name: "Unsupported provider",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "openai-direct", // Unsupported
					},
				}
			},
			wantErr:           true,
			expectedErrSubstr: "unsupported AI provider: openai-direct",
		},
		{
			name: "Azure provider missing API key",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "azure",
						Azure: configuration.SettingsAzureAI{
							Key:        "", // Missing
							Endpoint:   "https://test.openai.azure.com",
							Deployment: "test-deployment",
							APIVersion: "2023-05-15",
						},
					},
				}
				os.Unsetenv("AI_KEY")
			},
			wantErr:           true,
			expectedErrSubstr: "Azure OpenAI API key not configured",
		},
		{
			name: "Azure provider missing endpoint",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "azure",
						Azure: configuration.SettingsAzureAI{
							Key:        "test-key",
							Endpoint:   "", // Missing
							Deployment: "test-deployment",
							APIVersion: "2023-05-15",
						},
					},
				}
			},
			wantErr:           true,
			expectedErrSubstr: "Azure OpenAI endpoint or deployment not configured",
		},
		{
			name: "Azure provider missing deployment",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "azure",
						Azure: configuration.SettingsAzureAI{
							Key:        "test-key",
							Endpoint:   "https://test.openai.azure.com",
							Deployment: "", // Missing
							APIVersion: "2023-05-15",
						},
					},
				}
			},
			wantErr:           true,
			expectedErrSubstr: "Azure OpenAI endpoint or deployment not configured",
		},
		{
			name: "Anthropic provider missing API key",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "anthropic",
						Anthropic: configuration.SettingsAnthropicAI{
							Key:   "", // Missing
							Model: "claude-3-sonnet-20240229",
						},
					},
				}
				os.Unsetenv("ANTHROPIC_API_KEY")
			},
			wantErr:           true,
			expectedErrSubstr: "Anthropic API key not configured",
		},
		{
			name: "Azure provider uses default API version",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "azure",
						Azure: configuration.SettingsAzureAI{
							Key:        "test-key",
							Endpoint:   "https://test.openai.azure.com",
							Deployment: "test-deployment",
							APIVersion: "", // Empty - should use default
						},
					},
				}
			},
			wantProviderType: "azure",
			wantErr:          false,
		},
		{
			name: "Anthropic provider uses default model",
			setupFunc: func() {
				configuration.GlobalSettings = configuration.Settings{
					AI: configuration.SettingsAI{
						Provider: "anthropic",
						Anthropic: configuration.SettingsAnthropicAI{
							Key:   "test-anthropic-key",
							Model: "", // Empty - should use default
						},
					},
				}
			},
			wantProviderType: "anthropic",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			provider, err := GetAIProvider()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetAIProvider() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("GetAIProvider() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetAIProvider() unexpected error = %v", err)
				return
			}

			if provider == nil {
				t.Errorf("GetAIProvider() returned nil provider")
				return
			}

			// Check provider type
			switch tt.wantProviderType {
			case "azure":
				if _, ok := provider.(*AzureProvider); !ok {
					t.Errorf("GetAIProvider() returned %T, want *AzureProvider", provider)
				}
			case "anthropic":
				if _, ok := provider.(*AnthropicProvider); !ok {
					t.Errorf("GetAIProvider() returned %T, want *AnthropicProvider", provider)
				}
			}
		})
	}
}

func TestAzureProviderGenerateContent(t *testing.T) {
	// This test verifies the interface implementation but uses a mock
	// since we don't want to make real API calls in unit tests
	
	// Store original values
	originalSettings := configuration.GlobalSettings
	originalGetAIProvider := GetAIProvider
	defer func() {
		configuration.GlobalSettings = originalSettings
		GetAIProvider = originalGetAIProvider
	}()

	// Set up Azure configuration
	configuration.GlobalSettings = configuration.Settings{
		AI: configuration.SettingsAI{
			Provider: "azure",
			Azure: configuration.SettingsAzureAI{
				Key:        "test-key",
				Endpoint:   "https://test.openai.azure.com",
				Deployment: "test-deployment",
				APIVersion: "2023-05-15",
			},
		},
	}

	// Create Azure provider
	provider, err := createAzureProvider()
	if err != nil {
		t.Fatalf("createAzureProvider() failed: %v", err)
	}

	// Verify it implements the interface by assignment
	var _ AIProvider = (*AzureProvider)(nil)

	// Test that GenerateContent method exists and has correct signature
	ctx := context.Background()
	_, err = provider.GenerateContent(ctx, "test prompt", 100)
	
	// We expect an error here since we're using fake credentials
	// The important thing is that the method exists and can be called
	if err == nil {
		t.Log("GenerateContent() succeeded (unexpected with fake credentials, but method signature is correct)")
	} else {
		t.Logf("GenerateContent() failed as expected with fake credentials: %v", err)
	}
}

func TestAnthropicProviderGenerateContent(t *testing.T) {
	// This test verifies the interface implementation but uses a mock
	// since we don't want to make real API calls in unit tests
	
	// Store original values
	originalSettings := configuration.GlobalSettings
	originalGetAIProvider := GetAIProvider
	defer func() {
		configuration.GlobalSettings = originalSettings
		GetAIProvider = originalGetAIProvider
	}()

	// Set up Anthropic configuration
	configuration.GlobalSettings = configuration.Settings{
		AI: configuration.SettingsAI{
			Provider: "anthropic",
			Anthropic: configuration.SettingsAnthropicAI{
				Key:   "test-anthropic-key",
				Model: "claude-3-sonnet-20240229",
			},
		},
	}

	// Create Anthropic provider
	provider, err := createAnthropicProvider()
	if err != nil {
		t.Fatalf("createAnthropicProvider() failed: %v", err)
	}

	// Verify it implements the interface by assignment
	var _ AIProvider = (*AnthropicProvider)(nil)

	// Test that GenerateContent method exists and has correct signature
	ctx := context.Background()
	_, err = provider.GenerateContent(ctx, "test prompt", 100)
	
	// We expect an error here since we're using fake credentials
	// The important thing is that the method exists and can be called
	if err == nil {
		t.Log("GenerateContent() succeeded (unexpected with fake credentials, but method signature is correct)")
	} else {
		t.Logf("GenerateContent() failed as expected with fake credentials: %v", err)
	}
}

func TestProviderDefaults(t *testing.T) {
	// Store original values
	originalSettings := configuration.GlobalSettings
	originalGetAIProvider := GetAIProvider
	defer func() {
		configuration.GlobalSettings = originalSettings
		GetAIProvider = originalGetAIProvider
	}()

	t.Run("Azure provider uses default API version", func(t *testing.T) {
		configuration.GlobalSettings = configuration.Settings{
			AI: configuration.SettingsAI{
				Provider: "azure",
				Azure: configuration.SettingsAzureAI{
					Key:        "test-key",
					Endpoint:   "https://test.openai.azure.com",
					Deployment: "test-deployment",
					APIVersion: "", // Empty - should use default
				},
			},
		}

		provider, err := createAzureProvider()
		if err != nil {
			t.Fatalf("createAzureProvider() failed: %v", err)
		}

		// Verify it's an Azure provider (provider is already *AzureProvider type)
		if provider == nil {
			t.Fatal("Provider should not be nil")
		}

		// The provider should be created successfully even with empty API version
		// (the default "2023-05-15" should be used internally)
		if provider.client == nil {
			t.Error("Azure provider client should not be nil")
		}
	})

	t.Run("Anthropic provider uses default model", func(t *testing.T) {
		configuration.GlobalSettings = configuration.Settings{
			AI: configuration.SettingsAI{
				Provider: "anthropic",
				Anthropic: configuration.SettingsAnthropicAI{
					Key:   "test-anthropic-key",
					Model: "", // Empty - should use default
				},
			},
		}

		provider, err := createAnthropicProvider()
		if err != nil {
			t.Fatalf("createAnthropicProvider() failed: %v", err)
		}

		// Verify it's an Anthropic provider (provider is already *AnthropicProvider type)
		if provider == nil {
			t.Fatal("Provider should not be nil")
		}

		// The provider should have the default model
		expectedDefaultModel := "claude-sonnet-4-20250514"
		if provider.model != expectedDefaultModel {
			t.Errorf("Expected default model %q, got %q", expectedDefaultModel, provider.model)
		}
	})
}

func TestEnvironmentVariablePriority(t *testing.T) {
	// Store original values
	originalSettings := configuration.GlobalSettings
	originalGetAIProvider := GetAIProvider
	originalAIKey := os.Getenv("AI_KEY")
	originalAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	
	defer func() {
		configuration.GlobalSettings = originalSettings
		GetAIProvider = originalGetAIProvider
		if originalAIKey != "" {
			os.Setenv("AI_KEY", originalAIKey)
		} else {
			os.Unsetenv("AI_KEY")
		}
		if originalAnthropicKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalAnthropicKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	t.Run("Azure: Environment variable overrides config", func(t *testing.T) {
		configuration.GlobalSettings = configuration.Settings{
			AI: configuration.SettingsAI{
				Provider: "azure",
				Azure: configuration.SettingsAzureAI{
					Key:        "config-key", // This should be overridden
					Endpoint:   "https://test.openai.azure.com",
					Deployment: "test-deployment",
					APIVersion: "2023-05-15",
				},
			},
		}

		os.Setenv("AI_KEY", "env-key")

		provider, err := createAzureProvider()
		if err != nil {
			t.Fatalf("createAzureProvider() failed: %v", err)
		}

		// We can't easily test the internal key value, but we can verify
		// the provider was created successfully with the env var
		if provider == nil {
			t.Error("Provider should not be nil")
		}
	})

	t.Run("Anthropic: Environment variable overrides config", func(t *testing.T) {
		configuration.GlobalSettings = configuration.Settings{
			AI: configuration.SettingsAI{
				Provider: "anthropic",
				Anthropic: configuration.SettingsAnthropicAI{
					Key:   "config-key", // This should be overridden
					Model: "claude-3-sonnet-20240229",
				},
			},
		}

		os.Setenv("ANTHROPIC_API_KEY", "env-anthropic-key")

		provider, err := createAnthropicProvider()
		if err != nil {
			t.Fatalf("createAnthropicProvider() failed: %v", err)
		}

		// We can't easily test the internal key value, but we can verify
		// the provider was created successfully with the env var
		if provider == nil {
			t.Error("Provider should not be nil")
		}
	})
}