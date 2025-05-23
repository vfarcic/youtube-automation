package app

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/cli"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/ui"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"
	"devopstoolkit/youtube-automation/pkg/utils"

	"github.com/charmbracelet/huh"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MenuHandler handles the main menu and navigation logic
type MenuHandler struct {
	confirmer    Confirmer
	getDirsFunc  func() ([]Directory, error)
	dirSelector  DirectorySelector
	uiRenderer   *ui.Renderer
	videoManager *video.Manager
	filesystem   *filesystem.Operations
}

// Constants for menu indices
const indexCreateVideo = 0
const indexListVideos = 1

const (
	actionEdit = iota
	actionDelete
	actionMoveFiles
)
const actionReturn = 99

// ChooseIndex displays the main menu and handles user selection
func (m *MenuHandler) ChooseIndex() error {
	var selectedIndex int
	yaml := storage.YAML{IndexPath: "index.yaml"}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("What do you want to do?").
				Options(m.getIndexOptions()...).
				Value(&selectedIndex),
		),
	)
	err := form.Run()
	if err != nil {
		return fmt.Errorf("failed to run main menu form: %w", err)
	}
	switch selectedIndex {
	case indexCreateVideo:
		index := yaml.GetIndex()
		item, err := m.ChooseCreateVideoAndHandleError()
		if err != nil {
			return fmt.Errorf("error in create video choice: %w", err)
		}
		if len(item.Category) > 0 && len(item.Name) > 0 {
			index = append(index, item)
			yaml.WriteIndex(index)
		}
	case indexListVideos:
		for {
			index := yaml.GetIndex()
			returnVal, err := m.ChooseVideosPhaseAndHandleError(index)
			if err != nil {
				return fmt.Errorf("error in list videos phase: %w", err)
			}
			if returnVal {
				break
			}
		}
	case actionReturn:
		return nil
	}
	return nil
}

// GetPhaseText returns formatted text for a phase with completion status
func (m *MenuHandler) GetPhaseText(text string, task storage.Tasks) string {
	text = fmt.Sprintf("%s (%d/%d)", text, task.Completed, task.Total)
	if task.Completed == task.Total && task.Total > 0 {
		return ui.GreenStyle.Render(text)
	}
	return ui.OrangeStyle.Render(text)
}

// ChooseCreateVideo handles video creation workflow
func (m *MenuHandler) ChooseCreateVideoAndHandleError() (storage.VideoIndex, error) {
	var name, category string
	save := true
	fields, err := cli.GetCreateVideoFields(&name, &category, &save)
	if err != nil {
		return storage.VideoIndex{}, fmt.Errorf("error getting video fields: %w", err)
	}
	form := huh.NewForm(huh.NewGroup(fields...))
	err = form.Run()
	if err != nil {
		return storage.VideoIndex{}, fmt.Errorf("form run failed: %w", err)
	}
	vi := storage.VideoIndex{
		Name:     name,
		Category: category,
	}
	if !save {
		return vi, nil
	}
	dirPath := m.filesystem.GetDirPath(vi.Category)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if mkDirErr := os.Mkdir(dirPath, 0755); mkDirErr != nil {
			return storage.VideoIndex{}, fmt.Errorf("failed to create directory %s: %w", dirPath, mkDirErr)
		}
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
	filePath := m.filesystem.GetFilePath(vi.Category, vi.Name, "md")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		f, errCreate := os.Create(filePath)
		if errCreate != nil {
			return storage.VideoIndex{}, fmt.Errorf("failed to create script file %s: %w", filePath, errCreate)
		}
		defer f.Close()
		if _, writeErr := f.Write([]byte(scriptContent)); writeErr != nil {
			return storage.VideoIndex{}, fmt.Errorf("failed to write to script file %s: %w", filePath, writeErr)
		}
		return vi, nil
	}
	return storage.VideoIndex{}, nil
}

// ChooseVideosPhase handles the video phase selection workflow
func (m *MenuHandler) ChooseVideosPhaseAndHandleError(vi []storage.VideoIndex) (bool, error) {
	if len(vi) == 0 {
		fmt.Println(ui.ErrorStyle.Render("No videos found. Create a video first."))
		return true, nil
	}

	phases := map[int]int{
		workflow.PhaseIdeas:            0,
		workflow.PhaseStarted:          0,
		workflow.PhaseMaterialDone:     0,
		workflow.PhaseEditRequested:    0,
		workflow.PhasePublishPending:   0,
		workflow.PhasePublished:        0,
		workflow.PhaseDelayed:          0,
		workflow.PhaseSponsoredBlocked: 0,
	}

	// Count videos in each phase using video manager
	for _, video := range vi {
		currentPhase := m.videoManager.GetVideoPhase(video)
		phases[currentPhase]++
	}

	var selectedPhase int
	var options []huh.Option[int]

	// Add options in the original order, only if there are videos in that phase
	if text, count := m.getPhaseColoredTextWithCount(phases, workflow.PhasePublished, "Published"); count > 0 {
		options = append(options, huh.NewOption(text, workflow.PhasePublished))
	}
	if text, count := m.getPhaseColoredTextWithCount(phases, workflow.PhasePublishPending, "Pending publish"); count > 0 {
		options = append(options, huh.NewOption(text, workflow.PhasePublishPending))
	}
	if text, count := m.getPhaseColoredTextWithCount(phases, workflow.PhaseEditRequested, "Edit requested"); count > 0 {
		options = append(options, huh.NewOption(text, workflow.PhaseEditRequested))
	}
	if text, count := m.getPhaseColoredTextWithCount(phases, workflow.PhaseMaterialDone, "Material done"); count > 0 {
		options = append(options, huh.NewOption(text, workflow.PhaseMaterialDone))
	}
	if text, count := m.getPhaseColoredTextWithCount(phases, workflow.PhaseStarted, "Started"); count > 0 {
		options = append(options, huh.NewOption(text, workflow.PhaseStarted))
	}
	if text, count := m.getPhaseColoredTextWithCount(phases, workflow.PhaseDelayed, "Delayed"); count > 0 {
		options = append(options, huh.NewOption(text, workflow.PhaseDelayed))
	}
	if text, count := m.getPhaseColoredTextWithCount(phases, workflow.PhaseSponsoredBlocked, "Sponsored blocked"); count > 0 {
		options = append(options, huh.NewOption(text, workflow.PhaseSponsoredBlocked))
	}
	if text, count := m.getPhaseColoredTextWithCount(phases, workflow.PhaseIdeas, "Ideas"); count > 0 {
		options = append(options, huh.NewOption(text, workflow.PhaseIdeas))
	}

	options = append(options, huh.NewOption("Return", actionReturn))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("From which phase would you like to list the videos?").
				Options(options...).
				Value(&selectedPhase),
		),
	)
	err := form.Run()
	if err != nil {
		return false, fmt.Errorf("failed to run video phase form: %w", err)
	}

	if selectedPhase == actionReturn {
		return true, nil
	}

	if err := m.ChooseVideosAndHandleError(vi, selectedPhase, nil); err != nil {
		return false, fmt.Errorf("error in choose videos: %w", err)
	}
	return false, nil
}

// ChooseVideos handles video selection and actions for a specific phase
func (m *MenuHandler) ChooseVideosAndHandleError(vi []storage.VideoIndex, phase int, input *bytes.Buffer) error {
	yaml := storage.YAML{}
	var videosInPhase []storage.Video

	// Filter videos by phase using video manager
	for i, video := range vi {
		currentPhase := m.videoManager.GetVideoPhase(video)
		if currentPhase == phase {
			videoPath := m.filesystem.GetFilePath(video.Category, video.Name, "yaml")
			fullVideo := yaml.GetVideo(videoPath)
			fullVideo.Index = i // Store the index for potential deletion
			fullVideo.Name = video.Name
			fullVideo.Category = video.Category
			fullVideo.Path = videoPath
			videosInPhase = append(videosInPhase, fullVideo)
		}
	}

	if len(videosInPhase) == 0 {
		fmt.Println(ui.ErrorStyle.Render("No videos found in this phase."))
		return nil
	}

	// Sort videos by date
	sort.Slice(videosInPhase, func(i, j int) bool {
		date1, _ := time.Parse("2006-01-02T15:04", videosInPhase[i].Date)
		date2, _ := time.Parse("2006-01-02T15:04", videosInPhase[j].Date)
		return date1.Before(date2)
	})

	// Create video selection options
	var videoOptions []huh.Option[storage.Video]
	for _, video := range videosInPhase {
		displayTitle := m.getVideoTitleForDisplay(video, phase, time.Now())
		videoOptions = append(videoOptions, huh.NewOption(displayTitle, video))
	}
	videoOptions = append(videoOptions, huh.NewOption("Return", storage.Video{Name: "return"}))

	var selectedVideo storage.Video
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[storage.Video]().
				Title("Select a video:").
				Options(videoOptions...).
				Value(&selectedVideo),
		),
	)

	if input != nil {
		form = form.WithInput(input)
	}

	err := form.Run()
	if err != nil {
		return fmt.Errorf("failed to run video selection form: %w", err)
	}

	if selectedVideo.Name == "return" {
		return nil
	}

	// Now show action options for the selected video
	actionOptions := cli.GetActionOptions()
	var selectedAction int

	actionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title(fmt.Sprintf("What would you like to do with '%s'?", selectedVideo.Name)).
				Options(actionOptions...).
				Value(&selectedAction),
		),
	)

	err = actionForm.Run()
	if err != nil {
		return fmt.Errorf("failed to run action form: %w", err)
	}

	switch selectedAction {
	case actionEdit:
		// TODO: Implement edit functionality - would need to migrate phase-specific logic
		fmt.Println(ui.OrangeStyle.Render("Edit functionality not yet implemented in reorganized structure"))
	case actionDelete:
		deleted, errDel := m.handleDeleteVideoActionAndHandleError(selectedVideo, vi)
		if errDel != nil {
			return fmt.Errorf("error deleting video: %w", errDel)
		}
		if deleted {
			fmt.Println(ui.ConfirmationStyle.Render(fmt.Sprintf("Video '%s' deleted successfully.", selectedVideo.Name)))
		}
	case actionMoveFiles:
		// TODO: Implement move functionality
		fmt.Println(ui.OrangeStyle.Render("Move functionality not yet implemented in reorganized structure"))
	case actionReturn:
		return nil
	}
	return nil
}

// handleDeleteVideoAction handles video deletion workflow
func (m *MenuHandler) handleDeleteVideoActionAndHandleError(selectedVideo storage.Video, allVideoIndices []storage.VideoIndex) (bool, error) {
	confirmMsg := fmt.Sprintf("Are you sure you want to delete video '%s' and its associated files (.md, .yaml)?", selectedVideo.Name)

	confirmed := utils.ConfirmAction(confirmMsg)
	if confirmed {
		mdPath := strings.ReplaceAll(selectedVideo.Path, ".yaml", ".md")

		// Delete both files
		yamlErr := os.Remove(selectedVideo.Path)
		mdErr := os.Remove(mdPath)

		var deletionErrors []string
		if yamlErr != nil {
			deletionErrors = append(deletionErrors, fmt.Sprintf("YAML file (%s): %v", selectedVideo.Path, yamlErr))
		}
		if mdErr != nil {
			deletionErrors = append(deletionErrors, fmt.Sprintf("MD file (%s): %v", mdPath, mdErr))
		}

		if len(deletionErrors) > 0 {
			return false, fmt.Errorf("errors during file deletion: %s", strings.Join(deletionErrors, "; "))
		}

		// Remove from index
		if selectedVideo.Index >= 0 && selectedVideo.Index < len(allVideoIndices) {
			updatedIndices := append(allVideoIndices[:selectedVideo.Index], allVideoIndices[selectedVideo.Index+1:]...)
			yaml := storage.YAML{IndexPath: "index.yaml"}
			if errWrite := yaml.WriteIndex(updatedIndices); errWrite != nil {
				return false, fmt.Errorf("failed to write updated index: %w", errWrite)
			}
		}

		return true, nil
	}

	fmt.Println(ui.OrangeStyle.Render("Deletion cancelled."))
	return false, nil
}

// getPhaseColoredText returns colored text based on phase status
func (m *MenuHandler) getPhaseColoredText(phases map[int]int, phase int, title string) string {
	if phase != actionReturn {
		title = fmt.Sprintf("%s (%d)", title, phases[phase])
		if phase == workflow.PhasePublished {
			return ui.GreenStyle.Render(title)
		} else if phase == workflow.PhasePublishPending && phases[phase] > 0 {
			return ui.GreenStyle.Render(title)
		} else if phase == workflow.PhaseEditRequested && phases[phase] > 0 {
			return ui.GreenStyle.Render(title)
		} else if phase == workflow.PhaseMaterialDone && phases[phase] >= 3 {
			return ui.GreenStyle.Render(title)
		} else if phase == workflow.PhaseIdeas && phases[phase] >= 3 {
			return ui.GreenStyle.Render(title)
		} else if phase == workflow.PhaseStarted && phases[phase] >= 3 {
			return ui.GreenStyle.Render(title)
		} else {
			return ui.OrangeStyle.Render(title)
		}
	}
	return title
}

// getPhaseColoredTextWithCount returns colored text and count for a phase
func (m *MenuHandler) getPhaseColoredTextWithCount(phases map[int]int, phase int, title string) (string, int) {
	count := phases[phase]
	if count > 0 {
		title = fmt.Sprintf("%s (%d)", title, count)
		if phase == workflow.PhasePublished {
			return ui.GreenStyle.Render(title), count
		} else if phase == workflow.PhasePublishPending && count > 0 {
			return ui.GreenStyle.Render(title), count
		} else if phase == workflow.PhaseEditRequested && count > 0 {
			return ui.GreenStyle.Render(title), count
		} else if phase == workflow.PhaseMaterialDone && count >= 3 {
			return ui.GreenStyle.Render(title), count
		} else if phase == workflow.PhaseIdeas && count >= 3 {
			return ui.GreenStyle.Render(title), count
		} else if phase == workflow.PhaseStarted && count >= 3 {
			return ui.GreenStyle.Render(title), count
		} else {
			return ui.OrangeStyle.Render(title), count
		}
	}
	return title, count
}

// doGetAvailableDirectories implements directory listing functionality
func (m *MenuHandler) doGetAvailableDirectories() ([]Directory, error) {
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

// SelectDirectory implements DirectorySelector interface
func (m *MenuHandler) SelectDirectory(input *bytes.Buffer) (Directory, error) {
	availableDirs, err := m.getDirsFunc()
	if err != nil {
		return Directory{}, fmt.Errorf("failed to get available directories: %w", err)
	}

	if len(availableDirs) == 0 {
		return Directory{}, errors.New("no available directories to choose from")
	}

	var selectedDir Directory
	huhOptions := m.toHuhOptionsDirectory(availableDirs)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Directory]().
				Title("Select target directory").
				Options(huhOptions...).
				Value(&selectedDir),
		),
	)
	form = form.WithTheme(cli.GetCustomHuhTheme())
	if input != nil {
		form = form.WithInput(input)
	}

	if err := form.Run(); err != nil {
		return Directory{}, err
	}

	return selectedDir, nil
}

// toHuhOptionsDirectory converts directories to huh options
func (m *MenuHandler) toHuhOptionsDirectory(dirs []Directory) []huh.Option[Directory] {
	options := make([]huh.Option[Directory], len(dirs))
	for i, dir := range dirs {
		options[i] = huh.NewOption(dir.Name, dir)
	}
	return options
}

// getIndexOptions returns the main menu options
func (m *MenuHandler) getIndexOptions() []huh.Option[int] {
	return []huh.Option[int]{
		huh.NewOption("Create Video", indexCreateVideo),
		huh.NewOption("List Videos", indexListVideos),
		huh.NewOption("Exit", actionReturn),
	}
}

// getVideoTitleForDisplay returns a formatted video title for display
func (m *MenuHandler) getVideoTitleForDisplay(video storage.Video, currentPhase int, referenceTime time.Time) string {
	title := video.Name
	isSponsored := len(video.Sponsorship.Amount) > 0 && video.Sponsorship.Amount != "-" && video.Sponsorship.Amount != "N/A"
	isBlocked := len(video.Sponsorship.Blocked) > 0 && video.Sponsorship.Blocked != "-" && video.Sponsorship.Blocked != "N/A"

	// Default style (no special styling)
	styledTitle := title
	var isFarFuture bool = false

	if video.Date != "" {
		var dateErr error
		isFarFuture, dateErr = utils.IsFarFutureDate(video.Date, "2006-01-02T15:04", referenceTime)
		if dateErr != nil {
			log.Printf("Error checking if date is far future for video '%s': %v", video.Name, dateErr)
		}
	}

	// Apply styling based on phase and conditions
	if currentPhase == workflow.PhaseStarted && isFarFuture {
		// Use cyan style for far future videos in Started phase
		styledTitle = ui.FarFutureStyle.Render(title)
	} else if isSponsored && !isBlocked {
		// Use orange style for sponsored but not blocked videos
		styledTitle = ui.OrangeStyle.Render(title)
	} else {
		// Default styling (no special color)
		styledTitle = title
	}

	// Add bracket information based on status
	if isBlocked {
		blockDisplay := video.Sponsorship.Blocked
		if blockDisplay == "" || blockDisplay == "-" || blockDisplay == "N/A" {
			blockDisplay = "B"
		}
		styledTitle = fmt.Sprintf("%s (%s)", styledTitle, blockDisplay)
	} else {
		if len(video.Date) > 0 {
			styledTitle = fmt.Sprintf("%s (%s)", styledTitle, video.Date)
		}
		if isSponsored {
			styledTitle = fmt.Sprintf("%s (S)", styledTitle)
		}
	}

	// Add special category markers
	if video.Category == "ama" {
		styledTitle = fmt.Sprintf("%s (AMA)", styledTitle)
	}

	return styledTitle
}
