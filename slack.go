package main

import (
	"github.com/atotto/clipboard"
)

func postSlack(videoId string, posted bool) bool {
	if len(videoId) == 0 {
		println(redStyle.Render("\nPlease upload the video first."))
		return false
	}
	if !posted {
		clipboard.WriteAll(getYouTubeURL(videoId))
		println(orangeStyle.Render("\nThe URL of the video has be copied to clipboard. Please paste it into Slack manually."))
	}
	return getInputFromBool(posted)
}
