package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

var confirmationStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(lipgloss.Color("#006E14")).
	PaddingTop(1).
	PaddingBottom(1).
	PaddingLeft(5).
	PaddingRight(5).
	MarginTop(1).
	MarginBottom(1)

var errorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(lipgloss.Color("1")).
	PaddingTop(1).
	PaddingBottom(1).
	PaddingLeft(5).
	PaddingRight(5).
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

const indexCreateVideo = 0
const indexListVideos = 1

const actionEdit = 0
const actionDelete = 1
const actionReturn = 99

func (c *Choices) ChooseIndex() {
	var selectedIndex int
	yaml := YAML{IndexPath: "index.yaml"}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("What do you want to do?").
				Options(c.getIndexOptions()...).
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
	case actionReturn:
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
	errorMsg := ""
	for !returnVar {
		const phaseInit = 0
		const phaseWork = 1
		const phaseDefine = 2
		const phaseEdit = 3
		const phasePublish = 4
		var selected int
		title := "Which type of tasks would you like to work on?"
		if len(errorMsg) > 0 {
			title = fmt.Sprintf("%s\n%s", errorStyle.Render(errorMsg), title)
		}
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title(title).
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
		errorMsg = ""
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
				errorMsg = err.Error()
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
	var name, category string
	save := true
	fields, err := c.getCreateVideoFields(&name, &category, &save)
	if err != nil {
		panic(err)
	}
	form := huh.NewForm(huh.NewGroup(fields...))
	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}
	vi := VideoIndex{
		Name:     name,
		Category: category,
	}
	if !save {
		return vi
	}
	dirPath := c.GetDirPath(vi.Category)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0755)
	}
	scriptContent := `## Intro

FIXME: Explanation...

FIXME: This is...

FIXME: It's supposed to...

## Setup

FIXME:

## FIXME:

FIXME:

## FIXME: Pros and Cons

TODO: Header: Cons; Items: FIXME:

TODO: Header: Pros; Items: FIXME:

## Destroy

FIXME:
`
	filePath := c.GetFilePath(vi.Category, vi.Name, "md")
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
	explainedOrig := video.Explained
	if len(video.Gist) == 0 {
		video.Gist = strings.Replace(video.Path, ".yaml", ".md", 1)
	}
	sponsoredEmailsTitle, _ := c.ColorFromSponsoredEmails("Sponsorship emails (comma separated)", video.Sponsorship.Amount, video.Sponsorship.Emails)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title(c.ColorFromString("Project name", video.ProjectName)).Value(&video.ProjectName),
			huh.NewInput().Title(c.ColorFromString("Project URL", video.ProjectURL)).Value(&video.ProjectURL),
			huh.NewInput().Title(c.ColorFromString("Sponsorship amount", video.Sponsorship.Amount)).Value(&video.Sponsorship.Amount),
			huh.NewInput().Title(sponsoredEmailsTitle).Value(&video.Sponsorship.Emails),
			huh.NewInput().Title(c.ColorFromStringInverse("Sponsorship blocked", video.Sponsorship.Blocked)).Value(&video.Sponsorship.Blocked),
			huh.NewInput().Title(c.ColorFromString("Publish date (e.g., 2030-01-21T16:00)", video.Date)).Value(&video.Date),
			huh.NewConfirm().Title(c.ColorFromBool("Delayed", !video.Delayed)).Value(&video.Delayed),
			huh.NewInput().Title(c.ColorFromString("Gist path", video.Gist)).Value(&video.Gist),
			huh.NewConfirm().Title(c.ColorFromBool("Explain Project", video.Explained)).Value(&video.Explained),
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
		),
	)
	err := form.Run()
	if err != nil {
		return Video{}, err
	}
	// TODO: Remove
	if len(video.Sponsorship.Amount) == 0 {
		video.Sponsorship.Amount = video.Sponsored
	}
	if len(video.Sponsorship.Blocked) == 0 {
		video.Sponsorship.Blocked = video.SponsorshipBlocked
	}
	video.Init.Completed, video.Init.Total = c.Count([]interface{}{
		video.ProjectName,
		video.ProjectURL,
		video.Sponsorship.Amount,
		video.Gist,
		video.Explained,
		video.Date,
	})
	if _, completed := c.ColorFromSponsoredEmails("Sponsorship emails (comma separated)", video.Sponsorship.Amount, video.Sponsorship.Emails); completed {
		video.Init.Completed++
	}
	if video.Sponsorship.Blocked == "" {
		video.Init.Completed++
	}
	if !video.Delayed {
		video.Init.Completed++
	}
	if !explainedOrig && video.Explained {
		cmd := exec.Command("fabric", "--pattern", "explain_project_dot", video.ProjectURL)
		output, err := cmd.Output()
		if err != nil {
			return video, err
		}
		content, err := os.ReadFile(video.Gist)
		if err != nil {
			return video, err
		}
		contentWithExplanation := strings.Replace(string(content), "FIXME: Explanation...", string(output), 1)
		os.WriteFile(video.Gist, []byte(contentWithExplanation), 0644)
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
			huh.NewConfirm().Title(c.ColorFromBool("Talking head done", video.Head)).Value(&video.Head),
			huh.NewConfirm().Title(c.ColorFromBool("Screen done", video.Screen)).Value(&video.Screen),
			huh.NewText().Lines(3).CharLimit(10000).Title(c.ColorFromString("Related videos", video.RelatedVideos)).Value(&video.RelatedVideos),
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

func (c *Choices) ChooseFabric(video *Video, field *string, fieldName, pattern string, addToField bool) error {
	askAgain := true
	content, err := os.ReadFile(video.Gist)
	if err != nil {
		return err
	}
	firstIteration := true
	output := ""
	for askAgain || firstIteration {
		askAgain = false
		if firstIteration {
			firstIteration = false
		} else {
			cmd := exec.Command("fabric", "--pattern", pattern, string(content))
			outputBytes, err := cmd.Output()
			if err != nil {
				return err
			}
			output = string(outputBytes)
			output = strings.ReplaceAll(output, "TAGS:", "")
			if addToField {
				*field = output
			}
		}
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewText().Lines(20).CharLimit(10000).Title(c.ColorFromString(fieldName, *field)).Value(field),
				huh.NewText().Lines(20).CharLimit(10000).Title("AI Responses").Value(&output),
				huh.NewConfirm().Affirmative("Ask").Negative("Save & Continue").Value(&askAgain),
			).Title(fieldName),
		)
		err = form.Run()
		if err != nil {
			return err
		}
	}
	yaml := YAML{}
	yaml.WriteVideo(*video, video.Path)
	return nil
}

// TODO: Remove
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
				huh.NewText().Lines(20).CharLimit(10000).Title(c.ColorFromString(fieldName, *field)).Value(field),
				huh.NewText().Lines(20).CharLimit(10000).Title("AI Responses").Value(&history),
				huh.NewInput().Title(c.ColorFromString("Question", *field)).Value(&question),
				huh.NewConfirm().Affirmative("Ask").Negative("Save & Continue").Value(&askAgain),
			).Title(fieldName),
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
	if err := c.ChooseFabric(&video, &video.Title, "Title", "title_dot", false); err != nil {
		return video, err
	}

	// Description
	if err := c.ChooseFabric(&video, &video.Description, "Description", "description_dot", true); err != nil {
		return video, err
	}

	// Highlight
	if err := c.ChooseFabric(&video, &video.Highlight, "Highlight", "highlight_dot", true); err != nil {
		return video, err
	}

	// Tags
	// TODO: Change to Fabric
	if err := c.ChooseFabric(&video, &video.Tags, "Tags", "tags_dot", true); err != nil {
		return video, err
	}
	// err = c.ChooseDefineAI(
	// 	&video,
	// 	&video.Tags,
	// 	"Tags",
	// 	fmt.Sprintf("Write comma separated tags for a youtube video with the description \"%s\". Do not exceeed 450 characters.", video.Description),
	// )
	// if err != nil {
	// 	return video, err
	// }
	// Description tags
	// TODO: Change to Fabric
	err := c.ChooseDefineAI(
		&video,
		&video.DescriptionTags,
		"Description Tags",
		fmt.Sprintf("Write up to 4 tags separated with # for a youtube video with the description \"%s\"", video.Description),
	)
	if err != nil {
		return video, err
	}
	// Tweet
	// TODO: Change to Fabric
	err = c.ChooseDefineAI(
		&video,
		&video.Tweet,
		"Tweet",
		fmt.Sprintf("Write a Tweet about a YouTube video with the title \"%s\". Include @DevOpsToolkit into it. Use [YouTube Link] as a placeholder for the link to the vidfeo", video.Title),
	)
	if err != nil {
		return video, err
	}

	// Animations
	generateAnimations := true
	for generateAnimations {
		generateAnimations = false
		video.Animations = strings.TrimSpace(video.Animations)
		formAnimations := huh.NewForm(
			huh.NewGroup(
				huh.NewText().Lines(40).CharLimit(10000).Title(c.ColorFromString("Animations", video.Animations)).Value(&video.Animations).Editor("vi"),
				huh.NewConfirm().Affirmative("Generate").Negative("Continue").Value(&generateAnimations),
			).Title("Animations"),
		)
		err = formAnimations.Run()
		if err != nil {
			return Video{}, err
		}
		if generateAnimations {
			video.Animations = ""
			video.Timecodes = ""
			repo := Repo{}
			linesSlice, sectionsSlice, err := repo.GetAnimations(video.Gist)
			if err != nil {
				panic(err)
			}
			if err != nil {
				return Video{}, err
			}
			for _, line := range linesSlice {
				video.Animations = fmt.Sprintf("%s\n- %s", video.Animations, line)
			}
			if len(video.Timecodes) == 0 {
				video.Timecodes = "00:00 TODO:"
				for _, section := range sectionsSlice {
					video.Timecodes = fmt.Sprintf("%s\nTODO:TODO %s", video.Timecodes, strings.TrimLeft(section, "Section: "))
				}
			}
		}
	}
	// Thumbnail
	save := true
	requestThumbnailOrig := video.RequestThumbnail
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title(c.ColorFromBool("Thumbnail request", video.RequestThumbnail)).Value(&video.RequestThumbnail),
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
		video.Gist,
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
	timeCodesTitle := "Timecodes"
	if strings.Contains(video.Timecodes, "TODO:") {
		timeCodesTitle = redStyle.Render(timeCodesTitle)
	} else {
		timeCodesTitle = greenStyle.Render(timeCodesTitle)
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title(c.ColorFromString("Thumbnail 1 Path", video.Thumbnail)).Value(&video.Thumbnail),
			huh.NewInput().Title(c.ColorFromString("Thumbnail 2 Path", video.Thumbnail02)).Value(&video.Thumbnail02),
			huh.NewInput().Title(c.ColorFromString("Thumbnail 3 Path", video.Thumbnail03)).Value(&video.Thumbnail03),
			huh.NewInput().Title(c.ColorFromString("Members (comma separated)", video.Members)).Value(&video.Members),
			huh.NewConfirm().Title(c.ColorFromBool("Edit Request", video.RequestEdit)).Value(&video.RequestEdit),
			huh.NewText().Lines(5).CharLimit(10000).Title(timeCodesTitle).Value(&video.Timecodes),
			huh.NewConfirm().Title(c.ColorFromBool("Movie Done", video.Movie)).Value(&video.Movie),
			huh.NewConfirm().Title(c.ColorFromBool("Slides Done", video.Slides)).Value(&video.Slides),
			huh.NewConfirm().Title(c.ColorFromBool("Short done", video.Short)).Value(&video.Short),
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
		video.Thumbnail02,
		video.Thumbnail03,
		video.Members,
		video.RequestEdit,
		video.Movie,
		video.Slides,
		video.Short,
	})
	video.Edit.Total++
	if !strings.Contains(video.Timecodes, "TODO:") {
		video.Edit.Completed++
	}
	if !requestEditOrig && video.RequestEdit {
		email := NewEmail(settings.Email.Password)
		if err = email.SendEdit(settings.Email.From, settings.Email.EditTo, video); err != nil {
			return video, err
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
	if video.NotifiedSponsors || len(video.Sponsorship.Amount) == 0 || video.Sponsorship.Amount == "N/A" || video.Sponsorship.Amount == "-" {
		sponsorsNotifyText = greenStyle.Render(sponsorsNotifyText)
	} else {
		sponsorsNotifyText = redStyle.Render(sponsorsNotifyText)
	}
	createHugo := video.HugoPath != ""
	fields := []huh.Field{
		huh.NewConfirm().Title(c.ColorFromBool("Create Hugo Post", createHugo)).Value(&createHugo),
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
			video.HugoPath,
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
		if video.NotifiedSponsors || len(video.Sponsorship.Amount) == 0 || video.Sponsorship.Amount == "N/A" || video.Sponsorship.Amount == "-" {
			video.Publish.Completed++
		}
		if createHugo && len(video.HugoPath) == 0 {
			hugo := Hugo{}
			video.HugoPath, err = hugo.Post(video.Gist, video.Title, video.Date)
			if err != nil {
				return Video{}, err
			}
		} else if !createHugo {
			video.HugoPath = ""
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
			postLinkedIn(video.Tweet, video.VideoId)
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
			email.SendSponsors(settings.Email.From, video.Sponsorship.Emails, video.VideoId, video.Sponsorship.Amount)
		}
		if !save {
			break
		}
		yaml := YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, nil
}

func (c *Choices) ColorFromSponsoredEmails(title, sponsored string, sponsoredEmails string) (string, bool) {
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
	// TODO: Remove
	if len(video.Sponsorship.Blocked) == 0 {
		video.Sponsorship.Blocked = video.SponsorshipBlocked
	}
	if video.Delayed {
		return videosPhaseDelayed
	} else if len(video.Sponsorship.Blocked) > 0 {
		return videosPhaseSponsoredBlocked
	} else if len(video.Repo) > 0 {
		return videosPhasePublished
	} else if len(video.UploadVideo) > 0 && len(video.Tweet) > 0 {
		return videosPhasePublishPending
	} else if video.RequestEdit {
		return videosPhaseEditRequested
	} else if video.Code && video.Screen && video.Head && video.Diagrams {
		return videosPhaseMaterialDone
	} else if len(video.Date) > 0 {
		return videosPhaseStarted
	} else {
		return videosPhaseIdeas
	}
}

func (c *Choices) ChooseVideos(vi []VideoIndex, phase int) {
	var selectedVideo Video
	var selectedAction int
	options := huh.NewOptions[Video]()
	sortedVideos := []Video{}
	for i := range vi {
		videoPhase := c.GetVideoPhase(vi[i])
		if videoPhase == phase {
			yaml := YAML{}
			path := c.GetFilePath(vi[i].Category, vi[i].Name, "yaml")
			video := yaml.GetVideo(path)
			video.Name = vi[i].Name
			video.Path = path
			video.Index = i
			sortedVideos = append(sortedVideos, video)
		}
	}
	sort.Slice(sortedVideos, func(i, j int) bool {
		date1, _ := time.Parse("2006-01-02T15:04", sortedVideos[i].Date)
		date2, _ := time.Parse("2006-01-02T15:04", sortedVideos[j].Date)
		return date1.Before(date2)
	})
	for _, video := range sortedVideos {
		title := video.Name
		if len(video.Sponsorship.Blocked) > 0 && video.Sponsorship.Blocked != "-" && video.Sponsorship.Blocked != "N/A" {
			title = fmt.Sprintf("%s (%s)", title, video.Sponsorship.Blocked)
		} else {
			if len(video.Date) > 0 {
				title = fmt.Sprintf("%s (%s)", title, video.Date)
			}
			if len(video.Sponsorship.Amount) > 0 && video.Sponsorship.Amount != "-" && video.Sponsorship.Amount != "N/A" {
				title = fmt.Sprintf("%s (sponsored)", title)
			}
		}

		options = append(options, huh.NewOption(title, video))
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Video]().
				Title("Which video would you like to work on?").
				Options(options...).
				Value(&selectedVideo),
			huh.NewSelect[int]().
				Title("What would you like to do with the video?").
				Options(c.getActionOptions()...).
				Value(&selectedAction),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	switch selectedAction {
	case actionEdit:
		choices := Choices{}
		choices.ChoosePhase(selectedVideo)
	case actionDelete:
		shPath := strings.ReplaceAll(selectedVideo.Path, ".yaml", ".md")
		if os.Remove(shPath) != nil {
			panic(err)
		}
		os.Remove(selectedVideo.Path)
		// selectedVideoIndex = vi[len(vi)-1]
		vi = append(vi[:selectedVideo.Index], vi[selectedVideo.Index+1:]...)
	case actionReturn:
		return
	}
	yaml := YAML{IndexPath: "index.yaml"}
	yaml.WriteIndex(vi)
}

func (c *Choices) IsEmpty(str string) error {
	if len(str) == 0 {
		return errors.New("Required")
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
		} else if phase == videosPhaseStarted && phases[phase] >= 3 {
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

func (c *Choices) getCreateVideoFields(name, category *string, save *bool) ([]huh.Field, error) {
	categories, err := c.getCategories()
	if err != nil {
		return nil, err
	}
	return []huh.Field{
		huh.NewInput().Prompt("Name: ").Value(name),
		huh.NewSelect[string]().Title("Category").Options(categories...).Value(category),
		huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(save),
	}, nil
}

func (c *Choices) getCategories() ([]huh.Option[string], error) {
	files, err := os.ReadDir("manuscript")
	if err != nil {
		return nil, err
	}
	options := huh.NewOptions[string]()
	for _, file := range files {
		if file.IsDir() {
			caser := cases.Title(language.AmericanEnglish)
			categoryKey := strings.ReplaceAll(file.Name(), "-", " ")
			categoryKey = caser.String(categoryKey)
			options = append(options, huh.NewOption(categoryKey, file.Name()))
		}
	}
	return options, nil
}

func (c *Choices) getIndexOptions() []huh.Option[int] {
	return []huh.Option[int]{
		huh.NewOption("List Videos", indexListVideos),
		huh.NewOption("Create Video", indexCreateVideo),
		huh.NewOption("Exit", actionReturn),
	}
}

func (c *Choices) getActionOptions() []huh.Option[int] {
	return []huh.Option[int]{
		huh.NewOption("Edit", actionEdit),
		huh.NewOption("Delete", actionDelete),
		// TODO: Add the option to move video files to a different directory
		huh.NewOption("Return", actionReturn),
	}
}
