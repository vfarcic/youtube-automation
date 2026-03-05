package main

import (
	"fmt"
	"os"

	"devopstoolkit/youtube-automation/internal/api"
	"devopstoolkit/youtube-automation/internal/app"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/platform/bluesky"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/video"
)

var version = "dev" // Will be overwritten by linker flags during release build

func main() {
	// Check for version flag before anything else
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println(version)
		os.Exit(0)
	}

	// Parse CLI arguments and load configuration
	configuration.GetArgs()

	// Serve mode: start the HTTP API server
	if configuration.IsServeMode() {
		fsOps := filesystem.NewOperations()
		videoManager := video.NewManager(fsOps.GetFilePath)
		videoService := service.NewVideoService("index.yaml", fsOps, videoManager)

		srv := api.NewServer(videoService, videoManager)
		if err := srv.Start(configuration.GetServeHost(), configuration.GetServePort()); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Validate Bluesky configuration if identifier is set
	if configuration.GlobalSettings.Bluesky.Identifier != "" {
		config := bluesky.Config{
			Identifier: configuration.GlobalSettings.Bluesky.Identifier,
			Password:   configuration.GlobalSettings.Bluesky.Password,
			URL:        configuration.GlobalSettings.Bluesky.URL,
		}

		if err := bluesky.ValidateConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Bluesky configuration error: %s\n", err)
			os.Exit(1)
		}
	}

	// Start the CLI application
	application := app.New()
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
