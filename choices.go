package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
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
const videosPhaseDelayed = 5
const videosPhaseSponsoredBlocked = 6
const videosPhaseIdeas = 7

// TODO: Remove
const videosPhaseReturn = 8

const phasePrePublish = 0
const phasePublish = 1

const actionEdit = 0
const actionDelete = 1
const actionReturn = -1

const prePublishProjectName = 0
const prePublishProjectURL = 1
const prePublishSponsored = 2
const prePublishSponsoredEmails = 3
const prePublishSponsorshipBlocked = 4
const prePublishSubject = 5
const prePublishDate = 6
const prePublishDelayed = 7
const prePublishCode = 8
const prePublishScreen = 9
const prePublishHead = 10
const prePublishRelatedVideos = 11
const prePublishThumbnails = 12
const prePublishDiagrams = 13
const prePublishLocation = 14
const prePublishTagline = 15
const prePublishTaglineIdeas = 16
const prePublishOtherLogos = 17
const prePublishScreenshots = 18
const prePublishModifyTitle = 19
const prePublishModifyDescription = 20
const prePublishModifyTags = 21
const prePublishModifyDescriptionTags = 22
const prePublishRequestThumbnail = 23
const prePublishMembers = 24
const prePublishAnimations = 25
const prePublishRequestEdit = 26
const prePublishThumbnail = 27
const prePublishGotMovie = 28
const prePublishTimecodes = 29
const prePublishSlides = 30
const prePublishGist = 31
const prePublishPlaylists = 32

// TODO: Remove
const prePublishReturn = 33

const publishUploadVideo = 0
const publishModifyTweet = 1
const publishTweetPosted = 2
const publishLinkedInPosted = 3
const publishSlackPosted = 4
const publishRedditPosted = 5
const publishHNPosted = 6
const publishTCPosted = 7
const publishYouTubeHighlight = 8
const publishYouTubeComment = 9
const publishYouTubeCommentReply = 10
const publishGDE = 11
const publishTwitterSpace = 12
const publishRepo = 13
const publishNotifySponsors = 14

// TODO: Remove
const publishReturn = 15

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
	var selectedIndex int
	yaml := YAML{IndexPath: "index.yaml"}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("What do you want to do?").
				Options(
					huh.NewOption("Create a video", indexCreateVideo),
					huh.NewOption("List videos", indexListVideos),
					huh.NewOption("Exit", indexExit),
				).
				Value(&selectedIndex),
		),
	)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	switch selectedIndex {
	case indexCreateVideo:
		index := yaml.GetIndex()
		item := c.ChooseCreateVideo()
		if len(item.Category) > 0 && len(item.Name) > 0 {
			index = append(index, item)
			yaml.WriteIndex(index)
		}
	case indexListVideos:
		for {
			index := yaml.GetIndex()
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
	var selected int
	prePublishText := fmt.Sprintf("Pre-publish (%d/%d)", video.PrePublish.Completed, video.PrePublish.Total)
	if video.PrePublish.Completed == video.PrePublish.Total && video.PrePublish.Total > 0 {
		prePublishText = greenStyle.Render(prePublishText)
	} else {
		prePublishText = orangeStyle.Render(prePublishText)
	}
	publishText := fmt.Sprintf("Publish (%d/%d)", video.Publish.Completed, video.Publish.Total)
	if video.Publish.Completed == video.Publish.Total && video.Publish.Total > 0 {
		publishText = greenStyle.Render(publishText)
	} else {
		publishText = orangeStyle.Render(publishText)
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Which type of tasks would you like to work on?").
				Options(
					huh.NewOption(prePublishText, phasePrePublish),
					huh.NewOption(publishText, phasePublish),
					huh.NewOption("Return", actionReturn),
				).
				Value(&selected),
		),
	)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	returnVar := false
	switch selected {
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
	case actionReturn:
		returnVar = true
	}
	return video, returnVar
}

func (c *Choices) ChooseCreateVideo() VideoIndex {
	var name string
	var category string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Prompt("Name: ").Value(&name).Validate(c.IsEmpty),
			huh.NewInput().Prompt("Category: ").Value(&category).Validate(c.IsEmpty),
		),
	)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	vi := VideoIndex{
		Name:     name,
		Category: category,
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
# [[title]] #

# Additional Info:
# - [[additional-info]]

#########
# Intro #
#########

# TODO: Title screen

#########
# Setup #
#########

# TODO:

##########
# TODO:: #
##########

# TODO:

#######################
# TODO: Pros and Cons #
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

// TODO: Refactor
func (c *Choices) ChoosePrePublish(video Video) (Video, bool, error) {
	// var selected int
	// sponsorshipBlockedTitle, sponsorshipBlockedCompleted := c.GetOptionTextFromString("Sponsorship blocked?", video.SponsorshipBlocked)
	// if len(video.SponsorshipBlocked) > 0 {
	// 	sponsorshipBlockedTitle = redStyle.Render(sponsorshipBlockedTitle)
	// } else {
	// 	sponsorshipBlockedTitle = greenStyle.Render(sponsorshipBlockedTitle)
	// }
	// form := huh.NewForm(
	// 	huh.NewGroup(
	// 		huh.NewSelect[int]().
	// 			Title("Which pre-publish task would you like to work on?").
	// 			Options(
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Project name", video.ProjectName)), prePublishProjectName),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Project URL", video.ProjectURL)), prePublishProjectURL),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Sponsorship", video.Sponsored)), prePublishSponsored),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromSponsoredEmails("Sponsorship emails", video.Sponsored, video.SponsoredEmails)), prePublishSponsoredEmails),
	// 				huh.NewOption(c.Colorize(sponsorshipBlockedTitle, sponsorshipBlockedCompleted), prePublishSponsorshipBlocked),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Subject", video.Subject)), prePublishSubject),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Publish date", video.Date)), prePublishSponsorshipBlocked),
	// 				huh.NewOption(c.Colorize("Delayed?", !video.Delayed), prePublishDelayed),
	// 				huh.NewOption(c.Colorize("Code?", video.Code), prePublishCode),
	// 				huh.NewOption(c.Colorize("Screen?", video.Screen), prePublishScreen),
	// 				huh.NewOption(c.Colorize("Talking head?", video.Head), prePublishHead),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Related videos", video.RelatedVideos)), prePublishRelatedVideos),
	// 				huh.NewOption(c.Colorize("Thumbnails?", video.Thumbnails), prePublishThumbnails),
	// 				huh.NewOption(c.Colorize("Files location", video.Diagrams), prePublishDiagrams),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Files location", video.Location)), prePublishLocation),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Tagline", video.Tagline)), prePublishTagline),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Tagline ideas", video.TaglineIdeas)), prePublishTaglineIdeas),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Other logos", video.OtherLogos)), prePublishOtherLogos),
	// 				huh.NewOption(c.Colorize("Screenshots?", video.Screenshots), prePublishScreenshots),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Title (write/modify)", video.Title)), prePublishModifyTitle),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Description (write/modify)", video.Description)), prePublishModifyDescription),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Tags (write/modify)", video.Tags)), prePublishModifyTags),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Write/modify description tags", video.DescriptionTags)), prePublishModifyDescriptionTags),
	// 				huh.NewOption(c.Colorize("Thumbnail request", video.RequestThumbnail), prePublishRequestThumbnail),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Members", video.Members)), prePublishMembers),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Animations", video.Animations)), prePublishAnimations),
	// 				huh.NewOption(c.Colorize("Edit (request)", video.RequestEdit), prePublishRequestEdit),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Thumbnail?", video.Thumbnail)), prePublishThumbnail),
	// 				huh.NewOption(c.Colorize("Movie?", video.Movie), prePublishGotMovie),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Timecodes", video.Timecodes)), prePublishTimecodes),
	// 				huh.NewOption(c.Colorize("Slides?", video.Slides), prePublishSlides),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromString("Gist", video.Gist)), prePublishGist),
	// 				huh.NewOption(c.Colorize(c.GetOptionTextFromPlaylists("Playlists", video.Playlists)), prePublishPlaylists),
	// 				huh.NewOption("Save and return", actionReturn),
	// 			).
	// 			Value(&selected),
	// 	),
	// )
	// err := form.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	returnVar := false
	sponsorshipBlockedTask := getChoiceTextFromString("Sponsorship blocked?", video.SponsorshipBlocked)
	if len(video.SponsorshipBlocked) > 0 {
		sponsorshipBlockedTask.Title = redStyle.Render(sponsorshipBlockedTask.Title)
	} else {
		sponsorshipBlockedTask.Title = greenStyle.Render(sponsorshipBlockedTask.Title)
	}
	sponsorshipBlockedTask.Completed = !sponsorshipBlockedTask.Completed
	tasks := map[int]Task{
		prePublishProjectName:           colorize(getChoiceTextFromString("Project name", video.ProjectName)),
		prePublishProjectURL:            colorize(getChoiceTextFromString("Project URL", video.ProjectURL)),
		prePublishSponsored:             colorize(getChoiceTextFromString("Sponsorship", video.Sponsored)),
		prePublishSponsoredEmails:       colorize(getChoiceTextFromSponsoredEmails("Sponsorship emails", video.Sponsored, video.SponsoredEmails)),
		prePublishSponsorshipBlocked:    sponsorshipBlockedTask,
		prePublishSubject:               colorize(getChoiceTextFromString("Subject", video.Subject)),
		prePublishDate:                  colorize(getChoiceTextFromString("Publish date", video.Date)),
		prePublishDelayed:               colorize(getChoiceTextFromBool("Delayed?", !video.Delayed)),
		prePublishCode:                  colorize(getChoiceTextFromBool("Code?", video.Code)),
		prePublishScreen:                colorize(getChoiceTextFromBool("Screen?", video.Screen)),
		prePublishHead:                  colorize(getChoiceTextFromBool("Talking head?", video.Head)),
		prePublishRelatedVideos:         colorize(getChoiceTextFromString("Related videos", video.RelatedVideos)),
		prePublishThumbnails:            colorize(getChoiceTextFromBool("Thumbnails?", video.Thumbnails)),
		prePublishDiagrams:              colorize(getChoiceTextFromBool("Diagrams?", video.Diagrams)),
		prePublishLocation:              colorize(getChoiceTextFromString("Files location", video.Location)),
		prePublishTagline:               colorize(getChoiceTextFromString("Tagline", video.Tagline)),
		prePublishTaglineIdeas:          colorize(getChoiceTextFromString("Tagline ideas", video.TaglineIdeas)),
		prePublishOtherLogos:            colorize(getChoiceTextFromString("Other logos", video.OtherLogos)),
		prePublishScreenshots:           colorize(getChoiceTextFromBool("Screenshots?", video.Screenshots)),
		prePublishModifyTitle:           colorize(getChoiceTextFromString("Title (write/modify)", video.Title)),
		prePublishModifyDescription:     colorize(getChoiceTextFromString("Description (write/modify)", video.Description)),
		prePublishModifyTags:            colorize(getChoiceTextFromString("Tags (write/modify)", video.Tags)),
		prePublishModifyDescriptionTags: colorize(getChoiceTextFromString("Write/modify description tags", video.DescriptionTags)),
		prePublishRequestThumbnail:      colorize(getChoiceTextFromBool("Thumbnail request", video.RequestThumbnail)),
		prePublishMembers:               colorize(getChoiceTextFromString("Members", video.Members)),
		prePublishAnimations:            colorize(getChoiceTextFromString("Animations", video.Animations)),
		prePublishRequestEdit:           colorize(getChoiceTextFromBool("Edit (request)", video.RequestEdit)),
		prePublishThumbnail:             colorize(getChoiceTextFromString("Thumbnail?", video.Thumbnail)),
		prePublishGotMovie:              colorize(getChoiceTextFromBool("Movie?", video.Movie)),
		prePublishTimecodes:             colorize(getChoiceTextFromString("Timecodes", video.Timecodes)),
		prePublishSlides:                colorize(getChoiceTextFromBool("Slides?", video.Slides)),
		prePublishGist:                  colorize(getChoiceTextFromString("Gist", video.Gist)),
		prePublishPlaylists:             colorize(getChoiceTextFromPlaylists("Playlists", video.Playlists)),
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
	err = error(nil)
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
	case prePublishDelayed:
		video.Delayed = getInputFromBool(video.Delayed)
	case prePublishCode:
		video.Code = getInputFromBool(video.Code)
	case prePublishScreen:
		video.Screen = getInputFromBool(video.Screen)
	case prePublishHead:
		video.Head = getInputFromBool(video.Head)
	case prePublishRelatedVideos:
		video.RelatedVideos = getInputFromTextArea("What are the related videos?", video.RelatedVideos, 100)
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
	case prePublishModifyTitle:
		video.Title, err = modifyTextArea(video.Title, "Rewrite the title:", "")
	case prePublishModifyDescription:
		video.Description, err = modifyTextArea(video.Description, "Rewrite video description:", "")
	case prePublishModifyTags:
		video.Tags, err = modifyTextArea(video.Tags, "Write tags:", "")
	case prePublishModifyDescriptionTags:
		video.DescriptionTags, err = modifyDescriptionTagsX(video.Tags, video.DescriptionTags, "Write description tags (max 4):", "")
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
	case prePublishSlides:
		video.Slides = getInputFromBool(video.Slides)
	case prePublishGist:
		video.Gist, err = getInputFromString("Where is the gist?", video.Gist)
		if len(video.Gist) > 0 {
			repo := Repo{}
			video.GistUrl, err = repo.Gist(video.Gist, video.Title, video.ProjectName, video.ProjectURL, video.RelatedVideos)
		}
	case prePublishPlaylists:
		video.Playlists = getPlaylists()
	case prePublishReturn:
		returnVar = true
	}
	return video, returnVar, err
}

func (c *Choices) ChooseVideosPhase(vi []VideoIndex) bool {
	var selection int
	phases := make(map[int]int)
	for i := range vi {
		phase := c.GetVideoPhase(vi[i])
		phases[phase] = phases[phase] + 1
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("From which phase would you like to list the videos?").
				Options(
					huh.NewOption(c.GetPhaseColoredText(phases, videosPhasePublished, "Published"), videosPhasePublished),
					huh.NewOption(c.GetPhaseColoredText(phases, videosPhasePublishPending, "Pending publish"), videosPhasePublishPending),
					huh.NewOption(c.GetPhaseColoredText(phases, videosPhaseEditRequested, "Edit requested"), videosPhaseEditRequested),
					huh.NewOption(c.GetPhaseColoredText(phases, videosPhaseMaterialDone, "Material done"), videosPhaseMaterialDone),
					huh.NewOption(c.GetPhaseColoredText(phases, videosPhaseStarted, "Started"), videosPhaseStarted),
					huh.NewOption(c.GetPhaseColoredText(phases, videosPhaseDelayed, "Delayed"), videosPhaseDelayed),
					huh.NewOption(c.GetPhaseColoredText(phases, videosPhaseSponsoredBlocked, "Sponsored blocked"), videosPhaseSponsoredBlocked),
					huh.NewOption(c.GetPhaseColoredText(phases, videosPhaseIdeas, "Ideas"), videosPhaseIdeas),
					huh.NewOption("Return", videosPhaseReturn),
				).
				Value(&selection),
		),
	)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	if selection == videosPhaseReturn {
		return true
	}
	c.ChooseVideos(vi, selection)
	return false
}

func (c *Choices) GetVideoPhase(vi VideoIndex) int {
	yaml := YAML{}
	video := yaml.GetVideo(c.GetFilePath(vi.Category, vi.Name, "yaml"))
	if video.Delayed {
		return videosPhaseDelayed
	} else if len(video.SponsorshipBlocked) > 0 {
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
	} else {
		return videosPhaseIdeas
	}
}

func (c *Choices) ChooseVideos(vi []VideoIndex, phase int) {
	var selectedVideo int
	var selectedAction int
	options := huh.NewOptions[int]()
	for i := range vi {
		videoPhase := c.GetVideoPhase(vi[i])
		if videoPhase == phase {
			title := vi[i].Name
			yaml := YAML{}
			path := c.GetFilePath(vi[i].Category, vi[i].Name, "yaml")
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

			options = append(options, huh.NewOption(title, i))
		}
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Which video would you like to work on?").
				Options(options...).
				Value(&selectedVideo),
			huh.NewSelect[int]().
				Title("What would you like to do with the video?").
				Options(
					huh.NewOption("Edit", actionEdit),
					huh.NewOption("Delete", actionDelete),
					huh.NewOption("Return", actionReturn),
				).
				Value(&selectedAction),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	if selectedVideo == actionReturn {
		return
	}

	selectedVideoIndex := vi[selectedVideo]
	switch selectedAction {
	case actionEdit:
		returnVal := false
		path := c.GetFilePath(selectedVideoIndex.Category, selectedVideoIndex.Name, "yaml")
		yaml := YAML{}
		video := yaml.GetVideo(path)
		for !returnVal {
			choices := Choices{}
			video, returnVal = choices.ChoosePhase(video)
			yaml.WriteVideo(video, path)
		}
	case actionDelete:
		if os.Remove(c.GetFilePath(selectedVideoIndex.Category, selectedVideoIndex.Name, "sh")) != nil {
			panic(err)
		}
		os.Remove(c.GetFilePath(selectedVideoIndex.Category, selectedVideoIndex.Name, "yaml"))
		selectedVideoIndex = vi[len(vi)-1]
		vi = append(vi[:selectedVideo], vi[selectedVideo+1:]...)
	}
	yaml := YAML{IndexPath: "index.yaml"}
	yaml.WriteIndex(vi)
}

func (c *Choices) ChoosePublish(video Video) (Video, bool, error) {
	returnVar := false
	tasks := map[int]Task{
		publishUploadVideo:         colorize(getChoiceTextFromString("Upload video", video.UploadVideo)),
		publishModifyTweet:         colorize(getChoiceTextFromString("Tweet (write/modify)", video.Tweet)),
		publishTweetPosted:         colorize(getChoiceTextFromBool("Twitter post (MANUAL)", video.TweetPosted)),                // TODO:
		publishLinkedInPosted:      colorize(getChoiceTextFromBool("LinkedIn  post (MANUAL)", video.LinkedInPosted)),           // TODO:
		publishSlackPosted:         colorize(getChoiceTextFromBool("Slack post (MANUAL)", video.SlackPosted)),                  // TODO:
		publishRedditPosted:        colorize(getChoiceTextFromBool("Reddit post (MANUAL)", video.RedditPosted)),                // TODO:
		publishHNPosted:            colorize(getChoiceTextFromBool("Hacker News post (MANUAL)", video.HNPosted)),               // TODO:
		publishTCPosted:            colorize(getChoiceTextFromBool("Technology Conversations post (MANUAL)", video.TCPosted)),  // TODO:
		publishYouTubeHighlight:    colorize(getChoiceTextFromBool("YouTube Highlight (MANUAL)", video.YouTubeHighlight)),      // TODO:
		publishYouTubeComment:      colorize(getChoiceTextFromBool("Pinned comment (MANUAL)", video.YouTubeComment)),           // TODO:
		publishYouTubeCommentReply: colorize(getChoiceTextFromBool("Replies to comments (MANUAL)", video.YouTubeCommentReply)), // TODO:
		publishGDE:                 colorize(getChoiceTextFromBool("https://gde.advocu.com post (MANUAL)", video.GDE)),         // TODO:
		publishTwitterSpace:        colorize(getChoiceTextFromBool("Twitter Spaces post (MANUAL)", video.TwitterSpace)),        // TODO:
		publishRepo:                colorize(getChoiceTextFromString("Code repo", video.Repo)),                                 // TODO:
		publishNotifySponsors:      colorize(getChoiceNotifySponsors("Sponsors (notify)", video.Sponsored, video.NotifiedSponsors)),
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
	case publishModifyTweet:
		video.Tweet, err = modifyTextArea(video.Tweet, "Modify tweet:", "")
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

func (c *Choices) IsEmpty(str string) error {
	if len(str) == 0 {
		return errors.New("Required!")
	}
	return nil
}

func (c *Choices) GetPhaseColoredText(phases map[int]int, phase int, title string) string {
	if phase != videosPhaseReturn {
		title = fmt.Sprintf("%s (%d)", title, phases[phase])
		if phase == videosPhasePublished {
			return greenStyle.Render(title)
		} else if phase == videosPhasePublishPending && phases[phase] > 0 {
			return greenStyle.Render(title)
		} else if phase == videosPhaseEditRequested && phases[phase] > 0 {
			return greenStyle.Render(title)
		} else if (phase == videosPhaseMaterialDone || phase == videosPhaseIdeas) && phases[phase] >= 3 {
			return greenStyle.Render(title)
		} else if phase == videosPhaseStarted && phases[phase] > 0 {
			return greenStyle.Render(title)
		} else {
			return orangeStyle.Render(title)
		}
	}
	return title
}

func (c *Choices) GetOptionTextFromString(title, value string) (string, bool) {
	valueLength := len(value)
	completed := false
	if valueLength > 100 {
		value = fmt.Sprintf("%s...", value[0:100])
	}
	value = strings.ReplaceAll(value, "\n", " ")
	if value != "" && value != "-" && value != "N/A" {
		title = fmt.Sprintf("%s (%s)", title, value)
	}
	if len(value) > 0 {
		completed = true
	}
	return title, completed
}

func (c *Choices) GetOptionTextFromSponsoredEmails(title, sponsored string, sponsoredEmails []string) (string, bool) {
	completed := false
	if len(sponsoredEmails) > 0 {
		emailsText := strings.Join(sponsoredEmails, ", ")
		title = fmt.Sprintf("%s (%s)", title, emailsText)
		if len(title) > 100 {
			title = fmt.Sprintf("%s...", title[0:100])
		}
		completed = true
	} else if len(sponsored) == 0 || sponsored == "N/A" || sponsored == "-" {
		completed = true
	}
	return title, completed
}

func (c *Choices) Colorize(title string, completed bool) string {
	if completed {
		return greenStyle.Render(title)
	}
	return orangeStyle.Render(title)
}

func (c *Choices) GetOptionTextFromPlaylists(title string, values []Playlist) (string, bool) {
	completed := false
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
		title = fmt.Sprintf("%s (%s)", title, value)
	}
	if value != "" {
		completed = true
	}
	return title, completed
}
