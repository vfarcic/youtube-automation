package main

import (
	"fmt"
	"os"
)

func main() {
	// Parse CLI arguments and load configuration
	getArgs()

	// Validate Bluesky configuration if identifier is set
	if settings.Bluesky.Identifier != "" {
		config := BlueskyConfig{
			Identifier: settings.Bluesky.Identifier,
			Password:   settings.Bluesky.Password,
			URL:        settings.Bluesky.URL,
		}

		if err := ValidateBlueskyConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Bluesky configuration error: %s\n", err)
			os.Exit(1)
		}
	}

	choices := Choices{}
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
