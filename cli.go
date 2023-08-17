package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "youtube-release",
	Short: "youtube-release is a super fancy CLI for releasing YouTube videos.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

type Settings struct {
	path             string
	fromEmail        string
	toThumbnailEmail string
	toEditEmail      string
}

var settings Settings

func init() {
	rootCmd.Flags().StringVar(&settings.path, "path", "", "Path to the YAML file where the video info will be stored. Defaults to video.yaml in the current directory.")
	rootCmd.MarkFlagRequired("path")
	rootCmd.Flags().StringVar(&settings.fromEmail, "from-email", "", "From which email to send messages.")
	rootCmd.MarkFlagRequired("from-email")
	rootCmd.Flags().StringVar(&settings.toThumbnailEmail, "to-thumbnail-email", "", "To which email to send requests for thumbnails.")
	rootCmd.MarkFlagRequired("to-thumbnail-email")
	rootCmd.Flags().StringVar(&settings.toEditEmail, "to-edit-email", "", "To which email to send requests for edits.")
	rootCmd.MarkFlagRequired("to-edit-email")
}

func getArgs() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing the CLI '%s'", err)
		os.Exit(1)
	}
}
