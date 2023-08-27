package main

import "fmt"

func postReddit(title, videoId string, posted bool) bool {
	if len(title) == 0 {
		errorMessage = "Please generate the title first."
		return false
	}
	if len(videoId) == 0 {
		errorMessage = "Please upload video first."
		return false
	}
	if !posted {
		confirmationMessage = "Use the following information to post it to https://reddit.com manually."
		confirmationMessage += fmt.Sprintf("\n\nTitle:\n%s\n\nURL:\n%s", title, getYouTubeURL(videoId))
	}
	return getInputFromBool(posted)
}
