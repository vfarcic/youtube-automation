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

var path string

func init() {
	rootCmd.Flags().StringVar(&path, "path", "video.yaml", "Path to the YAML file where the video info will be stored. Defaults to video.yaml in the current directory.")
	rootCmd.MarkFlagRequired("path")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing the CLI '%s'", err)
		os.Exit(1)
	}
}
