package platform

import (
	"github.com/atotto/clipboard"
)

func PostSlack(videoId string, getYouTubeURL func(string) string, confirmationStyle interface{ Render(...string) string }) {
	clipboard.WriteAll(getYouTubeURL(videoId))
	println(confirmationStyle.Render("The video URL has been copied to clipboard. Please paste it into Slack manually."))
}