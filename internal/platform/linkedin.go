package platform

import (
	"strings"

	"github.com/atotto/clipboard"
)

func PostLinkedIn(message, videoId string, getYouTubeURL func(string) string, confirmationStyle interface{ Render(...string) string }) {
	message = strings.ReplaceAll(message, "[YouTube Link]", getYouTubeURL(videoId))
	clipboard.WriteAll(message)
	println(confirmationStyle.Render("The message has be copied to clipboard. Please paste it into LinkedIn manually."))
}