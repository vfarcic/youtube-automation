package main

import (
	"github.com/atotto/clipboard"
)

type Twitter struct{}

func (t *Twitter) Post(message string, posted bool) bool {
	if len(message) == 0 {
		println(redStyle.Render("\nPlease generate Tweet first."))
		return false
	}
	if !posted {
		clipboard.WriteAll(message)
		println(orangeStyle.Render("\nThe tweet has be copied to clipboard. Please paste it into Twitter manually."))
	}
	return getInputFromBool(posted)
}

func (t *Twitter) PostSpace(videoID string, posted bool) bool {
	if len(videoID) == 0 {
		println(redStyle.Render("\nUpload the video first."))
		return false
	}
	if !posted {
		clipboard.WriteAll(getYouTubeURL(videoID))
		println(orangeStyle.Render("\nThe video URL has be copied to clipboard. Please paste it into Twitter manually."))
	}
	return getInputFromBool(posted)
}
