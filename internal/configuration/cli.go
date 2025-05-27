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
	Key        string `yaml:"key"`
	Endpoint   string `yaml:"endpoint"`
	Deployment string `yaml:"deployment"`
	APIVersion string `yaml:"apiVersion,omitempty"`
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
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Endpoint, "ai-endpoint", GlobalSettings.AI.Endpoint, "AI endpoint. Only Azure OpenAI is currently supported. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Key, "ai-key", GlobalSettings.AI.Key, "AI key. Only Azure OpenAI is currently supported. Environment variable `AI_KEY` is supported as well. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Deployment, "ai-deployment", GlobalSettings.AI.Deployment, "AI Deployment. Only Azure OpenAI is currently supported. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.APIVersion, "ai-api-version", GlobalSettings.AI.APIVersion, "Azure OpenAI API Version (e.g., 2023-05-15). Defaults to a common version if not set.")
	RootCmd.Flags().StringVar(&GlobalSettings.YouTube.APIKey, "youtube-api-key", GlobalSettings.YouTube.APIKey, "YouTube API key. Environment variable `YOUTUBE_API_KEY` is supported as well. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Hugo.Path, "hugo-path", GlobalSettings.Hugo.Path, "Path to the repo with Hugo posts. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.Identifier, "bluesky-identifier", GlobalSettings.Bluesky.Identifier, "Bluesky username/identifier (e.g., username.bsky.social)")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.Password, "bluesky-password", GlobalSettings.Bluesky.Password, "Bluesky password. Environment variable `BLUESKY_PASSWORD` is supported as well.")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.URL, "bluesky-url", GlobalSettings.Bluesky.URL, "Bluesky API URL")
	RootCmd.Flags().StringVar(&GlobalSettings.VideoDefaults.Language, "video-defaults-language", "", "Default language for videos (e.g., 'en', 'es')")
	RootCmd.Flags().StringVar(&GlobalSettings.VideoDefaults.AudioLanguage, "video-defaults-audio-language", "", "Default audio language for videos (e.g., 'en', 'es')")
	RootCmd.Flags().IntVar(&GlobalSettings.API.Port, "api-port", GlobalSettings.API.Port, "Port for REST API server")
	RootCmd.Flags().BoolVar(&GlobalSettings.API.Enabled, "api-enabled", GlobalSettings.API.Enabled, "Enable REST API server")

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

	if GlobalSettings.AI.Endpoint == "" {
		RootCmd.MarkFlagRequired("ai-endpoint")
	}

	if envAIKey := os.Getenv("AI_KEY"); envAIKey != "" {
		GlobalSettings.AI.Key = envAIKey
	} else if GlobalSettings.AI.Key == "" {
		RootCmd.MarkFlagRequired("ai-key")
	}

	if GlobalSettings.AI.Deployment == "" {
		RootCmd.MarkFlagRequired("ai-deployment")
	}

	// Default API version if not set
	if GlobalSettings.AI.APIVersion == "" {
		GlobalSettings.AI.APIVersion = "2023-05-15" // Defaulting to a common version
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
