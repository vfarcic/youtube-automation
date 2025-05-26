package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"devopstoolkit/youtube-automation/internal/api"
	"devopstoolkit/youtube-automation/internal/app"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/platform/bluesky"
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

	// Check if API server should be started
	if configuration.GlobalSettings.API.Enabled {
		startAPIServer()
	} else {
		// Start the CLI application
		application := app.New()
		if err := application.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
			os.Exit(1)
		}
	}
}

func startAPIServer() {
	server := api.NewServer()
	
	// Create a channel to listen for interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	var wg sync.WaitGroup
	
	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Starting YouTube Automation API server on port %d", configuration.GlobalSettings.API.Port)
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	<-c
	log.Println("Received interrupt signal, shutting down...")
	
	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Shutdown the server
	if err := server.Stop(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}
	
	wg.Wait()
	log.Println("Server shutdown complete")
}
