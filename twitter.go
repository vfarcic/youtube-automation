package main

import (
	"github.com/atotto/clipboard"
)

type Twitter struct{}

func (t *Twitter) Post(message string) {
	clipboard.WriteAll(message)
	println(confirmationStyle.Render("The tweet has be copied to clipboard. Please paste it into Twitter manually."))
}

func (t *Twitter) PostSpace(videoID string) {
	clipboard.WriteAll(getYouTubeURL(videoID))
	println(confirmationStyle.Render("The video URL has be copied to clipboard. Please paste it into Twitter manually."))
}
