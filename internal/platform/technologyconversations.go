package platform

import "fmt"

func PostTechnologyConversations(title, description, videoId, gist, projectName, projectURL, relatedVideos string, getAdditionalInfoFromPath func(string, string, string, string) string, confirmationStyle interface{ Render(...string) string }) {
	message := "Use the following information to post it to https://wordpress.com/posts/technologyconversations.com manually."
	message += fmt.Sprintf("\n\nTitle:\n%s", title)
	message += fmt.Sprintf("\n\nDescription:\n%s", description)
	message += fmt.Sprintf("\n\nVideo ID:\n%s", videoId)
	message += fmt.Sprintf("\n\nAdditional info:\n%s", getAdditionalInfoFromPath(gist, projectName, projectURL, relatedVideos))
	println(confirmationStyle.Render(message))
}