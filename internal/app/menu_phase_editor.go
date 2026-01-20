package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/calendar"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/constants"
	"devopstoolkit/youtube-automation/internal/dubbing"
	"devopstoolkit/youtube-automation/internal/notification"
	"devopstoolkit/youtube-automation/internal/platform"
	"devopstoolkit/youtube-automation/internal/platform/bluesky"
	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/slack"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/charmbracelet/huh"
)

// handleEditVideoPhases presents a menu to choose which aspect of a video to edit.
func (m *MenuHandler) handleEditVideoPhases(videoToEdit storage.Video) error {
	for {
		var selectedEditPhase int // Keep original variable name for minimal diff

		// Use shared video manager for consistent progress calculations
		initCompleted, initTotal := m.videoManager.CalculateInitialDetailsProgress(videoToEdit)
		workCompleted, workTotal := m.videoManager.CalculateWorkProgressProgress(videoToEdit)
		defineCompleted, defineTotal := m.videoManager.CalculateDefinePhaseCompletion(videoToEdit)
		editCompleted, editTotal := m.videoManager.CalculatePostProductionProgress(videoToEdit)
		publishCompleted, publishTotal := m.videoManager.CalculatePublishingProgress(videoToEdit)
		dubbingCompleted, dubbingTotal := m.videoManager.CalculateDubbingProgress(videoToEdit)
		postPublishCompleted, postPublishTotal := m.videoManager.CalculatePostPublishProgress(videoToEdit)
		analysisCompleted, analysisTotal := m.videoManager.CalculateAnalysisProgress(videoToEdit)

		editPhaseOptions := []huh.Option[int]{
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleInitialDetails, initCompleted, initTotal), editPhaseInitial),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleWorkProgress, workCompleted, workTotal), editPhaseWork),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleDefinition, defineCompleted, defineTotal), editPhaseDefinition),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitlePostProduction, editCompleted, editTotal), editPhasePostProduction),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleUpload, publishCompleted, publishTotal), editPhasePublishing),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleDubbing, dubbingCompleted, dubbingTotal), editPhaseDubbing),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitlePostPublish, postPublishCompleted, postPublishTotal), editPhasePostPublish),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleAnalysis, analysisCompleted, analysisTotal), editPhaseAnalysis),
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
			applyRandomTiming := false // Track if user wants to apply random timing

			// Auto-populate Gist path if empty, similar to old logic
			if len(updatedVideo.Gist) == 0 && updatedVideo.Path != "" {
				updatedVideo.Gist = strings.Replace(updatedVideo.Path, ".yaml", ".md", 1)
			}

			initialFormFields := []huh.Field{
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleProjectName, updatedVideo.ProjectName)).Value(&updatedVideo.ProjectName),
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleProjectURL, updatedVideo.ProjectURL)).Value(&updatedVideo.ProjectURL),
				huh.NewInput().Title(m.colorTitleStringInverse(constants.FieldTitleSponsorshipName, updatedVideo.Sponsorship.Name)).Value(&updatedVideo.Sponsorship.Name),
				huh.NewInput().Title(m.colorTitleStringInverse(constants.FieldTitleSponsorshipURL, updatedVideo.Sponsorship.URL)).Value(&updatedVideo.Sponsorship.URL),
				huh.NewInput().Title(m.colorTitleSponsorshipAmount(constants.FieldTitleSponsorshipAmount, updatedVideo.Sponsorship.Amount)).Value(&updatedVideo.Sponsorship.Amount),
				huh.NewInput().Title(m.colorTitleSponsoredEmails(constants.FieldTitleSponsorshipEmails, updatedVideo.Sponsorship.Amount, updatedVideo.Sponsorship.Emails)).Value(&updatedVideo.Sponsorship.Emails),
				huh.NewInput().Title(m.colorTitleStringInverse(constants.FieldTitleSponsorshipBlocked, updatedVideo.Sponsorship.Blocked)).Value(&updatedVideo.Sponsorship.Blocked),
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitlePublishDate, updatedVideo.Date)).Value(&updatedVideo.Date),
				huh.NewConfirm().
					Title("Apply Random Timing?").
					Description("Pick a random timing recommendation from settings.yaml").
					Affirmative("Yes").
					Negative("No").
					Value(&applyRandomTiming),
				huh.NewConfirm().Title(m.colorTitleBoolInverse(constants.FieldTitleDelayed, updatedVideo.Delayed)).Value(&updatedVideo.Delayed), // True means NOT delayed, so inverse logic for green
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleGistPath, updatedVideo.Gist)).Value(&updatedVideo.Gist),
				huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
			}

			phaseInitialForm := huh.NewForm(huh.NewGroup(initialFormFields...))
			err = phaseInitialForm.Run()

			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render(MessageInitialDetailsEditCancelled))
					continue // Continue the loop to re-select edit phase
				}
				return fmt.Errorf("%s: %w", ErrorRunInitialDetailsForm, err)
			}

			// Handle random timing application if user requested it
			if applyRandomTiming && save {
				// Load timing recommendations from settings
				recommendations := m.settings.Timing.Recommendations

				if len(recommendations) == 0 {
					fmt.Println(m.orangeStyle.Render("⚠️  No timing recommendations found in settings.yaml"))
					fmt.Println(m.normalStyle.Render("   Run 'Analyze → Timing' to generate recommendations first."))
				} else {
					// Apply random timing
					originalDate := updatedVideo.Date
					newDateStr, selectedRec, timingErr := ApplyRandomTiming(updatedVideo.Date, recommendations)
					if timingErr != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error applying random timing: %v", timingErr)))
					} else {
						// Update the video date
						updatedVideo.Date = newDateStr

						// Show user what changed
						fmt.Println(m.normalStyle.Render("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
						fmt.Println(m.confirmationStyle.Render("🎲 Random timing applied:"))
						fmt.Println(m.normalStyle.Render(fmt.Sprintf("   %s %s UTC", selectedRec.Day, selectedRec.Time)))
						fmt.Println(m.normalStyle.Render(fmt.Sprintf("   Reasoning: %s", selectedRec.Reasoning)))
						fmt.Println(m.normalStyle.Render(""))

						// Format dates with day of week
						originalDateFormatted := originalDate
						newDateFormatted := newDateStr
						if parsedOriginal, parseErr := time.Parse("2006-01-02T15:04", originalDate); parseErr == nil {
							originalDateFormatted = fmt.Sprintf("%s, %s", parsedOriginal.Format("Monday"), originalDate)
						}
						if parsedNew, parseErr := time.Parse("2006-01-02T15:04", newDateStr); parseErr == nil {
							newDateFormatted = fmt.Sprintf("%s, %s", parsedNew.Format("Monday"), newDateStr)
						}

						fmt.Println(m.normalStyle.Render(fmt.Sprintf("📅 Original date: %s", originalDateFormatted)))
						fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("📅 New date:      %s", newDateFormatted)))

						// Parse dates to show week boundaries
						if parsedDate, parseErr := time.Parse("2006-01-02T15:04", newDateStr); parseErr == nil {
							monday, sunday := GetWeekBoundaries(parsedDate)
							fmt.Println(m.normalStyle.Render(fmt.Sprintf("   (Same week: Monday %s - Sunday %s)",
								monday.Format("Jan 2"), sunday.Format("Jan 2, 2006"))))
						}
						fmt.Println(m.normalStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
					}
				}
			}

			if save {
				yaml := storage.YAML{}

				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("%s: %w", ErrorSaveInitialDetails, err)
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' %s.", updatedVideo.Name, MessageInitialDetailsUpdated)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render(MessageChangesNotSavedInitialDetails))
			}

		case editPhaseWork:
			save := true
			workFormFields := []huh.Field{
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleCodeDone, updatedVideo.Code)).Value(&updatedVideo.Code),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleTalkingHeadDone, updatedVideo.Head)).Value(&updatedVideo.Head),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleScreenRecordingDone, updatedVideo.Screen)).Value(&updatedVideo.Screen),
				huh.NewText().Lines(3).CharLimit(1000).Title(m.colorTitleString(constants.FieldTitleRelatedVideos, updatedVideo.RelatedVideos)).Value(&updatedVideo.RelatedVideos),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleThumbnailsDone, updatedVideo.Thumbnails)).Value(&updatedVideo.Thumbnails),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleDiagramsDone, updatedVideo.Diagrams)).Value(&updatedVideo.Diagrams),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleScreenshotsDone, updatedVideo.Screenshots)).Value(&updatedVideo.Screenshots),
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleFilesLocation, updatedVideo.Location)).Value(&updatedVideo.Location),
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleTagline, updatedVideo.Tagline)).Value(&updatedVideo.Tagline),
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleTaglineIdeas, updatedVideo.TaglineIdeas)).Value(&updatedVideo.TaglineIdeas),
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleOtherLogos, updatedVideo.OtherLogos)).Value(&updatedVideo.OtherLogos),
				huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
			}

			phaseWorkForm := huh.NewForm(huh.NewGroup(workFormFields...))
			err = phaseWorkForm.Run()

			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render(MessageWorkProgressEditCancelled))
					continue // Continue the loop to re-select edit phase
				}
				return fmt.Errorf("%s: %w", ErrorRunWorkProgressForm, err)
			}

			if save {
				yaml := storage.YAML{}
				// No longer store calculated values - both CLI and API use real-time calculations only
				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("%s: %w", ErrorSaveWorkProgress, err)
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' %s.", updatedVideo.Name, MessageWorkProgressUpdated)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render(MessageChangesNotSavedWorkProgress))
			}

		case editPhaseDefinition:
			updatedVideo, err = m.editPhaseDefinition(updatedVideo, m.settings) // updatedVideo was videoToEdit
			if err != nil {
				return fmt.Errorf("%s: %w", ErrorDefinitionPhase, err)
			}
			videoToEdit = updatedVideo // Persist changes for the next loop iteration

		case editPhasePostProduction:
			save := true
			originalRequestEditStatus := updatedVideo.RequestEdit

			// --- Thumbnail Management Section ---
			// Initialize thumbnail variables from struct
			thumbOriginal := updatedVideo.Thumbnail
			thumbSubtle := ""
			thumbBold := ""

			// Load existing variants if present
			for _, v := range updatedVideo.ThumbnailVariants {
				switch v.Type {
				case "original":
					thumbOriginal = v.Path
				case "subtle":
					thumbSubtle = v.Path
				case "bold":
					thumbBold = v.Path
				}
			}

			// Loop for interactive thumbnail management
			thumbnailDone := false
			const (
				actionThumbContinue = 0
				actionThumbGenerate = 1
			)

			for !thumbnailDone {
				var thumbAction int

				thumbForm := huh.NewForm(
					huh.NewGroup(
						huh.NewNote().Title("Thumbnail Management"),
						huh.NewInput().Title(m.colorTitleString(constants.FieldTitleThumbnailPath, thumbOriginal)).Value(&thumbOriginal).Description("Path to the original thumbnail"),
						huh.NewInput().Title("Thumbnail (Subtle)").Value(&thumbSubtle).Description("Path to the subtle variation"),
						huh.NewInput().Title("Thumbnail (Bold)").Value(&thumbBold).Description("Path to the bold variation"),
						huh.NewSelect[int]().
							Title("Action").
							Options(
								huh.NewOption("Save & Continue to Details", actionThumbContinue),
								huh.NewOption("Generate Variation Prompts (AI)", actionThumbGenerate),
							).
							Value(&thumbAction),
					),
				)

				err := thumbForm.Run()
				if err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						fmt.Println(m.orangeStyle.Render("Thumbnail editing cancelled. Returning to menu."))
						return nil // Return to main menu
					}
					return fmt.Errorf("error in thumbnail form: %w", err)
				}

				switch thumbAction {
				case actionThumbGenerate:
					if thumbOriginal == "" {
						fmt.Println(m.errorStyle.Render("Please enter an Original Thumbnail path first."))
						continue
					}
					// Check if file exists
					if _, err := os.Stat(thumbOriginal); os.IsNotExist(err) {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Original thumbnail file not found: %s", thumbOriginal)))
						continue
					}

					fmt.Println(m.normalStyle.Render("Analyzing thumbnail and generating variations..."))
					ctx := context.Background()
					variations, err := ai.GenerateThumbnailVariations(ctx, thumbOriginal)
					if err != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to generate variations: %v", err)))
					} else {
						fmt.Println("")
						fmt.Println(m.greenStyle.Render("✓ Variations generated! Copy these prompts:"))
						fmt.Println("")
						fmt.Println(m.normalStyle.Render("--- Subtle Variation ---"))
						fmt.Println(variations.Subtle)
						fmt.Println("")
						fmt.Println(m.normalStyle.Render("--- Bold Variation ---"))
						fmt.Println(variations.Bold)
						fmt.Println("")
						fmt.Println(m.orangeStyle.Render("Press Enter to continue..."))
						fmt.Scanln() // Wait for user acknowledgement
					}

				case actionThumbContinue:
					// Update the video struct with the final thumbnail values
					updatedVideo.Thumbnail = thumbOriginal
					updatedVideo.ThumbnailVariants = []storage.ThumbnailVariant{
						{Index: 1, Type: "original", Path: thumbOriginal},
						{Index: 2, Type: "subtle", Path: thumbSubtle},
						{Index: 3, Type: "bold", Path: thumbBold},
					}
					thumbnailDone = true
				}
			}

			// --- Shorts Analysis Section ---
			shortsDone := false
			const (
				actionShortsContinue = 0
				actionShortsAnalyze  = 1
			)

			// Show current shorts count if any
			currentShortsCount := len(updatedVideo.Shorts)
			shortsStatus := "No Shorts analyzed yet"
			if currentShortsCount > 0 {
				shortsStatus = fmt.Sprintf("%d Shorts selected", currentShortsCount)
			}

			for !shortsDone {
				var shortsAction int

				shortsForm := huh.NewForm(
					huh.NewGroup(
						huh.NewNote().Title("YouTube Shorts").Description(shortsStatus),
						huh.NewSelect[int]().
							Title("Action").
							Options(
								huh.NewOption("Save & Continue to Details", actionShortsContinue),
								huh.NewOption("Analyze Manuscript for Shorts (AI)", actionShortsAnalyze),
							).
							Value(&shortsAction),
					),
				)

				err := shortsForm.Run()
				if err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						fmt.Println(m.orangeStyle.Render("Shorts analysis cancelled. Returning to menu."))
						return nil
					}
					return fmt.Errorf("error in shorts form: %w", err)
				}

				switch shortsAction {
				case actionShortsAnalyze:
					selectedShorts, analysisErr := m.HandleAnalyzeShorts(&updatedVideo)
					if analysisErr != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Shorts analysis failed: %v", analysisErr)))
						continue
					}
					if len(selectedShorts) > 0 {
						updatedVideo.Shorts = selectedShorts
						shortsStatus = fmt.Sprintf("%d Shorts selected", len(selectedShorts))
						fmt.Println(m.greenStyle.Render(fmt.Sprintf("✓ %d Shorts selected", len(selectedShorts))))
					}

				case actionShortsContinue:
					// Save shorts to YAML immediately (like Thumbnail section)
					if len(updatedVideo.Shorts) > 0 {
						yaml := storage.YAML{}
						if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
							fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to save shorts: %v", err)))
							continue
						}
						fmt.Println(m.greenStyle.Render(fmt.Sprintf("✓ %d Shorts saved to video YAML", len(updatedVideo.Shorts))))
						videoToEdit = updatedVideo // Persist changes for consistency
					}
					shortsDone = true
				}
			}

			// --- Rest of Post-Production Form ---
			timeCodesTitle := constants.FieldTitleTimecodes
			if strings.Contains(updatedVideo.Timecodes, "FIXME:") {
				timeCodesTitle = m.orangeStyle.Render(timeCodesTitle)
			} else {
				timeCodesTitle = m.greenStyle.Render(timeCodesTitle)
			}

			editFormFields := []huh.Field{
				huh.NewNote().Title("Post-Production Details"),
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleMembers, updatedVideo.Members)).Value(&updatedVideo.Members),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleRequestEdit, updatedVideo.RequestEdit)).Value(&updatedVideo.RequestEdit),
				huh.NewText().Lines(5).CharLimit(10000).Title(timeCodesTitle).Value(&updatedVideo.Timecodes),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleMovieDone, updatedVideo.Movie)).Value(&updatedVideo.Movie),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleSlidesDone, updatedVideo.Slides)).Value(&updatedVideo.Slides),
				huh.NewConfirm().Affirmative("Save All").Negative("Cancel").Value(&save),
			}

			phaseEditForm := huh.NewForm(huh.NewGroup(editFormFields...))
			err = phaseEditForm.Run()
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render(MessagePostProductionEditCancelled))
					continue
				}
				return fmt.Errorf("%s: %w", ErrorRunPostProductionForm, err)
			}

			if save {
				yaml := storage.YAML{}
				// Save the updated video (which now includes new thumbnail variants and other fields)
				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("%s: %w", ErrorSavePostProductionDetails, err)
				}

				if !originalRequestEditStatus && updatedVideo.RequestEdit {
					if configuration.GlobalSettings.Email.Password == "" {
						log.Println(m.errorStyle.Render("Email password not configured. Cannot send edit request email."))
					} else {
						emailService := notification.NewEmail(configuration.GlobalSettings.Email.Password)
						if emailErr := emailService.SendEdit(configuration.GlobalSettings.Email.From, configuration.GlobalSettings.Email.EditTo, updatedVideo); emailErr != nil {
							log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to send edit request email: %v", emailErr)))
						} else {
							fmt.Println(m.confirmationStyle.Render("Edit request email sent."))
						}
					}
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' %s.", updatedVideo.Name, MessagePostProductionUpdated)))
				videoToEdit = updatedVideo // Persist changes to the original reference for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render(MessageChangesNotSavedPostProduction))
			}

		case editPhasePublishing:
			save := true
			var uploadTrigger bool       // Declare uploadTrigger here
			var createCalendarEvent bool // Manual calendar event creation trigger (always defaults to false)
			var uploadShortsTrigger bool // Trigger for uploading shorts
			// Store original values to detect changes for actions
			originalHugoPath := updatedVideo.HugoPath
			// If VideoId is empty, createHugo will be false, also influencing the title color.
			createHugo := updatedVideo.HugoPath != "" && updatedVideo.VideoId != ""

			// Find pending shorts (those without YouTubeID)
			pendingShortIndices := []int{}
			for i, short := range updatedVideo.Shorts {
				if short.YouTubeID == "" {
					pendingShortIndices = append(pendingShortIndices, i)
				}
			}
			// Create slice for short file paths
			shortFilePaths := make([]string, len(pendingShortIndices))
			for i, shortIdx := range pendingShortIndices {
				shortFilePaths[i] = updatedVideo.Shorts[shortIdx].FilePath
			}

			publishingFormFields := []huh.Field{
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleVideoFilePath, updatedVideo.UploadVideo)).Value(&updatedVideo.UploadVideo),
				huh.NewConfirm().Title(m.colorTitleString(constants.FieldTitleUploadToYouTube, updatedVideo.VideoId)).Value(&uploadTrigger),
				huh.NewNote().Title(m.colorTitleString(constants.FieldTitleCurrentVideoID, updatedVideo.VideoId)).Description(updatedVideo.VideoId),
			}
			// Show calendar button unless calendar integration is disabled in settings
			if !configuration.GlobalSettings.Calendar.Disabled {
				publishingFormFields = append(publishingFormFields,
					huh.NewConfirm().Title("Create Calendar Event").Description("Create a Google Calendar reminder for video release").Value(&createCalendarEvent))
			}
			// The m.colorTitleBool will show orange if createHugo is false (e.g. no VideoId)
			// The action logic below also prevents Hugo creation if VideoId is missing.
			publishingFormFields = append(publishingFormFields,
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleCreateHugo, createHugo)).Value(&createHugo))

			// Add shorts upload fields if there are pending shorts and main video is uploaded
			if len(pendingShortIndices) > 0 && updatedVideo.VideoId != "" {
				for i, shortIdx := range pendingShortIndices {
					short := updatedVideo.Shorts[shortIdx]
					fieldTitle := fmt.Sprintf("Short %d: %s", i+1, short.Title)
					publishingFormFields = append(publishingFormFields,
						huh.NewInput().Title(fieldTitle).Value(&shortFilePaths[i]).Placeholder("Leave empty to skip"))
				}

				publishingFormFields = append(publishingFormFields,
					huh.NewConfirm().Title("Upload Shorts").Description("Upload shorts with specified file paths").Value(&uploadShortsTrigger))
			}

			publishingFormFields = append(publishingFormFields,
				huh.NewConfirm().Affirmative("Save & Process Actions").Negative("Cancel").Value(&save))

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

				// Action: Hugo Post
				// createHugo will be false if VideoId is empty due to its initialization.
				// The additional updatedVideo.VideoId != "" check here is for extra safety but might be redundant.
				if createHugo && updatedVideo.VideoId != "" && updatedVideo.HugoPath == "" && originalHugoPath == "" { // Create new Hugo post only if VideoId is present
					hugoPublisher := publishing.Hugo{}
					createdPath, hugoErr := hugoPublisher.Post(updatedVideo.Gist, updatedVideo.GetUploadTitle(), updatedVideo.Date, updatedVideo.VideoId)
					if hugoErr != nil {
						log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to create Hugo post: %v", hugoErr)))
						updatedVideo.HugoPath = originalHugoPath // Revert intent
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
						log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to upload video from path: %s. YouTube API might have returned an empty ID or an error occurred.", updatedVideo.UploadVideo)))
						return fmt.Errorf("failed to upload video from path: %s", updatedVideo.UploadVideo)
					} else {
						updatedVideo.VideoId = newVideoID // Store the new video ID
						fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video uploaded successfully. New Video ID: %s", updatedVideo.VideoId)))
						// Thumbnail upload should happen AFTER successful video upload and ID retrieval
						if updatedVideo.Thumbnail != "" { // User provided/confirmed a thumbnail path
							if tnErr := publishing.UploadThumbnail(updatedVideo); tnErr != nil {
								log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to upload thumbnail: %v", tnErr)))
							} else {
								fmt.Println(m.confirmationStyle.Render("Thumbnail uploaded."))
							}
						}
						fmt.Println(m.orangeStyle.Render("Manual YouTube Studio Actions Needed: End screen, Playlists, Language, Monetization"))
					}
				}

				// Action: Manual Calendar Event Creation (independent of upload flow)
				if createCalendarEvent && updatedVideo.VideoId != "" && updatedVideo.Date != "" {
					if publishTime, parseErr := time.Parse("2006-01-02T15:04", updatedVideo.Date); parseErr == nil {
						calService, calErr := calendar.NewCalendarService(context.Background())
						if calErr != nil {
							log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to initialize calendar service: %v", calErr)))
						} else {
							youtubeURL := publishing.GetYouTubeURL(updatedVideo.VideoId)
							_, eventErr := calService.CreateVideoReleaseEvent(context.Background(), updatedVideo.GetUploadTitle(), youtubeURL, publishTime)
							if eventErr != nil {
								log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to create calendar event: %v", eventErr)))
							} else {
								fmt.Println(m.confirmationStyle.Render("Calendar event created for video release."))
							}
						}
					} else {
						log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to parse publish date for calendar event: %v", parseErr)))
					}
				} else if createCalendarEvent {
					// User selected to create calendar event but prerequisites are missing
					if updatedVideo.VideoId == "" {
						log.Print(m.errorStyle.Render("Cannot create calendar event: Video ID is missing. Upload video first."))
					}
					if updatedVideo.Date == "" {
						log.Print(m.errorStyle.Render("Cannot create calendar event: Publish date is not set."))
					}
				}

				// Action: Upload Shorts if requested
				if uploadShortsTrigger && len(pendingShortIndices) > 0 && updatedVideo.VideoId != "" {
					// Parse main video date for scheduling
					mainVideoDate, dateErr := time.Parse("2006-01-02T15:04", updatedVideo.Date)
					if dateErr != nil {
						mainVideoDate, dateErr = time.Parse("2006-01-02T15:04:05Z", updatedVideo.Date)
					}
					if dateErr != nil {
						log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to parse main video date for shorts scheduling: %v", dateErr)))
					} else {
						// Calculate scheduled dates
						schedules := publishing.CalculateShortsSchedule(mainVideoDate, len(pendingShortIndices))

						// Validate file paths exist before uploading
						allPathsValid := true
						for i, filePath := range shortFilePaths {
							if filePath != "" {
								if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
									log.Print(m.errorStyle.Render(fmt.Sprintf("File does not exist for short %d: %s", i+1, filePath)))
									allPathsValid = false
								}
							}
						}

						if allPathsValid {
							uploadedCount := 0
							for i, shortIdx := range pendingShortIndices {
								filePath := shortFilePaths[i]
								if filePath == "" {
									fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Skipping short %d (no file path)", i+1)))
									continue
								}

								short := &updatedVideo.Shorts[shortIdx]
								short.FilePath = filePath
								short.ScheduledDate = publishing.FormatScheduleISO(schedules[i])

								fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Uploading short %d: %s", i+1, short.Title)))

								youtubeID, uploadErr := publishing.UploadShort(filePath, *short, updatedVideo.VideoId)
								if uploadErr != nil {
									log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to upload short %d: %v", i+1, uploadErr)))
									continue
								}

								short.YouTubeID = youtubeID
								uploadedCount++
								fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Short %d uploaded. YouTube ID: %s", i+1, youtubeID)))
							}

							if uploadedCount > 0 {
								fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("\n%d short(s) uploaded successfully.", uploadedCount)))
								fmt.Println(m.orangeStyle.Render("Manual YouTube Studio Action: Set 'Related Video' link for each Short"))
							}
						}
					}
				}

				// --- End of Actions Section (for Publishing Phase) ---

				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save publishing details: %w", err)
				}

				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' publishing details updated and actions processed.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for publishing."))
			}

		case editPhaseDubbing:
			// Handle Dubbing phase - Spanish dubbing workflow for long-form and shorts
			dubbingDone := false

			for !dubbingDone {
				// Action constants: 0 = long-form, 1-999 = shorts dubbing (by index+1),
				// 1000+ = general actions, 2000+ = upload shorts (2000 + index)
				const (
					actionDubbingLongForm    = 0
					actionDubbingCheckStatus = 1000
					actionDubbingBack        = 1001
					actionDubbingTranslate   = 1002
					actionDubbingUploadAll   = 1004
				)

				// Helper to get status text for a dubbing key
				getDubbingStatus := func(key string) string {
					if updatedVideo.Dubbing == nil {
						return "Not started"
					}
					info, ok := updatedVideo.Dubbing[key]
					if !ok || info.DubbingStatus == "" {
						return "Not started"
					}
					switch info.DubbingStatus {
					case "dubbing":
						return "Dubbing..."
					case "dubbed":
						return "Dubbed"
					case "failed":
						return "Failed"
					default:
						return info.DubbingStatus
					}
				}

				// Build status description
				var statusLines []string
				statusLines = append(statusLines, fmt.Sprintf("Long-form: %s", getDubbingStatus("es")))
				for i, short := range updatedVideo.Shorts {
					shortKey := fmt.Sprintf("es:short%d", i+1)
					statusLines = append(statusLines, fmt.Sprintf("Short %d: \"%s\" [%s]", i+1, short.Title, getDubbingStatus(shortKey)))
				}

				var selectedAction int
				var dubbingFormFields []huh.Field

				dubbingFormFields = append(dubbingFormFields,
					huh.NewNote().
						Title("Spanish Dubbing").
						Description(fmt.Sprintf("Video: %s\n\n%s", updatedVideo.Name, strings.Join(statusLines, "\n"))))

				// Build options
				options := []huh.Option[int]{}

				// Helper to determine source type
				getSourceLabel := func(localPath, youtubeID string) string {
					if localPath != "" {
						if _, err := os.Stat(localPath); err == nil {
							return "[Local]"
						}
					}
					if youtubeID != "" {
						return "[YouTube]"
					}
					return "[No source]"
				}

				// Long-form video option
				longFormStatus := getDubbingStatus("es")
				if longFormStatus == "Not started" || longFormStatus == "Failed" {
					sourceLabel := getSourceLabel(updatedVideo.UploadVideo, updatedVideo.VideoId)
					options = append(options, huh.NewOption(fmt.Sprintf("Dub Long-form Video %s", sourceLabel), actionDubbingLongForm))
				}

				// Short options
				for i, short := range updatedVideo.Shorts {
					shortKey := fmt.Sprintf("es:short%d", i+1)
					shortStatus := getDubbingStatus(shortKey)
					if shortStatus == "Not started" || shortStatus == "Failed" {
						sourceLabel := getSourceLabel(short.FilePath, short.YouTubeID)
						label := fmt.Sprintf("Dub Short %d %s: \"%s\"", i+1, sourceLabel, short.Title)
						if len(label) > 60 {
							label = label[:57] + "..."
						}
						options = append(options, huh.NewOption(label, i+1)) // shorts use index+1 as action
					}
				}

				// Check if any dubbing is in progress or needs download retry
				// (dubbed but missing DubbedVideoPath means download failed or was from YouTube URL)
				hasInProgress := false
				if updatedVideo.Dubbing != nil {
					for _, info := range updatedVideo.Dubbing {
						if info.DubbingStatus == "dubbing" ||
							(info.DubbingStatus == "dubbed" && info.DubbedVideoPath == "") {
							hasInProgress = true
							break
						}
					}
				}
				if hasInProgress {
					options = append(options, huh.NewOption("Check Status / Retry Download", actionDubbingCheckStatus))
				}

				// Always show translate option if there's a title to translate
				if updatedVideo.GetUploadTitle() != "" {
					translateLabel := "Translate Metadata"
					// Check if already translated
					if updatedVideo.Dubbing != nil {
						if info, ok := updatedVideo.Dubbing["es"]; ok && info.Title != "" {
							translateLabel = m.greenStyle.Render("Translate Metadata (done)")
						}
					}
					options = append(options, huh.NewOption(translateLabel, actionDubbingTranslate))
				}

				// Show upload option when dubbed items exist
				if updatedVideo.Dubbing != nil {
					// Count uploadable items (dubbed + has file + not yet uploaded)
					// Long-form also requires translated title
					canUploadLongForm := false
					uploadableShortCount := 0
					allUploaded := true

					if info, ok := updatedVideo.Dubbing["es"]; ok {
						canUploadLongForm = info.DubbingStatus == "dubbed" &&
							info.DubbedVideoPath != "" &&
							info.Title != "" &&
							info.UploadedVideoID == ""
						if info.DubbingStatus == "dubbed" && info.UploadedVideoID == "" {
							allUploaded = false
						}
					}

					// Check shorts - no Title requirement (will use original title)
					for i := range updatedVideo.Shorts {
						shortKey := fmt.Sprintf("es:short%d", i+1)
						if shortInfo, ok := updatedVideo.Dubbing[shortKey]; ok {
							if shortInfo.DubbingStatus == "dubbed" &&
								shortInfo.DubbedVideoPath != "" &&
								shortInfo.UploadedVideoID == "" {
								uploadableShortCount++
								allUploaded = false
							}
						}
					}

					totalUploadable := 0
					if canUploadLongForm {
						totalUploadable++
					}
					totalUploadable += uploadableShortCount

					if totalUploadable > 0 {
						uploadAllLabel := fmt.Sprintf("Upload All to YouTube (%d items)", totalUploadable)
						options = append(options, huh.NewOption(uploadAllLabel, actionDubbingUploadAll))
					} else if !allUploaded {
						// Some items dubbed but missing requirements (e.g., long-form missing translated title)
						options = append(options, huh.NewOption("Upload All to YouTube (translate metadata first)", actionDubbingUploadAll))
					} else {
						// Check if anything was ever dubbed and uploaded
						hasDubbedItems := false
						if info, ok := updatedVideo.Dubbing["es"]; ok && info.DubbingStatus == "dubbed" {
							hasDubbedItems = true
						}
						if hasDubbedItems && allUploaded {
							options = append(options, huh.NewOption(m.greenStyle.Render("Upload All to YouTube (done)"), actionDubbingUploadAll))
						}
					}
				}

				options = append(options, huh.NewOption("Back", actionDubbingBack))

				dubbingFormFields = append(dubbingFormFields,
					huh.NewSelect[int]().
						Title("Action").
						Options(options...).
						Value(&selectedAction))

				dubbingForm := huh.NewForm(huh.NewGroup(dubbingFormFields...))

				if err := dubbingForm.Run(); err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						fmt.Println(m.orangeStyle.Render("Dubbing cancelled."))
						dubbingDone = true
						continue
					}
					return fmt.Errorf("error in dubbing form: %w", err)
				}

				// Handle actions
				if selectedAction == actionDubbingBack {
					dubbingDone = true
					continue
				}

				if selectedAction == actionDubbingCheckStatus {
					// Check status for all in-progress jobs
					apiKey := os.Getenv("ELEVENLABS_API_KEY")
					if apiKey == "" {
						fmt.Println(m.errorStyle.Render("ELEVENLABS_API_KEY environment variable not set."))
						continue
					}

					client := dubbing.NewClient(apiKey, dubbing.Config{})
					ctx := context.Background()

					for key, info := range updatedVideo.Dubbing {
						if info.DubbingStatus != "dubbing" {
							continue
						}

						fmt.Println(m.normalStyle.Render(fmt.Sprintf("Checking %s...", key)))

						job, err := client.GetDubbingStatus(ctx, info.DubbingID)
						if err != nil {
							fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to check %s: %v", key, err)))
							continue
						}

						info.DubbingStatus = job.Status
						if job.Status == dubbing.StatusFailed {
							info.DubbingError = job.Error
							fmt.Println(m.errorStyle.Render(fmt.Sprintf("%s failed: %s", key, job.Error)))
						} else if job.Status == dubbing.StatusDubbed {
							fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("%s complete! Downloading...", key)))

							// Determine source file path for output naming
							var sourcePath string
							if key == "es" {
								sourcePath = updatedVideo.UploadVideo
							} else if strings.HasPrefix(key, "es:short") {
								// Extract short index
								shortIdxStr := strings.TrimPrefix(key, "es:short")
								shortIdx, parseErr := strconv.Atoi(shortIdxStr)
								if parseErr != nil {
									fmt.Println(m.errorStyle.Render(fmt.Sprintf("Invalid short key format: %s", key)))
									continue
								}
								if shortIdx > 0 && shortIdx <= len(updatedVideo.Shorts) {
									sourcePath = updatedVideo.Shorts[shortIdx-1].FilePath
								}
							}

							if sourcePath != "" {
								dir := filepath.Dir(sourcePath)
								ext := filepath.Ext(sourcePath)
								base := strings.TrimSuffix(filepath.Base(sourcePath), ext)
								outputPath := filepath.Join(dir, base+"_es"+ext)

								err := client.DownloadDubbedAudio(ctx, info.DubbingID, "es", outputPath)
								if err != nil {
									fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to download %s: %v", key, err)))
								} else {
									info.DubbedVideoPath = outputPath
									fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Downloaded to: %s", outputPath)))
								}
							}
						} else {
							fmt.Println(m.normalStyle.Render(fmt.Sprintf("%s: %s", key, job.Status)))
						}

						updatedVideo.Dubbing[key] = info
					}

					// Save updated statuses
					yaml := storage.YAML{}
					if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to save status: %v", err)))
					}
					videoToEdit = updatedVideo
					continue
				}

				if selectedAction == actionDubbingTranslate {
					// Translate metadata using AI (long-form + shorts)
					fmt.Println(m.normalStyle.Render("Translating metadata to Spanish..."))

					title := updatedVideo.GetUploadTitle()
					if title == "" {
						fmt.Println(m.errorStyle.Render("No title available to translate. Please set a title first."))
						continue
					}

					// Collect short titles for translation
					var shortTitles []string
					for _, short := range updatedVideo.Shorts {
						shortTitles = append(shortTitles, short.Title)
					}

					input := ai.VideoMetadataInput{
						Title:       title,
						Description: updatedVideo.Description,
						Tags:        updatedVideo.Tags,
						Timecodes:   updatedVideo.Timecodes,
						ShortTitles: shortTitles,
					}

					ctx := context.Background()
					output, err := ai.TranslateVideoMetadata(ctx, input, "Spanish")
					if err != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Translation failed: %v", err)))
						continue
					}

					// Save translated fields to DubbingInfo
					if updatedVideo.Dubbing == nil {
						updatedVideo.Dubbing = make(map[string]storage.DubbingInfo)
					}

					// Save long-form translations
					info := updatedVideo.Dubbing["es"]
					info.Title = output.Title
					info.Description = output.Description
					info.Tags = output.Tags
					info.Timecodes = output.Timecodes
					updatedVideo.Dubbing["es"] = info

					// Save translated short titles
					for i, translatedTitle := range output.ShortTitles {
						if i >= len(updatedVideo.Shorts) {
							break
						}
						shortKey := fmt.Sprintf("es:short%d", i+1)
						shortInfo := updatedVideo.Dubbing[shortKey]
						shortInfo.Title = translatedTitle
						updatedVideo.Dubbing[shortKey] = shortInfo
					}

					// Save to YAML
					yaml := storage.YAML{}
					if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to save translations: %v", err)))
						continue
					}

					fmt.Println(m.confirmationStyle.Render("Translation complete!"))
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Title: %s", output.Title)))
					if output.Description != "" {
						// Show first 100 chars of description
						descPreview := output.Description
						if len(descPreview) > 100 {
							descPreview = descPreview[:100] + "..."
						}
						fmt.Println(m.normalStyle.Render(fmt.Sprintf("Description: %s", descPreview)))
					}
					if output.Tags != "" {
						fmt.Println(m.normalStyle.Render(fmt.Sprintf("Tags: %s", output.Tags)))
					}
					if output.Timecodes != "" {
						fmt.Println(m.normalStyle.Render("Timecodes: translated"))
					}
					if len(output.ShortTitles) > 0 {
						fmt.Println(m.normalStyle.Render(fmt.Sprintf("Short titles: %d translated", len(output.ShortTitles))))
						for i, st := range output.ShortTitles {
							fmt.Println(m.normalStyle.Render(fmt.Sprintf("  Short %d: %s", i+1, st)))
						}
					}

					videoToEdit = updatedVideo
					continue
				}

				if selectedAction == actionDubbingUploadAll {
					// Upload all dubbed videos (long-form + shorts) sequentially
					fmt.Println(m.normalStyle.Render("Uploading all dubbed videos to YouTube..."))
					fmt.Println()

					uploadCount := 0
					failCount := 0

					// Upload long-form if ready
					if info, ok := updatedVideo.Dubbing["es"]; ok {
						canUpload := info.DubbingStatus == "dubbed" &&
							info.DubbedVideoPath != "" &&
							info.Title != "" &&
							info.UploadedVideoID == ""
						if canUpload {
							fmt.Println(m.normalStyle.Render("Uploading long-form video..."))
							videoID, err := publishing.UploadDubbedVideo(&updatedVideo, "es")
							if err != nil {
								fmt.Println(m.errorStyle.Render(fmt.Sprintf("  Failed: %v", err)))
								failCount++
							} else {
								info.UploadedVideoID = videoID
								updatedVideo.Dubbing["es"] = info
								fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("  Done! ID: %s", videoID)))
								uploadCount++
							}
						}
					}

					// Upload shorts if ready (no Title requirement - will use original title)
					for i := range updatedVideo.Shorts {
						shortKey := fmt.Sprintf("es:short%d", i+1)
						if shortInfo, ok := updatedVideo.Dubbing[shortKey]; ok {
							canUpload := shortInfo.DubbingStatus == "dubbed" &&
								shortInfo.DubbedVideoPath != "" &&
								shortInfo.UploadedVideoID == ""
							if canUpload {
								fmt.Println(m.normalStyle.Render(fmt.Sprintf("Uploading short %d...", i+1)))
								videoID, err := publishing.UploadDubbedShort(&updatedVideo, i)
								if err != nil {
									fmt.Println(m.errorStyle.Render(fmt.Sprintf("  Failed: %v", err)))
									failCount++
								} else {
									shortInfo.UploadedVideoID = videoID
									updatedVideo.Dubbing[shortKey] = shortInfo
									fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("  Done! ID: %s", videoID)))
									uploadCount++
								}
							}
						}
					}

					// Save all changes to YAML
					yaml := storage.YAML{}
					if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to save upload info: %v", err)))
					}

					fmt.Println()
					if failCount == 0 {
						fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("All uploads complete! %d videos uploaded.", uploadCount)))
					} else {
						fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Uploads finished: %d succeeded, %d failed.", uploadCount, failCount)))
					}

					videoToEdit = updatedVideo
					continue
				}

				// Start dubbing for long-form or a short
				var youtubeID string
				var localFilePath string
				var dubbingKey string

				if selectedAction == actionDubbingLongForm {
					youtubeID = updatedVideo.VideoId
					localFilePath = updatedVideo.UploadVideo
					dubbingKey = "es"
				} else if selectedAction >= 1 && selectedAction <= len(updatedVideo.Shorts) {
					shortIdx := selectedAction - 1
					youtubeID = updatedVideo.Shorts[shortIdx].YouTubeID
					localFilePath = updatedVideo.Shorts[shortIdx].FilePath
					dubbingKey = fmt.Sprintf("es:short%d", selectedAction)
				} else {
					continue
				}

				// Check if local file exists
				useLocalFile := false
				if localFilePath != "" {
					if _, err := os.Stat(localFilePath); err == nil {
						useLocalFile = true
					}
				}

				// Validate we have either local file or YouTube ID
				if !useLocalFile && youtubeID == "" {
					fmt.Println(m.errorStyle.Render("No local video file found and video not published on YouTube. Please provide a local file path or publish to YouTube first."))
					continue
				}

				// Get API key and start dubbing
				apiKey := os.Getenv("ELEVENLABS_API_KEY")
				if apiKey == "" {
					fmt.Println(m.errorStyle.Render("ELEVENLABS_API_KEY environment variable not set."))
					continue
				}

				dubbingConfig := dubbing.Config{
					TestMode:            m.settings.ElevenLabs.TestMode,
					StartTime:           m.settings.ElevenLabs.StartTime,
					EndTime:             m.settings.ElevenLabs.EndTime,
					NumSpeakers:         m.settings.ElevenLabs.NumSpeakers,
					DropBackgroundAudio: m.settings.ElevenLabs.DropBackgroundAudio,
				}
				client := dubbing.NewClient(apiKey, dubbingConfig)

				ctx := context.Background()
				var job *dubbing.DubbingJob
				var err error

				if useLocalFile {
					// Try local file first (with auto-compression for files >1GB)
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Dubbing %s from local file:", dubbingKey)))
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("  %s", localFilePath)))

					// Check if compression will be needed
					needsCompression, compErr := dubbing.NeedsCompression(localFilePath)
					if compErr != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to check file size: %v", compErr)))
						continue
					}
					if needsCompression {
						fmt.Println(m.normalStyle.Render("Step 1/2: Compressing video (file >1GB)... this may take a few minutes"))
					} else {
						fmt.Println(m.normalStyle.Render("Step 1/1: Uploading to ElevenLabs..."))
					}

					job, err = client.CreateDubFromFile(ctx, localFilePath, "en", "es")

					if err == nil && needsCompression {
						fmt.Println(m.confirmationStyle.Render("Compression complete, upload finished."))
					}
				} else {
					// Fall back to YouTube URL
					youtubeURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", youtubeID)
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Starting dubbing for %s from YouTube...", dubbingKey)))
					job, err = client.CreateDubFromURL(ctx, youtubeURL, "en", "es")
				}

				if err != nil {
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to start dubbing: %v", err)))
					continue
				}

				// Initialize dubbing map if nil
				if updatedVideo.Dubbing == nil {
					updatedVideo.Dubbing = make(map[string]storage.DubbingInfo)
				}

				// Store dubbing info
				updatedVideo.Dubbing[dubbingKey] = storage.DubbingInfo{
					DubbingID:     job.DubbingID,
					DubbingStatus: "dubbing",
				}

				// Save immediately
				yaml := storage.YAML{}
				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to save dubbing info: %v", err)))
					continue
				}

				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Dubbing started! Job ID: %s", job.DubbingID)))
				if job.ExpectedDuration > 0 {
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Expected duration: %.0f seconds", job.ExpectedDuration)))
				}
				videoToEdit = updatedVideo
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
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleDOTPosted, updatedVideo.DOTPosted)).Value(&updatedVideo.DOTPosted),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleBlueSkyPosted, updatedVideo.BlueSkyPosted)).Value(&updatedVideo.BlueSkyPosted),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleLinkedInPosted, updatedVideo.LinkedInPosted)).Value(&updatedVideo.LinkedInPosted),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleSlackPosted, updatedVideo.SlackPosted)).Value(&updatedVideo.SlackPosted),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleYouTubeHighlight, updatedVideo.YouTubeHighlight)).Value(&updatedVideo.YouTubeHighlight),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleYouTubeComment, updatedVideo.YouTubeComment)).Value(&updatedVideo.YouTubeComment),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleYouTubeCommentReply, updatedVideo.YouTubeCommentReply)).Value(&updatedVideo.YouTubeCommentReply),
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleGDEPosted, updatedVideo.GDE)).Value(&updatedVideo.GDE),
				huh.NewInput().Title(m.colorTitleString(constants.FieldTitleCodeRepository, updatedVideo.Repo)).Value(&updatedVideo.Repo),
				huh.NewConfirm().Title(sponsorsNotifyText).Value(&updatedVideo.NotifiedSponsors),
				huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(&save),
			}

			postPublishingForm := huh.NewForm(
				huh.NewGroup(postPublishingFormFields...),
			)
			err = postPublishingForm.Run()

			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(m.orangeStyle.Render("Post-Publish details editing cancelled."))
					continue
				}
				log.Print(m.errorStyle.Render(fmt.Sprintf("Error running post-publish details form: %v", err)))
				return err
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
						emailService.SendSponsors(configuration.GlobalSettings.Email.From, updatedVideo.Sponsorship.Emails, updatedVideo.VideoId, updatedVideo.Sponsorship.Amount, updatedVideo.GetUploadTitle())
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
					if errSl := slack.LoadAndValidateSlackConfig(""); errSl != nil {
						log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to load Slack configuration: %v", errSl)))
						updatedVideo.SlackPosted = false
						return fmt.Errorf("failed to load Slack configuration: %w", errSl)
					} else {
						slackService, errSlSvc := slack.NewSlackService(slack.GlobalSlackConfig)
						if errSlSvc != nil {
							log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to create Slack service: %v", errSlSvc)))
							updatedVideo.SlackPosted = false
							return fmt.Errorf("failed to create Slack service: %w", errSlSvc)
						} else {
							errSlPost := slackService.PostVideo(&updatedVideo, updatedVideo.Path)
							if errSlPost != nil {
								log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to post video to Slack: %v", errSlPost)))
								updatedVideo.SlackPosted = false
								return fmt.Errorf("failed to post video to Slack: %w", errSlPost)
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
						log.Print(m.errorStyle.Render("BlueSky credentials not configured. Cannot post to BlueSky."))
						updatedVideo.BlueSkyPosted = false // Revert intent
					} else {
						bsConfig := bluesky.Config{
							Identifier: configuration.GlobalSettings.Bluesky.Identifier,
							Password:   configuration.GlobalSettings.Bluesky.Password,
							URL:        configuration.GlobalSettings.Bluesky.URL,
						}
						if bsErr := bluesky.SendPost(bsConfig, updatedVideo.Tweet, updatedVideo.VideoId, updatedVideo.Thumbnail); bsErr != nil {
							log.Print(m.errorStyle.Render(fmt.Sprintf("Failed to post to BlueSky: %v", bsErr)))
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
					log.Println(m.orangeStyle.Render(fmt.Sprintf("TODO: Implement repository update for %s with title %s, videoId %s", updatedVideo.Repo, updatedVideo.GetUploadTitle(), updatedVideo.VideoId)))
				} else if updatedVideo.Repo != originalRepo && (updatedVideo.Repo == "" || updatedVideo.Repo == "N/A") { // User cleared repo or set to N/A
					// Just save the cleared/N/A state, no specific action beyond that.
				}

				// --- End of Actions Section for Post-Publish ---

				if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
					return fmt.Errorf("failed to save post-publish details: %w", err)
				}
				fmt.Println(m.confirmationStyle.Render(fmt.Sprintf("Video '%s' post-publish details updated.", updatedVideo.Name)))
				videoToEdit = updatedVideo // Persist changes for the next loop iteration
			} else {
				fmt.Println(m.orangeStyle.Render("Changes not saved for post-publish details."))
			}

		case editPhaseAnalysis:
			// Handle Analysis phase - edit title A/B test share percentages
			if len(updatedVideo.Titles) == 0 {
				fmt.Println(m.errorStyle.Render("No titles found. Please add titles in the Definition phase first."))
				continue
			}

			// Create string variables for form input (huh.NewInput requires *string)
			share1Str, share2Str, share3Str := "", "", ""
			title1Text, title2Text, title3Text := "", "", ""

			// Load current values and convert to strings
			for _, t := range updatedVideo.Titles {
				switch t.Index {
				case 1:
					title1Text = t.Text
					if t.Share > 0 {
						share1Str = fmt.Sprintf("%.2f", t.Share)
					}
				case 2:
					title2Text = t.Text
					if t.Share > 0 {
						share2Str = fmt.Sprintf("%.2f", t.Share)
					}
				case 3:
					title3Text = t.Text
					if t.Share > 0 {
						share3Str = fmt.Sprintf("%.2f", t.Share)
					}
				}
			}

			// Build form fields dynamically based on which titles exist
			var formFields []huh.Field
			formFields = append(formFields, huh.NewNote().
				Title("Title A/B Test Results").
				Description("Enter watch time share percentages from YouTube Studio A/B test results"))

			if title1Text != "" {
				formFields = append(formFields,
					huh.NewNote().Description(fmt.Sprintf("Title 1: %s", title1Text)),
					huh.NewInput().
						Title("Watch Time Share % (Title 1)").
						Value(&share1Str).
						Placeholder("0.0"),
				)
			}

			if title2Text != "" {
				formFields = append(formFields,
					huh.NewNote().Description(fmt.Sprintf("Title 2: %s", title2Text)),
					huh.NewInput().
						Title("Watch Time Share % (Title 2)").
						Value(&share2Str).
						Placeholder("0.0"),
				)
			}

			if title3Text != "" {
				formFields = append(formFields,
					huh.NewNote().Description(fmt.Sprintf("Title 3: %s", title3Text)),
					huh.NewInput().
						Title("Watch Time Share % (Title 3)").
						Value(&share3Str).
						Placeholder("0.0"),
				)
			}

			analysisForm := huh.NewForm(huh.NewGroup(formFields...))

			if err := analysisForm.Run(); err != nil {
				if err == huh.ErrUserAborted {
					fmt.Println(m.orangeStyle.Render("Analysis editing cancelled."))
					continue
				}
				return fmt.Errorf("failed to run analysis form: %w", err)
			}

			// Parse string inputs to float64 and update the title shares
			for i := range updatedVideo.Titles {
				var shareValue float64
				var parseErr error

				switch updatedVideo.Titles[i].Index {
				case 1:
					if share1Str != "" {
						shareValue, parseErr = strconv.ParseFloat(share1Str, 64)
						if parseErr != nil {
							fmt.Println(m.errorStyle.Render(fmt.Sprintf("Invalid share value for Title 1: %s", share1Str)))
							continue
						}
						updatedVideo.Titles[i].Share = shareValue
					}
				case 2:
					if share2Str != "" {
						shareValue, parseErr = strconv.ParseFloat(share2Str, 64)
						if parseErr != nil {
							fmt.Println(m.errorStyle.Render(fmt.Sprintf("Invalid share value for Title 2: %s", share2Str)))
							continue
						}
						updatedVideo.Titles[i].Share = shareValue
					}
				case 3:
					if share3Str != "" {
						shareValue, parseErr = strconv.ParseFloat(share3Str, 64)
						if parseErr != nil {
							fmt.Println(m.errorStyle.Render(fmt.Sprintf("Invalid share value for Title 3: %s", share3Str)))
							continue
						}
						updatedVideo.Titles[i].Share = shareValue
					}
				}
			}

			// Save the video
			yaml := storage.YAML{}
			if err := yaml.WriteVideo(updatedVideo, updatedVideo.Path); err != nil {
				return fmt.Errorf("failed to save analysis data: %w", err)
			}

			fmt.Println(m.greenStyle.Render("✓ Analysis data saved successfully"))
			videoToEdit = updatedVideo // Update the local copy for the next iteration

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
