package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"devopstoolkit/youtube-automation/internal/api"
	"devopstoolkit/youtube-automation/internal/app"
	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/auth"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/frontend"
	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/notification"
	gitpkg "devopstoolkit/youtube-automation/internal/git"
	"devopstoolkit/youtube-automation/internal/platform/bluesky"
	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/service"
	slackpkg "devopstoolkit/youtube-automation/internal/slack"
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

		fsOps := filesystem.NewOperationsWithBaseDir(dataDir, "manuscript")
		aspectSvcForProgress := aspect.NewService()
		videoManager := video.NewManager(fsOps.GetFilePath, aspectSvcForProgress)
		videoService := service.NewVideoService(filepath.Join(dataDir, "index.yaml"), fsOps, videoManager)

		if gitSync != nil {
			videoService.SetOnMutate(gitSync.CommitAndPush)
		}

		aspectSvc := aspect.NewService()

		distFS, _ := fs.Sub(frontend.DistFS, "dist")
		srv := api.NewServer(videoService, videoManager, aspectSvc, fsOps, &api.DefaultAIService{}, configuration.GetAPIToken(), distFS)

		// Publishing: configure YouTube upload, Hugo, social media
		{
			bsCfg := bluesky.GetConfig(
				configuration.GlobalSettings.Bluesky.Identifier,
				configuration.GlobalSettings.Bluesky.Password,
				configuration.GlobalSettings.Bluesky.URL,
			)
			hugo := &publishing.Hugo{}
			var slackSvc *slackpkg.SlackService
			if slackpkg.GlobalSlackConfig.Token != "" {
				if svc, err := slackpkg.NewSlackService(slackpkg.GlobalSlackConfig); err == nil {
					slackSvc = svc
					slog.Info("Slack posting enabled")
				} else {
					slog.Warn("Slack service creation failed", "error", err)
				}
			}
			pubSvc := api.NewDefaultPublishingService(bsCfg, hugo, slackSvc)
			srv.SetPublishingService(pubSvc)
			slog.Info("Publishing service configured")
		}

		// Email: configure action button email sending
		if configuration.GlobalSettings.Email.Password != "" {
			emailSvc := notification.NewEmail(configuration.GlobalSettings.Email.Password)
			srv.SetEmailService(emailSvc, &configuration.GlobalSettings.Email)
			slog.Info("Email notifications enabled for action buttons")
		}

		// Google Drive: configure thumbnail upload if credentials are set
		gdriveCfg := configuration.GlobalSettings.GDrive
		if gdriveCfg.CredentialsFile != "" {
			tokenFile := gdriveCfg.TokenFile
			if tokenFile == "" {
				tokenFile = "gdrive-go.json"
			}
			callbackPort := gdriveCfg.CallbackPort
			if callbackPort == 0 {
				callbackPort = 8092
			}
			authCfg := auth.OAuthConfig{
				CredentialsFile: gdriveCfg.CredentialsFile,
				TokenFileName:   tokenFile,
				CallbackPort:    callbackPort,
				Scopes:          []string{"https://www.googleapis.com/auth/drive"},
			}
			httpClient, err := auth.GetClient(context.Background(), authCfg)
			if err != nil {
				slog.Warn("Google Drive auth failed, thumbnail uploads disabled", "error", err)
			} else {
				ds, err := gdrive.NewDriveService(context.Background(), httpClient)
				if err != nil {
					slog.Warn("Google Drive service creation failed", "error", err)
				} else {
					srv.SetDriveService(ds, gdriveCfg.FolderID)
					slog.Info("Google Drive thumbnail uploads enabled")
				}
			}
		}

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
