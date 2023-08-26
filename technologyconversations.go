package main

import "fmt"

func postTechnologyConversations(title, description, videoId, gist, relatedVideos string, posted bool) bool {
	if len(title) == 0 {
		println(redStyle.Render("\nPlease generate the title first."))
		return false
	}
	if len(description) == 0 {
		println(redStyle.Render("\nPlease generate the description first."))
		return false
	}
	if len(videoId) == 0 {
		println(redStyle.Render("\nPlease upload video first."))
		return false
	}
	if len(gist) == 0 {
		println(redStyle.Render("\nPlease set the Gist first."))
		return false
	}
	if !posted {
		println(orangeStyle.Render("\nUse the following information to post it to https://wordpress.com/posts/technologyconversations.com manually."))
		println(fmt.Sprintf("Title:\n%s", title))
		println(fmt.Sprintf("Description:\n%s", description))
		println(fmt.Sprintf("Video ID:\n%s", videoId))
		println(fmt.Sprintf("Additional info:\n%s", getAdditionalInfo(gist, relatedVideos)))
	}
	return getInputFromBool(posted)
}
