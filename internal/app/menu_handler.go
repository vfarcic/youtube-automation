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
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/slack"
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
	videoService      *service.VideoService
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
	editPhasePostPublish
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
			if len(field.(string)) > 0 && field.(string) != "-" { // Field is complete if not empty and not just a dash
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
	if len(value) > 0 && value != "-" { // Green if not empty and not just a dash
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
		index, err := yaml.GetIndex()
		if err != nil {
			return fmt.Errorf("failed to get video index for create: %w", err)
		}
		var item storage.VideoIndex
		item, err = m.ChooseCreateVideoAndHandleError()
		if err != nil {
			return fmt.Errorf("error in create video choice: %w", err)
		}
		if len(item.Category) > 0 && len(item.Name) > 0 {
			index = append(index, item)
			yaml.WriteIndex(index)
		}
	case indexListVideos:
		for {
			var index []storage.VideoIndex
			index, err = yaml.GetIndex()
			if err != nil {
				return fmt.Errorf("failed to get video index for list: %w", err)
			}
			var returnVal bool
			returnVal, err = m.ChooseVideosPhaseAndHandleError(index)
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
	
	// Use the service to create the video with all the proper logic
	return m.videoService.CreateVideo(name, category)
}

// ChooseVideosPhase handles the video phase selection workflow
func (m *MenuHandler) ChooseVideosPhaseAndHandleError(vi []storage.VideoIndex) (bool, error) {
	if len(vi) == 0 {
		fmt.Println(m.errorStyle.Render("No videos found. Create a video first."))
		return true, nil
	}

	// Get phase counts from the service
	phases, err := m.videoService.GetVideoPhases()
	if err != nil {
		return false, fmt.Errorf("failed to get video phases: %w", err)
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
	err = form.Run()
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
	// Use the service to get videos by phase
	videosInPhase, err := m.videoService.GetVideosByPhase(phase)
	if err != nil {
		return fmt.Errorf("failed to get videos for phase: %w", err)
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

	err = form.Run()
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

		// Use the service to move the video
		moveErr := m.videoService.MoveVideo(selectedVideo.Name, selectedVideo.Category, targetDir.Path)
		if moveErr != nil {
			log.Printf(m.errorStyle.Render(fmt.Sprintf("Error moving video files for '%s': %v", selectedVideo.Name, moveErr)))
		} else {
			fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' files moved to %s", selectedVideo.Name, targetDir.Path)))
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
		// Use the service to delete the video
		if err := m.videoService.DeleteVideo(selectedVideo.Name, selectedVideo.Category); err != nil {
			return false, fmt.Errorf("failed to delete video: %w", err)
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
	// Corrected definition: isSponsored is true if Amount is a positive indicator (not empty, "-", or "N/A")
	isSponsored := len(video.Sponsorship.Amount) > 0 && video.Sponsorship.Amount != "-" && video.Sponsorship.Amount != "N/A"
	// Corrected definition: isBlocked is true if Blocked field has any content at all.
	isBlocked := len(video.Sponsorship.Blocked) > 0

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
	if isBlocked { // True if Blocked field is non-empty (e.g., "Reason", "-", "N/A")
		blockDisplay := video.Sponsorship.Blocked
		if blockDisplay == "-" || blockDisplay == "N/A" { // Standardize specific placeholders to (B)
			blockDisplay = "B"
		}
		// If video.Sponsorship.Blocked was an actual reason like "Legal", blockDisplay remains "Legal".
		// If it was "-" or "N/A", blockDisplay is now "B".
		styledTitle = fmt.Sprintf("%s (%s)", styledTitle, blockDisplay)
	} else { // Blocked field is empty, so it's not blocked
		if len(video.Date) > 0 {
			styledTitle = fmt.Sprintf("%s (%s)", styledTitle, video.Date)
		}
		if isSponsored { // isSponsored is false if Amount is empty, "-", or "N/A"
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
			huh.NewOption(m.getEditPhaseOptionText("Work In Progress", videoToEdit.Work.Completed, videoToEdit.Work.Total), editPhaseWork),
			huh.NewOption(m.getEditPhaseOptionText("Definition", videoToEdit.Define.Completed, videoToEdit.Define.Total), editPhaseDefinition),
			huh.NewOption(m.getEditPhaseOptionText("Post-Production", videoToEdit.Edit.Completed, videoToEdit.Edit.Total), editPhasePostProduction),
			huh.NewOption(m.getEditPhaseOptionText("Publishing Details", videoToEdit.Publish.Completed, videoToEdit.Publish.Total), editPhasePublishing),
			huh.NewOption(m.getEditPhaseOptionText("Post-Publish Details", videoToEdit.PostPublish.Completed, videoToEdit.PostPublish.Total), editPhasePostPublish),
			huh.NewOption("Return to Video List", actionReturn),
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
			var uploadTrigger bool // Declare uploadTrigger here
			// Store original values to detect changes for actions
			originalHugoPath := updatedVideo.HugoPath
			// If VideoId is empty, createHugo will be false, also influencing the title color.
			createHugo := updatedVideo.HugoPath != "" && updatedVideo.VideoId != ""

			publishingFormFields := []huh.Field{
				huh.NewInput().Title(m.colorTitleString("Video File Path", updatedVideo.UploadVideo)).Value(&updatedVideo.UploadVideo),
				huh.NewConfirm().Title(m.colorTitleString("Upload Video to YouTube?", updatedVideo.VideoId)).Value(&uploadTrigger),
				huh.NewNote().Title(m.colorTitleString("Current YouTube Video ID", updatedVideo.VideoId)).Description(updatedVideo.VideoId),
				// The m.colorTitleBool will show orange if createHugo is false (e.g. no VideoId)
				// The action logic below also prevents Hugo creation if VideoId is missing.
				huh.NewConfirm().Title(m.colorTitleBool("Create/Update Hugo Post", createHugo)).Value(&createHugo),
				huh.NewConfirm().Affirmative("Save & Process Actions").Negative("Cancel").Value(&save),
			}

			phasePublishingForm := huh.NewForm(
				huh.NewGroup(publishingFormFields...),
			)
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
				// var uploadTrigger bool // Moved declaration to the top of the case block

				// Action: Hugo Post
				// createHugo will be false if VideoId is empty due to its initialization.
				// The additional updatedVideo.VideoId != "" check here is for extra safety but might be redundant.
				if createHugo && updatedVideo.VideoId != "" && updatedVideo.HugoPath == "" && originalHugoPath == "" { // Create new Hugo post only if VideoId is present
					hugoPublisher := publishing.Hugo{}
					createdPath, hugoErr := hugoPublisher.Post(updatedVideo.Gist, updatedVideo.Title, updatedVideo.Date, updatedVideo.VideoId)
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

				// Action: Upload Video to YouTube if requested
				if uploadTrigger && updatedVideo.UploadVideo != "" {
					fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Attempting to upload video: %s", updatedVideo.UploadVideo)))
					newVideoID := publishing.UploadVideo(&updatedVideo) // Pass the whole struct
					if newVideoID == "" {
						log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to upload video from path: %s. YouTube API might have returned an empty ID or an error occurred.", updatedVideo.UploadVideo)))
						// Potentially revert uploadTrigger or handle error more explicitly.
						// For now, if upload fails, newVideoID will be empty, and updatedVideo.VideoId won't be set with a new ID.
						// We might want to return an error here to prevent saving if upload was critical.
						return fmt.Errorf("failed to upload video from path: %s", updatedVideo.UploadVideo)
					} else {
						updatedVideo.VideoId = newVideoID // Store the new video ID
						fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video uploaded successfully. New Video ID: %s", updatedVideo.VideoId)))
						// Thumbnail upload should happen AFTER successful video upload and ID retrieval
						if updatedVideo.Thumbnail != "" { // User provided/confirmed a thumbnail path
							// No need for tempVideoForThumbnail, updatedVideo now has the correct VideoId
							if tnErr := publishing.UploadThumbnail(updatedVideo); tnErr != nil { // Use updatedVideo directly
								log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to upload thumbnail: %v", tnErr)))
								// This error is non-critical to the video upload itself, so we log and continue.
								// Consider if this should be a return fmt.Errorf(...)
							} else {
								fmt.Println(m.confirmationStyle.Render("Thumbnail uploaded."))
							}
						}
						fmt.Println(m.orangeStyle.Render("Manual YouTube Studio Actions Needed: End screen, Playlists, Language, Monetization"))
					}
				}
				// Ensure updatedVideo.VideoId is current, even if no new upload happened but it was pre-filled or changed manually.
				// The following block that derived VideoId from a URL is now removed as UploadVideo is a path.

				// --- End of Actions Section (for Publishing Phase) ---

				// Now, calculate completion based on the *actual* state of updatedVideo after actions.
				updatedVideo.Publish.Completed, updatedVideo.Publish.Total = m.countCompletedTasks([]interface{}{
					updatedVideo.UploadVideo, // Done if path exists
					updatedVideo.HugoPath,    // Done if path exists (and VideoId was present for creation)
				})

				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save publishing details: %w", err)
				}

				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' publishing details updated and actions processed.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for publishing."))
			}

		case editPhasePostPublish: // New case for Post-Publish Details
			save := true
			updatedVideo := videoToEdit                               // Work with a copy
			originalNotifiedSponsors := updatedVideo.NotifiedSponsors // Capture original state
			originalBlueSkyPosted := updatedVideo.BlueSkyPosted       // Capture original state
			originalLinkedInPosted := updatedVideo.LinkedInPosted     // Capture original state
			originalSlackPosted := updatedVideo.SlackPosted           // Capture original state
			originalRepo := updatedVideo.Repo                         // Capture original state for Repo

			// Define sponsorsNotifyText for this scope (editPhasePostPublish)
			sponsorsNotifyText := "Notify Sponsors"
			if updatedVideo.NotifiedSponsors || len(updatedVideo.Sponsorship.Amount) == 0 || updatedVideo.Sponsorship.Amount == "N/A" || updatedVideo.Sponsorship.Amount == "-" {
				sponsorsNotifyText = m.greenStyle.Render(sponsorsNotifyText)
			} else {
				sponsorsNotifyText = m.orangeStyle.Render(sponsorsNotifyText)
			}

			// Define fields for the Post-Publish Details form
			postPublishingFormFields := []huh.Field{
				huh.NewNote().Title("Post-Publish Details"),
				huh.NewConfirm().Title(m.colorTitleBool("BlueSky Post Sent", updatedVideo.BlueSkyPosted)).Value(&updatedVideo.BlueSkyPosted),
				huh.NewConfirm().Title(m.colorTitleBool("LinkedIn Post Sent", updatedVideo.LinkedInPosted)).Value(&updatedVideo.LinkedInPosted),
				huh.NewConfirm().Title(m.colorTitleBool("Slack Post Sent", updatedVideo.SlackPosted)).Value(&updatedVideo.SlackPosted),
				huh.NewConfirm().Title(m.colorTitleBool("YouTube Highlight Created", updatedVideo.YouTubeHighlight)).Value(&updatedVideo.YouTubeHighlight),
				huh.NewConfirm().Title(m.colorTitleBool("YouTube Pinned Comment Added", updatedVideo.YouTubeComment)).Value(&updatedVideo.YouTubeComment),
				huh.NewConfirm().Title(m.colorTitleBool("Replied to YouTube Comments", updatedVideo.YouTubeCommentReply)).Value(&updatedVideo.YouTubeCommentReply),
				huh.NewConfirm().Title(m.colorTitleBool("GDE Advocu Post Sent", updatedVideo.GDE)).Value(&updatedVideo.GDE),
				huh.NewInput().Title(m.colorTitleString("Code Repository URL", updatedVideo.Repo)).Value(&updatedVideo.Repo),
				huh.NewConfirm().Title(sponsorsNotifyText).Value(&updatedVideo.NotifiedSponsors), // Use sponsorsNotifyText here
				huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
			}

			postPublishingForm := huh.NewForm(
				huh.NewGroup(postPublishingFormFields...),
			)
			err = postPublishingForm.Run() // Corrected: was phasePublishingForm.Run()

			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render("Post-Publish details editing cancelled.")) // Corrected message
					continue                                                                     // Go back to phase selection
				}
				log.Printf(m.errorStyle.Render(fmt.Sprintf("Error running post-publish details form: %v", err))) // Corrected message
				return err                                                                                       // Return on other errors
			}

			if save {
				yaml := storage.YAML{}

				// --- Actions Section for Post-Publish Phase ---

				// Action: Notify Sponsors (if changed from false to true in this phase)
				if !originalNotifiedSponsors && updatedVideo.NotifiedSponsors && len(updatedVideo.Sponsorship.Emails) > 0 {
					if configuration.GlobalSettings.Email.Password == "" {
						log.Println(m.errorStyle.Render("Email password not configured. Cannot send sponsor notification."))
						updatedVideo.NotifiedSponsors = false // Revert intent
					} else {
						emailService := notification.NewEmail(configuration.GlobalSettings.Email.Password)
						emailService.SendSponsors(configuration.GlobalSettings.Email.From, updatedVideo.Sponsorship.Emails, updatedVideo.VideoId, updatedVideo.Sponsorship.Amount)
						fmt.Println(m.confirmationStyle.Render("Sponsor notification email sent."))
					}
				} else if originalNotifiedSponsors && !updatedVideo.NotifiedSponsors { // User deselected in this phase
					// No action needed other than saving the state
				}

				// Action: LinkedIn Post (if changed from false to true in this phase)
				if !originalLinkedInPosted && updatedVideo.LinkedInPosted && updatedVideo.Tweet != "" && updatedVideo.VideoId != "" {
					platform.PostLinkedIn(updatedVideo.Tweet, updatedVideo.VideoId, publishing.GetYouTubeURL, m.confirmationStyle)
					fmt.Println(m.confirmationStyle.Render("LinkedIn post triggered."))
				} else if originalLinkedInPosted && !updatedVideo.LinkedInPosted { // User deselected
					// No action needed
				}

				// Action: Slack Post (if changed from false to true in this phase)
				if !originalSlackPosted && updatedVideo.SlackPosted && updatedVideo.VideoId != "" {
					if errSl := slack.LoadAndValidateSlackConfig(""); errSl != nil { // Renamed err to errSl
						log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to load Slack configuration: %v", errSl)))
						updatedVideo.SlackPosted = false                                   // Revert intent
						return fmt.Errorf("failed to load Slack configuration: %w", errSl) // Return error
					} else {
						slackService, errSlSvc := slack.NewSlackService(slack.GlobalSlackConfig) // Renamed err to errSlSvc
						if errSlSvc != nil {
							log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to create Slack service: %v", errSlSvc)))
							updatedVideo.SlackPosted = false                                  // Revert intent
							return fmt.Errorf("failed to create Slack service: %w", errSlSvc) // Return error
						} else {
							errSlPost := slackService.PostVideo(&updatedVideo, updatedVideo.Path) // Renamed err to errSlPost
							if errSlPost != nil {
								log.Printf(m.errorStyle.Render(fmt.Sprintf("Failed to post video to Slack: %v", errSlPost)))
								updatedVideo.SlackPosted = false                                  // Revert intent
								return fmt.Errorf("failed to post video to Slack: %w", errSlPost) // Return error
							} else {
								fmt.Println(m.confirmationStyle.Render("Successfully posted to Slack."))
							}
						}
					}
				} else if originalSlackPosted && !updatedVideo.SlackPosted { // User deselected
					// No action needed
				}

				// Action: BlueSky Post (if changed from false to true in this phase)
				if !originalBlueSkyPosted && updatedVideo.BlueSkyPosted && updatedVideo.Tweet != "" && updatedVideo.VideoId != "" {
					if configuration.GlobalSettings.Bluesky.Identifier == "" || configuration.GlobalSettings.Bluesky.Password == "" {
						log.Printf(m.errorStyle.Render("BlueSky credentials not configured. Cannot post to BlueSky."))
						updatedVideo.BlueSkyPosted = false // Revert intent
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
							updatedVideo.BlueSkyPosted = false // Revert intent
						} else {
							fmt.Println(m.confirmationStyle.Render("Posted to BlueSky."))
						}
					}
				} else if originalBlueSkyPosted && !updatedVideo.BlueSkyPosted { // User deselected
					// No action needed
				}

				// Action: Repo Update (if changed meaningfully in this phase)
				if updatedVideo.Repo != originalRepo && updatedVideo.Repo != "" && updatedVideo.Repo != "N/A" {
					log.Println(m.orangeStyle.Render(fmt.Sprintf("TODO: Implement repository update for %s with title %s, videoId %s", updatedVideo.Repo, updatedVideo.Title, updatedVideo.VideoId)))
					// If this had a real implementation that could fail:
					// if repoUpdateErr != nil { updatedVideo.Repo = originalRepo; log.Printf(...); /* return if critical */ }
				} else if updatedVideo.Repo != originalRepo && (updatedVideo.Repo == "" || updatedVideo.Repo == "N/A") { // User cleared repo or set to N/A
					// Just save the cleared/N/A state, no specific action beyond that.
				}

				// --- End of Actions Section for Post-Publish ---

				// Update task counts for PostPublish phase
				updatedVideo.PostPublish.Completed, updatedVideo.PostPublish.Total = m.countCompletedTasks([]interface{}{
					updatedVideo.BlueSkyPosted,
					updatedVideo.LinkedInPosted,
					updatedVideo.SlackPosted,
					updatedVideo.YouTubeHighlight,
					updatedVideo.YouTubeComment,
					updatedVideo.YouTubeCommentReply,
					updatedVideo.GDE,
					updatedVideo.Repo, // Ensure Repo is counted for completion
					// NotifiedSponsors has special logic
				})
				// Special logic for NotifiedSponsors completion
				updatedVideo.PostPublish.Total++
				if updatedVideo.NotifiedSponsors || len(updatedVideo.Sponsorship.Amount) == 0 || updatedVideo.Sponsorship.Amount == "N/A" || updatedVideo.Sponsorship.Amount == "-" {
					updatedVideo.PostPublish.Completed++
				}

				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save post-publish details: %w", err)
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' post-publish details updated.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for post-publish details."))
			}

		case actionReturn:
			return nil // Return from editing this video
		}

		if err != nil {
			// Log or display error from phase handlers if any
			// For now, just return it, but might want to allow continuing the loop
			return fmt.Errorf("error during edit phase '%d': %w", selectedEditPhase, err)
		}
		// Loop continues to allow editing other phases or returning
	}
}
