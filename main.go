package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	getArgs()
	// TODO: Add the screen to:
	// - add new videos
	// - list videos by category
	// - edit videos
	yaml := YAML{}
	video := yaml.GetVideo(settings.Path)
	for {
		choices := Choices{}
		video = choices.ChoosePhase(video)
		yaml.WriteVideo(video, settings.Path)
	}
}

func getChoiceTextFromString(title, value string) Task {
	task := Task{Title: title, Completed: false}
	valueLength := len(value)
	if valueLength > 100 {
		value = fmt.Sprintf("%s...", value[0:100])
	}
	value = strings.ReplaceAll(value, "\n", " ")
	if value != "" && value != "-" && value != "N/A" {
		task.Title = fmt.Sprintf("%s (%s)", task.Title, value)
	}
	if len(value) > 0 {
		task.Completed = true
	}
	return task
}

func colorize(task Task) Task {
	if task.Completed {
		task.Title = greenStyle.Render(task.Title)
	} else {
		task.Title = orangeStyle.Render(task.Title)
	}
	return task
}

func getChoiceTextFromSponsoredEmails(title, sponsored string, sponsoredEmails []string) Task {
	task := Task{Title: title, Completed: false}
	if len(sponsoredEmails) > 0 {
		emailsText := strings.Join(sponsoredEmails, ", ")
		task.Title = fmt.Sprintf("%s (%s)", task.Title, emailsText)
		if len(task.Title) > 100 {
			task.Title = fmt.Sprintf("%s...", task.Title[0:100])
		}
		task.Completed = true
	} else if len(sponsored) == 0 || sponsored == "N/A" || sponsored == "-" {
		task.Completed = true
	}
	return task
}

func getChoiceTextFromPlaylists(title string, values []Playlist) Task {
	task := Task{Title: title, Completed: false}
	value := ""
	for i := range values {
		value = fmt.Sprintf("%s, %s", values[i].Title, value)
	}
	valueLength := len(value)
	if valueLength > 100 {
		value = fmt.Sprintf("%s...", value[0:100])
	}
	value = strings.TrimRight(value, ", ")
	value = strings.ReplaceAll(value, "\n", " ")
	if value != "" && value != "-" && value != "N/A" {
		task.Title = fmt.Sprintf("%s (%s)", task.Title, value)
	}
	if value != "" {
		task.Completed = true
	}
	return task
}

func getChoiceTextFromBool(title string, value bool) Task {
	return Task{Title: title, Completed: value}
}

func getChoiceThumbnail(value bool, from, to string, video Video) bool {
	if value {
		return false
	}
	if sendThumbnailEmail(from, to, video) != nil {
		return false
	}
	return true
}

func getChoiceNotifySponsors(title, sponsored string, notified bool) Task {
	task := Task{Title: title, Completed: false}
	if notified || len(sponsored) == 0 || sponsored == "N/A" || sponsored == "-" {
		task.Completed = true
	}
	return task
}

func requestEdit(value bool, from, to string, video Video) bool {
	if value {
		return false
	}
	if sendEditEmail(from, to, video) != nil {
		return false
	}
	return true
}

func notifySponsors(to []string, videoID, sponsorshipPrice string, value bool) bool {
	if value {
		return false
	}
	sendSponsorsEmail(settings.Email.From, to, videoID, sponsorshipPrice)
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
	confirmationMessage = `Following should be set manually:
- End screen
- Playlists
- Tags
- Language
- Monetization`
	return video.UploadVideo, video.VideoId
}

func writeSponsoredEmails(emails []string) []string {
	emailsString := ""
	for i := range emails {
		emailsString = fmt.Sprintf("%s\n%s", emailsString, emails[i])
	}
	emailsString, _ = modifyTextArea(emailsString, "Write emails that should be sent to sponsors separate with new lines:", "")
	return deleteEmpty(strings.Split(emailsString, "\n"))
}

func getPlaylists() []Playlist {
	choices := make(map[int]string)
	index := 0
	for _, item := range getYouTubePlaylists() {
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

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
