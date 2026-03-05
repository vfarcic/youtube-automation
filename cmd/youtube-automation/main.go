package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"devopstoolkit/youtube-automation/internal/api"
	"devopstoolkit/youtube-automation/internal/app"
	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/frontend"
	gitpkg "devopstoolkit/youtube-automation/internal/git"
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
		dataDir := configuration.GetDataDir()
		gitCfg := configuration.GlobalSettings.Git

		// Git sync: clone/pull data repo, or ensure data-dir exists
		var gitSync *gitpkg.SyncManager
		if gitCfg.RepoURL != "" {
			gitSync = gitpkg.NewSyncManager(gitCfg.RepoURL, gitCfg.Branch, dataDir, gitCfg.Token)
			if err := gitSync.InitialSync(); err != nil {
				fmt.Fprintf(os.Stderr, "Git sync failed: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create data directory: %v\n", err)
				os.Exit(1)
			}
		}

		fsOps := filesystem.NewOperationsWithBaseDir(filepath.Join(dataDir, "manuscript"))
		videoManager := video.NewManager(fsOps.GetFilePath)
		videoService := service.NewVideoService(filepath.Join(dataDir, "index.yaml"), fsOps, videoManager)

		if gitSync != nil {
			videoService.SetOnMutate(gitSync.CommitAndPush)
		}

		aspectSvc := aspect.NewService()

		distFS, _ := fs.Sub(frontend.DistFS, "dist")
		srv := api.NewServer(videoService, videoManager, aspectSvc, fsOps, configuration.GetAPIToken(), distFS)
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
