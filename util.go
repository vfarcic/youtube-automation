package main

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var redStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("1"))

var greenStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("2"))

var orangeStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("3"))

const choiceProjectName = 0
const choiceProjectURL = 1
const choiceSponsored = 2
const choiceSubject = 3
const choiceDate = 4
const choiceCode = 5
const choiceScreen = 6
const choiceHead = 7
const choiceThumbnails = 8
const choiceLocation = 9
const choiceTagline = 10
const choiceTaglineIdeas = 11
const choiceOtherLogos = 12
const choiceScreenshots = 13
const choiceGenerateTitle = 14
const choiceModifyTitle = 15
const choiceGenerateDescription = 16
const choiceModifyDescription = 17
const choiceGenerateTags = 18
const choiceModifyTags = 19
const choiceModifyDescriptionTags = 20
const choiceRequestThumbnail = 21
const choiceMembers = 22
const choiceAnimations = 23
const choiceRequestEdit = 24
const choiceThumbnail = 25
const choiceGotMovie = 26
const choiceTimecodes = 27
const choiceGist = 28
const choiceRelatedVideos = 29
const choicePlaylists = 30
const choiceUploadVideo = 31
const choiceGenerateTweet = 32
const choiceModifyTweet = 33
const choiceExit = 34

type Video struct {
	ProjectName      string
	ProjectURL       string
	Sponsored        string
	Subject          string
	Date             string
	Code             bool
	Screen           bool
	Head             bool
	Thumbnails       bool
	Title            string
	Description      string
	Tags             string
	DescriptionTags  string
	Location         string
	Tagline          string
	TaglineIdeas     string
	OtherLogos       string
	Screenshots      bool
	RequestThumbnail bool
	Thumbnail        string
	Members          string
	Animations       string
	RequestEdit      bool
	Movie            bool
	Timecodes        string
	Gist             string
	RelatedVideos    string
	Playlists        []Playlist
	UploadVideo      string
	VideoId          string
	Tweet            string
}

type Playlist struct {
	Title string
	Id    string
}

func modifyChoice(video Video) (Video, error) {
	choices := map[int]string{
		choiceProjectName:           getChoiceTextFromString("Set project name", video.ProjectName),
		choiceProjectURL:            getChoiceTextFromString("Set project URL", video.ProjectURL),
		choiceSponsored:             getChoiceTextFromString("Set sponsorship", video.Sponsored),
		choiceSubject:               getChoiceTextFromString("Set the subject", video.Subject),
		choiceDate:                  getChoiceTextFromString("Set publish date", video.Date),
		choiceCode:                  getChoiceTextFromBool("Wrote code?", video.Code),
		choiceScreen:                getChoiceTextFromBool("Recorded screen?", video.Screen),
		choiceHead:                  getChoiceTextFromBool("Recorded talking head?", video.Head),
		choiceThumbnails:            getChoiceTextFromBool("Downloaded thumbnails?", video.Thumbnails),
		choiceLocation:              getChoiceTextFromString("Set files location", video.Location),
		choiceTagline:               getChoiceTextFromString("Set tagline", video.Tagline),
		choiceTaglineIdeas:          getChoiceTextFromString("Set tagline ideas", video.TaglineIdeas),
		choiceOtherLogos:            getChoiceTextFromString("Set other logos", video.OtherLogos),
		choiceScreenshots:           getChoiceTextFromBool("Created screenshots?", video.Screenshots),
		choiceGenerateTitle:         getChoiceTextFromString("Generate title", video.Title),
		choiceModifyTitle:           getChoiceTextFromString("Write/modify title", video.Title),
		choiceGenerateDescription:   getChoiceTextFromString("Generate description", video.Description),
		choiceModifyDescription:     getChoiceTextFromString("Write/modify description", video.Description),
		choiceGenerateTags:          getChoiceTextFromString("Generate tags", video.Tags),
		choiceModifyTags:            getChoiceTextFromString("Write/modify tags", video.Tags),
		choiceModifyDescriptionTags: getChoiceTextFromString("Write/modify description tags", video.DescriptionTags),
		choiceRequestThumbnail:      getChoiceTextFromBool("Request thumbnail", video.RequestThumbnail),
		choiceMembers:               getChoiceTextFromString("Set members", video.Members),
		choiceAnimations:            getChoiceTextFromString("Write/modify animations", video.Animations),
		choiceRequestEdit:           getChoiceTextFromBool("Request edit", video.RequestEdit),
		choiceThumbnail:             getChoiceTextFromString("Set thumbnail?", video.Thumbnail),
		choiceGotMovie:              getChoiceTextFromBool("Got movie?", video.Movie),
		choiceTimecodes:             getChoiceTextFromString("Set timecodes", video.Timecodes),
		choiceGist:                  getChoiceTextFromString("Set gist", video.Gist),
		choiceRelatedVideos:         getChoiceTextFromString("Set related videos", video.RelatedVideos),
		choicePlaylists:             getChoiceTextFromPlaylists("Set playlists", video.Playlists),
		choiceUploadVideo:           getChoiceTextFromString("Upload video", video.UploadVideo),
		choiceGenerateTweet:         getChoiceTextFromString("Generate Tweet", video.Tweet),
		choiceModifyTweet:           getChoiceTextFromString("Write/modify Tweet", video.Tweet),
		choiceExit:                  "Exit",
	}
	println()
	choice, _ := getChoice(choices, "What would you like to do?")
	err := error(nil)
	switch choice {
	case choiceProjectName:
		video.ProjectName, err = getInputFromString("Set project name)", video.ProjectName)
	case choiceProjectURL:
		video.ProjectURL, err = getInputFromString("Set project URL", video.ProjectURL)
	case choiceSponsored:
		video.Sponsored, err = getInputFromString("Sponsorship amount ('-' or 'N/A' if not sponsored)", video.Sponsored)
	case choiceSubject:
		video.Subject, err = getInputFromString("What is the subject of the video?", video.Subject)
	case choiceDate:
		video.Date, err = getInputFromString("What is the publish of the video (e.g., 2030-01-21T16:00)?", video.Date)
	case choiceCode:
		video.Code = getInputFromBool(video.Code)
	case choiceScreen:
		video.Screen = getInputFromBool(video.Screen)
	case choiceHead:
		video.Head = getInputFromBool(video.Head)
	case choiceThumbnails:
		video.Thumbnails = getInputFromBool(video.Thumbnails)
	case choiceLocation:
		video.Location, err = getInputFromString("Where are files located?", video.Location)
	case choiceTagline:
		video.Tagline, err = getInputFromString("What is the tagline?", video.Tagline)
	case choiceTaglineIdeas:
		video.TaglineIdeas, err = getInputFromString("What are the tagline ideas?", video.TaglineIdeas)
	case choiceOtherLogos:
		video.OtherLogos, err = getInputFromString("What are the other logos?", video.OtherLogos)
	case choiceScreenshots:
		video.Screenshots = getInputFromBool(video.Screenshots)
	case choiceGenerateTitle:
		video, err = generateTitle(video)
	case choiceModifyTitle:
		video.Title, err = modifyTextArea(video.Title, "Rewrite the title:", "Title was not generated!")
	case choiceGenerateDescription:
		video, err = generateDescription(video)
	case choiceModifyDescription:
		video.Description, err = modifyTextArea(video.Description, "Modify video description:", "Description was not generated!")
	case choiceGenerateTags:
		video.Tags, err = generateTags(video.Title)
	case choiceModifyTags:
		video.Tags, err = modifyTextArea(video.Tags, "Modify tags:", "Tags were not generated!")
	case choiceModifyDescriptionTags:
		video.DescriptionTags, err = modifyDescriptionTags(video.Tags, video.DescriptionTags, "Modify description tags (max 4):", "Description tags were not generated!")
	case choiceRequestThumbnail:
		video.RequestThumbnail = getChoiceThumbnail(video.RequestThumbnail, settings.Email.From, settings.Email.ThumbnailTo, video)
	case choiceThumbnail:
		video.Thumbnail, err = setThumbnail(video.Thumbnail)
	case choiceMembers:
		video.Members, err = getInputFromString("Who are new members?", video.Members)
	case choiceAnimations:
		video.Animations, err = modifyAnimations(video)
	case choiceRequestEdit:
		video.RequestEdit = getChoiceEdit(video.RequestEdit, settings.Email.From, settings.Email.EditTo, video)
	case choiceGotMovie:
		video.Movie = getInputFromBool(video.Movie)
	case choiceTimecodes:
		video.Timecodes = getInputFromTextArea("What are timecodes?", video.Timecodes, 20)
	case choiceGist:
		video.Gist, err = getInputFromString("Where is the gist?", video.Gist)
	case choiceRelatedVideos:
		video.RelatedVideos = getInputFromTextArea("What are the related videos?", video.RelatedVideos, 20)
	case choicePlaylists:
		video.Playlists = getChoicePlaylists()
	case choiceUploadVideo:
		video.UploadVideo, video.VideoId = getChoiceUploadVideo(video)
	case choiceGenerateTweet:
		video.Tweet, err = generateTweet(video.Title, video.VideoId)
	case choiceModifyTweet:
		video.Tweet, err = modifyTextArea(video.Tweet, "Modify tweet:", "Tweet was not generated!")
	case choiceExit:
		os.Exit(0)
	}
	return video, err
}
