package main

import "fmt"

func postTechnologyConversations(title, description, videoId, gist, projectName, projectURL, relatedVideos string, posted bool) bool {
	// if len(title) == 0 {
	// 	errorMessage = "Please generate the title first."
	// 	return false
	// }
	// if len(description) == 0 {
	// 	errorMessage = "Please generate the description first."
	// 	return false
	// }
	// if len(videoId) == 0 {
	// 	errorMessage = "Please upload video first."
	// 	return false
	// }
	// if len(gist) == 0 {
	// 	errorMessage = "Please set the Gist first."
	// 	return false
	// }
	if !posted {
		confirmationMessage = "Use the following information to post it to https://wordpress.com/posts/technologyconversations.com manually."
		confirmationMessage += fmt.Sprintf("\n\nTitle:\n%s", title)
		confirmationMessage += fmt.Sprintf("\n\nDescription:\n%s", description)
		confirmationMessage += fmt.Sprintf("\n\nVideo ID:\n%s", videoId)
		confirmationMessage += fmt.Sprintf("\n\nAdditional info:\n%s", getAdditionalInfo(gist, projectName, projectURL, relatedVideos))
	}
	return getInputFromBool(posted)
}
