package main

import (
	"github.com/atotto/clipboard"
)

func postSlack(videoId string, posted bool) bool {
	if len(videoId) == 0 {
		errorMessage = "Please upload video first."
		return false
	}
	if !posted {
		clipboard.WriteAll(getYouTubeURL(videoId))
		confirmationMessage = "The video URL has be copied to clipboard. Please paste it into Slack manually."
	}
	return getInputFromBool(posted)
}
