package main

import "fmt"

func postHackerNews(title, videoId string) {
	message := fmt.Sprintf(
		"Use the following information to post it to https://news.ycombinator.com/submit manually.\n\nTitle:\n%s\nURL:\n%s",
		title,
		getYouTubeURL(videoId),
	)
	println(confirmationStyle.Render(message))
}
