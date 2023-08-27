package main

import "fmt"

func postHackerNews(title, videoId string, posted bool) bool {
	if len(title) == 0 {
		errorMessage = "Please generate the title first."
		return false
	}
	if len(videoId) == 0 {
		errorMessage = "Please upload video first."
		return false
	}
	if !posted {
		confirmationMessage = fmt.Sprintf(
			"Use the following information to post it to https://news.ycombinator.com/submit manually.\n\nTitle:\n%s\nURL:\n%s",
			title, getYouTubeURL(videoId))
	}
	return getInputFromBool(posted)
}
