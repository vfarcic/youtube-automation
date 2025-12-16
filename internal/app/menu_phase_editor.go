package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/calendar"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/constants"
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
		postPublishCompleted, postPublishTotal := m.videoManager.CalculatePostPublishProgress(videoToEdit)
		analysisCompleted, analysisTotal := m.videoManager.CalculateAnalysisProgress(videoToEdit)

		editPhaseOptions := []huh.Option[int]{
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleInitialDetails, initCompleted, initTotal), editPhaseInitial),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleWorkProgress, workCompleted, workTotal), editPhaseWork),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitleDefinition, defineCompleted, defineTotal), editPhaseDefinition),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitlePostProduction, editCompleted, editTotal), editPhasePostProduction),
			huh.NewOption(m.getEditPhaseOptionText(constants.PhaseTitlePublishingDetails, publishCompleted, publishTotal), editPhasePublishing),
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
			// Store original values to detect changes for actions
			originalHugoPath := updatedVideo.HugoPath
			// If VideoId is empty, createHugo will be false, also influencing the title color.
			createHugo := updatedVideo.HugoPath != "" && updatedVideo.VideoId != ""

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
				huh.NewConfirm().Title(m.colorTitleBool(constants.FieldTitleCreateHugo, createHugo)).Value(&createHugo),
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
					createdPath, hugoErr := hugoPublisher.Post(updatedVideo.Gist, updatedVideo.Title, updatedVideo.Date, updatedVideo.VideoId)
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

				// --- End of Actions Section (for Publishing Phase) ---

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
					log.Println(m.orangeStyle.Render(fmt.Sprintf("TODO: Implement repository update for %s with title %s, videoId %s", updatedVideo.Repo, updatedVideo.Title, updatedVideo.VideoId)))
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

	return nil
}
