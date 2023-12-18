package main

import (
	"github.com/atotto/clipboard"
)

func postSlack(videoId string) {
	clipboard.WriteAll(getYouTubeURL(videoId))
	println(confirmationStyle.Render("The video URL has been copied to clipboard. Please paste it into Slack manually."))
}
