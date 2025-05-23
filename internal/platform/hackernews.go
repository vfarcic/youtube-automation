package platform

import "fmt"

func PostHackerNews(title, videoId string, getYouTubeURL func(string) string, confirmationStyle interface{ Render(...string) string }) {
	message := fmt.Sprintf(
		"Use the following information to post it to https://news.ycombinator.com/submit manually.\n\nTitle:\n%s\nURL:\n%s",
		title,
		getYouTubeURL(videoId),
	)
	println(confirmationStyle.Render(message))
}