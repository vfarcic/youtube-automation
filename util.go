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
const choiceSponsoredEmails = 3
const choiceSubject = 4
const choiceDate = 5
const choiceCode = 6
const choiceScreen = 7
const choiceHead = 8
const choiceThumbnails = 9
const choiceLocation = 10
const choiceTagline = 11
const choiceTaglineIdeas = 12
const choiceOtherLogos = 13
const choiceScreenshots = 14
const choiceGenerateTitle = 15
const choiceModifyTitle = 16
const choiceGenerateDescription = 17
const choiceModifyDescription = 18
const choiceGenerateTags = 19
const choiceModifyTags = 20
const choiceModifyDescriptionTags = 21
const choiceRequestThumbnail = 22
const choiceMembers = 23
const choiceAnimations = 24
const choiceRequestEdit = 25
const choiceThumbnail = 26
const choiceGotMovie = 27
const choiceTimecodes = 28
const choiceGist = 29
const choiceRelatedVideos = 30
const choicePlaylists = 31
const choiceUploadVideo = 32
const choiceGenerateTweet = 33
const choiceModifyTweet = 34
const choiceTweetPosted = 35
const choiceLinkedInPosted = 36
const choiceSlackPosted = 37
const choiceRedditPosted = 38
const choiceHNPosted = 39
const choiceTCPosted = 40
const choiceYouTubeHighlight = 41
const choiceYouTubeComment = 42
const choiceYouTubeCommentReply = 43
const choiceSlides = 44
const choiceGDE = 45
const choiceRepoReadme = 46
const choiceTwitterSpace = 47
const choiceNotifySponsors = 48
const choiceExit = 49

type Video struct {
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

type Playlist struct {
	Title string
	Id    string
}

func modifyChoice(video Video) (Video, error) {
	choices := map[int]string{
		choiceProjectName:           getChoiceTextFromString("Set project name", video.ProjectName),
		choiceProjectURL:            getChoiceTextFromString("Set project URL", video.ProjectURL),
		choiceSponsored:             getChoiceTextFromString("Set sponsorship", video.Sponsored),
		choiceSponsoredEmails:       getChoiceTextFromSponsoredEmails("Set sponsorship emails", video.Sponsored, video.SponsoredEmails),
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
		choiceTweetPosted:           getChoiceTextFromBool("Post to Tweeter (MANUAL)", video.TweetPosted),
		choiceLinkedInPosted:        getChoiceTextFromBool("Post to LinkedIn (MANUAL)", video.LinkedInPosted),
		choiceSlackPosted:           getChoiceTextFromBool("Post to Slack (MANUAL)", video.SlackPosted),
		choiceRedditPosted:          getChoiceTextFromBool("Post to Reddit (MANUAL)", video.RedditPosted),
		choiceHNPosted:              getChoiceTextFromBool("Post to Hacker News (MANUAL)", video.HNPosted),
		choiceTCPosted:              getChoiceTextFromBool("Post to Technology Conversations (MANUAL)", video.TCPosted),
		choiceYouTubeHighlight:      getChoiceTextFromBool("Set as YouTube Highlight (MANUAL)", video.YouTubeHighlight),
		choiceYouTubeComment:        getChoiceTextFromBool("Write pinned comment (MANUAL)", video.YouTubeComment),
		choiceYouTubeCommentReply:   getChoiceTextFromBool("Write replies to comments (MANUAL)", video.YouTubeCommentReply),
		choiceSlides:                getChoiceTextFromBool("Added to slides?", video.Slides),
		choiceGDE:                   getChoiceTextFromBool("Add to https://gde.advocu.com (MANUAL)", video.GDE),
		choiceRepoReadme:            getChoiceTextFromBool("Update repo README (MANUAL)", video.RepoReadme),
		choiceTwitterSpace:          getChoiceTextFromBool("Post to a Twitter Spaces (MANUAL)", video.TwitterSpace),
		choiceNotifySponsors:        getChoiceNotifySponsors("Notify sponsors", video.Sponsored, video.NotifiedSponsors),
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
	case choiceSponsoredEmails:
		video.SponsoredEmails = writeSponsoredEmails(video.SponsoredEmails)
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
		video.RequestEdit = requestEdit(video.RequestEdit, settings.Email.From, settings.Email.EditTo, video)
	case choiceGotMovie:
		video.Movie = getInputFromBool(video.Movie)
	case choiceTimecodes:
		video.Timecodes = getInputFromTextArea("What are timecodes?", video.Timecodes, 20)
	case choiceGist:
		video.Gist, err = getInputFromString("Where is the gist?", video.Gist)
	case choiceRelatedVideos:
		video.RelatedVideos = getInputFromTextArea("What are the related videos?", video.RelatedVideos, 20)
	case choicePlaylists:
		video.Playlists = getPlaylists()
	case choiceUploadVideo: // TODO: Finish implementing End screen, Playlists, Tags, Language, License, Monetization
		video.UploadVideo, video.VideoId = getChoiceUploadVideo(video)
	case choiceGenerateTweet:
		video.Tweet, err = generateTweet(video.Title, video.VideoId)
	case choiceModifyTweet:
		video.Tweet, err = modifyTextArea(video.Tweet, "Modify tweet:", "Tweet was not generated!")
	case choiceTweetPosted: // TODO: Automate
		twitter := Twitter{}
		video.TweetPosted = twitter.Post(video.Tweet, video.TweetPosted)
	case choiceLinkedInPosted: // TODO: Automate
		video.LinkedInPosted = postLinkedIn(video.Tweet, video.LinkedInPosted)
	case choiceSlackPosted: // TODO: Automate
		video.SlackPosted = postSlack(video.VideoId, video.SlackPosted)
	case choiceRedditPosted: // TODO: Automate
		video.RedditPosted = postReddit(video.Title, video.VideoId, video.RedditPosted)
	case choiceHNPosted: // TODO: Automate
		video.HNPosted = postHackerNews(video.Title, video.VideoId, video.HNPosted)
	case choiceTCPosted: // TODO: Automate
		video.TCPosted = postTechnologyConversations(video.Title, video.Description, video.VideoId, video.Gist, video.RelatedVideos, video.TCPosted)
	case choiceYouTubeHighlight: // TODO: Automate
		video.YouTubeHighlight = getInputFromBool(video.YouTubeHighlight)
	case choiceYouTubeComment: // TODO: Automate
		video.YouTubeComment = getInputFromBool(video.YouTubeComment)
	case choiceYouTubeCommentReply: // TODO: Automate
		video.YouTubeCommentReply = getInputFromBool(video.YouTubeCommentReply)
	case choiceSlides: // TODO: Automate
		video.Slides = getInputFromBool(video.Slides)
	case choiceGDE: // TODO: Automate
		video.GDE = getInputFromBool(video.GDE)
	case choiceRepoReadme: // TODO: Automate
		video.RepoReadme = getInputFromBool(video.RepoReadme)
	case choiceTwitterSpace:
		twitter := Twitter{}
		video.TwitterSpace = twitter.PostSpace(video.VideoId, video.TwitterSpace)
	case choiceNotifySponsors:
		video.NotifiedSponsors = notifySponsors(video.SponsoredEmails, video.VideoId, video.Sponsored, video.NotifiedSponsors)
	case choiceExit:
		os.Exit(0)
	}
	return video, err
}
