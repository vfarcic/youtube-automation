package main

import (
	"fmt"
	"os"

	"devopstoolkitseries/youtube-automation/internal/configuration"
	"devopstoolkitseries/youtube-automation/pkg/bluesky"
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

	choices := NewChoices()
	for {
		choices.ChooseIndex()
	}
}

// func deleteEmpty(s []string) []string {
// 	var r []string
// 	for _, str := range s {
// 		if str != "" {
// 			r = append(r, strings.TrimSpace(str))
// 		}
// 	}
// 	return r
// }
