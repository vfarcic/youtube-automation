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
	Email Email
}

type Email struct {
	From        string
	ThumbnailTo string
	EditTo      string
	FinanceTo   string
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
}

func getArgs() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing the CLI '%s'", err)
		os.Exit(1)
	}
}
