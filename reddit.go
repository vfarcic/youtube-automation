package main

import "fmt"

func postReddit(title, videoId string) {
	message := "Use the following information to post it to https://reddit.com manually."
	message += fmt.Sprintf("\n\nTitle:\n%s\n\nURL:\n%s", title, getYouTubeURL(videoId))
	println(confirmationStyle.Render(message))
}
