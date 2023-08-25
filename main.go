package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// - [ ] Publish on YouTube
// - [ ] Publish on Twitter
// - [ ] Publish on LinkedIn
// - [ ] Publish on Slack
// - [ ] Publish on Reddit
// - [ ] Hacker News
// - [ ] Publish on TechnologyConversations.com
// - [ ] Add to YT spotlight
// - [ ] Add a comment to the video
// - [ ] Respond to comments
// - [ ] Add to slides
// - [ ] Add to https://gde.advocu.com
// - [ ] Modify repo README.md
// - [ ] Publish on a Twitter space
// - [ ] Convert to Crossplane
// - [ ] Email
// - [ ] Top X?
// - [ ] Add to cncf-demo

func main() {
	getArgs()
	video := readYaml(settings.Path)
	var err error
	for {
		video, err = modifyChoice(video)
		if err != nil {
			println(fmt.Sprintf("\n%s", err.Error()))
			continue
		}
		writeYaml(video, settings.Path)
	}
}

func getChoiceTextFromString(choice, value string) string {
	valueLength := len(value)
	if valueLength > 100 {
		value = fmt.Sprintf("%s...", value[0:100])
	}
	text := choice
	value = strings.ReplaceAll(value, "\n", " ")
	if value != "" && value != "-" && value != "N/A" {
		text = fmt.Sprintf("%s (%s)", text, value)
	}
	if value == "" {
		return orangeStyle.Render(text)
	}
	return greenStyle.Render(text)
}

func getChoiceTextFromPlaylists(choice string, values []Playlist) string {
	value := ""
	for i := range values {
		value = fmt.Sprintf("%s, %s", values[i].Title, value)
	}
	valueLength := len(value)
	if valueLength > 100 {
		value = fmt.Sprintf("%s...", value[0:100])
	}
	value = strings.TrimRight(value, ", ")
	text := choice
	value = strings.ReplaceAll(value, "\n", " ")
	if value != "" && value != "-" && value != "N/A" {
		text = fmt.Sprintf("%s (%s)", text, value)
	}
	if value == "" {
		return orangeStyle.Render(text)
	}
	return greenStyle.Render(text)
}

func getChoiceTextFromBool(choice string, value bool) string {
	if !value {
		return orangeStyle.Render(choice)
	}
	return greenStyle.Render(choice)
}

func getChoiceThumbnail(value bool, from, to string, video Video) bool {
	if value {
		return false
	}
	sendThumbnailEmail(from, to, video)
	return true
}

func getChoiceEdit(value bool, from, to string, video Video) bool {
	if value {
		return false
	}
	sendEditEmail(from, to, video)
	return true
}

func getChoiceUploadVideo(video Video) (string, string) {
	if len(video.UploadVideo) > 0 {
		return "", ""
	}
	video.UploadVideo, _ = getInputFromString("What is the path to the video?", video.UploadVideo)
	video.VideoId = uploadVideo(video)
	uploadThumbnail(video)
	// err := setPlaylists(video)
	// if err != nil {
	// 	println(redStyle.Render(fmt.Sprintf("Error setting playlists: %s", err.Error())))
	// }
	println(redStyle.Render(`Following should be set manually:
	- End screen
	- Playlists
	- Tags
	- Language
	- License
	- Monetization
	`))
	return video.UploadVideo, video.VideoId
}

func getChoicePlaylists() []Playlist {
	choices := make(map[int]string)
	index := 0
	for _, item := range getPlaylists() {
		choices[index] = item
		index += 1
	}
	selectedMap := getChoices(choices, "Select playlists")
	selected := []Playlist{}
	for _, value := range selectedMap {
		if len(value) > 0 {
			id := strings.Split(value, " - ")[1]
			title := strings.Split(value, " - ")[0]
			playlist := Playlist{Title: title, Id: id}
			selected = append(selected, playlist)
		}
	}
	return selected
}

func modifyAnimations(video Video) (string, error) {
	println()
	if len(video.ProjectName) == 0 {
		return video.Animations, fmt.Errorf(redStyle.Render("Project name was not specified!"))
	}
	if len(video.ProjectURL) == 0 {
		return video.Animations, fmt.Errorf(redStyle.Render("Project URL was not specified!"))
	}
	if len(video.Title) == 0 {
		return video.Animations, fmt.Errorf(redStyle.Render("Video title was not specified!"))
	}
	title := `Write animation bullets.

Suggested bullets:
- Thumbnails: ([[TODO]]) + text "The link is in the description" + an arrow pointing below
- Logo: [[TODO]]
- Section: [[TODO]]
- Text: [[TODO]]
- Text: [[TODO]] (big)
- Plug: [[TODO]] + logo + URL ([[TODO]]) (use their website for animations or screenshots; make it look different from the main video; I'll let you know where to put it once the main video is ready)
- Diagram: [[TODO]]
- Header: Cons; Items: [[TODO]]
- Header: Pros; Items: [[TODO]]
`
	if len(video.Animations) == 0 {
		video.Animations = fmt.Sprintf(`- Animation: Subscribe (anywhere in the video)
- Animation: Like (anywhere in the video)
- Lower third: Viktor Farcic (anywhere in the video)
- Animation: Join the channel (anywhere in the video)
- Animation: Sponsor the channel (anywhere in the video)
- Lower third: %s + logo + URL (%s) (add to a few places when I mention %s)
- Text: Gist with the commands + an arrow pointing below (add shortly after we start showing the code)
- Title roll: %s
- Member shoutouts: Thanks a ton to the new members for supporting the channel: %s
- Outro roll
`,
			video.ProjectName,
			video.ProjectURL,
			video.ProjectName,
			video.Title,
			video.Members)
	}
	return getInputFromTextArea(title, video.Animations, 100), nil
}

func setThumbnail(path string) (string, error) {
	if len(path) == 0 {
		path = fmt.Sprintf("%s/", filepath.Dir(settings.Path))
	}
	path, err := getInputFromString("What is the path to the thumbnail?", path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf(redStyle.Render("File does not exist!"))
	}
	return path, nil
}
