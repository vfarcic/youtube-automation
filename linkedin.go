package main

import (
	"strings"

	"github.com/atotto/clipboard"
)

func postLinkedIn(message, videoId string) {
	message = strings.ReplaceAll(message, "[YouTube Link]", getYouTubeURL(videoId))
	clipboard.WriteAll(message)
	println(confirmationStyle.Render("The message has be copied to clipboard. Please paste it into LinkedIn manually."))
}
