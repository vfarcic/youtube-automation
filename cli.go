package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// osExit is a variable to allow mocking os.Exit in tests
var osExit = os.Exit

var rootCmd = &cobra.Command{
	Use:   "youtube-release",
	Short: "youtube-release is a super fancy CLI for releasing YouTube videos.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

type Settings struct {
	Email   SettingsEmail   `yaml:"email"`
	AI      SettingsAI      `yaml:"ai"`
	YouTube SettingsYouTube `yaml:"youtube"`
	Hugo    SettingsHugo    `yaml:"hugo"`
	Bluesky SettingsBluesky `yaml:"bluesky"`
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
}

type SettingsYouTube struct {
	APIKey string `yaml:"apiKey"`
}

type SettingsBluesky struct {
	Identifier string `yaml:"identifier"`
	Password   string `yaml:"password"`
	URL        string `yaml:"url"`
}

var settings Settings

func init() {
	// Load settings from YAML file
	yamlFile, err := os.ReadFile("settings.yaml")
	if err == nil {
		if err := yaml.Unmarshal(yamlFile, &settings); err != nil {
			fmt.Printf("Error parsing config file: %s\n", err)
		}
	}

	// Define command-line flags
	rootCmd.Flags().StringVar(&settings.Email.From, "email-from", settings.Email.From, "From which email to send messages. (required)")
	rootCmd.Flags().StringVar(&settings.Email.ThumbnailTo, "email-thumbnail-to", settings.Email.ThumbnailTo, "To which email to send requests for thumbnails. (required)")
	rootCmd.Flags().StringVar(&settings.Email.EditTo, "email-edit-to", settings.Email.EditTo, "To which email to send requests for edits. (required)")
	rootCmd.Flags().StringVar(&settings.Email.FinanceTo, "email-finance-to", settings.Email.FinanceTo, "To which email to send emails related to finances. (required)")
	rootCmd.Flags().StringVar(&settings.Email.Password, "email-password", settings.Email.Password, "Email server password. Environment variable `EMAIL_PASSWORD` is supported as well. (required)")
	rootCmd.Flags().StringVar(&settings.AI.Endpoint, "ai-endpoint", settings.AI.Endpoint, "AI endpoint. Only Azure OpenAI is currently supported. (required)")
	rootCmd.Flags().StringVar(&settings.AI.Key, "ai-key", settings.AI.Key, "AI key. Only Azure OpenAI is currently supported. Environment variable `AI_KEY` is supported as well. (required)")
	rootCmd.Flags().StringVar(&settings.AI.Deployment, "ai-deployment", settings.AI.Deployment, "AI Deployment. Only Azure OpenAI is currently supported. (required)")
	rootCmd.Flags().StringVar(&settings.YouTube.APIKey, "youtube-api-key", settings.YouTube.APIKey, "YouTube API key. Environment variable `YOUTUBE_API_KEY` is supported as well. (required)")
	rootCmd.Flags().StringVar(&settings.Hugo.Path, "hugo-path", settings.Hugo.Path, "Path to the repo with Hugo posts. (required)")
	rootCmd.Flags().StringVar(&settings.Bluesky.Identifier, "bluesky-identifier", settings.Bluesky.Identifier, "Bluesky username/identifier (e.g., username.bsky.social)")
	rootCmd.Flags().StringVar(&settings.Bluesky.Password, "bluesky-password", settings.Bluesky.Password, "Bluesky password. Environment variable `BLUESKY_PASSWORD` is supported as well.")
	rootCmd.Flags().StringVar(&settings.Bluesky.URL, "bluesky-url", settings.Bluesky.URL, "Bluesky API URL")
	if settings.Bluesky.URL == "" {
		settings.Bluesky.URL = "https://bsky.social/xrpc"
	}

	// Check required fields and environment variables
	if settings.Email.From == "" {
		rootCmd.MarkFlagRequired("email-from")
	}
	if settings.Email.ThumbnailTo == "" {
		rootCmd.MarkFlagRequired("email-thumbnail-to")
	}
	if settings.Email.EditTo == "" {
		rootCmd.MarkFlagRequired("email-edit-to")
	}
	if settings.Email.FinanceTo == "" {
		rootCmd.MarkFlagRequired("email-finance-to")
	}

	// Check environment variables
	if envPassword := os.Getenv("EMAIL_PASSWORD"); envPassword != "" {
		settings.Email.Password = envPassword
	} else if settings.Email.Password == "" {
		rootCmd.MarkFlagRequired("email-password")
	}

	if settings.AI.Endpoint == "" {
		rootCmd.MarkFlagRequired("ai-endpoint")
	}

	if envAIKey := os.Getenv("AI_KEY"); envAIKey != "" {
		settings.AI.Key = envAIKey
	} else if settings.AI.Key == "" {
		rootCmd.MarkFlagRequired("ai-key")
	}

	if settings.AI.Deployment == "" {
		rootCmd.MarkFlagRequired("ai-deployment")
	}

	if envYouTubeKey := os.Getenv("YOUTUBE_API_KEY"); envYouTubeKey != "" {
		settings.YouTube.APIKey = envYouTubeKey
	} else if settings.YouTube.APIKey == "" {
		rootCmd.MarkFlagRequired("youtube-api-key")
	}

	if settings.Hugo.Path == "" {
		rootCmd.MarkFlagRequired("hugo-path")
	}

	// Bluesky validation
	if settings.Bluesky.Identifier != "" {
		envBlueskyPassword := os.Getenv("BLUESKY_PASSWORD")
		if envBlueskyPassword != "" {
			settings.Bluesky.Password = envBlueskyPassword
		} else if settings.Bluesky.Password == "" {
			rootCmd.MarkFlagRequired("bluesky-password")
		}
	} else if envBlueskyPassword := os.Getenv("BLUESKY_PASSWORD"); envBlueskyPassword != "" {
		settings.Bluesky.Password = envBlueskyPassword
	}
}

func getArgs() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing the CLI '%s'", err)
		osExit(1)
	}
}
