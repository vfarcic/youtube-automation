package configuration

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// osExit is a variable to allow mocking os.Exit in tests
var osExit = os.Exit

var RootCmd = &cobra.Command{
	Use:   "youtube-release",
	Short: "youtube-release is a super fancy CLI for releasing YouTube videos.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

type Settings struct {
	Email         SettingsEmail         `yaml:"email"`
	AI            SettingsAI            `yaml:"ai"`
	YouTube       SettingsYouTube       `yaml:"youtube"`
	Hugo          SettingsHugo          `yaml:"hugo"`
	Bluesky       SettingsBluesky       `yaml:"bluesky"`
	VideoDefaults SettingsVideoDefaults `yaml:"videoDefaults"`
	API           SettingsAPI           `yaml:"api"`
	Slack         SettingsSlack         `yaml:"slack"`
	Timing        TimingConfig          `yaml:"timing"`
	Calendar      SettingsCalendar      `yaml:"calendar"`
	Shorts        ShortsConfig          `yaml:"shorts"`
}

type SettingsEmail struct {
	From        string `yaml:"from"`
	ThumbnailTo string `yaml:"thumbnailTo"`
	EditTo      string `yaml:"editTo"`
	FinanceTo   string `yaml:"financeTo"`
	Password    string `yaml:"password"`
}

type SettingsHugo struct {
	Path string `yaml:"path"`
}

type SettingsAI struct {
	Provider  string              `yaml:"provider"`
	Azure     SettingsAzureAI     `yaml:"azure"`
	Anthropic SettingsAnthropicAI `yaml:"anthropic"`
}

type SettingsAzureAI struct {
	Key        string `yaml:"key"`
	Endpoint   string `yaml:"endpoint"`
	Deployment string `yaml:"deployment"`
	APIVersion string `yaml:"apiVersion,omitempty"`
}

type SettingsAnthropicAI struct {
	Key   string `yaml:"key"`
	Model string `yaml:"model"`
}

type SettingsYouTube struct {
	APIKey    string `yaml:"apiKey"`
	ChannelId string `yaml:"channelId"`
}

type SettingsBluesky struct {
	Identifier string `yaml:"identifier"`
	Password   string `yaml:"password"`
	URL        string `yaml:"url"`
}

type SettingsVideoDefaults struct {
	Language      string `yaml:"language"`
	AudioLanguage string `yaml:"audioLanguage"`
}

type SettingsAPI struct {
	Port    int  `yaml:"port"`
	Enabled bool `yaml:"enabled"`
}

type SettingsSlack struct {
	TargetChannelIDs []string `yaml:"targetChannelIDs"`
}

// SettingsCalendar holds Google Calendar integration settings
type SettingsCalendar struct {
	Disabled bool `yaml:"disabled"` // Set to true to disable calendar event creation (enabled by default)
}

// TimingRecommendation represents a single timing recommendation
// for video publishing based on audience behavior and performance data
type TimingRecommendation struct {
	Day       string `yaml:"day" json:"day"`             // "Monday", "Tuesday", etc.
	Time      string `yaml:"time" json:"time"`           // "16:00", "09:00", etc. (UTC)
	Reasoning string `yaml:"reasoning" json:"reasoning"` // Why this slot recommended
}

// TimingConfig holds timing recommendations for video publishing
type TimingConfig struct {
	Recommendations []TimingRecommendation `yaml:"recommendations" json:"recommendations"`
}

// ShortsConfig holds configuration for YouTube Shorts identification and scheduling
type ShortsConfig struct {
	MaxWords       int `yaml:"maxWords" json:"maxWords"`             // Maximum word count for a Short segment (default: 150)
	CandidateCount int `yaml:"candidateCount" json:"candidateCount"` // Number of Short candidates to generate (default: 10)
}

var GlobalSettings Settings

func init() {
	// Load settings from YAML file
	yamlFile, err := os.ReadFile("settings.yaml")
	if err == nil {
		if err := yaml.Unmarshal(yamlFile, &GlobalSettings); err != nil {
			fmt.Printf("Error parsing config file: %s\n", err)
		}
	}

	// Define command-line flags
	RootCmd.Flags().StringVar(&GlobalSettings.Email.From, "email-from", GlobalSettings.Email.From, "From which email to send messages. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Email.ThumbnailTo, "email-thumbnail-to", GlobalSettings.Email.ThumbnailTo, "To which email to send requests for thumbnails. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Email.EditTo, "email-edit-to", GlobalSettings.Email.EditTo, "To which email to send requests for edits. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Email.FinanceTo, "email-finance-to", GlobalSettings.Email.FinanceTo, "To which email to send emails related to finances. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Email.Password, "email-password", GlobalSettings.Email.Password, "Email server password. Environment variable `EMAIL_PASSWORD` is supported as well. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Provider, "ai-provider", GlobalSettings.AI.Provider, "AI provider (azure or anthropic). Defaults to azure for backward compatibility.")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Azure.Endpoint, "ai-endpoint", GlobalSettings.AI.Azure.Endpoint, "AI endpoint. For Azure OpenAI. (required for azure)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Azure.Key, "ai-key", GlobalSettings.AI.Azure.Key, "AI key. Environment variable `AI_KEY` is supported as well. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Azure.Deployment, "ai-deployment", GlobalSettings.AI.Azure.Deployment, "AI Deployment. For Azure OpenAI. (required for azure)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Azure.APIVersion, "ai-api-version", GlobalSettings.AI.Azure.APIVersion, "Azure OpenAI API Version (e.g., 2023-05-15). Defaults to a common version if not set.")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Anthropic.Key, "anthropic-key", GlobalSettings.AI.Anthropic.Key, "Anthropic API key. Environment variable `ANTHROPIC_API_KEY` is supported as well. (required for anthropic)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Anthropic.Model, "anthropic-model", GlobalSettings.AI.Anthropic.Model, "Anthropic model (e.g., claude-3-sonnet-20240229). (required for anthropic)")
	RootCmd.Flags().StringVar(&GlobalSettings.YouTube.APIKey, "youtube-api-key", GlobalSettings.YouTube.APIKey, "YouTube API key. Environment variable `YOUTUBE_API_KEY` is supported as well. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Hugo.Path, "hugo-path", GlobalSettings.Hugo.Path, "Path to the repo with Hugo posts. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.Identifier, "bluesky-identifier", GlobalSettings.Bluesky.Identifier, "Bluesky username/identifier (e.g., username.bsky.social)")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.Password, "bluesky-password", GlobalSettings.Bluesky.Password, "Bluesky password. Environment variable `BLUESKY_PASSWORD` is supported as well.")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.URL, "bluesky-url", GlobalSettings.Bluesky.URL, "Bluesky API URL")
	RootCmd.Flags().StringVar(&GlobalSettings.VideoDefaults.Language, "video-defaults-language", "", "Default language for videos (e.g., 'en', 'es')")
	RootCmd.Flags().StringVar(&GlobalSettings.VideoDefaults.AudioLanguage, "video-defaults-audio-language", "", "Default audio language for videos (e.g., 'en', 'es')")
	RootCmd.Flags().IntVar(&GlobalSettings.API.Port, "api-port", GlobalSettings.API.Port, "Port for REST API server")
	RootCmd.Flags().BoolVar(&GlobalSettings.API.Enabled, "api-enabled", GlobalSettings.API.Enabled, "Enable REST API server")
	RootCmd.Flags().BoolVar(&GlobalSettings.Calendar.Disabled, "calendar-disabled", GlobalSettings.Calendar.Disabled, "Disable Google Calendar event creation after video upload")

	// Default Bluesky URL if not set by file or flag
	if GlobalSettings.Bluesky.URL == "" {
		GlobalSettings.Bluesky.URL = "https://bsky.social/xrpc"
	}

	// Added for PRD: Automated Video Language Setting
	// Default video language if not set by file or flag (flag default is 'en')
	if GlobalSettings.VideoDefaults.Language == "" {
		GlobalSettings.VideoDefaults.Language = "en"
	}
	if GlobalSettings.VideoDefaults.AudioLanguage == "" {
		GlobalSettings.VideoDefaults.AudioLanguage = "en"
	}

	// Default API settings
	if GlobalSettings.API.Port == 0 {
		GlobalSettings.API.Port = 8080
	}

	// Default Shorts settings
	if GlobalSettings.Shorts.MaxWords == 0 {
		GlobalSettings.Shorts.MaxWords = 150
	}
	if GlobalSettings.Shorts.CandidateCount == 0 {
		GlobalSettings.Shorts.CandidateCount = 10
	}

	// Calendar settings: enabled by default, set calendar.disabled: true to disable

	// Check required fields and environment variables
	if GlobalSettings.Email.From == "" {
		RootCmd.MarkFlagRequired("email-from")
	}
	if GlobalSettings.Email.ThumbnailTo == "" {
		RootCmd.MarkFlagRequired("email-thumbnail-to")
	}
	if GlobalSettings.Email.EditTo == "" {
		RootCmd.MarkFlagRequired("email-edit-to")
	}
	if GlobalSettings.Email.FinanceTo == "" {
		RootCmd.MarkFlagRequired("email-finance-to")
	}

	// Check environment variables
	if envPassword := os.Getenv("EMAIL_PASSWORD"); envPassword != "" {
		GlobalSettings.Email.Password = envPassword
	} else if GlobalSettings.Email.Password == "" {
		RootCmd.MarkFlagRequired("email-password")
	}

	// Default to azure provider for backward compatibility
	if GlobalSettings.AI.Provider == "" {
		GlobalSettings.AI.Provider = "azure"
	}

	// Provider-specific validation
	switch GlobalSettings.AI.Provider {
	case "azure":
		if GlobalSettings.AI.Azure.Endpoint == "" {
			RootCmd.MarkFlagRequired("ai-endpoint")
		}

		if envAIKey := os.Getenv("AI_KEY"); envAIKey != "" {
			GlobalSettings.AI.Azure.Key = envAIKey
		} else if GlobalSettings.AI.Azure.Key == "" {
			RootCmd.MarkFlagRequired("ai-key")
		}

		if GlobalSettings.AI.Azure.Deployment == "" {
			RootCmd.MarkFlagRequired("ai-deployment")
		}

		// Default API version if not set
		if GlobalSettings.AI.Azure.APIVersion == "" {
			GlobalSettings.AI.Azure.APIVersion = "2023-05-15" // Defaulting to a common version
		}

	case "anthropic":
		if envAnthropicKey := os.Getenv("ANTHROPIC_API_KEY"); envAnthropicKey != "" {
			GlobalSettings.AI.Anthropic.Key = envAnthropicKey
		} else if GlobalSettings.AI.Anthropic.Key == "" {
			RootCmd.MarkFlagRequired("anthropic-key")
		}

		if GlobalSettings.AI.Anthropic.Model == "" {
			GlobalSettings.AI.Anthropic.Model = "claude-3-sonnet-20240229" // Default model
		}

	default:
		fmt.Printf("Unsupported AI provider: %s. Supported providers: azure, anthropic\n", GlobalSettings.AI.Provider)
		osExit(1)
	}

	if envYouTubeKey := os.Getenv("YOUTUBE_API_KEY"); envYouTubeKey != "" {
		GlobalSettings.YouTube.APIKey = envYouTubeKey
	} else if GlobalSettings.YouTube.APIKey == "" {
		RootCmd.MarkFlagRequired("youtube-api-key")
	}

	if GlobalSettings.Hugo.Path == "" {
		RootCmd.MarkFlagRequired("hugo-path")
	}

	// Bluesky validation
	if GlobalSettings.Bluesky.Identifier != "" {
		envBlueskyPassword := os.Getenv("BLUESKY_PASSWORD")
		if envBlueskyPassword != "" {
			GlobalSettings.Bluesky.Password = envBlueskyPassword
		} else if GlobalSettings.Bluesky.Password == "" {
			RootCmd.MarkFlagRequired("bluesky-password")
		}
	} else if envBlueskyPassword := os.Getenv("BLUESKY_PASSWORD"); envBlueskyPassword != "" {
		GlobalSettings.Bluesky.Password = envBlueskyPassword
	}
}

func GetArgs() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing the CLI '%s'", err)
		osExit(1)
	}
}
