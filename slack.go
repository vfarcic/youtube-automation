package main

import (
	"log"
	
	"devopstoolkitseries/youtube-automation/internal/configuration"
	"devopstoolkitseries/youtube-automation/internal/storage"
	"devopstoolkitseries/youtube-automation/pkg/slack"
	
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
)

// Define styles for consistent UI rendering
var (
	confirmationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))  // Green text for success
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))   // Red text for errors
)

func postSlack(videoId string) {
	// For backward compatibility - still copy to clipboard as backup
	clipboard.WriteAll(getYouTubeURL(videoId))
	
	// Check if Slack token is configured
	if configuration.GlobalSettings.Slack.Token == "" {
		println(confirmationStyle.Render("The video URL has been copied to clipboard. Please paste it into Slack manually."))
		println("(Configure a Slack token to enable automatic posting)")
		return
	}
	
	// Create a dummy video object with just the VideoId if we can't find it
	video := storage.Video{
		VideoId: videoId,
		Title:   "YouTube Video",
	}
	
	// Try to find the complete video by ID for better formatting
	if completeVideo, found := findVideoByID(videoId); found {
		video = completeVideo
	}
	
	// Post to Slack with the API
	config := slack.Config{
		Token:         configuration.GlobalSettings.Slack.Token,
		DefaultChannel: configuration.GlobalSettings.Slack.DefaultChannel,
		Reactions:     configuration.GlobalSettings.Slack.Reactions,
	}
	
	if err := slack.PostMessage(config, video); err != nil {
		log.Printf("Failed to post to Slack: %v", err)
		println(errorStyle.Render("Failed to automatically post to Slack. The video URL has been copied to clipboard instead. Please paste it manually."))
	} else {
		// Update the video metadata with posting status if we found the complete video
		if video.Path != "" {
			yaml := storage.YAML{}
			yaml.WriteVideo(video, video.Path)
		}
		
		println(confirmationStyle.Render("Successfully posted to Slack."))
	}
}

// findVideoByID locates a video in the index based on its VideoId
func findVideoByID(videoId string) (storage.Video, bool) {
	// Load all videos from the index
	yaml := storage.YAML{IndexPath: "index.yaml"}
	videos := yaml.GetIndex()
	
	for _, vi := range videos {
		videoPath := "manuscript/" + vi.Category + "/" + vi.Name + ".yaml"
		video := yaml.GetVideo(videoPath)
		if video.VideoId == videoId {
			return video, true
		}
	}
	
	return storage.Video{}, false
}
