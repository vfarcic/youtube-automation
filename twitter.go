package main

import (
	"strings"

	"github.com/atotto/clipboard"
)

type Twitter struct{}

func (t *Twitter) Post(message, videoId string) {
	message = strings.ReplaceAll(message, "[YouTube Link]", getYouTubeURL(videoId))
	clipboard.WriteAll(message)
	println(confirmationStyle.Render("The tweet has be copied to clipboard. Please paste it into Twitter manually."))
}
