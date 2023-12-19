package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
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

const videosPhasePublished = 0
const videosPhasePublishPending = 1
const videosPhaseEditRequested = 2
const videosPhaseMaterialDone = 3
const videosPhaseStarted = 4
const videosPhaseDelayed = 5
const videosPhaseSponsoredBlocked = 6
const videosPhaseIdeas = 7

const actionReturn = 99

func (c *Choices) ChooseIndex() {
	const indexCreateVideo = 0
	const indexListVideos = 1
	const indexExit = 2
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

func (c *Choices) GetPhaseText(text string, task Tasks) string {
	text = fmt.Sprintf("%s (%d/%d)", text, task.Completed, task.Total)
	if task.Completed == task.Total && task.Total > 0 {
		return greenStyle.Render(text)
	}
	return orangeStyle.Render(text)
}

func (c *Choices) ChoosePhase(video Video) {
	returnVar := false
	for returnVar == false {
		const phaseInit = 0
		const phaseWork = 1
		const phaseDefine = 2
		const phaseEdit = 3
		const phasePublish = 4
		var selected int
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Which type of tasks would you like to work on?").
					Options(
						huh.NewOption(c.GetPhaseText("Initialize", video.Init), phaseInit),
						huh.NewOption(c.GetPhaseText("Work", video.Work), phaseWork),
						huh.NewOption(c.GetPhaseText("Define", video.Define), phaseDefine),
						huh.NewOption(c.GetPhaseText("Edit", video.Edit), phaseEdit),
						huh.NewOption(c.GetPhaseText("Publish", video.Publish), phasePublish),
						huh.NewOption("Return", actionReturn),
					).
					Value(&selected),
			),
		)
		err := form.Run()
		if err != nil {
			log.Fatal(err)
		}
		switch selected {
		case phaseInit:
			var err error
			if video, err = c.ChooseInit(video); err != nil {
				panic(err)
			}
		case phaseWork:
			var err error
			if video, err = c.ChooseWork(video); err != nil {
				panic(err)
			}
		case phaseDefine:
			var err error
			if video, err = c.ChooseDefine(video); err != nil {
				panic(err)
			}
		case phaseEdit:
			var err error
			if video, err = c.ChooseEdit(video); err != nil {
				panic(err)
			}
		case phasePublish:
			var err error
			if video, err = c.ChoosePublish(video); err != nil {
				panic(err)
			}
		case actionReturn:
			returnVar = true
		}
	}
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
			panic(err)
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

func (c *Choices) Count(fields []interface{}) (green, all int) {
	for _, field := range fields {
		valueType := reflect.TypeOf(field)
		if valueType.Kind() == reflect.String && len(field.(string)) > 0 {
			green++
		} else if valueType.Kind() == reflect.Bool && field.(bool) {
			green++
		} else if valueType.Kind() == reflect.Slice && reflect.Indirect(reflect.ValueOf(field)).Len() > 0 {
			green++
		}
		all++
	}
	return green, all
}

func (c *Choices) ChooseInit(video Video) (Video, error) {
	save := true
	sponsoredEmailsString := strings.Join(video.SponsoredEmails, ", ")
	sponsoredEmailsTitle, _ := c.ColorFromSponsoredEmails("Sponsorship emails (comma separated)", video.Sponsored, video.SponsoredEmails)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title(c.ColorFromString("Project name", video.ProjectName)).Value(&video.ProjectName),
			huh.NewInput().Title(c.ColorFromString("Project URL", video.ProjectURL)).Value(&video.ProjectURL),
			huh.NewInput().Title(c.ColorFromString("Sponsorship amount", video.Sponsored)).Value(&video.Sponsored),
			huh.NewInput().Title(sponsoredEmailsTitle).Value(&sponsoredEmailsString),
			huh.NewInput().Title(c.ColorFromStringInverse("Sponsorship blocked", video.SponsorshipBlocked)).Value(&video.SponsorshipBlocked),
			huh.NewInput().Title(c.ColorFromString("Subject", video.Subject)).Value(&video.Subject),
			huh.NewInput().Title(c.ColorFromString("Publish date (e.g., 2030-01-21T16:00)", video.Date)).Value(&video.Date),
			huh.NewConfirm().Title(c.ColorFromBool("Delayed", !video.Delayed)).Value(&video.Delayed),
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
		),
	)
	err := form.Run()
	if err != nil {
		return Video{}, err
	}
	video.SponsoredEmails = deleteEmpty(strings.Split(sponsoredEmailsString, ","))
	video.Init.Completed, video.Init.Total = c.Count([]interface{}{
		video.ProjectName,
		video.ProjectURL,
		video.Sponsored,
		video.Subject,
		video.Date,
	})
	if _, completed := c.ColorFromSponsoredEmails("Sponsorship emails (comma separated)", video.Sponsored, video.SponsoredEmails); completed {
		video.Init.Completed++
	}
	if video.SponsorshipBlocked == "" {
		video.Init.Completed++
	}
	if !video.Delayed {
		video.Init.Completed++
	}
	video.Init.Total += 3
	if save {
		yaml := YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, err
}

func (c *Choices) ChooseWork(video Video) (Video, error) {
	save := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title(c.ColorFromBool("Code done", video.Code)).Value(&video.Code),
			huh.NewConfirm().Title(c.ColorFromBool("Screen done", video.Screen)).Value(&video.Screen),
			huh.NewConfirm().Title(c.ColorFromBool("Talking head done", video.Head)).Value(&video.Head),
			huh.NewText().Lines(3).Title(c.ColorFromString("Related videos", video.RelatedVideos)).Value(&video.RelatedVideos),
			huh.NewConfirm().Title(c.ColorFromBool("Thumbnails done", video.Thumbnails)).Value(&video.Thumbnails),
			huh.NewConfirm().Title(c.ColorFromBool("Diagrams done", video.Diagrams)).Value(&video.Diagrams),
			huh.NewInput().Title(c.ColorFromString("Files location", video.Location)).Value(&video.Location),
			huh.NewInput().Title(c.ColorFromString("Tagline", video.Tagline)).Value(&video.Tagline),
			huh.NewInput().Title(c.ColorFromString("Tagline ideas", video.TaglineIdeas)).Value(&video.TaglineIdeas),
			huh.NewInput().Title(c.ColorFromString("Other logos", video.OtherLogos)).Value(&video.OtherLogos),
			huh.NewConfirm().Title(c.ColorFromBool("Screenshots done", video.Screenshots)).Value(&video.Screenshots),
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
		),
	)
	err := form.Run()
	if err != nil {
		return Video{}, err
	}
	video.Work.Completed, video.Work.Total = c.Count([]interface{}{
		video.Code,
		video.Screen,
		video.Head,
		video.RelatedVideos,
		video.Thumbnails,
		video.Diagrams,
		video.Location,
		video.Tagline,
		video.TaglineIdeas,
		video.OtherLogos,
		video.Screenshots,
	})
	if save {
		yaml := YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, err
}

func (c *Choices) ChooseDefineAI(video *Video, field *string, fieldName string, initialQuestion string) error {
	firstIteration := true
	askAgain := true
	question := ""
	chat := NewAIChatYouTube(settings.AI.Endpoint, settings.AI.Key, settings.AI.Deployment)
	history := ""
	defer chat.Close()
	for askAgain || firstIteration {
		askAgain = false
		if firstIteration {
			firstIteration = false
			question = initialQuestion
		} else {
			responses, err := chat.Chat(question)
			if err != nil {
				log.Fatal(err)
			}
			for _, resp := range responses {
				resp = strings.ReplaceAll(resp, "\"", "")
				*field = resp
				history = fmt.Sprintf("%s\n%s\n---\n", history, resp)
			}
			question = ""
		}
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewText().Lines(20).CharLimit(10000).Title("AI Responses").Value(&history),
				huh.NewText().Lines(20).CharLimit(1000).Title(c.ColorFromString(fmt.Sprintf("Current %s", fieldName), *field)).Value(field),
				huh.NewInput().Title(c.ColorFromString("Question", *field)).Value(&question),
				huh.NewConfirm().Affirmative("Ask").Negative("Save & Continue").Value(&askAgain),
			),
		)
		err := form.Run()
		if err != nil {
			return err
		}
	}
	yaml := YAML{}
	yaml.WriteVideo(*video, video.Path)
	return nil
}

func (c *Choices) ChooseDefine(video Video) (Video, error) {
	// Title
	err := c.ChooseDefineAI(
		&video,
		&video.Title,
		"Title",
		fmt.Sprintf("Write a title for a youtube video about %s", video.Subject),
	)
	if err != nil {
		return video, err
	}

	// Description
	err = c.ChooseDefineAI(
		&video,
		&video.Description,
		"Description",
		fmt.Sprintf("Write a short description for a youtube video with the title \"%s\"", video.Title),
	)
	if err != nil {
		return video, err
	}
	// Tags
	err = c.ChooseDefineAI(
		&video,
		&video.Tags,
		"Tags",
		fmt.Sprintf("Write comma separated tags for a youtube video with the description \"%s\"", video.Description),
	)
	// Description tags
	err = c.ChooseDefineAI(
		&video,
		&video.DescriptionTags,
		"Description Tags",
		fmt.Sprintf("Write up to 4 tags separated with # for a youtube video with the description \"%s\"", video.Description),
	)
	if err != nil {
		return video, err
	}
	// Tweet
	err = c.ChooseDefineAI(
		&video,
		&video.Tweet,
		"Tweet",
		fmt.Sprintf("Write a Tweet about a YouTube video with the title \"%s\". Include @DevOpsToolkit into it.", video.Title),
	)
	if err != nil {
		return video, err
	}

	// The rest
	save := true
	requestThumbnailOrig := video.RequestThumbnail
	animationsPlaceHolder := fmt.Sprintf(`- Animation: Subscribe (anywhere in the video)
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
		video.Members,
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().Lines(10).Title(c.ColorFromString("Animations", video.Animations)).Value(&video.Animations).Editor("vi").Placeholder(animationsPlaceHolder),
			huh.NewConfirm().Title(c.ColorFromBool("Thumbnail requested", video.RequestThumbnail)).Value(&video.RequestThumbnail),
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
		),
	)
	err = form.Run()
	if err != nil {
		return Video{}, err
	}
	video.Define.Completed, video.Define.Total = c.Count([]interface{}{
		video.Title,
		video.Description,
		video.Tags,
		video.DescriptionTags,
		video.RequestThumbnail,
		video.Animations,
		video.Tweet,
	})
	if !requestThumbnailOrig && video.RequestThumbnail {
		email := NewEmail(settings.Email.Password)
		if email.SendThumbnail(settings.Email.From, settings.Email.ThumbnailTo, video) != nil {
			panic(err)
		}
	}
	if save {
		yaml := YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, err
}

func (c *Choices) ChooseEdit(video Video) (Video, error) {
	save := true
	requestEditOrig := video.RequestEdit
	gistOrig := video.Gist
	playlistOptions := huh.NewOptions[string]()
	for key, value := range getYouTubePlaylists() {
		selected := false
		for _, existing := range video.Playlists {
			if key == existing.Id {
				selected = true
			}
		}
		playlistOptions = append(playlistOptions, huh.NewOption(value, value).Selected(selected))
	}
	var playlists []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title(c.ColorFromString("Thumbnail path", video.Thumbnail)).Value(&video.Thumbnail),
			huh.NewInput().Title(c.ColorFromString("Members (comma separated)", video.Members)).Value(&video.Members),
			huh.NewConfirm().Title(c.ColorFromBool("Edit requested", video.RequestEdit)).Value(&video.RequestEdit),
			huh.NewText().Lines(5).Title(c.ColorFromString("Timecodes", video.Timecodes)).Value(&video.Timecodes),
			huh.NewConfirm().Title(c.ColorFromBool("Movie done", video.Movie)).Value(&video.Movie),
			huh.NewConfirm().Title(c.ColorFromBool("Slides done", video.Slides)).Value(&video.Slides),
			huh.NewInput().Title(c.ColorFromString("Gist path", video.Gist)).Value(&video.Gist),
			huh.NewMultiSelect[string]().Title("Playlists").Options(playlistOptions...).Value(&playlists),
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
		),
	)
	err := form.Run()
	if err != nil {
		return Video{}, err
	}
	if save {
		yaml := YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	video.Edit.Completed, video.Edit.Total = c.Count([]interface{}{
		video.Thumbnail,
		video.Members,
		video.RequestEdit,
		video.Timecodes,
		video.Movie,
		video.Slides,
		video.Gist,
		video.Playlists,
	})
	if !requestEditOrig && video.RequestEdit {
		email := NewEmail(settings.Email.Password)
		if email.SendEdit(settings.Email.From, settings.Email.EditTo, video) != nil {
			panic(err)
		}
	}
	if len(gistOrig) == 0 && len(video.Gist) > 0 {
		repo := Repo{}
		video.GistUrl, err = repo.Gist(video.Gist, video.Title, video.ProjectName, video.ProjectURL, video.RelatedVideos)
		if err != nil {
			panic(err)
		}
	}
	video.Playlists = []Playlist{}
	if len(playlists) > 0 {
		for _, value := range playlists {
			if len(value) > 0 {
				id := strings.Split(value, " - ")[1]
				title := strings.Split(value, " - ")[0]
				playlist := Playlist{Title: title, Id: id}
				video.Playlists = append(video.Playlists, playlist)
			}
		}
	}
	if save {
		yaml := YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, err
}

func (c *Choices) ChoosePublish(video Video) (Video, error) {
	save := true
	sponsorsNotifyText := "Sponsors notify"
	notifiedSponsorsOrig := video.NotifiedSponsors
	if video.NotifiedSponsors || len(video.Sponsored) == 0 || video.Sponsored == "N/A" || video.Sponsored == "-" {
		sponsorsNotifyText = greenStyle.Render(sponsorsNotifyText)
	} else {
		sponsorsNotifyText = redStyle.Render(sponsorsNotifyText)
	}
	fields := []huh.Field{
		huh.NewInput().Title(c.ColorFromString("Upload video", video.UploadVideo)).Value(&video.UploadVideo),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("Twitter post", video.TweetPosted)).Value(&video.TweetPosted),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("LinkedIn post", video.LinkedInPosted)).Value(&video.LinkedInPosted),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("Slack post", video.SlackPosted)).Value(&video.SlackPosted),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("Reddit post", video.RedditPosted)).Value(&video.RedditPosted),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("Hacker News post", video.HNPosted)).Value(&video.HNPosted),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("Technology Conversations post", video.TCPosted)).Value(&video.TCPosted),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("YouTube Highlight", video.YouTubeHighlight)).Value(&video.YouTubeHighlight),
		huh.NewConfirm().Title(c.ColorFromBool("Pinned comment", video.YouTubeComment)).Value(&video.YouTubeComment),
		huh.NewConfirm().Title(c.ColorFromBool("Replies to comments", video.YouTubeCommentReply)).Value(&video.YouTubeCommentReply),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("https://gde.advocu.com post", video.GDE)).Value(&video.GDE),
		// TODO: Automate
		huh.NewConfirm().Title(c.ColorFromBool("Twitter Spaces post", video.TwitterSpace)).Value(&video.TwitterSpace),
		huh.NewInput().Title(c.ColorFromString("Code repo", video.Repo)).Value(&video.Repo),
		huh.NewConfirm().Title(sponsorsNotifyText).Value(&video.NotifiedSponsors),
	}
	for index := range fields {
		uploadVideoOrig := video.UploadVideo
		tweetPostedOrig := video.TweetPosted
		linkedInPostedOrig := video.LinkedInPosted
		slackPostedOrig := video.SlackPosted
		redditPostedOrig := video.RedditPosted
		hnPostedOrig := video.HNPosted
		tcPosted := video.TCPosted
		twitterSpaceOrig := video.TwitterSpace
		repoOrig := video.Repo
		form := huh.NewForm(
			huh.NewGroup(
				fields[index],
				huh.NewConfirm().Affirmative("Save & continue").Negative("Cancel").Value(&save),
			),
		)
		err := form.Run()
		if err != nil {
			return Video{}, err
		}
		video.Publish.Completed, video.Publish.Total = c.Count([]interface{}{
			video.UploadVideo,
			video.TweetPosted,
			video.LinkedInPosted,
			video.SlackPosted,
			video.RedditPosted,
			video.HNPosted,
			video.TCPosted,
			video.YouTubeHighlight,
			video.YouTubeComment,
			video.YouTubeCommentReply,
			video.GDE,
			video.TwitterSpace,
			video.Repo,
		})
		video.Publish.Total++
		if video.NotifiedSponsors || len(video.Sponsored) == 0 || video.Sponsored == "N/A" || video.Sponsored == "-" {
			video.Publish.Completed++
		}
		if len(uploadVideoOrig) == 0 && len(video.UploadVideo) > 0 {
			video.VideoId = uploadVideo(video)
			uploadThumbnail(video)
			// TODO: Automate
			println(confirmationStyle.Render(`Following should be set manually:
- End screen
- Playlists
- Language
- Monetization`))
		}
		twitter := Twitter{}
		if !tweetPostedOrig && len(video.Tweet) > 0 && video.TweetPosted {
			twitter.Post(video.Tweet, video.VideoId)
		}
		if !linkedInPostedOrig && len(video.Tweet) > 0 && video.LinkedInPosted {
			postLinkedIn(video.Tweet)
		}
		if !slackPostedOrig && len(video.VideoId) > 0 && video.SlackPosted {
			postSlack(video.VideoId)
		}
		if !redditPostedOrig && len(video.VideoId) > 0 && video.RedditPosted {
			postReddit(video.Title, video.VideoId)
		}
		if !hnPostedOrig && len(video.VideoId) > 0 && video.HNPosted {
			postHackerNews(video.Title, video.VideoId)
		}
		if !tcPosted && len(video.VideoId) > 0 && video.TCPosted {
			postTechnologyConversations(video.Title, video.Description, video.VideoId, video.Gist, video.ProjectName, video.ProjectURL, video.RelatedVideos)
		}
		if !twitterSpaceOrig && len(video.VideoId) > 0 && video.TwitterSpace {
			twitter.PostSpace(video.VideoId)
		}
		if len(repoOrig) == 0 && len(video.Repo) > 0 && video.Repo != "N/A" {
			repo := Repo{}
			repo.Update(video.Repo, video.Title, video.VideoId)
		}
		if !notifiedSponsorsOrig && video.NotifiedSponsors {
			email := NewEmail(settings.Email.Password)
			email.SendSponsors(settings.Email.From, video.SponsoredEmails, video.VideoId, video.Sponsored)
		}
		if !save {
			break
		}
		yaml := YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, nil
}

func (c *Choices) ColorFromSponsoredEmails(title, sponsored string, sponsoredEmails []string) (string, bool) {
	if len(sponsored) == 0 || sponsored == "N/A" || sponsored == "-" || len(sponsoredEmails) > 0 {
		return greenStyle.Render(title), true
	}
	return redStyle.Render(title), false
}

func (c *Choices) ColorFromString(title, value string) string {
	if len(value) > 0 {
		return greenStyle.Render(title)
	}
	return redStyle.Render(title)
}

func (c *Choices) ColorFromStringInverse(title, value string) string {
	if len(value) > 0 {
		return redStyle.Render(title)
	}
	return greenStyle.Render(title)
}

func (c *Choices) ColorFromBool(title string, value bool) string {
	if value {
		return greenStyle.Render(title)
	}
	return redStyle.Render(title)
}

func (c *Choices) ChooseVideosPhase(vi []VideoIndex) bool {
	var selection int
	phases := make(map[int]int)
	for i := range vi {
		phase := c.GetVideoPhase(vi[i])
		phases[phase] = phases[phase] + 1
	}
	options := huh.NewOptions[int]()
	if text, count := c.GetPhaseColoredText(phases, videosPhasePublished, "Published"); count > 0 {
		options = append(options, huh.NewOption(text, videosPhasePublished))
	}
	if text, count := c.GetPhaseColoredText(phases, videosPhasePublishPending, "Pending publish"); count > 0 {
		options = append(options, huh.NewOption(text, videosPhasePublishPending))
	}
	if text, count := c.GetPhaseColoredText(phases, videosPhaseEditRequested, "Edit requested"); count > 0 {
		options = append(options, huh.NewOption(text, videosPhaseEditRequested))
	}
	if text, count := c.GetPhaseColoredText(phases, videosPhaseMaterialDone, "Material done"); count > 0 {
		options = append(options, huh.NewOption(text, videosPhaseMaterialDone))
	}
	if text, count := c.GetPhaseColoredText(phases, videosPhaseStarted, "Started"); count > 0 {
		options = append(options, huh.NewOption(text, videosPhaseStarted))
	}
	if text, count := c.GetPhaseColoredText(phases, videosPhaseDelayed, "Delayed"); count > 0 {
		options = append(options, huh.NewOption(text, videosPhaseDelayed))
	}
	if text, count := c.GetPhaseColoredText(phases, videosPhaseSponsoredBlocked, "Sponsored blocked"); count > 0 {
		options = append(options, huh.NewOption(text, videosPhaseSponsoredBlocked))
	}
	if text, count := c.GetPhaseColoredText(phases, videosPhaseIdeas, "Ideas"); count > 0 {
		options = append(options, huh.NewOption(text, videosPhaseIdeas))
	}
	options = append(options, huh.NewOption("Return", actionReturn))
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("From which phase would you like to list the videos?").
				Options(options...).
				Value(&selection),
		),
	)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	if selection == actionReturn {
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
	const actionEdit = 0
	const actionDelete = 1
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
		path := c.GetFilePath(selectedVideoIndex.Category, selectedVideoIndex.Name, "yaml")
		yaml := YAML{}
		video := yaml.GetVideo(path)
		video.Path = path
		choices := Choices{}
		choices.ChoosePhase(video)
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

func (c *Choices) IsEmpty(str string) error {
	if len(str) == 0 {
		return errors.New("Required!")
	}
	return nil
}

func (c *Choices) GetPhaseColoredText(phases map[int]int, phase int, title string) (string, int) {
	if phase != actionReturn {
		title = fmt.Sprintf("%s (%d)", title, phases[phase])
		if phase == videosPhasePublished {
			return greenStyle.Render(title), phases[phase]
		} else if phase == videosPhasePublishPending && phases[phase] > 0 {
			return greenStyle.Render(title), phases[phase]
		} else if phase == videosPhaseEditRequested && phases[phase] > 0 {
			return greenStyle.Render(title), phases[phase]
		} else if (phase == videosPhaseMaterialDone || phase == videosPhaseIdeas) && phases[phase] >= 3 {
			return greenStyle.Render(title), phases[phase]
		} else if phase == videosPhaseStarted && phases[phase] > 0 {
			return greenStyle.Render(title), phases[phase]
		} else {
			return orangeStyle.Render(title), phases[phase]
		}
	}
	return title, phases[phase]
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
