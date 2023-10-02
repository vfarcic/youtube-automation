package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Choices struct{}

var redStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("1"))

var greenStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("2"))

var orangeStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("3"))

var errorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(lipgloss.Color("#B60A02")).
	PaddingTop(1).
	PaddingBottom(1).
	PaddingLeft(2).
	PaddingRight(2).
	MarginTop(1).
	MarginBottom(1)

var confirmationStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(lipgloss.Color("#006E14")).
	PaddingTop(1).
	PaddingBottom(1).
	MarginTop(1).
	MarginBottom(1)

var titleStyle = lipgloss.NewStyle().
	MarginTop(2).
	MarginBottom(1)

const indexCreateVideo = 0
const indexListVideos = 1
const indexExit = 2

const videosPhasePublished = 0
const videosPhasePublishPending = 1
const videosPhaseEditRequested = 2
const videosPhaseMaterialDone = 3
const videosPhaseStarted = 4
const videosPhaseSponsoredBlocked = 5
const videosPhaseIdeas = 6
const videosPhaseReturn = 7

const phasePrePublish = 0
const phasePublish = 1
const phaseReturn = 2

const prePublishProjectName = 0
const prePublishProjectURL = 1
const prePublishSponsored = 2
const prePublishSponsoredEmails = 3
const prePublishSponsorshipBlocked = 4
const prePublishSubject = 5
const prePublishDate = 6
const prePublishCode = 7
const prePublishScreen = 8
const prePublishHead = 9
const prePublishRelatedVideos = 10
const prePublishThumbnails = 11
const prePublishDiagrams = 12
const prePublishLocation = 13
const prePublishTagline = 14
const prePublishTaglineIdeas = 15
const prePublishOtherLogos = 16
const prePublishScreenshots = 17
const prePublishGenerateTitle = 18
const prePublishModifyTitle = 19
const prePublishGenerateDescription = 20
const prePublishModifyDescription = 21
const prePublishGenerateTags = 22
const prePublishModifyTags = 23
const prePublishModifyDescriptionTags = 24
const prePublishRequestThumbnail = 25
const prePublishMembers = 26
const prePublishAnimations = 27
const prePublishRequestEdit = 28
const prePublishThumbnail = 29
const prePublishGotMovie = 30
const prePublishTimecodes = 31
const prePublishGist = 32
const prePublishPlaylists = 33
const prePublishReturn = 34

const publishUploadVideo = 0
const publishGenerateTweet = 1
const publishModifyTweet = 2
const publishTweetPosted = 3
const publishLinkedInPosted = 4
const publishSlackPosted = 5
const publishRedditPosted = 6
const publishHNPosted = 7
const publishTCPosted = 8
const publishYouTubeHighlight = 9
const publishYouTubeComment = 10
const publishYouTubeCommentReply = 11
const publishSlides = 12
const publishGDE = 13
const publishTwitterSpace = 14
const publishRepo = 15
const publishNotifySponsors = 16
const publishReturn = 17

type Tasks struct {
	Completed int
	Total     int
}

type Task struct {
	Title     string
	Completed bool
	Counter   int
	Index     int
}

type Playlist struct {
	Title string
	Id    string
}

func (c *Choices) ChooseIndex() {
	yaml := YAML{IndexPath: "index.yaml"}
	index := yaml.GetIndex()
	tasks := map[int]Task{
		indexCreateVideo: {Title: "Create a video"},
		indexListVideos:  {Title: "List videos"},
		indexExit:        {Title: "Exit"},
	}
	option, _ := getChoice(tasks, titleStyle.Render("What would you like to do?"))
	switch option {
	case indexCreateVideo:
		item := c.ChooseCreateVideo()
		if len(item.Category) > 0 && len(item.Name) > 0 {
			index = append(index, item)
			yaml.WriteIndex(index)
		}
	case indexListVideos:
		for {
			returnVal := c.ChooseVideosPhase(index)
			if returnVal {
				break
			}
		}
	case indexExit:
		os.Exit(0)
	}
}

func (c *Choices) ChoosePhase(video Video) (Video, bool) {
	returnVar := false
	prePublish := Task{
		Title:     "Pre-publish",
		Completed: video.PrePublish.Completed == video.PrePublish.Total && video.PrePublish.Total > 0,
	}
	if video.PrePublish.Total > 0 {
		prePublish.Title = fmt.Sprintf("%s (%d/%d)", prePublish.Title, video.PrePublish.Completed, video.PrePublish.Total)
	}
	publish := Task{
		Title:     "Publish",
		Completed: video.Publish.Completed == video.Publish.Total && video.Publish.Total > 0,
	}
	if video.Publish.Total > 0 {
		publish.Title = fmt.Sprintf("%s (%d/%d)", publish.Title, video.Publish.Completed, video.Publish.Total)
	}
	tasks := map[int]Task{
		phasePrePublish: colorize(prePublish),
		phasePublish:    colorize(publish),
		phaseReturn:     {Title: "Return"},
	}
	option, _ := getChoice(tasks, titleStyle.Render("Would you like to work on pre-publish or publish tasks?"))
	switch option {
	case phasePrePublish:
		var err error
		for !returnVar {
			video, returnVar, err = c.ChoosePrePublish(video)
			if err != nil {
				errorMessage = err.Error()
				continue
			}
		}
	case phasePublish:
		var err error
		for !returnVar {
			video, returnVar, err = c.ChoosePublish(video)
			if err != nil {
				errorMessage = err.Error()
				continue
			}
		}
	case phaseReturn:
		returnVar = true
	}
	return video, returnVar
}

func (c *Choices) ChooseCreateVideo() VideoIndex {
	const name = "1. Name"
	const category = "2. Category"
	qa := map[string]string{
		name:     "",
		category: "",
	}
	m, _ := getMultipleInputsFromString(qa)
	vi := VideoIndex{}
	for k, v := range m {
		switch k {
		case name:
			vi.Name = v
		case category:
			vi.Category = v
		}
	}
	if vi.Name == "" || vi.Category == "" {
		errorMessage = "Name and category are required!"
		return vi
	}
	dirPath := c.GetDirPath(vi.Category)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0755)
	}
	scriptContent := `
#############
# [[title]] #
# [[url]]   #
#############

# Additional Info:
# - [[additional-info]]

#########
# Intro #
#########

# TODO: Intro (Sara)

# TODO: Title screen

#########
# Setup #
#########

# TODO:

##########
# TODO:: #
##########

# TODO: Gist

# TODO: Commands

#######################
# TODO: Pros And Cons #
#######################

# Cons:
# - TODO:

# Pros:
# - TODO:

###########
# Destroy #
###########

# TODO:
`
	filePath := c.GetFilePath(vi.Category, vi.Name, "sh")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		f, err := os.Create(filePath)
		if err != nil {
			errorMessage = err.Error()
			return VideoIndex{}
		}
		defer f.Close()
		f.Write([]byte(scriptContent))
		return vi
	}
	return VideoIndex{}
}

func (c *Choices) GetDirPath(category string) string {
	return fmt.Sprintf("manuscript/%s", strings.ReplaceAll(strings.ToLower(category), " ", "-"))
}

func (c *Choices) GetFilePath(category, name, extension string) string {
	dirPath := c.GetDirPath(category)
	filePath := fmt.Sprintf("%s/%s.%s", dirPath, strings.ToLower(name), extension)
	filePath = strings.ReplaceAll(filePath, " ", "-")
	filePath = strings.ReplaceAll(filePath, "?", "")
	return filePath
}

func (c *Choices) ChoosePrePublish(video Video) (Video, bool, error) {
	openAI := OpenAI{}
	returnVar := false
	sponsorshipBlockedTask := getChoiceTextFromString("Sponsorship blocked?", video.SponsorshipBlocked)
	if len(video.SponsorshipBlocked) > 0 {
		sponsorshipBlockedTask.Title = redStyle.Render(sponsorshipBlockedTask.Title)
	} else {
		sponsorshipBlockedTask.Title = greenStyle.Render(sponsorshipBlockedTask.Title)
	}
	sponsorshipBlockedTask.Completed = !sponsorshipBlockedTask.Completed
	tasks := map[int]Task{
		prePublishProjectName:           colorize(getChoiceTextFromString("Set project name", video.ProjectName)),
		prePublishProjectURL:            colorize(getChoiceTextFromString("Set project URL", video.ProjectURL)),
		prePublishSponsored:             colorize(getChoiceTextFromString("Set sponsorship", video.Sponsored)),
		prePublishSponsoredEmails:       colorize(getChoiceTextFromSponsoredEmails("Set sponsorship emails", video.Sponsored, video.SponsoredEmails)),
		prePublishSponsorshipBlocked:    sponsorshipBlockedTask,
		prePublishSubject:               colorize(getChoiceTextFromString("Set the subject", video.Subject)),
		prePublishDate:                  colorize(getChoiceTextFromString("Set publish date", video.Date)),
		prePublishCode:                  colorize(getChoiceTextFromBool("Wrote code?", video.Code)),
		prePublishScreen:                colorize(getChoiceTextFromBool("Recorded screen?", video.Screen)),
		prePublishHead:                  colorize(getChoiceTextFromBool("Recorded talking head?", video.Head)),
		prePublishRelatedVideos:         colorize(getChoiceTextFromString("Set related videos", video.RelatedVideos)),
		prePublishThumbnails:            colorize(getChoiceTextFromBool("Downloaded thumbnails?", video.Thumbnails)),
		prePublishDiagrams:              colorize(getChoiceTextFromBool("Created diagrams?", video.Diagrams)),
		prePublishLocation:              colorize(getChoiceTextFromString("Set files location", video.Location)),
		prePublishTagline:               colorize(getChoiceTextFromString("Set tagline", video.Tagline)),
		prePublishTaglineIdeas:          colorize(getChoiceTextFromString("Set tagline ideas", video.TaglineIdeas)),
		prePublishOtherLogos:            colorize(getChoiceTextFromString("Set other logos", video.OtherLogos)),
		prePublishScreenshots:           colorize(getChoiceTextFromBool("Created screenshots?", video.Screenshots)),
		prePublishGenerateTitle:         colorize(getChoiceTextFromString("Generate title", video.Title)),
		prePublishModifyTitle:           colorize(getChoiceTextFromString("Write/modify title", video.Title)),
		prePublishGenerateDescription:   colorize(getChoiceTextFromString("Generate description", video.Description)),
		prePublishModifyDescription:     colorize(getChoiceTextFromString("Write/modify description", video.Description)),
		prePublishGenerateTags:          colorize(getChoiceTextFromString("Generate tags", video.Tags)),
		prePublishModifyTags:            colorize(getChoiceTextFromString("Write/modify tags", video.Tags)),
		prePublishModifyDescriptionTags: colorize(getChoiceTextFromString("Write/modify description tags", video.DescriptionTags)),
		prePublishRequestThumbnail:      colorize(getChoiceTextFromBool("Request thumbnail", video.RequestThumbnail)),
		prePublishMembers:               colorize(getChoiceTextFromString("Set members", video.Members)),
		prePublishAnimations:            colorize(getChoiceTextFromString("Write/modify animations", video.Animations)),
		prePublishRequestEdit:           colorize(getChoiceTextFromBool("Request edit", video.RequestEdit)),
		prePublishThumbnail:             colorize(getChoiceTextFromString("Got thumbnail?", video.Thumbnail)),
		prePublishGotMovie:              colorize(getChoiceTextFromBool("Got movie?", video.Movie)),
		prePublishTimecodes:             colorize(getChoiceTextFromString("Set timecodes", video.Timecodes)),
		prePublishGist:                  colorize(getChoiceTextFromString("Set gist", video.Gist)),
		prePublishPlaylists:             colorize(getChoiceTextFromPlaylists("Set playlists", video.Playlists)),
		prePublishReturn:                {Title: "Save and return"},
	}
	completed := 0
	for _, task := range tasks {
		if task.Completed {
			completed++
		}
	}
	video.PrePublish = Tasks{Total: len(tasks) - 1, Completed: completed}
	choice, _ := getChoice(tasks, titleStyle.Render("Which pre-publish task would you like to work on?"))
	err := error(nil)
	switch choice {
	case prePublishProjectName:
		video.ProjectName, err = getInputFromString("Set project name", video.ProjectName)
	case prePublishProjectURL:
		video.ProjectURL, err = getInputFromString("Set project URL", video.ProjectURL)
	case prePublishSponsored:
		video.Sponsored, err = getInputFromString("Sponsorship amount ('-' or 'N/A' if not sponsored)", video.Sponsored)
	case prePublishSponsoredEmails:
		video.SponsoredEmails = writeSponsoredEmails(video.SponsoredEmails)
	case prePublishSponsorshipBlocked:
		video.SponsorshipBlocked, err = getInputFromString(video.SponsorshipBlocked, video.SponsorshipBlocked)
	case prePublishSubject:
		video.Subject, err = getInputFromString("What is the subject of the video?", video.Subject)
	case prePublishDate:
		video.Date, err = getInputFromString("What is the publish of the video (e.g., 2030-01-21T16:00)?", video.Date)
	case prePublishCode:
		video.Code = getInputFromBool(video.Code)
	case prePublishScreen:
		video.Screen = getInputFromBool(video.Screen)
	case prePublishHead:
		video.Head = getInputFromBool(video.Head)
	case prePublishRelatedVideos:
		video.RelatedVideos = getInputFromTextArea("What are the related videos?", video.RelatedVideos, 20)
	case prePublishThumbnails:
		video.Thumbnails = getInputFromBool(video.Thumbnails)
	case prePublishDiagrams:
		video.Diagrams = getInputFromBool(video.Diagrams)
	case prePublishLocation:
		video.Location, err = getInputFromString("Where are files located?", video.Location)
	case prePublishTagline:
		video.Tagline, err = getInputFromString("What is the tagline?", video.Tagline)
	case prePublishTaglineIdeas:
		video.TaglineIdeas, err = getInputFromString("What are the tagline ideas?", video.TaglineIdeas)
	case prePublishOtherLogos:
		video.OtherLogos, err = getInputFromString("What are the other logos?", video.OtherLogos)
	case prePublishScreenshots:
		video.Screenshots = getInputFromBool(video.Screenshots)
	case prePublishGenerateTitle:
		video, err = openAI.GenerateTitle(video)
	case prePublishModifyTitle:
		video.Title, err = modifyTextArea(video.Title, "Rewrite the title:", "Title was not generated!")
	case prePublishGenerateDescription:
		video, err = openAI.GenerateDescription(video)
	case prePublishModifyDescription:
		video.Description, err = modifyTextArea(video.Description, "Modify video description:", "Description was not generated!")
	case prePublishGenerateTags: // TODO: Add default tags like "viktor farcic", "DevOps", etc.
		video.Tags, err = openAI.GenerateTags(video.Title)
	case prePublishModifyTags:
		video.Tags, err = modifyTextArea(video.Tags, "Modify tags:", "Tags were not generated!")
	case prePublishModifyDescriptionTags:
		video.DescriptionTags, err = modifyDescriptionTagsX(video.Tags, video.DescriptionTags, "Modify description tags (max 4):", "Description tags were not generated!")
	case prePublishRequestThumbnail:
		video.RequestThumbnail = getChoiceThumbnail(video.RequestThumbnail, settings.Email.From, settings.Email.ThumbnailTo, video)
	case prePublishThumbnail:
		video.Thumbnail, err = setThumbnail(video.Thumbnail)
	case prePublishMembers:
		video.Members, err = getInputFromString("Who are new members?", video.Members)
	case prePublishAnimations:
		video.Animations, err = modifyAnimations(video)
	case prePublishRequestEdit:
		video.RequestEdit = requestEdit(video.RequestEdit, settings.Email.From, settings.Email.EditTo, video)
	case prePublishGotMovie:
		video.Movie = getInputFromBool(video.Movie)
	case prePublishTimecodes:
		video.Timecodes = getInputFromTextArea("What are timecodes?", video.Timecodes, 20)
	case prePublishGist: // TODO: Ask for the Gist path, create it, and store both the path and the URL.
		video.Gist, err = getInputFromString("Where is the gist?", video.Gist)
	case prePublishPlaylists:
		video.Playlists = getPlaylists()
	case prePublishReturn:
		returnVar = true
	}
	return video, returnVar, err
}

func (c *Choices) ChooseVideosPhase(vi []VideoIndex) bool {
	tasks := map[int]Task{
		videosPhasePublished:        {Title: "Published"},
		videosPhasePublishPending:   {Title: "Pending publish"},
		videosPhaseEditRequested:    {Title: "Edit requested"},
		videosPhaseMaterialDone:     {Title: "Material done"},
		videosPhaseStarted:          {Title: "Started"},
		videosPhaseSponsoredBlocked: {Title: "Sponsored blocked"},
		videosPhaseIdeas:            {Title: "Ideas"},
		videosPhaseReturn:           {Title: "Return"},
	}
	for i := range vi {
		phase := c.GetVideoPhase(vi[i])
		task := tasks[phase]
		task.Counter++
		tasks[phase] = task
	}
	for key := range tasks {
		task := tasks[key]
		if key != videosPhaseReturn {
			task.Title = fmt.Sprintf("%s (%d)", task.Title, task.Counter)
			if key == videosPhasePublished {
				task.Title = greenStyle.Render(task.Title)
			} else if key == videosPhasePublishPending && task.Counter > 0 {
				task.Title = greenStyle.Render(task.Title)
			} else if key == videosPhaseEditRequested && task.Counter > 0 {
				task.Title = greenStyle.Render(task.Title)
			} else if key == videosPhaseMaterialDone && task.Counter >= 3 {
				task.Title = greenStyle.Render(task.Title)
			} else if task.Counter == 0 {
				task.Title = greenStyle.Render(task.Title)
			} else {
				task.Title = orangeStyle.Render(task.Title)
			}
			tasks[key] = task
		}
	}
	choice, _ := getChoice(tasks, titleStyle.Render("From which phase would you like to list the videos?"))
	if choice == videosPhaseReturn {
		return true
	}
	c.ChooseVideos(vi, choice)
	return false
}

func (c *Choices) GetVideoPhase(vi VideoIndex) int {
	yaml := YAML{}
	video := yaml.GetVideo(c.GetFilePath(vi.Category, vi.Name, "yaml"))
	if len(video.SponsorshipBlocked) > 0 {
		return videosPhaseSponsoredBlocked
	} else if len(video.Repo) > 0 {
		return videosPhasePublished
	} else if len(video.UploadVideo) > 0 && len(video.Tweet) > 0 {
		return videosPhasePublishPending
	} else if video.RequestEdit {
		return videosPhaseEditRequested
	} else if video.Code && video.Screen && video.Head && video.Thumbnails && video.Diagrams {
		return videosPhaseMaterialDone
	} else if len(video.Date) > 0 {
		return videosPhaseStarted
		// TODO: Remove
		// } else if len(vi.Phase) == 0 {
	} else {
		return videosPhaseIdeas
	}
	// return videosPhaseReturn
}

func (c *Choices) ChooseVideos(vi []VideoIndex, phase int) {
	tasks := make(map[int]Task)
	index := 0
	for i := range vi {
		videoIndex := vi[i]
		videoPhase := c.GetVideoPhase(videoIndex)
		if videoPhase == phase {
			title := videoIndex.Name
			yaml := YAML{}
			path := c.GetFilePath(videoIndex.Category, videoIndex.Name, "yaml")
			video := yaml.GetVideo(path)
			if len(video.SponsorshipBlocked) > 0 && video.SponsorshipBlocked != "-" && video.SponsorshipBlocked != "N/A" {
				title = fmt.Sprintf("%s (%s)", title, video.SponsorshipBlocked)
			} else {
				if len(video.Date) > 0 {
					title = fmt.Sprintf("%s (%s)", title, video.Date)
				}
				if len(video.Sponsored) > 0 && video.Sponsored != "-" && video.Sponsored != "N/A" {
					title = fmt.Sprintf("%s (sponsored)", title)
				}
			}
			tasks[index] = Task{Title: title, Index: i}
			index++
		}
	}
	tasks[len(tasks)] = Task{Title: "Return"}
	choice, _ := getChoice(tasks, titleStyle.Render("Which video would you like to work on?"))
	selectedTask := tasks[choice]
	if choice == len(tasks)-1 {
		return
	}
	selectedVideoIndex := vi[selectedTask.Index]
	selectedPhase := c.GetVideoPhase(selectedVideoIndex)
	video, _ := c.VideoTasks(selectedVideoIndex, selectedPhase)
	if len(video.Name) == 0 {
		err := os.Remove(c.GetFilePath(selectedVideoIndex.Category, selectedVideoIndex.Name, "sh"))
		if err != nil {
			errorMessage = err.Error()
		}
		os.Remove(c.GetFilePath(selectedVideoIndex.Category, selectedVideoIndex.Name, "yaml"))
		selectedVideoIndex = vi[len(vi)-1]
		vi = vi[:len(vi)-1]
	} else {
		selectedVideoIndex = video
	}
	yaml := YAML{IndexPath: "index.yaml"}
	yaml.WriteIndex(vi)
}

func (c *Choices) VideoTasks(vi VideoIndex, phase int) (VideoIndex, bool) {
	const edit = 0
	const delete = 1
	const back = 2
	tasks := map[int]Task{
		edit:   {Title: "Edit"},
		delete: {Title: "Delete (TODO: fix it)"},
		back:   {Title: "Return"},
	}
	question := fmt.Sprintf("What would you like to do with '%s'?", vi.Name)
	choice, _ := getChoice(tasks, titleStyle.Render(question))
	switch choice {
	case edit:
		returnVal := false
		path := c.GetFilePath(vi.Category, vi.Name, "yaml")
		yaml := YAML{}
		video := yaml.GetVideo(path)
		for !returnVal {
			choices := Choices{}
			video, returnVal = choices.ChoosePhase(video)
			yaml.WriteVideo(video, path)
		}
	case delete:
		return VideoIndex{}, true
	case back:
		return vi, true
	}
	return vi, false
}

func (c *Choices) ChoosePublish(video Video) (Video, bool, error) {
	openAI := OpenAI{}
	returnVar := false
	tasks := map[int]Task{
		publishUploadVideo: colorize(getChoiceTextFromString("Upload video", video.UploadVideo)),
		// TODO: Add a new option to Update the Gist with Gist and Video URL
		publishGenerateTweet:       colorize(getChoiceTextFromString("Generate Tweet", video.Tweet)),
		publishModifyTweet:         colorize(getChoiceTextFromString("Write/modify Tweet", video.Tweet)),
		publishTweetPosted:         colorize(getChoiceTextFromBool("Post to Twitter (MANUAL)", video.TweetPosted)),                   // TODO:
		publishLinkedInPosted:      colorize(getChoiceTextFromBool("Post to LinkedIn (MANUAL)", video.LinkedInPosted)),               // TODO:
		publishSlackPosted:         colorize(getChoiceTextFromBool("Post to Slack (MANUAL)", video.SlackPosted)),                     // TODO:
		publishRedditPosted:        colorize(getChoiceTextFromBool("Post to Reddit (MANUAL)", video.RedditPosted)),                   // TODO:
		publishHNPosted:            colorize(getChoiceTextFromBool("Post to Hacker News (MANUAL)", video.HNPosted)),                  // TODO:
		publishTCPosted:            colorize(getChoiceTextFromBool("Post to Technology Conversations (MANUAL)", video.TCPosted)),     // TODO:
		publishYouTubeHighlight:    colorize(getChoiceTextFromBool("Set as YouTube Highlight (MANUAL)", video.YouTubeHighlight)),     // TODO:
		publishYouTubeComment:      colorize(getChoiceTextFromBool("Write pinned comment (MANUAL)", video.YouTubeComment)),           // TODO:
		publishYouTubeCommentReply: colorize(getChoiceTextFromBool("Write replies to comments (MANUAL)", video.YouTubeCommentReply)), // TODO:
		publishSlides:              colorize(getChoiceTextFromBool("Added to slides?", video.Slides)),
		publishGDE:                 colorize(getChoiceTextFromBool("Add to https://gde.advocu.com (MANUAL)", video.GDE)),     // TODO:
		publishTwitterSpace:        colorize(getChoiceTextFromBool("Post to a Twitter Spaces (MANUAL)", video.TwitterSpace)), // TODO:
		publishRepo:                colorize(getChoiceTextFromString("Set code repo", video.Repo)),                           // TODO:
		publishNotifySponsors:      colorize(getChoiceNotifySponsors("Notify sponsors", video.Sponsored, video.NotifiedSponsors)),
		publishReturn:              {Title: "Save and return"},
	}
	completed := 0
	for _, task := range tasks {
		if task.Completed {
			completed++
		}
	}
	video.Publish = Tasks{Total: len(tasks) - 1, Completed: completed}
	choice, _ := getChoice(tasks, titleStyle.Render("Which publish task would you like to work on?"))
	err := error(nil)
	switch choice {
	case publishUploadVideo: // TODO: Finish implementing End screen, Playlists, Tags, Language, License, Monetization
		video.UploadVideo, video.VideoId = getChoiceUploadVideo(video)
	case publishGenerateTweet:
		video.Tweet, err = openAI.GenerateTweet(video.Title, video.VideoId)
	case publishModifyTweet:
		video.Tweet, err = modifyTextArea(video.Tweet, "Modify tweet:", "Tweet was not generated!")
	case publishTweetPosted: // TODO: Automate
		twitter := Twitter{}
		video.TweetPosted = twitter.Post(video.Tweet, video.TweetPosted)
	case publishLinkedInPosted: // TODO: Automate
		video.LinkedInPosted = postLinkedIn(video.Tweet, video.LinkedInPosted)
	case publishSlackPosted: // TODO: Automate
		video.SlackPosted = postSlack(video.VideoId, video.SlackPosted)
	case publishRedditPosted: // TODO: Automate
		video.RedditPosted = postReddit(video.Title, video.VideoId, video.RedditPosted)
	case publishHNPosted: // TODO: Automate
		video.HNPosted = postHackerNews(video.Title, video.VideoId, video.HNPosted)
	case publishTCPosted: // TODO: Automate
		video.TCPosted = postTechnologyConversations(video.Title, video.Description, video.VideoId, video.Gist, video.ProjectName, video.ProjectURL, video.RelatedVideos, video.TCPosted)
	case publishYouTubeHighlight: // TODO: Automate
		video.YouTubeHighlight = getInputFromBool(video.YouTubeHighlight)
	case publishYouTubeComment: // TODO: Automate
		video.YouTubeComment = getInputFromBool(video.YouTubeComment)
	case publishYouTubeCommentReply: // TODO: Automate
		video.YouTubeCommentReply = getInputFromBool(video.YouTubeCommentReply)
	case publishSlides: // TODO: Automate
		video.Slides = getInputFromBool(video.Slides)
	case publishGDE: // TODO: Automate
		video.GDE = getInputFromBool(video.GDE)
	case publishTwitterSpace:
		twitter := Twitter{}
		video.TwitterSpace = twitter.PostSpace(video.VideoId, video.TwitterSpace)
	case publishRepo:
		if video.Repo != "N/A" {
			repo := Repo{}
			video.Repo, _ = repo.Update(video.Repo, video.Title, video.VideoId)
		}
	case publishNotifySponsors:
		video.NotifiedSponsors = notifySponsors(video.SponsoredEmails, video.VideoId, video.Sponsored, video.NotifiedSponsors)
	case publishReturn:
		returnVar = true
	}
	return video, returnVar, err
}
