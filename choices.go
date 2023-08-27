package main

import (
	"fmt"
	"os"

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

const phasePrePublish = 0
const phasePublish = 1
const phaseExit = 2

const prePublishProjectName = 0
const prePublishProjectURL = 1
const prePublishSponsored = 2
const prePublishSponsoredEmails = 3
const prePublishSubject = 4
const prePublishDate = 5
const prePublishCode = 6
const prePublishScreen = 7
const prePublishHead = 8
const prePublishThumbnails = 9
const prePublishLocation = 10
const prePublishTagline = 11
const prePublishTaglineIdeas = 12
const prePublishOtherLogos = 13
const prePublishScreenshots = 14
const prePublishGenerateTitle = 15
const prePublishModifyTitle = 16
const prePublishGenerateDescription = 17
const prePublishModifyDescription = 18
const prePublishGenerateTags = 19
const prePublishModifyTags = 20
const prePublishModifyDescriptionTags = 21
const prePublishRequestThumbnail = 22
const prePublishMembers = 23
const prePublishAnimations = 24
const prePublishRequestEdit = 25
const prePublishThumbnail = 26
const prePublishGotMovie = 27
const prePublishTimecodes = 28
const prePublishGist = 29
const prePublishRelatedVideos = 30
const prePublishPlaylists = 31
const prePublishReturn = 32

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
const publishRepoReadme = 14
const publishTwitterSpace = 15
const publishNotifySponsors = 16
const publishReturn = 17

type Video struct {
	PrePublish          Tasks
	Publish             Tasks
	ProjectName         string
	ProjectURL          string
	Sponsored           string
	SponsoredEmails     []string
	Subject             string
	Date                string
	Code                bool
	Screen              bool
	Head                bool
	Thumbnails          bool
	Title               string
	Description         string
	Tags                string
	DescriptionTags     string
	Location            string
	Tagline             string
	TaglineIdeas        string
	OtherLogos          string
	Screenshots         bool
	RequestThumbnail    bool
	Thumbnail           string
	Members             string
	Animations          string
	RequestEdit         bool
	Movie               bool
	Timecodes           string
	Gist                string
	RelatedVideos       string
	Playlists           []Playlist
	UploadVideo         string
	VideoId             string
	Tweet               string
	TweetPosted         bool
	LinkedInPosted      bool
	SlackPosted         bool
	RedditPosted        bool
	HNPosted            bool
	TCPosted            bool
	YouTubeHighlight    bool
	YouTubeComment      bool
	YouTubeCommentReply bool
	Slides              bool
	GDE                 bool
	RepoReadme          bool
	TwitterSpace        bool
	NotifiedSponsors    bool
}

type Tasks struct {
	Completed int
	Total     int
}

type Task struct {
	Title     string
	Completed bool
}

type Playlist struct {
	Title string
	Id    string
}

func (c *Choices) ChoosePhase(video Video) Video {
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
		phaseExit:       {Title: "Exit"},
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
			writeYaml(video, settings.Path)
		}
	case phasePublish:
		var err error
		for !returnVar {
			video, returnVar, err = c.ChoosePublish(video)
			if err != nil {
				errorMessage = err.Error()
				continue
			}
			writeYaml(video, settings.Path)
		}
	case phaseExit:
		os.Exit(0)
	}
	return video
}

func (c *Choices) ChoosePrePublish(video Video) (Video, bool, error) {
	openAI := OpenAI{}
	returnVar := false
	tasks := map[int]Task{
		prePublishProjectName:           colorize(getChoiceTextFromString("Set project name", video.ProjectName)),
		prePublishProjectURL:            colorize(getChoiceTextFromString("Set project URL", video.ProjectURL)),
		prePublishSponsored:             colorize(getChoiceTextFromString("Set sponsorship", video.Sponsored)),
		prePublishSponsoredEmails:       colorize(getChoiceTextFromSponsoredEmails("Set sponsorship emails", video.Sponsored, video.SponsoredEmails)),
		prePublishSubject:               colorize(getChoiceTextFromString("Set the subject", video.Subject)),
		prePublishDate:                  colorize(getChoiceTextFromString("Set publish date", video.Date)),
		prePublishCode:                  colorize(getChoiceTextFromBool("Wrote code?", video.Code)),
		prePublishScreen:                colorize(getChoiceTextFromBool("Recorded screen?", video.Screen)),
		prePublishHead:                  colorize(getChoiceTextFromBool("Recorded talking head?", video.Head)),
		prePublishThumbnails:            colorize(getChoiceTextFromBool("Downloaded thumbnails?", video.Thumbnails)),
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
		prePublishThumbnail:             colorize(getChoiceTextFromString("Set thumbnail?", video.Thumbnail)),
		prePublishGotMovie:              colorize(getChoiceTextFromBool("Got movie?", video.Movie)),
		prePublishTimecodes:             colorize(getChoiceTextFromString("Set timecodes", video.Timecodes)),
		prePublishGist:                  colorize(getChoiceTextFromString("Set gist", video.Gist)),
		prePublishRelatedVideos:         colorize(getChoiceTextFromString("Set related videos", video.RelatedVideos)),
		prePublishPlaylists:             colorize(getChoiceTextFromPlaylists("Set playlists", video.Playlists)),
		prePublishReturn:                {Title: "Return"},
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
	case prePublishThumbnails:
		video.Thumbnails = getInputFromBool(video.Thumbnails)
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
	case prePublishRelatedVideos:
		video.RelatedVideos = getInputFromTextArea("What are the related videos?", video.RelatedVideos, 20)
	case prePublishPlaylists:
		video.Playlists = getPlaylists()
	case prePublishReturn:
		returnVar = true
	}
	return video, returnVar, err
}

func (c *Choices) ChoosePublish(video Video) (Video, bool, error) {
	openAI := OpenAI{}
	returnVar := false
	tasks := map[int]Task{
		publishUploadVideo: colorize(getChoiceTextFromString("Upload video", video.UploadVideo)),
		// TODO: Add new option to Update the Gist with Gist and Video URL
		publishGenerateTweet:       colorize(getChoiceTextFromString("Generate Tweet", video.Tweet)),
		publishModifyTweet:         colorize(getChoiceTextFromString("Write/modify Tweet", video.Tweet)),
		publishTweetPosted:         colorize(getChoiceTextFromBool("Post to Tweeter (MANUAL)", video.TweetPosted)),
		publishLinkedInPosted:      colorize(getChoiceTextFromBool("Post to LinkedIn (MANUAL)", video.LinkedInPosted)),
		publishSlackPosted:         colorize(getChoiceTextFromBool("Post to Slack (MANUAL)", video.SlackPosted)),
		publishRedditPosted:        colorize(getChoiceTextFromBool("Post to Reddit (MANUAL)", video.RedditPosted)),
		publishHNPosted:            colorize(getChoiceTextFromBool("Post to Hacker News (MANUAL)", video.HNPosted)),
		publishTCPosted:            colorize(getChoiceTextFromBool("Post to Technology Conversations (MANUAL)", video.TCPosted)),
		publishYouTubeHighlight:    colorize(getChoiceTextFromBool("Set as YouTube Highlight (MANUAL)", video.YouTubeHighlight)),
		publishYouTubeComment:      colorize(getChoiceTextFromBool("Write pinned comment (MANUAL)", video.YouTubeComment)),
		publishYouTubeCommentReply: colorize(getChoiceTextFromBool("Write replies to comments (MANUAL)", video.YouTubeCommentReply)),
		publishSlides:              colorize(getChoiceTextFromBool("Added to slides?", video.Slides)),
		publishGDE:                 colorize(getChoiceTextFromBool("Add to https://gde.advocu.com (MANUAL)", video.GDE)),
		publishRepoReadme:          colorize(getChoiceTextFromBool("Update repo README (MANUAL)", video.RepoReadme)),
		publishTwitterSpace:        colorize(getChoiceTextFromBool("Post to a Twitter Spaces (MANUAL)", video.TwitterSpace)),
		publishNotifySponsors:      colorize(getChoiceNotifySponsors("Notify sponsors", video.Sponsored, video.NotifiedSponsors)),
		publishReturn:              {Title: "Return"},
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
		video.TCPosted = postTechnologyConversations(video.Title, video.Description, video.VideoId, video.Gist, video.RelatedVideos, video.TCPosted)
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
	case publishRepoReadme: // TODO: Automate
		video.RepoReadme = getInputFromBool(video.RepoReadme)
	case publishTwitterSpace:
		twitter := Twitter{}
		video.TwitterSpace = twitter.PostSpace(video.VideoId, video.TwitterSpace)
	case publishNotifySponsors:
		video.NotifiedSponsors = notifySponsors(video.SponsoredEmails, video.VideoId, video.Sponsored, video.NotifiedSponsors)
	case publishReturn:
		returnVar = true
	}
	return video, returnVar, err
}
