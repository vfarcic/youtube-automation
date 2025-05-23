package app

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/cli"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/notification"
	"devopstoolkit/youtube-automation/internal/platform"
	"devopstoolkit/youtube-automation/internal/platform/bluesky"
	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/ui"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"
	"devopstoolkit/youtube-automation/pkg/utils"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ErrExitApplication is a sentinel error to signal application termination.
var ErrExitApplication = errors.New("user requested application exit")

// MenuHandler handles the main menu and navigation logic
type MenuHandler struct {
	confirmer         Confirmer
	dirSelector       DirectorySelector
	uiRenderer        *ui.Renderer
	videoManager      *video.Manager
	filesystem        *filesystem.Operations
	greenStyle        lipgloss.Style
	orangeStyle       lipgloss.Style
	redStyle          lipgloss.Style
	farFutureStyle    lipgloss.Style
	confirmationStyle lipgloss.Style
	errorStyle        lipgloss.Style
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

// Constants for edit phases
const (
	editPhaseInitial = iota
	editPhaseWork
	editPhaseDefinition
	editPhasePostProduction
	editPhasePublishing
	// actionReturn can be reused for returning from this sub-menu
)

// Helper function to count completed tasks based on old logic
func (m *MenuHandler) countCompletedTasks(fields []interface{}) (completed int, total int) {
	for _, field := range fields {
		valueType := reflect.TypeOf(field)
		if valueType == nil { // Handle cases where a field might be nil unexpectedly
			total++
			continue
		}
		switch valueType.Kind() {
		case reflect.String:
			if len(field.(string)) > 0 && field.(string) != "N/A" && field.(string) != "-" {
				completed++
			}
		case reflect.Bool:
			if field.(bool) {
				completed++
			}
		case reflect.Slice:
			// Assuming non-empty slice means task related to it is done
			if reflect.ValueOf(field).Len() > 0 {
				completed++
			}
		}
		total++
	}
	return completed, total
}

// Helper for form titles based on string value
func (m *MenuHandler) colorTitleString(title, value string) string {
	if len(value) > 0 && value != "N/A" && value != "-" {
		return m.greenStyle.Render(title)
	}
	return m.orangeStyle.Render(title)
}

// Helper for form titles based on boolean value
func (m *MenuHandler) colorTitleBool(title string, value bool) string {
	if value {
		return m.greenStyle.Render(title)
	}
	return m.orangeStyle.Render(title)
}

// Helper for form titles for Sponsorship Amount (green if any text is present)
func (m *MenuHandler) colorTitleSponsorshipAmount(title, value string) string {
	if len(value) > 0 {
		return m.greenStyle.Render(title)
	}
	return m.orangeStyle.Render(title)
}

// Helper for form titles for sponsored emails
func (m *MenuHandler) colorTitleSponsoredEmails(title, sponsoredAmount, sponsoredEmails string) string {
	if len(sponsoredAmount) == 0 || sponsoredAmount == "N/A" || sponsoredAmount == "-" || len(sponsoredEmails) > 0 {
		return m.greenStyle.Render(title)
	}
	return m.orangeStyle.Render(title) // Was RedStyle, now consistency with orangeStyle
}

// Helper for form titles based on string value (inverse logic: green if empty)
func (m *MenuHandler) colorTitleStringInverse(title, value string) string {
	if len(value) > 0 {
		return m.orangeStyle.Render(title)
	}
	return m.greenStyle.Render(title)
}

// Helper for form titles based on boolean value (inverse logic: green if false)
func (m *MenuHandler) colorTitleBoolInverse(title string, value bool) string {
	if !value {
		return m.greenStyle.Render(title)
	}
	return m.orangeStyle.Render(title)
}

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
		return ErrExitApplication
	}
	return nil
}

// GetPhaseText returns formatted text for a phase with completion status
func (m *MenuHandler) GetPhaseText(text string, task storage.Tasks) string {
	text = fmt.Sprintf("%s (%d/%d)", text, task.Completed, task.Total)
	if task.Completed == task.Total && task.Total > 0 {
		return m.greenStyle.Render(text)
	}
	return m.orangeStyle.Render(text)
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
		fmt.Println(m.errorStyle.Render("No videos found. Create a video first."))
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
		fmt.Println(m.errorStyle.Render("No videos found in this phase."))
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
		// Call the new phase selection handler
		if err := m.handleEditVideoPhases(selectedVideo); err != nil {
			// Log or display error, then return to allow ChooseVideosAndHandleError to go back to phase list
			log.Printf(m.errorStyle.Render(fmt.Sprintf("Error during video edit phases: %v", err)))
		}
		// After returning from handleEditVideoPhases, the current switch case ends,
		// and ChooseVideosAndHandleError will return nil, which causes ChooseVideosPhaseAndHandleError
		// to loop again, effectively showing the list of videos in the current phase.

	case actionDelete:
		deleted, errDel := m.handleDeleteVideoActionAndHandleError(selectedVideo, vi)
		if errDel != nil {
			return fmt.Errorf("error deleting video: %w", errDel)
		}
		if deleted {
			fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' deleted successfully.", selectedVideo.Name)))
		}
	case actionMoveFiles:
		targetDir, selErr := m.dirSelector.SelectDirectory(input)
		if selErr != nil {
			if errors.Is(selErr, huh.ErrUserAborted) {
				fmt.Println(m.orangeStyle.Render("Move video action cancelled."))
			} else {
				log.Printf(m.errorStyle.Render("Error selecting target directory: %v"), selErr)
			}
			return nil
		}

		currentYAMLPath := selectedVideo.Path
		ext := filepath.Ext(currentYAMLPath)
		videoBaseFileName := strings.TrimSuffix(filepath.Base(currentYAMLPath), ext)
		currentMDPath := strings.TrimSuffix(currentYAMLPath, ext) + ".md"

		newYAMLPath, _, moveErr := utils.MoveVideoFiles(currentYAMLPath, currentMDPath, targetDir.Path, videoBaseFileName)
		if moveErr != nil {
			log.Printf(m.errorStyle.Render(fmt.Sprintf("Error moving video files for '%s': %v", selectedVideo.Name, moveErr)))
		} else {
			fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' files moved to %s", selectedVideo.Name, targetDir.Path)))

			currentVideoIndex := -1
			for i, vid := range vi {
				if vid.Name == selectedVideo.Name && vid.Category == selectedVideo.Category {
					currentVideoIndex = i
					break
				}
			}

			if currentVideoIndex != -1 {
				vi[currentVideoIndex].Category = filepath.Base(targetDir.Path)
				yamlStorage := storage.YAML{IndexPath: "index.yaml"}
				if errWrite := yamlStorage.WriteIndex(vi); errWrite != nil {
					log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to update index.yaml after moving video: %v", errWrite)))
				} else {
					fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' category updated in index.yaml to '%s'.", selectedVideo.Name, vi[currentVideoIndex].Category)))
					selectedVideo.Category = vi[currentVideoIndex].Category
					selectedVideo.Path = newYAMLPath
				}
			} else {
				log.Printf(m.orangeStyle.Render(fmt.Sprintf("Could not find video '%s' in the current list to update its category after moving.", selectedVideo.Name)))
			}
		}
	case actionReturn:
		return nil
	}
	return nil
}

// handleDeleteVideoAction handles video deletion workflow
func (m *MenuHandler) handleDeleteVideoActionAndHandleError(selectedVideo storage.Video, allVideoIndices []storage.VideoIndex) (bool, error) {
	confirmMsg := fmt.Sprintf("Are you sure you want to delete video '%s' and its associated files (.md, .yaml)?", selectedVideo.Name)

	confirmed := utils.ConfirmAction(confirmMsg, nil)
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

	fmt.Println(m.orangeStyle.Render("Deletion cancelled."))
	return false, nil
}

// getPhaseColoredText returns colored text based on phase status
func (m *MenuHandler) getPhaseColoredText(phases map[int]int, phase int, title string) string {
	if phase != actionReturn {
		title = fmt.Sprintf("%s (%d)", title, phases[phase])
		if phase == workflow.PhasePublished {
			return m.greenStyle.Render(title)
		} else if phase == workflow.PhasePublishPending && phases[phase] > 0 {
			return m.greenStyle.Render(title)
		} else if phase == workflow.PhaseEditRequested && phases[phase] > 0 {
			return m.greenStyle.Render(title)
		} else if phase == workflow.PhaseMaterialDone && phases[phase] >= 3 {
			return m.greenStyle.Render(title)
		} else if phase == workflow.PhaseIdeas && phases[phase] >= 3 {
			return m.greenStyle.Render(title)
		} else if phase == workflow.PhaseStarted && phases[phase] >= 3 {
			return m.greenStyle.Render(title)
		} else {
			return m.orangeStyle.Render(title)
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
			return m.greenStyle.Render(title), count
		} else if phase == workflow.PhasePublishPending && count > 0 {
			return m.greenStyle.Render(title), count
		} else if phase == workflow.PhaseEditRequested && count > 0 {
			return m.greenStyle.Render(title), count
		} else if phase == workflow.PhaseMaterialDone && count >= 3 {
			return m.greenStyle.Render(title), count
		} else if phase == workflow.PhaseIdeas && count >= 3 {
			return m.greenStyle.Render(title), count
		} else if phase == workflow.PhaseStarted && count >= 3 {
			return m.greenStyle.Render(title), count
		} else {
			return m.orangeStyle.Render(title), count
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
	availableDirs, err := m.doGetAvailableDirectories()
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
		styledTitle = m.farFutureStyle.Render(title)
	} else if isSponsored && !isBlocked {
		// Use orange style for sponsored but not blocked videos
		styledTitle = m.orangeStyle.Render(title)
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

// getEditPhaseOptionText returns a colored string for an edit phase menu option.
func (m *MenuHandler) getEditPhaseOptionText(phaseName string, completed, total int) string {
	text := fmt.Sprintf("%s (%d/%d)", phaseName, completed, total)
	if total > 0 && completed == total {
		return m.greenStyle.Render(text)
	}
	return m.orangeStyle.Render(text)
}

// handleEditVideoPhases presents a menu to choose which aspect of a video to edit.
func (m *MenuHandler) handleEditVideoPhases(videoToEdit storage.Video) error {
	for {
		var selectedEditPhase int
		editPhaseOptions := []huh.Option[int]{
			huh.NewOption(m.getEditPhaseOptionText("Initial Details", videoToEdit.Init.Completed, videoToEdit.Init.Total), editPhaseInitial),
			huh.NewOption(m.getEditPhaseOptionText("Work Progress", videoToEdit.Work.Completed, videoToEdit.Work.Total), editPhaseWork),
			huh.NewOption(m.getEditPhaseOptionText("Definition", videoToEdit.Define.Completed, videoToEdit.Define.Total), editPhaseDefinition),
			huh.NewOption(m.getEditPhaseOptionText("Post-Production", videoToEdit.Edit.Completed, videoToEdit.Edit.Total), editPhasePostProduction),
			huh.NewOption(m.getEditPhaseOptionText("Publishing Details", videoToEdit.Publish.Completed, videoToEdit.Publish.Total), editPhasePublishing),
			huh.NewOption("Return", actionReturn),
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title(fmt.Sprintf("Editing video: %s. Select aspect to edit:", videoToEdit.Name)).
					Options(editPhaseOptions...).
					Value(&selectedEditPhase),
			),
		)

		runErr := form.Run()
		if runErr != nil {
			if errors.Is(runErr, huh.ErrUserAborted) {
				fmt.Println(m.orangeStyle.Render("Edit phase selection cancelled."))
				return nil // Return to previous menu (video list)
			}
			return fmt.Errorf("failed to run edit phase form: %w", runErr)
		}

		var err error
		updatedVideo := videoToEdit // Work with a copy that can be updated by phase handlers

		switch selectedEditPhase {
		case editPhaseInitial:
			save := true
			// Auto-populate Gist path if empty, similar to old logic
			if len(updatedVideo.Gist) == 0 && updatedVideo.Path != "" {
				updatedVideo.Gist = strings.Replace(updatedVideo.Path, ".yaml", ".md", 1)
			}

			initialFormFields := []huh.Field{
				huh.NewInput().Title(m.colorTitleString("Project Name", updatedVideo.ProjectName)).Value(&updatedVideo.ProjectName),
				huh.NewInput().Title(m.colorTitleString("Project URL", updatedVideo.ProjectURL)).Value(&updatedVideo.ProjectURL),
				huh.NewInput().Title(m.colorTitleSponsorshipAmount("Sponsorship Amount", updatedVideo.Sponsorship.Amount)).Value(&updatedVideo.Sponsorship.Amount),
				huh.NewInput().Title(m.colorTitleSponsoredEmails("Sponsorship Emails (comma separated)", updatedVideo.Sponsorship.Amount, updatedVideo.Sponsorship.Emails)).Value(&updatedVideo.Sponsorship.Emails),
				huh.NewInput().Title(m.colorTitleStringInverse("Sponsorship Blocked Reason", updatedVideo.Sponsorship.Blocked)).Value(&updatedVideo.Sponsorship.Blocked),
				huh.NewInput().Title(m.colorTitleString("Publish Date (YYYY-MM-DDTHH:MM)", updatedVideo.Date)).Value(&updatedVideo.Date),
				huh.NewConfirm().Title(m.colorTitleBoolInverse("Delayed", updatedVideo.Delayed)).Value(&updatedVideo.Delayed), // True means NOT delayed, so inverse logic for green
				huh.NewInput().Title(m.colorTitleString("Gist Path (.md file)", updatedVideo.Gist)).Value(&updatedVideo.Gist),
				huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
			}

			phaseInitialForm := huh.NewForm(huh.NewGroup(initialFormFields...))
			err = phaseInitialForm.Run()

			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render("Initial details edit cancelled."))
					continue // Continue the loop to re-select edit phase
				}
				return fmt.Errorf("failed to run initial details edit form: %w", err)
			}

			if save {
				yaml := storage.YAML{}

				// Revised completion calculation for Init phase
				var completedCount, totalCount int

				// General fields whose standard check is okay via m.countCompletedTasks:
				generalFields := []interface{}{
					updatedVideo.ProjectName,
					updatedVideo.ProjectURL,
					updatedVideo.Gist,
					updatedVideo.Date,
				}
				c, t := m.countCompletedTasks(generalFields)
				completedCount += c
				totalCount += t

				// Specifically handle Sponsorship.Amount itself (1 task)
				totalCount++                                  // Sponsorship.Amount is its own task
				if len(updatedVideo.Sponsorship.Amount) > 0 { // Old logic: done if not empty
					completedCount++
				}

				// Add the 3 special condition tasks to total
				totalCount += 3
				// Condition 1: Sponsorship Emails related (done if no amount, or N/A, or emails exist)
				if len(updatedVideo.Sponsorship.Amount) == 0 || updatedVideo.Sponsorship.Amount == "N/A" || updatedVideo.Sponsorship.Amount == "-" || len(updatedVideo.Sponsorship.Emails) > 0 {
					completedCount++
				}
				// Condition 2: Sponsorship Blocked (done if not blocked)
				if len(updatedVideo.Sponsorship.Blocked) == 0 {
					completedCount++
				}
				// Condition 3: Delayed (done if not delayed)
				if !updatedVideo.Delayed {
					completedCount++
				}

				updatedVideo.Init.Completed = completedCount
				updatedVideo.Init.Total = totalCount

				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save initial details: %w", err)
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' initial details updated.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for initial details."))
			}

		case editPhaseWork:
			save := true
			workFormFields := []huh.Field{
				huh.NewConfirm().Title(m.colorTitleBool("Code Done", updatedVideo.Code)).Value(&updatedVideo.Code),
				huh.NewConfirm().Title(m.colorTitleBool("Talking Head Done", updatedVideo.Head)).Value(&updatedVideo.Head),
				huh.NewConfirm().Title(m.colorTitleBool("Screen Recording Done", updatedVideo.Screen)).Value(&updatedVideo.Screen),
				huh.NewText().Lines(3).CharLimit(1000).Title(m.colorTitleString("Related Videos (comma separated)", updatedVideo.RelatedVideos)).Value(&updatedVideo.RelatedVideos),
				huh.NewConfirm().Title(m.colorTitleBool("Thumbnails Done", updatedVideo.Thumbnails)).Value(&updatedVideo.Thumbnails),
				huh.NewConfirm().Title(m.colorTitleBool("Diagrams Done", updatedVideo.Diagrams)).Value(&updatedVideo.Diagrams),
				huh.NewConfirm().Title(m.colorTitleBool("Screenshots Done", updatedVideo.Screenshots)).Value(&updatedVideo.Screenshots),
				huh.NewInput().Title(m.colorTitleString("Files Location (e.g., Google Drive link)", updatedVideo.Location)).Value(&updatedVideo.Location),
				huh.NewInput().Title(m.colorTitleString("Tagline", updatedVideo.Tagline)).Value(&updatedVideo.Tagline),
				huh.NewInput().Title(m.colorTitleString("Tagline Ideas", updatedVideo.TaglineIdeas)).Value(&updatedVideo.TaglineIdeas),
				huh.NewInput().Title(m.colorTitleString("Other Logos/Assets", updatedVideo.OtherLogos)).Value(&updatedVideo.OtherLogos),
				huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
			}

			phaseWorkForm := huh.NewForm(huh.NewGroup(workFormFields...))
			err = phaseWorkForm.Run()

			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render("Work progress edit cancelled."))
					continue // Continue the loop to re-select edit phase
				}
				return fmt.Errorf("failed to run work progress edit form: %w", err)
			}

			if save {
				yaml := storage.YAML{}
				updatedVideo.Work.Completed, updatedVideo.Work.Total = m.countCompletedTasks([]interface{}{
					updatedVideo.Code,
					updatedVideo.Head,
					updatedVideo.Screen,
					updatedVideo.RelatedVideos,
					updatedVideo.Thumbnails,
					updatedVideo.Diagrams,
					updatedVideo.Screenshots,
					updatedVideo.Location,
					updatedVideo.Tagline,
					updatedVideo.TaglineIdeas,
					updatedVideo.OtherLogos,
				})
				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save work progress: %w", err)
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' work progress updated.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for work progress."))
			}

		case editPhaseDefinition:
			save := true
			originalRequestThumbnailStatus := updatedVideo.RequestThumbnail

			definitionFormFields := []huh.Field{
				huh.NewText().Lines(1).CharLimit(200).Title(m.colorTitleString("Title", updatedVideo.Title)).Value(&updatedVideo.Title),
				huh.NewText().Lines(5).CharLimit(5000).Title(m.colorTitleString("Description", updatedVideo.Description)).Value(&updatedVideo.Description),
				huh.NewText().Lines(1).CharLimit(200).Title(m.colorTitleString("Highlight (Short Teaser)", updatedVideo.Highlight)).Value(&updatedVideo.Highlight),
				huh.NewText().Lines(2).CharLimit(500).Title(m.colorTitleString("Tags (comma separated)", updatedVideo.Tags)).Value(&updatedVideo.Tags),
				huh.NewText().Lines(2).CharLimit(500).Title(m.colorTitleString("Description Tags (comma separated)", updatedVideo.DescriptionTags)).Value(&updatedVideo.DescriptionTags),
				huh.NewText().Lines(3).CharLimit(280).Title(m.colorTitleString("Tweet Text", updatedVideo.Tweet)).Value(&updatedVideo.Tweet),
				// TODO: Add back interactive animation generation if needed. For now, direct text edit.
				huh.NewText().Lines(10).CharLimit(10000).Title(m.colorTitleString("Animations Script", updatedVideo.Animations)).Value(&updatedVideo.Animations),
				huh.NewConfirm().Title(m.colorTitleBool("Request Thumbnail Generation", updatedVideo.RequestThumbnail)).Value(&updatedVideo.RequestThumbnail),
				huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
			}

			phaseDefinitionForm := huh.NewForm(huh.NewGroup(definitionFormFields...))
			err = phaseDefinitionForm.Run()

			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render("Definition edit cancelled."))
					continue // Continue the loop to re-select edit phase
				}
				return fmt.Errorf("failed to run definition edit form: %w", err)
			}

			if save {
				yaml := storage.YAML{}
				updatedVideo.Define.Completed, updatedVideo.Define.Total = m.countCompletedTasks([]interface{}{
					updatedVideo.Title,
					updatedVideo.Description,
					updatedVideo.Highlight,
					updatedVideo.Tags,
					updatedVideo.DescriptionTags,
					updatedVideo.Tweet,
					updatedVideo.Animations,
					updatedVideo.RequestThumbnail,
					// Gist is part of definition completeness according to old logic
					updatedVideo.Gist,
				})
				// Note: Old logic for Define.Total was more complex due to fabric interactions.
				// Here, it's based on the direct fields in this simplified form.

				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save definition details: %w", err)
				}

				if !originalRequestThumbnailStatus && updatedVideo.RequestThumbnail {
					if configuration.GlobalSettings.Email.Password == "" {
						log.Println(m.errorStyle.Render("Email password not configured. Cannot send thumbnail request email."))
					} else {
						emailService := notification.NewEmail(configuration.GlobalSettings.Email.Password)
						if emailErr := emailService.SendThumbnail(configuration.GlobalSettings.Email.From, configuration.GlobalSettings.Email.ThumbnailTo, updatedVideo); emailErr != nil {
							log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to send thumbnail request email: %v", emailErr)))
						} else {
							fmt.Println(m.confirmationStyle.Render("Thumbnail request email sent."))
						}
					}
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' definition details updated.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for definition."))
			}

		case editPhasePostProduction: // This is where the existing ChooseEdit logic goes
			save := true
			originalRequestEditStatus := updatedVideo.RequestEdit

			timeCodesTitle := "Timecodes"
			if strings.Contains(updatedVideo.Timecodes, "FIXME:") {
				timeCodesTitle = m.orangeStyle.Render(timeCodesTitle)
			} else {
				timeCodesTitle = m.greenStyle.Render(timeCodesTitle)
			}

			editFormFields := []huh.Field{
				huh.NewInput().Title(m.colorTitleString("Thumbnail Path", updatedVideo.Thumbnail)).Value(&updatedVideo.Thumbnail),
				huh.NewInput().Title(m.colorTitleString("Members (comma separated)", updatedVideo.Members)).Value(&updatedVideo.Members),
				huh.NewConfirm().Title(m.colorTitleBool("Edit Request", updatedVideo.RequestEdit)).Value(&updatedVideo.RequestEdit),
				huh.NewText().Lines(5).CharLimit(10000).Title(timeCodesTitle).Value(&updatedVideo.Timecodes),
				huh.NewConfirm().Title(m.colorTitleBool("Movie Done", updatedVideo.Movie)).Value(&updatedVideo.Movie),
				huh.NewConfirm().Title(m.colorTitleBool("Slides Done", updatedVideo.Slides)).Value(&updatedVideo.Slides),
				huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
			}

			phaseEditForm := huh.NewForm(huh.NewGroup(editFormFields...))
			err = phaseEditForm.Run()
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render("Post-production edit cancelled."))
					// Continue the loop to re-select edit phase
					continue
				}
				return fmt.Errorf("failed to run post-production edit form: %w", err)
			}

			if save {
				yaml := storage.YAML{}
				updatedVideo.Edit.Completed, updatedVideo.Edit.Total = m.countCompletedTasks([]interface{}{
					updatedVideo.Thumbnail,
					updatedVideo.Members,
					updatedVideo.RequestEdit,
					updatedVideo.Movie,
					updatedVideo.Slides,
				})
				updatedVideo.Edit.Total++ // For Timecodes
				if !strings.Contains(updatedVideo.Timecodes, "FIXME:") {
					updatedVideo.Edit.Completed++
				}
				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save post-production details: %w", err)
				}

				if !originalRequestEditStatus && updatedVideo.RequestEdit {
					if configuration.GlobalSettings.Email.Password == "" {
						log.Println(m.errorStyle.Render("Email password not configured. Cannot send edit request email."))
					} else {
						emailService := notification.NewEmail(configuration.GlobalSettings.Email.Password)
						if emailErr := emailService.SendEdit(configuration.GlobalSettings.Email.From, configuration.GlobalSettings.Email.EditTo, updatedVideo); emailErr != nil {
							log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to send edit request email: %v", emailErr)))
						} else {
							fmt.Println(m.confirmationStyle.Render("Edit request email sent."))
						}
					}
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' post-production details updated.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes to the original reference for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for post-production."))
			}

		case editPhasePublishing:
			save := true
			// Store original values to detect changes for actions
			originalUploadVideo := updatedVideo.UploadVideo
			originalBlueSkyPosted := updatedVideo.BlueSkyPosted
			originalLinkedInPosted := updatedVideo.LinkedInPosted
			originalSlackPosted := updatedVideo.SlackPosted
			originalRepo := updatedVideo.Repo
			originalNotifiedSponsors := updatedVideo.NotifiedSponsors
			originalHugoPath := updatedVideo.HugoPath
			createHugo := updatedVideo.HugoPath != ""

			sponsorsNotifyText := "Notify Sponsors"
			if updatedVideo.NotifiedSponsors || len(updatedVideo.Sponsorship.Amount) == 0 || updatedVideo.Sponsorship.Amount == "N/A" || updatedVideo.Sponsorship.Amount == "-" {
				sponsorsNotifyText = m.greenStyle.Render(sponsorsNotifyText)
			} else {
				sponsorsNotifyText = m.orangeStyle.Render(sponsorsNotifyText) // Was RedStyle
			}

			publishingFormFields := []huh.Field{
				huh.NewConfirm().Title(m.colorTitleBool("Create/Update Hugo Post", createHugo)).Value(&createHugo),
				huh.NewInput().Title(m.colorTitleString("YouTube Video ID (after upload)", updatedVideo.UploadVideo)).Value(&updatedVideo.UploadVideo),
				huh.NewConfirm().Title(m.colorTitleBool("BlueSky Post Sent", updatedVideo.BlueSkyPosted)).Value(&updatedVideo.BlueSkyPosted),
				huh.NewConfirm().Title(m.colorTitleBool("LinkedIn Post Sent", updatedVideo.LinkedInPosted)).Value(&updatedVideo.LinkedInPosted),
				huh.NewConfirm().Title(m.colorTitleBool("Slack Post Sent", updatedVideo.SlackPosted)).Value(&updatedVideo.SlackPosted),
				huh.NewConfirm().Title(m.colorTitleBool("YouTube Highlight Created", updatedVideo.YouTubeHighlight)).Value(&updatedVideo.YouTubeHighlight),
				huh.NewConfirm().Title(m.colorTitleBool("YouTube Pinned Comment Added", updatedVideo.YouTubeComment)).Value(&updatedVideo.YouTubeComment),
				huh.NewConfirm().Title(m.colorTitleBool("Replied to YouTube Comments", updatedVideo.YouTubeCommentReply)).Value(&updatedVideo.YouTubeCommentReply),
				huh.NewConfirm().Title(m.colorTitleBool("GDE Advocu Post Sent", updatedVideo.GDE)).Value(&updatedVideo.GDE),
				huh.NewInput().Title(m.colorTitleString("Code Repository URL", updatedVideo.Repo)).Value(&updatedVideo.Repo),
				huh.NewConfirm().Title(sponsorsNotifyText).Value(&updatedVideo.NotifiedSponsors),
				huh.NewConfirm().Affirmative("Save & Process Actions").Negative("Cancel").Value(&save),
			}

			phasePublishingForm := huh.NewForm(huh.NewGroup(publishingFormFields...))
			err = phasePublishingForm.Run()

			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render("Publishing edit cancelled."))
					continue
				}
				return fmt.Errorf("failed to run publishing edit form: %w", err)
			}

			if save {
				yaml := storage.YAML{}

				// --- Actions Section ---
				// We will perform actions first, and only if they succeed, we keep the user's intent.
				// If an action fails, we revert the corresponding boolean in updatedVideo.

				// Action: Hugo Post
				if createHugo && updatedVideo.HugoPath == "" && originalHugoPath == "" { // Create new Hugo post
					hugoPublisher := publishing.Hugo{}
					createdPath, hugoErr := hugoPublisher.Post(updatedVideo.Gist, updatedVideo.Title, updatedVideo.Date)
					if hugoErr != nil {
						log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to create Hugo post: %v", hugoErr)))
						updatedVideo.HugoPath = originalHugoPath // Revert intent
						// No return here yet, we'll save the reverted state and then let the outer error handling catch it if needed
						// or decide later if this specific error should halt everything before save.
						// For now, let's allow other actions to proceed and save the partially successful state.
						// Qodo comment suggests this could be problematic. Let's stick to returning critical errors.
						return fmt.Errorf("failed to create Hugo post: %w", hugoErr)
					} else {
						updatedVideo.HugoPath = createdPath // Action succeeded, keep intent
					}
				} else if !createHugo { // User deselected Hugo creation
					updatedVideo.HugoPath = ""
				}

				// Action: Set VideoID from UploadVideo and Upload Thumbnail
				actualVideoID := updatedVideo.VideoId                            // Assume it might have been pre-filled
				if originalUploadVideo == "" && updatedVideo.UploadVideo != "" { // New video upload URL provided
					vidID := strings.ReplaceAll(updatedVideo.UploadVideo, "https://youtu.be/", "")
					vidID = strings.ReplaceAll(vidID, "https://www.youtube.com/watch?v=", "")
					if strings.Contains(vidID, "&") {
						vidID = vidID[:strings.Index(vidID, "&")]
					}
					actualVideoID = vidID // This is the derived VideoId
					// No direct failure possible here, but thumbnail upload depends on it
					fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Video ID derived as: %s", actualVideoID)))

					if updatedVideo.Thumbnail != "" { // User provided/confirmed a thumbnail path
						tempVideoForThumbnail := updatedVideo         // Create a temporary copy
						tempVideoForThumbnail.VideoId = actualVideoID // Ensure it has the ID for UploadThumbnail
						if tnErr := publishing.UploadThumbnail(tempVideoForThumbnail); tnErr != nil {
							log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to upload thumbnail: %v", tnErr)))
							// If thumbnail upload fails, what does it mean for `updatedVideo.UploadVideo` intent?
							// The VideoId might still be valid. For now, we only log and return, as per previous change.
							return fmt.Errorf("failed to upload thumbnail: %w", tnErr)
						} else {
							fmt.Println(m.confirmationStyle.Render("Thumbnail uploaded."))
						}
					}
					fmt.Println(m.orangeStyle.Render("Manual YouTube Studio Actions Needed: End screen, Playlists, Language, Monetization"))
				} // End of new video upload URL block
				updatedVideo.VideoId = actualVideoID // Ensure updatedVideo has the final video ID

				// Action: LinkedIn Post
				if !originalLinkedInPosted && updatedVideo.LinkedInPosted && updatedVideo.Tweet != "" && updatedVideo.VideoId != "" {
					platform.PostLinkedIn(updatedVideo.Tweet, updatedVideo.VideoId, publishing.GetYouTubeURL, m.confirmationStyle)
					// No programmatic error to catch here, it's a manual step. updatedVideo.LinkedInPosted reflects intent.
				} else if originalLinkedInPosted && !updatedVideo.LinkedInPosted { // User deselected
					updatedVideo.LinkedInPosted = false
				}

				// Action: Slack Post
				if !originalSlackPosted && updatedVideo.SlackPosted && updatedVideo.VideoId != "" {
					platform.PostSlack(updatedVideo.VideoId, publishing.GetYouTubeURL, m.confirmationStyle)
					// No programmatic error to catch here. updatedVideo.SlackPosted reflects intent.
				} else if originalSlackPosted && !updatedVideo.SlackPosted { // User deselected
					updatedVideo.SlackPosted = false
				}

				// Action: BlueSky Post
				if !originalBlueSkyPosted && updatedVideo.BlueSkyPosted && updatedVideo.Tweet != "" && updatedVideo.VideoId != "" {
					if configuration.GlobalSettings.Bluesky.Identifier == "" || configuration.GlobalSettings.Bluesky.Password == "" {
						log.Printf(m.errorStyle.Render("BlueSky credentials not configured. Cannot post to BlueSky."))
						updatedVideo.BlueSkyPosted = false // Revert intent as action cannot be performed
					} else {
						bsConfig := bluesky.Config{
							Identifier: configuration.GlobalSettings.Bluesky.Identifier,
							Password:   configuration.GlobalSettings.Bluesky.Password,
							URL:        configuration.GlobalSettings.Bluesky.URL,
						}
						bsPost := bluesky.Post{
							Text:          updatedVideo.Tweet,
							YouTubeURL:    publishing.GetYouTubeURL(updatedVideo.VideoId),
							VideoID:       updatedVideo.VideoId,
							ThumbnailPath: updatedVideo.Thumbnail,
						}
						if _, bsErr := bluesky.CreatePost(bsConfig, bsPost); bsErr != nil {
							log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to post to BlueSky: %v", bsErr)))
							updatedVideo.BlueSkyPosted = false                        // Revert intent
							return fmt.Errorf("failed to post to BlueSky: %w", bsErr) // Critical, return error
						} else {
							fmt.Println(m.confirmationStyle.Render("Posted to BlueSky."))
						}
					}
				} else if originalBlueSkyPosted && !updatedVideo.BlueSkyPosted { // User deselected
					updatedVideo.BlueSkyPosted = false
				}

				// Action: Repo Update (Conceptual)
				if originalRepo == "" && updatedVideo.Repo != "" && updatedVideo.Repo != "N/A" {
					log.Println(m.orangeStyle.Render(fmt.Sprintf("TODO: Implement repository update for %s with title %s, videoId %s", updatedVideo.Repo, updatedVideo.Title, updatedVideo.VideoId)))
					// If this had a real implementation that could fail:
					// if repoUpdateErr != nil { updatedVideo.Repo = originalRepo; log.Printf(...); /* return if critical */ }
				} else if originalRepo != "" && (updatedVideo.Repo == "" || updatedVideo.Repo == "N/A") { // User cleared repo
					updatedVideo.Repo = ""
				}

				// Action: Notify Sponsors
				if !originalNotifiedSponsors && updatedVideo.NotifiedSponsors && len(updatedVideo.Sponsorship.Emails) > 0 {
					if configuration.GlobalSettings.Email.Password == "" {
						log.Println(m.errorStyle.Render("Email password not configured. Cannot send sponsor notification."))
						updatedVideo.NotifiedSponsors = false // Revert intent
					} else {
						emailService := notification.NewEmail(configuration.GlobalSettings.Email.Password)
						// Assuming SendSponsors doesn't return an error that needs handling here, or we log it inside.
						emailService.SendSponsors(configuration.GlobalSettings.Email.From, updatedVideo.Sponsorship.Emails, updatedVideo.VideoId, updatedVideo.Sponsorship.Amount)
						fmt.Println(m.confirmationStyle.Render("Sponsor notification email sent."))
					}
				} else if originalNotifiedSponsors && !updatedVideo.NotifiedSponsors { // User deselected
					updatedVideo.NotifiedSponsors = false
				}

				// --- End of Actions Section ---

				// Now, calculate completion based on the *actual* state of updatedVideo after actions.
				updatedVideo.Publish.Completed, updatedVideo.Publish.Total = m.countCompletedTasks([]interface{}{
					updatedVideo.UploadVideo,    // This is the URL string, considered done if present
					updatedVideo.HugoPath,       // Considered done if path exists
					updatedVideo.BlueSkyPosted,  // Boolean, true if action succeeded
					updatedVideo.LinkedInPosted, // Boolean, true if action succeeded
					updatedVideo.SlackPosted,    // Boolean, true if action succeeded
					updatedVideo.YouTubeHighlight,
					updatedVideo.YouTubeComment,
					updatedVideo.YouTubeCommentReply,
					updatedVideo.GDE,
					updatedVideo.Repo, // URL string, considered done if present (and not N/A)
				})
				// Special logic for NotifiedSponsors completion (based on actual state)
				updatedVideo.Publish.Total++
				if updatedVideo.NotifiedSponsors || len(updatedVideo.Sponsorship.Amount) == 0 || updatedVideo.Sponsorship.Amount == "N/A" || updatedVideo.Sponsorship.Amount == "-" {
					updatedVideo.Publish.Completed++
				}

				// Finally, save the video with its actual state after all actions.
				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save publishing details: %w", err)
				}

				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' publishing details updated and actions processed.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for publishing."))
			}

		case actionReturn:
			return nil // Return to video list
		}

		if err != nil {
			// Log or display error from phase handlers if any
			// For now, just return it, but might want to allow continuing the loop
			return fmt.Errorf("error during edit phase '%d': %w", selectedEditPhase, err)
		}
		// Loop continues to allow editing other phases or returning
	}
}
