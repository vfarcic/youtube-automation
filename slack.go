package main

import (
	"log"
	
	"devopstoolkitseries/youtube-automation/internal/configuration"
	"devopstoolkitseries/youtube-automation/internal/storage"
	"devopstoolkitseries/youtube-automation/pkg/slack"
	
	"github.com/atotto/clipboard"
)

func postSlack(videoId string) {
	// For backward compatibility - still copy to clipboard as backup
	clipboard.WriteAll(getYouTubeURL(videoId))
	
	// Get the video from storage - if we can't find it, fall back to old behavior
	video, found := findVideoByID(videoId)
	if !found {
		println(confirmationStyle.Render("The video URL has been copied to clipboard. Please paste it into Slack manually."))
		return
	}
	
	// Try to post to Slack with the API if token is configured
	if configuration.GlobalSettings.Slack.Token != "" {
		config := slack.Config{
			Token:         configuration.GlobalSettings.Slack.Token,
			DefaultChannel: configuration.GlobalSettings.Slack.DefaultChannel,
			Reactions:     configuration.GlobalSettings.Slack.Reactions,
		}
		
		if err := slack.PostMessage(config, video); err != nil {
			log.Printf("Failed to post to Slack: %v", err)
			println(errorStyle.Render("Failed to automatically post to Slack. The video URL has been copied to clipboard instead. Please paste it manually."))
		} else {
			// Update the video metadata with posting status
			yaml := storage.YAML{}
			yaml.WriteVideo(video, video.Path)
			
			println(confirmationStyle.Render("Successfully posted to Slack."))
		}
	} else {
		println(confirmationStyle.Render("The video URL has been copied to clipboard. Please paste it into Slack manually."))
		println("(Configure a Slack token to enable automatic posting)")
	}
}

// findVideoByID locates a video in the index based on its VideoId
func findVideoByID(videoId string) (storage.Video, bool) {
	// Load all videos from the index
	yaml := storage.YAML{IndexPath: "index.yaml"}
	videos := yaml.GetIndex()
	
	for _, vi := range videos {
		video := yaml.GetVideo(GetVideoPath(vi.Category, vi.Name))
		if video.VideoId == videoId {
			return video, true
		}
	}
	
	return storage.Video{}, false
}

// GetVideoPath constructs the path to a video's YAML file
func GetVideoPath(category, name string) string {
	return "manuscript/" + category + "/" + name + ".yaml"
}
