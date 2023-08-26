package main

import "fmt"

func postHackerNews(title, videoId string, posted bool) bool {
	if len(title) == 0 {
		println(redStyle.Render("\nPlease generate the title first."))
		return false
	}
	if len(videoId) == 0 {
		println(redStyle.Render("\nPlease upload video first."))
		return false
	}
	if !posted {
		println(orangeStyle.Render("\nUse the following information to post it to https://news.ycombinator.com/submit manually."))
		println(fmt.Sprintf("Title:\n%s\nURL:\n%s", title, getYouTubeURL(videoId)))
	}
	return getInputFromBool(posted)
}
