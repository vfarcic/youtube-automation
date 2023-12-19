package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "youtube-release",
	Short: "youtube-release is a super fancy CLI for releasing YouTube videos.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

type Settings struct {
	Email   SettingsEmail
	AI      SettingsAI
	YouTube SettingsYouTube
}

type SettingsEmail struct {
	From        string
	ThumbnailTo string
	EditTo      string
	FinanceTo   string
	Password    string
}

type SettingsAI struct {
	Key        string
	Endpoint   string
	Deployment string
}

type SettingsYouTube struct {
	APIKey string
}

var settings Settings

func init() {
	viper.SetConfigFile("settings.yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
		return
	}

	rootCmd.Flags().StringVar(&settings.Email.From, "email-from", "", "From which email to send messages. (required)")
	rootCmd.Flags().StringVar(&settings.Email.ThumbnailTo, "email-thumbnail-to", "", "To which email to send requests for thumbnails. (required)")
	rootCmd.Flags().StringVar(&settings.Email.EditTo, "email-edit-to", "", "To which email to send requests for edits. (required)")
	rootCmd.Flags().StringVar(&settings.Email.FinanceTo, "email-finance-to", "", "To which email to send emails related to finances. (required)")
	rootCmd.Flags().StringVar(&settings.Email.Password, "email-password", "", "Email server password. Environment variable `EMAIL_PASSWORD` is supported as well. (required)")
	rootCmd.Flags().StringVar(&settings.AI.Endpoint, "ai-endpoint", "", "AI endpoint. Only Azure OpenAI is currently supported. (required)")
	rootCmd.Flags().StringVar(&settings.AI.Key, "ai-key", "", "AI key. Only Azure OpenAI is currently supported. Environment variable `AI_KEY` is supported as well. (required)")
	rootCmd.Flags().StringVar(&settings.AI.Deployment, "ai-deployment", "", "AI Deployment. Only Azure OpenAI is currently supported. (required)")
	rootCmd.Flags().StringVar(&settings.YouTube.APIKey, "youtube-api-key", "", "AI Deployment. Only Azure OpenAI is currently supported. (required)")
	if viper.IsSet("email.from") {
		settings.Email.From = viper.GetString("email.from")
	} else {
		rootCmd.MarkFlagRequired("email-from")
	}
	if viper.IsSet("email.thumbnailTo") {
		settings.Email.ThumbnailTo = viper.GetString("email.thumbnailTo")
	} else {
		rootCmd.MarkFlagRequired("email-thumbnail-to")
	}
	if viper.IsSet("email.editTo") {
		settings.Email.EditTo = viper.GetString("email.editTo")
	} else {
		rootCmd.MarkFlagRequired("email-edit-to")
	}
	if viper.IsSet("email.financeTo") {
		settings.Email.FinanceTo = viper.GetString("email.financeTo")
	} else {
		rootCmd.MarkFlagRequired("email-finance-to")
	}
	if len(os.Getenv("EMAIL_PASSWORD")) > 0 {
		settings.AI.Key = os.Getenv("EMAIL_PASSWORD")
	} else {
		rootCmd.MarkFlagRequired("email-password")
	}
	if viper.IsSet("ai.endpoint") {
		settings.AI.Endpoint = viper.GetString("ai.endpoint")
	} else {
		rootCmd.MarkFlagRequired("ai-endpoint")
	}
	if len(os.Getenv("AI_KEY")) > 0 {
		settings.AI.Key = os.Getenv("AI_KEY")
	} else {
		rootCmd.MarkFlagRequired("ai-key")
	}
	if viper.IsSet("ai.deployment") {
		settings.AI.Deployment = viper.GetString("ai.deployment")
	} else {
		rootCmd.MarkFlagRequired("ai-deployment")
	}
	if len(os.Getenv("YOUTUBE_API_KEY")) > 0 {
		settings.YouTube.APIKey = os.Getenv("YOUTUBE_API_KEY")
	} else {
		rootCmd.MarkFlagRequired("youtube-api-key")
	}
}

func getArgs() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing the CLI '%s'", err)
		os.Exit(1)
	}
}
