package main

import (
	"bytes"
	"errors"
	"fmt"

	// "io" // Comment out if not used elsewhere after removals
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"devopstoolkitseries/youtube-automation/internal/configuration"
	"devopstoolkitseries/youtube-automation/internal/notification"
	repository "devopstoolkitseries/youtube-automation/internal/repository"
	"devopstoolkitseries/youtube-automation/internal/storage"
	"devopstoolkitseries/youtube-automation/pkg/bluesky"
	"devopstoolkitseries/youtube-automation/pkg/utils"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Video struct is defined in yaml.go

// Directory represents a selectable directory option.
// Name is for display, Path is the actual file system path.
// Used by getAvailableDirectories and selectTargetDirectory.
type Directory struct {
	Name string
	Path string
}

// DirectorySelector defines an interface for selecting a directory.
// This allows for mocking in tests.
type DirectorySelector interface {
	SelectDirectory(input *bytes.Buffer) (Directory, error)
}

// confirmer defines an interface for confirming actions.
// This allows for mocking in tests.
type confirmer interface {
	Confirm(message string) bool
}

// defaultConfirmer is the default implementation of confirmer using utils.ConfirmAction.
type defaultConfirmer struct{}

func (dc defaultConfirmer) Confirm(message string) bool {
	return utils.ConfirmAction(message)
}

type Choices struct {
	confirmer   confirmer
	getDirsFunc func() ([]Directory, error)
	dirSelector DirectorySelector // New field for injecting directory selection behavior
}

// NewChoices creates a new instance of Choices with the default confirmer,
// default directory listing function, and default directory selector.
func NewChoices() *Choices {
	c := &Choices{confirmer: defaultConfirmer{}}
	c.getDirsFunc = c.doGetAvailableDirectories
	c.dirSelector = c // Choices implements DirectorySelector via selectTargetDirectory
	return c
}

var redStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("1"))

var greenStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("2"))

var orangeStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("3"))

var farFutureStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("6")) // Cyan

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

// getCustomHuhTheme creates a custom theme for huh forms.
// This theme allows pre-rendered styles for unselected options to show through
// and applies a distinct style for selected options.
func getCustomHuhTheme() *huh.Theme {
	theme := huh.ThemeCharm() // Start with a copy of the Charm theme

	// For UNSELECTED options:
	// Apply an empty style. This allows pre-rendered styles (like cyan for far-future videos)
	// from getVideoTitleForDisplay to be visible in the resting state.
	theme.Focused.UnselectedOption = lipgloss.NewStyle()

	// For SELECTED options (when hovered/active):
	// Apply a style that gives clear visual feedback and overrides pre-rendered styles.
	// Reverse(true) swaps the item's foreground and background colors.
	theme.Focused.SelectedOption = lipgloss.NewStyle().Reverse(true)

	// Ensure the selector (e.g., '>') remains styled as per Charm theme (fuchsia).
	// This might be inherited correctly, but re-asserting to be safe.
	theme.Focused.SelectSelector = lipgloss.NewStyle().Foreground(lipgloss.Color("#F780E2")).SetString("> ")

	return theme
}

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

const (
	actionEdit = iota
	actionDelete
	actionMoveFiles
)
const actionReturn = 99

func (c *Choices) ChooseIndex() {
	var selectedIndex int
	yaml := storage.YAML{IndexPath: "index.yaml"}
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

func (c *Choices) GetPhaseText(text string, task storage.Tasks) string {
	text = fmt.Sprintf("%s (%d/%d)", text, task.Completed, task.Total)
	if task.Completed == task.Total && task.Total > 0 {
		return greenStyle.Render(text)
	}
	return orangeStyle.Render(text)
}

func (c *Choices) ChoosePhase(video storage.Video) {
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
			var errInit error // Renamed to avoid conflict
			if video, errInit = c.ChooseInit(video); errInit != nil {
				panic(errInit)
			}
		case phaseWork:
			var errWork error // Renamed
			if video, errWork = c.ChooseWork(video); errWork != nil {
				panic(errWork)
			}
		case phaseDefine:
			var errDefine error // Renamed
			if video, errDefine = c.ChooseDefine(video); errDefine != nil {
				panic(errDefine)
			}
		case phaseEdit:
			var errEdit error // Renamed
			if video, errEdit = c.ChooseEdit(video); errEdit != nil {
				errorMsg = errEdit.Error()
			}
		case phasePublish:
			var errPublish error // Renamed
			if video, errPublish = c.ChoosePublish(video); errPublish != nil {
				panic(errPublish)
			}
		case actionReturn:
			returnVar = true
		}
	}
}

func (c *Choices) ChooseCreateVideo() storage.VideoIndex {
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
	vi := storage.VideoIndex{
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

FIXME: Shock

FIXME: Establish expectations

FIXME: What's the ending?

## Setup

FIXME:

## FIXME:

FIXME:

## FIXME: Pros and Cons

FIXME: Header: Cons; Items: FIXME:

FIXME: Header: Pros; Items: FIXME:

## Destroy

FIXME:
`
	filePath := c.GetFilePath(vi.Category, vi.Name, "md")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		f, errCreate := os.Create(filePath) // Renamed err
		if errCreate != nil {
			panic(errCreate)
		}
		defer f.Close()
		f.Write([]byte(scriptContent))
		return vi
	}
	return storage.VideoIndex{}
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

func (c *Choices) ChooseInit(video storage.Video) (storage.Video, error) {
	save := true
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
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
		),
	)
	err := form.Run()
	if err != nil {
		return storage.Video{}, err
	}
	video.Init.Completed, video.Init.Total = c.Count([]interface{}{
		video.ProjectName,
		video.ProjectURL,
		video.Sponsorship.Amount,
		video.Gist,
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
	video.Init.Total += 3
	if save {
		yaml := storage.YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, nil // Return nil error if successful
}

func (c *Choices) ChooseWork(video storage.Video) (storage.Video, error) {
	save := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title(c.ColorFromBool("Code done", video.Code)).Value(&video.Code),
			huh.NewConfirm().Title(c.ColorFromBool("Talking head done", video.Head)).Value(&video.Head),
			huh.NewConfirm().Title(c.ColorFromBool("Screen done", video.Screen)).Value(&video.Screen),
			huh.NewText().Lines(3).CharLimit(10000).Title(c.ColorFromString("Related videos", video.RelatedVideos)).Value(&video.RelatedVideos),
			huh.NewConfirm().Title(c.ColorFromBool("Thumbnails done", video.Thumbnails)).Value(&video.Thumbnails),
			huh.NewConfirm().Title(c.ColorFromBool("Diagrams done", video.Diagrams)).Value(&video.Diagrams),
			huh.NewConfirm().Title(c.ColorFromBool("Screenshots done", video.Screenshots)).Value(&video.Screenshots),
			huh.NewInput().Title(c.ColorFromString("Files location", video.Location)).Value(&video.Location),
			huh.NewInput().Title(c.ColorFromString("Tagline", video.Tagline)).Value(&video.Tagline),
			huh.NewInput().Title(c.ColorFromString("Tagline ideas", video.TaglineIdeas)).Value(&video.TaglineIdeas),
			huh.NewInput().Title(c.ColorFromString("Other logos", video.OtherLogos)).Value(&video.OtherLogos),
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
		),
	)
	err := form.Run()
	if err != nil {
		return storage.Video{}, err
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
		yaml := storage.YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, nil // Return nil error if successful
}

func (c *Choices) ChooseFabric(video *storage.Video, field *string, fieldName, pattern string, addToField bool) error {
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
			outputBytes, cmdErr := cmd.Output() // Renamed err
			if cmdErr != nil {
				return fmt.Errorf("%s\n%s", cmdErr.Error(), string(outputBytes))
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
		formErr := form.Run() // Renamed err
		if formErr != nil {
			return formErr
		}
	}
	yaml := storage.YAML{}
	yaml.WriteVideo(*video, video.Path)
	return nil
}

func (c *Choices) ChooseDefine(video storage.Video) (storage.Video, error) {
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
	if err := c.ChooseFabric(&video, &video.Tags, "Tags", "tags_dot", true); err != nil {
		return video, err
	}

	// Description tags
	if err := c.ChooseFabric(&video, &video.DescriptionTags, "Description Tags", "description_tags_dot", true); err != nil {
		return video, err
	}

	// Tweet
	if err := c.ChooseFabric(&video, &video.Tweet, "Tweet", "tweet", true); err != nil {
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
		err := formAnimations.Run()
		if err != nil {
			return storage.Video{}, err
		}
		if generateAnimations {
			video.Animations = ""
			video.Timecodes = ""
			repo := repository.Repo{}                                                     // UPDATED
			linesSlice, sectionsSlice, errGetAnimations := repo.GetAnimations(video.Gist) // Renamed err
			if errGetAnimations != nil {
				return storage.Video{}, fmt.Errorf("failed to get animations: %w", errGetAnimations)
			}
			for _, line := range linesSlice {
				video.Animations = fmt.Sprintf("%s\n- %s", video.Animations, line)
			}
			if len(video.Timecodes) == 0 {
				video.Timecodes = "00:00 FIXME:"
				for _, section := range sectionsSlice {
					video.Timecodes = fmt.Sprintf("%s\nFIXME:FIXME %s", video.Timecodes, strings.TrimLeft(section, "Section: "))
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
	err := form.Run()
	if err != nil {
		return storage.Video{}, err
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
		email := notification.NewEmail(configuration.GlobalSettings.Email.Password)
		if errSend := email.SendThumbnail(configuration.GlobalSettings.Email.From, configuration.GlobalSettings.Email.ThumbnailTo, video); errSend != nil {
			return storage.Video{}, fmt.Errorf("failed to send thumbnail email: %w", errSend)
		}
	}
	if save {
		yaml := storage.YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	return video, nil
}

func (c *Choices) ChooseEdit(video storage.Video) (storage.Video, error) {
	save := true
	requestEditOrig := video.RequestEdit
	timeCodesTitle := "Timecodes"
	if strings.Contains(video.Timecodes, "FIXME:") {
		timeCodesTitle = redStyle.Render(timeCodesTitle)
	} else {
		timeCodesTitle = greenStyle.Render(timeCodesTitle)
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title(c.ColorFromString("Thumbnail Path", video.Thumbnail)).Value(&video.Thumbnail),
			huh.NewInput().Title(c.ColorFromString("Members (comma separated)", video.Members)).Value(&video.Members),
			huh.NewConfirm().Title(c.ColorFromBool("Edit Request", video.RequestEdit)).Value(&video.RequestEdit),
			huh.NewText().Lines(5).CharLimit(10000).Title(timeCodesTitle).Value(&video.Timecodes),
			huh.NewConfirm().Title(c.ColorFromBool("Movie Done", video.Movie)).Value(&video.Movie),
			huh.NewConfirm().Title(c.ColorFromBool("Slides Done", video.Slides)).Value(&video.Slides),
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
		),
	)
	err := form.Run()
	if err != nil {
		return storage.Video{}, err
	}
	if save { // Save before sending email in case send fails
		yaml := storage.YAML{}
		yaml.WriteVideo(video, video.Path)
	}
	video.Edit.Completed, video.Edit.Total = c.Count([]interface{}{
		video.Thumbnail,
		video.Members,
		video.RequestEdit,
		video.Movie,
		video.Slides,
	})
	video.Edit.Total++
	if !strings.Contains(video.Timecodes, "FIXME:") {
		video.Edit.Completed++
	}
	if !requestEditOrig && video.RequestEdit {
		email := notification.NewEmail(configuration.GlobalSettings.Email.Password)
		if errSend := email.SendEdit(configuration.GlobalSettings.Email.From, configuration.GlobalSettings.Email.EditTo, video); errSend != nil {
			return video, errSend
		}
	}
	return video, nil
}

func (c *Choices) ChoosePublish(video storage.Video) (storage.Video, error) {
	var errorMsg string
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
		huh.NewConfirm().Title(c.ColorFromBool("BlueSky post", video.BlueSkyPosted)).Value(&video.BlueSkyPosted),
		huh.NewConfirm().Title(c.ColorFromBool("LinkedIn post", video.LinkedInPosted)).Value(&video.LinkedInPosted),
		huh.NewConfirm().Title(c.ColorFromBool("Slack post", video.SlackPosted)).Value(&video.SlackPosted),
		huh.NewConfirm().Title(c.ColorFromBool("Hacker News post", video.HNPosted)).Value(&video.HNPosted),
		huh.NewConfirm().Title(c.ColorFromBool("devopstoolkit.live", video.DOTPosted)).Value(&video.DOTPosted),
		huh.NewConfirm().Title(c.ColorFromBool("YouTube Highlight", video.YouTubeHighlight)).Value(&video.YouTubeHighlight),
		huh.NewConfirm().Title(c.ColorFromBool("Pinned comment", video.YouTubeComment)).Value(&video.YouTubeComment),
		huh.NewConfirm().Title(c.ColorFromBool("Replies to comments", video.YouTubeCommentReply)).Value(&video.YouTubeCommentReply),
		huh.NewConfirm().Title(c.ColorFromBool("https://gde.advocu.com post", video.GDE)).Value(&video.GDE),
		huh.NewInput().Title(c.ColorFromString("Code repo", video.Repo)).Value(&video.Repo),
		huh.NewConfirm().Title(sponsorsNotifyText).Value(&video.NotifiedSponsors),
	}
	for index := range fields {
		uploadVideoOrig := video.UploadVideo
		blueSkyPostedOrig := video.BlueSkyPosted
		linkedInPostedOrig := video.LinkedInPosted
		slackPostedOrig := video.SlackPosted
		hnPostedOrig := video.HNPosted
		dotPosted := video.DOTPosted
		repoOrig := video.Repo

		form := huh.NewForm(
			huh.NewGroup(
				fields[index],
				huh.NewConfirm().Affirmative("Save & continue").Negative("Cancel").Value(&save),
			),
		)
		err := form.Run()
		if err != nil {
			return storage.Video{}, err
		}

		video.Publish.Completed, video.Publish.Total = c.Count([]interface{}{
			video.UploadVideo,
			video.HugoPath,
			video.BlueSkyPosted,
			video.LinkedInPosted,
			video.SlackPosted,
			video.HNPosted,
			video.DOTPosted,
			video.YouTubeHighlight,
			video.YouTubeComment,
			video.YouTubeCommentReply,
			video.GDE,
			video.Repo,
		})
		video.Publish.Total++
		if video.NotifiedSponsors || len(video.Sponsorship.Amount) == 0 || video.Sponsorship.Amount == "N/A" || video.Sponsorship.Amount == "-" {
			video.Publish.Completed++
		}

		if createHugo && len(video.HugoPath) == 0 {
			hugo := Hugo{}
			var errHugo error
			video.HugoPath, errHugo = hugo.Post(video.Gist, video.Title, video.Date)
			if errHugo != nil {
				return storage.Video{}, errHugo
			}
		} else if !createHugo {
			video.HugoPath = ""
		}

		if len(uploadVideoOrig) == 0 && len(video.UploadVideo) > 0 {
			video.VideoId = uploadVideo(video) // uploadVideo only returns a string

			if err := uploadThumbnail(video); err != nil { // uploadThumbnail returns an error; call directly
				// Preserve existing behavior of printing to console, but also return the error
				errorMsg := fmt.Sprintf("Failed to upload thumbnail: %v", err)
				println(errorStyle.Render(errorMsg))
				return video, fmt.Errorf(errorMsg) // Return the error
			}

			println(confirmationStyle.Render(`Following should be set manually:
- End screen
- Playlists
- Language
- Monetization`))
		}
		if !linkedInPostedOrig && len(video.Tweet) > 0 && video.LinkedInPosted {
			postLinkedIn(video.Tweet, video.VideoId)
		}
		if !slackPostedOrig && len(video.VideoId) > 0 && video.SlackPosted {
			postSlack(video.VideoId)
		}
		if !hnPostedOrig && len(video.VideoId) > 0 && video.HNPosted {
			postHackerNews(video.Title, video.VideoId)
		}
		if !dotPosted && len(video.VideoId) > 0 && video.DOTPosted {
			postTechnologyConversations(video.Title, video.Description, video.VideoId, video.Gist, video.ProjectName, video.ProjectURL, video.RelatedVideos)
		}
		if !blueSkyPostedOrig && len(video.Tweet) > 0 && video.BlueSkyPosted {
			config := bluesky.Config{
				Identifier: configuration.GlobalSettings.Bluesky.Identifier,
				Password:   configuration.GlobalSettings.Bluesky.Password,
				URL:        configuration.GlobalSettings.Bluesky.URL,
			}
			if errBluesky := bluesky.SendPost(config, video.Tweet, video.VideoId, video.Thumbnail); errBluesky != nil {
				println(errorStyle.Render(fmt.Sprintf("Failed to post to Bluesky: %s", errBluesky.Error())))
			} else {
				println(confirmationStyle.Render("Successfully posted to Bluesky."))
			}
		}

		if len(repoOrig) == 0 && len(video.Repo) > 0 && video.Repo != "N/A" {
			fmt.Printf("Updating repository %s...\n", video.Repo)
			repo := repository.Repo{}
			if errUpdate := repo.Update(video.Repo, video.Title, video.VideoId); errUpdate != nil {
				errorMsg = fmt.Sprintf("Failed to update repository %s: %v", video.Repo, errUpdate)
				fmt.Println(errorStyle.Render(errorMsg))
			} else {
				fmt.Println(confirmationStyle.Render(fmt.Sprintf("Repository %s updated successfully.", video.Repo)))
			}
		}
		if !notifiedSponsorsOrig && video.NotifiedSponsors {
			email := notification.NewEmail(configuration.GlobalSettings.Email.Password)
			email.SendSponsors(configuration.GlobalSettings.Email.From, video.Sponsorship.Emails, video.VideoId, video.Sponsorship.Amount)
		}
		if !save {
			break
		}
		yaml := storage.YAML{}
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

func (c *Choices) ChooseVideosPhase(vi []storage.VideoIndex) bool {
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
	c.ChooseVideos(vi, selection, nil)
	return false
}

func (c *Choices) GetVideoPhase(vi storage.VideoIndex) int {
	yaml := storage.YAML{}
	video := yaml.GetVideo(c.GetFilePath(vi.Category, vi.Name, "yaml"))
	if video.Delayed {
		return videosPhaseDelayed
	} else if len(video.Sponsorship.Blocked) > 0 {
		return videosPhaseSponsoredBlocked
	} else if len(video.Repo) > 0 { // Assuming video.Repo is populated when published
		return videosPhasePublished
	} else if len(video.UploadVideo) > 0 && len(video.Tweet) > 0 { // Assuming these indicate pending publish
		return videosPhasePublishPending
	} else if video.RequestEdit {
		return videosPhaseEditRequested
	} else if video.Code && video.Screen && video.Head && video.Diagrams { // Assuming these are key for material done
		return videosPhaseMaterialDone
	} else if len(video.Date) > 0 { // Date implies started
		return videosPhaseStarted
	} else {
		return videosPhaseIdeas
	}
}

func (c *Choices) ChooseVideos(vi []storage.VideoIndex, phase int, input *bytes.Buffer) {
	var selectedVideo storage.Video
	var selectedAction int
	options := huh.NewOptions[storage.Video]()
	sortedVideos := []storage.Video{}
	for i := range vi {
		videoPhase := c.GetVideoPhase(vi[i])
		if videoPhase == phase {
			yaml := storage.YAML{}
			path := c.GetFilePath(vi[i].Category, vi[i].Name, "yaml")
			video := yaml.GetVideo(path)
			video.Name = vi[i].Name
			video.Path = path
			video.Index = i
			video.Category = vi[i].Category
			sortedVideos = append(sortedVideos, video)
		}
	}
	sort.Slice(sortedVideos, func(i, j int) bool {
		date1, _ := time.Parse("2006-01-02T15:04", sortedVideos[i].Date)
		date2, _ := time.Parse("2006-01-02T15:04", sortedVideos[j].Date)
		return date1.Before(date2)
	})
	for _, video := range sortedVideos {
		titleString := c.getVideoTitleForDisplay(video, phase, time.Now())
		options = append(options, huh.NewOption(titleString, video))
	}
	options = append(options, huh.NewOption("Return", storage.Video{Name: "Return"}))

	currentTheme := getCustomHuhTheme()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[storage.Video]().
				Title("Choose a video").
				Options(options...).
				Value(&selectedVideo),
		),
	).WithTheme(currentTheme)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	if selectedVideo.Name == "Return" {
		return // User chose to return from video selection
	}

	actionOptions := c.getActionOptions()
	actionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title(fmt.Sprintf("Choose action for %s", selectedVideo.Name)).
				Options(actionOptions...).
				Value(&selectedAction),
		),
	).WithTheme(currentTheme)

	err = actionForm.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println(orangeStyle.Render("Action selection cancelled."))
			return
		}
		log.Printf("Error running action form: %v", err)
		return
	}

	switch selectedAction {
	case actionEdit:
		choices := NewChoices() // Create new choices instance for fresh state if needed
		choices.ChoosePhase(selectedVideo)
	case actionDelete:
		var delErr error
		vi, delErr = c.handleDeleteVideoAction(selectedVideo, vi)
		if delErr != nil {
			log.Printf("Error during video deletion process: %v", delErr)
		}
	case actionMoveFiles:
		targetDir, selErr := c.dirSelector.SelectDirectory(input)
		if selErr != nil {
			if errors.Is(selErr, huh.ErrUserAborted) {
				fmt.Println(orangeStyle.Render("Move video action cancelled."))
			} else {
				log.Printf("Error selecting target directory: %v", selErr)
			}
		} else {
			currentYAMLPath := selectedVideo.Path
			ext := filepath.Ext(currentYAMLPath)
			videoBaseFileName := strings.TrimSuffix(filepath.Base(currentYAMLPath), ext)
			currentMDPath := strings.TrimSuffix(currentYAMLPath, ext) + ".md"

			newYAMLPath, _, moveErr := utils.MoveVideoFiles(currentYAMLPath, currentMDPath, targetDir.Path, videoBaseFileName)
			if moveErr != nil {
				log.Printf("Error moving video files for '%s': %v", selectedVideo.Name, moveErr)
			} else {
				newCategory := filepath.Base(targetDir.Path)
				updated := false
				for i, videoIdx := range vi {
					if videoIdx.Name == selectedVideo.Name && videoIdx.Category == selectedVideo.Category {
						vi[i].Category = newCategory
						updated = true
						break
					}
				}

				if updated {
					yamlOps := storage.YAML{IndexPath: "index.yaml"}
					yamlOps.WriteIndex(vi)
					fmt.Println(confirmationStyle.Render(fmt.Sprintf("Video '%s' moved to %s and index updated.", selectedVideo.Name, targetDir.Name)))
				} else {
					log.Printf("Could not find video '%s' in index to update its category after moving. Files moved to %s.", selectedVideo.Name, newYAMLPath)
				}
			}
		}
		return
	case actionReturn:
		return
	}
	yaml := storage.YAML{IndexPath: "index.yaml"}
	yaml.WriteIndex(vi)
}

func (c *Choices) getVideoTitleForDisplay(video storage.Video, currentPhase int, referenceTime time.Time) string {
	title := video.Name
	isSponsored := len(video.Sponsorship.Amount) > 0 && video.Sponsorship.Amount != "-" && video.Sponsorship.Amount != "N/A"
	isBlocked := len(video.Sponsorship.Blocked) > 0 && video.Sponsorship.Blocked != "-" && video.Sponsorship.Blocked != "N/A"

	displayStyle := lipgloss.NewStyle()
	var isFarFuture bool = false

	if video.Date != "" {
		var dateErr error
		isFarFuture, dateErr = utils.IsFarFutureDate(video.Date, "2006-01-02T15:04", referenceTime)
		if dateErr != nil {
			log.Printf("Error checking if date is far future for video '%s': %v", video.Name, dateErr)
		}
	}

	if currentPhase == videosPhaseStarted && isFarFuture {
		displayStyle = farFutureStyle
	} else if isSponsored && !isBlocked {
		displayStyle = orangeStyle
	}

	if isBlocked {
		blockDisplay := video.Sponsorship.Blocked
		if blockDisplay == "" || blockDisplay == "-" || blockDisplay == "N/A" {
			blockDisplay = "B"
		}
		title = fmt.Sprintf("%s (%s)", title, blockDisplay)
	} else {
		if len(video.Date) > 0 {
			title = fmt.Sprintf("%s (%s)", title, video.Date)
		}
		if isSponsored {
			title = fmt.Sprintf("%s (S)", title)
		}
	}

	if video.Category == "ama" {
		title = fmt.Sprintf("%s (AMA)", title)
	}

	return displayStyle.Render(title)
}

func (c *Choices) performVideoFileDeletions(yamlPath, mdPath string) (yamlError, mdError error) {
	if _, statErr := os.Stat(mdPath); statErr == nil {
		if rmErr := os.Remove(mdPath); rmErr != nil {
			mdError = fmt.Errorf("error deleting MD file %s: %w", mdPath, rmErr)
		}
	} else if !os.IsNotExist(statErr) {
		mdError = fmt.Errorf("error checking MD file %s: %w", mdPath, statErr)
	}

	if _, statErr := os.Stat(yamlPath); statErr == nil {
		if rmErr := os.Remove(yamlPath); rmErr != nil {
			yamlError = fmt.Errorf("error deleting YAML file %s: %w", yamlPath, rmErr)
		}
	} else if !os.IsNotExist(statErr) {
		yamlError = fmt.Errorf("error checking YAML file %s: %w", yamlPath, statErr)
	}

	return
}

func (c *Choices) handleDeleteVideoAction(selectedVideo storage.Video, allVideoIndices []storage.VideoIndex) ([]storage.VideoIndex, error) {
	confirmMsg := fmt.Sprintf("Are you sure you want to delete video '%s' and its associated files (.md, .yaml)?", selectedVideo.Name)

	if c.confirmer.Confirm(confirmMsg) {
		mdPath := strings.ReplaceAll(selectedVideo.Path, ".yaml", ".md")

		yamlErr, mdErr := c.performVideoFileDeletions(selectedVideo.Path, mdPath)

		if yamlErr != nil {
			log.Printf(yamlErr.Error())
		}
		if mdErr != nil {
			log.Printf(mdErr.Error())
		}

		if selectedVideo.Index < 0 || selectedVideo.Index >= len(allVideoIndices) {
			return allVideoIndices, fmt.Errorf("selected video index %d is out of bounds for video indices slice (len %d)", selectedVideo.Index, len(allVideoIndices))
		}

		updatedIndices := append(allVideoIndices[:selectedVideo.Index], allVideoIndices[selectedVideo.Index+1:]...)

		fmt.Println(confirmationStyle.Render(fmt.Sprintf("Video '%s' and associated files deleted.", selectedVideo.Name)))
		return updatedIndices, nil
	} else {
		fmt.Println(orangeStyle.Render("Deletion cancelled."))
		return allVideoIndices, nil
	}
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
		} else if phase == videosPhaseMaterialDone && phases[phase] >= 3 {
			return greenStyle.Render(title), phases[phase]
		} else if phase == videosPhaseIdeas && phases[phase] >= 3 {
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
		huh.NewOption("Create Video", indexCreateVideo),
		huh.NewOption("List Videos", indexListVideos),
		huh.NewOption("Exit", actionReturn),
	}
}

func (c *Choices) getActionOptions() []huh.Option[int] {
	return []huh.Option[int]{
		huh.NewOption("Edit", actionEdit),
		huh.NewOption("Delete", actionDelete),
		huh.NewOption("Move Video", actionMoveFiles),
		huh.NewOption("Return", actionReturn),
	}
}

func (c *Choices) getAvailableDirectories() ([]Directory, error) {
	return c.getDirsFunc()
}

func (c *Choices) doGetAvailableDirectories() ([]Directory, error) {
	var availableDirs []Directory
	manuscriptPath := "manuscript"

	files, err := os.ReadDir(manuscriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Directory{}, nil
		}
		return nil, fmt.Errorf("failed to read manuscript directory '%s': %w", manuscriptPath, err)
	}

	caser := cases.Title(language.AmericanEnglish)
	for _, file := range files {
		if file.IsDir() {
			displayName := caser.String(strings.ReplaceAll(file.Name(), "-", " "))
			dirPath := filepath.Join(manuscriptPath, file.Name())
			availableDirs = append(availableDirs, Directory{Name: displayName, Path: dirPath})
		}
	}

	sort.Slice(availableDirs, func(i, j int) bool {
		return availableDirs[i].Name < availableDirs[j].Name
	})

	return availableDirs, nil
}

func (c *Choices) toHuhOptionsDirectory(dirs []Directory) []huh.Option[Directory] {
	options := make([]huh.Option[Directory], len(dirs))
	for i, dir := range dirs {
		options[i] = huh.NewOption(dir.Name, dir)
	}
	return options
}

func (c *Choices) doSelectTargetDirectory(input *bytes.Buffer) (Directory, error) {
	availableDirs, err := c.getAvailableDirectories()
	if err != nil {
		return Directory{}, fmt.Errorf("failed to get available directories: %w", err)
	}

	if len(availableDirs) == 0 {
		return Directory{}, errors.New("no available directories to choose from")
	}

	var selectedDir Directory
	huhOptions := c.toHuhOptionsDirectory(availableDirs)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Directory]().
				Title("Select target directory").
				Options(huhOptions...).
				Value(&selectedDir),
		),
	)
	form = form.WithTheme(nil)
	if input != nil {
		form = form.WithInput(input)
	}

	if err := form.Run(); err != nil {
		return Directory{}, err
	}

	return selectedDir, nil
}

func (c *Choices) SelectDirectory(input *bytes.Buffer) (Directory, error) {
	return c.doSelectTargetDirectory(input)
}
