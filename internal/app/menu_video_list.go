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
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/workflow"
	"devopstoolkit/youtube-automation/pkg/utils"

	"github.com/charmbracelet/huh"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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
	case indexAnalyze:
		err = m.HandleAnalyzeMenu()
		if err != nil {
			return fmt.Errorf("error in analyze menu: %w", err)
		}
	case actionReturn:
		return ErrExitApplication
	}
	return nil
}

// GetPhaseText returns formatted text for a phase with completion status
func (m *MenuHandler) GetPhaseText(text string, completed, total int) string {
	text = fmt.Sprintf("%s (%d/%d)", text, completed, total)
	if completed == total && total > 0 {
		return m.greenStyle.Render(text)
	}
	return m.orangeStyle.Render(text)
}

// ChooseCreateVideo handles video creation workflow
func (m *MenuHandler) ChooseCreateVideoAndHandleError() (storage.VideoIndex, error) {
	var name, category, date string
	save := true
	fields, err := cli.GetCreateVideoFields(&name, &category, &date, &save)
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
	return m.videoService.CreateVideo(name, category, date)
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

	// Create video selection options using video names (strings are comparable)
	var videoOptions []huh.Option[string]
	for _, video := range videosInPhase {
		displayTitle := m.getVideoTitleForDisplay(video, phase, time.Now())
		videoOptions = append(videoOptions, huh.NewOption(displayTitle, video.Name))
	}
	videoOptions = append(videoOptions, huh.NewOption("Return", "return"))

	var selectedVideoName string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a video:").
				Options(videoOptions...).
				Value(&selectedVideoName),
		),
	)

	if input != nil {
		form = form.WithInput(input)
	}

	err = form.Run()
	if err != nil {
		return fmt.Errorf("failed to run video selection form: %w", err)
	}

	if selectedVideoName == "return" {
		return nil
	}

	// Look up the selected video by name
	var selectedVideo storage.Video
	for _, video := range videosInPhase {
		if video.Name == selectedVideoName {
			selectedVideo = video
			break
		}
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
			log.Print(m.errorStyle.Render(fmt.Sprintf("Error during video edit phases: %v", err)))
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
				log.Print(m.errorStyle.Render(fmt.Sprintf("Error selecting target directory: %v", selErr)))
			}
			return nil
		}

		// Use the service to move the video
		moveErr := m.videoService.MoveVideo(selectedVideo.Name, selectedVideo.Category, targetDir.Path)
		if moveErr != nil {
			log.Print(m.errorStyle.Render(fmt.Sprintf("Error moving video files for '%s': %v", selectedVideo.Name, moveErr)))
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
		huh.NewOption("Analyze", indexAnalyze),
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
	if total > 0 && completed == total { // Corrected logic from previous diff
		return m.greenStyle.Render(text)
	}
	return m.orangeStyle.Render(text)
}
