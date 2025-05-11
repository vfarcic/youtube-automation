package main

import (
	"fmt"
	"os"

	"devopstoolkitseries/youtube-automation/pkg/bluesky"
)

func main() {
	// Parse CLI arguments and load configuration
	getArgs()

	// Validate Bluesky configuration if identifier is set
	if settings.Bluesky.Identifier != "" {
		config := bluesky.Config{
			Identifier: settings.Bluesky.Identifier,
			Password:   settings.Bluesky.Password,
			URL:        settings.Bluesky.URL,
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
