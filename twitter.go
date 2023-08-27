package main

import (
	"github.com/atotto/clipboard"
)

type Twitter struct{}

func (t *Twitter) Post(message string, posted bool) bool {
	if len(message) == 0 {
		errorMessage = "Please generate Tweet first."
		return false
	}
	if !posted {
		clipboard.WriteAll(message)
		confirmationMessage = "The tweet has be copied to clipboard. Please paste it into Twitter manually."
	}
	return getInputFromBool(posted)
}

func (t *Twitter) PostSpace(videoID string, posted bool) bool {
	if len(videoID) == 0 {
		errorMessage = "Please upload video first."
		return false
	}
	if !posted {
		clipboard.WriteAll(getYouTubeURL(videoID))
		confirmationMessage = "The video URL has be copied to clipboard. Please paste it into Twitter manually."
	}
	return getInputFromBool(posted)
}
