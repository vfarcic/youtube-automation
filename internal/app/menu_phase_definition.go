package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/cli"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/constants"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/notification"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/charmbracelet/huh"
)

func (m *MenuHandler) editPhaseDefinition(videoToEdit storage.Video, settings configuration.Settings) (storage.Video, error) {
	fmt.Println(m.normalStyle.Render("\n--- Defining Video Details ---"))
	originalVideoForThisCall := videoToEdit // Snapshot for early abort
	yamlHelper := storage.YAML{}

	definitionFields := []struct {
		name                   string
		description            string
		isTitleField           bool
		isDescriptionField     bool
		isThumbnailField       bool
		isTagsField            bool
		isDescriptionTagsField bool
		isTweetField           bool
		isAnimationsField      bool // New field for Animations specific logic
		getValue               func() interface{}
		updateAction           func(newValue interface{})
		revertField            func(originalValue interface{})
	}{
		{
			name: "Titles", description: "Video titles for A/B testing (max 70 chars each). First title is uploaded to YouTube.", isTitleField: true,
			getValue: func() interface{} {
				// Return array of titles (not used directly, but keeps interface consistent)
				return videoToEdit.Titles
			},
			updateAction: func(newValue interface{}) {
				// Will be handled in the title field logic
			},
			revertField: func(originalValue interface{}) {
				videoToEdit.Titles = originalValue.([]storage.TitleVariant)
			},
		},
		{
			name: "Description", description: "Video description (max 5000 chars). Include keywords.", isDescriptionField: true,
			getValue:     func() interface{} { return videoToEdit.Description },
			updateAction: func(newValue interface{}) { videoToEdit.Description = newValue.(string) },
			revertField:  func(originalValue interface{}) { videoToEdit.Description = originalValue.(string) },
		},
		{
			name: "Tags", description: "Comma-separated tags (max 15 tags, 50 chars/tag, 450 total). e.g., golang,devops,tutorial.", isTagsField: true,
			getValue:     func() interface{} { return videoToEdit.Tags },
			updateAction: func(newValue interface{}) { videoToEdit.Tags = newValue.(string) },
			revertField:  func(originalValue interface{}) { videoToEdit.Tags = originalValue.(string) },
		},
		{
			name: "Description Tags", description: "Exactly three space-separated tags, each starting with # (e.g., #golang #devops #tutorial).", isDescriptionTagsField: true, // Updated name and set new flag
			getValue:     func() interface{} { return videoToEdit.DescriptionTags },
			updateAction: func(newValue interface{}) { videoToEdit.DescriptionTags = newValue.(string) },
			revertField: func(originalValue interface{}) {
				videoToEdit.DescriptionTags = originalValue.(string)
			},
		},
		{
			name: "Tweet", description: "Promotional tweet text (max 280 chars). Include [YOUTUBE] placeholder.", isTweetField: true, // Updated for AI
			getValue:     func() interface{} { return videoToEdit.Tweet },
			updateAction: func(newValue interface{}) { videoToEdit.Tweet = newValue.(string) },
			revertField:  func(originalValue interface{}) { videoToEdit.Tweet = originalValue.(string) },
		},
		{
			name: "Animations", description: "Script for animations, one per line, starting with '-'.", isAnimationsField: true, // Mark as Animations field
			getValue:     func() interface{} { return videoToEdit.Animations },
			updateAction: func(newValue interface{}) { videoToEdit.Animations = newValue.(string) },
			revertField:  func(originalValue interface{}) { videoToEdit.Animations = originalValue.(string) },
		},
		{
			name: "RequestThumbnail", description: "Request a custom thumbnail from the designer?", isThumbnailField: true,
			getValue:     func() interface{} { return videoToEdit.RequestThumbnail },
			updateAction: func(newValue interface{}) { videoToEdit.RequestThumbnail = newValue.(bool) },
			revertField:  func(originalValue interface{}) { videoToEdit.RequestThumbnail = originalValue.(bool) },
		},
		{
			name: "Gist", description: "Path to Gist/Markdown file for manuscript (relative to execution path).",
			getValue:     func() interface{} { return videoToEdit.Gist },
			updateAction: func(newValue interface{}) { videoToEdit.Gist = newValue.(string) },
			revertField:  func(originalValue interface{}) { videoToEdit.Gist = originalValue.(string) },
		},
	}

	const (
		generalActionSave    = 0
		generalActionAskAI   = 1
		generalActionSkip    = 2
		generalActionUnknown = -1
	)

	for fieldIdx, df := range definitionFields {
		originalFieldValue := df.getValue()
		initialRequestThumbnailForThisField := videoToEdit.RequestThumbnail
		var formError error

		if df.isTitleField {
			fieldSavedOrSkipped := false

			// Initialize title values once before the loop (with fallback to legacy Title field)
			title1, title2, title3 := "", "", ""
			if len(videoToEdit.Titles) > 0 {
				for _, t := range videoToEdit.Titles {
					switch t.Index {
					case 1:
						title1 = t.Text
					case 2:
						title2 = t.Text
					case 3:
						title3 = t.Text
					}
				}
			}

			for !fieldSavedOrSkipped {

				var selectedAction int = generalActionUnknown
				titleForm := huh.NewForm(
					huh.NewGroup(
						huh.NewNote().
							Title("Titles").
							Description(df.description),
						huh.NewInput().
							Title("✓ Title 1 (Uploaded to YouTube)").
							Value(&title1).
							CharLimit(70),
						huh.NewInput().
							Title("  Title 2 (A/B Test Variant - optional)").
							Value(&title2).
							CharLimit(70),
						huh.NewInput().
							Title("  Title 3 (A/B Test Variant - optional)").
							Value(&title3).
							CharLimit(70),
						huh.NewSelect[int]().
							Title("Action for Titles").
							Options(
								huh.NewOption("Save Titles & Continue", generalActionSave),
								huh.NewOption("Ask AI for Suggestions", generalActionAskAI),
								huh.NewOption("Continue Without Saving Titles", generalActionSkip),
							).
							Value(&selectedAction),
					),
				)

				formError = titleForm.Run()
				if formError != nil {
					if formError == huh.ErrUserAborted {
						fmt.Println(m.orangeStyle.Render("Action for 'Titles' aborted by user."))
						df.revertField(originalFieldValue)
						if fieldIdx == 0 {
							fmt.Println(m.normalStyle.Render(MessageDefinitionPhaseAborted))
							return originalVideoForThisCall, nil
						}
						fieldSavedOrSkipped = true
						continue
					}
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error in titles form: %v", formError)))
					return videoToEdit, formError
				}

				switch selectedAction {
				case generalActionSave:
					// Create a map of existing titles to preserve Share percentages
					existingShares := make(map[string]float64)
					for _, t := range originalFieldValue.([]storage.TitleVariant) {
						existingShares[t.Text] = t.Share
					}

					// Update Titles array, preserving Share percentages for matching titles
					videoToEdit.Titles = []storage.TitleVariant{}
					if title1 != "" {
						share := existingShares[title1] // Will be 0 if not found
						videoToEdit.Titles = append(videoToEdit.Titles, storage.TitleVariant{Index: 1, Text: title1, Share: share})
					}
					if title2 != "" {
						share := existingShares[title2]
						videoToEdit.Titles = append(videoToEdit.Titles, storage.TitleVariant{Index: 2, Text: title2, Share: share})
					}
					if title3 != "" {
						share := existingShares[title3]
						videoToEdit.Titles = append(videoToEdit.Titles, storage.TitleVariant{Index: 3, Text: title3, Share: share})
					}
					saveErr := yamlHelper.WriteVideo(videoToEdit, videoToEdit.Path)
					if saveErr != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error saving titles: %v", saveErr)))
						df.revertField(originalFieldValue)
						return videoToEdit, saveErr
					}
					fieldSavedOrSkipped = true
				case generalActionAskAI:
					fmt.Println(m.normalStyle.Render("Attempting to get AI title suggestions..."))
					if videoToEdit.Gist == "" {
						fmt.Fprintf(os.Stderr, "Manuscript/Gist path is not defined. Cannot fetch content for AI.\n")
					} else {
						aiConfig, cfgErr := ai.GetAIConfig()
						if cfgErr != nil {
							fmt.Fprintf(os.Stderr, "Error getting AI config: %v\n", cfgErr)
						} else {
							manuscriptContent, readErr := m.videoService.GetVideoManuscript(videoToEdit.Name, videoToEdit.Category)
							if readErr != nil {
								fmt.Fprintf(os.Stderr, "Error reading manuscript: %v\n", readErr)
							} else {
								suggestedTitles, suggErr := ai.SuggestTitles(context.Background(), manuscriptContent, aiConfig)
								if suggErr != nil {
									fmt.Fprintf(os.Stderr, "Error suggesting titles: %v\n", suggErr)
								} else if len(suggestedTitles) > 0 {
									// Multi-select 1-3 titles
									var selectedTitles []string
									options := []huh.Option[string]{}
									for _, sTitle := range suggestedTitles {
										options = append(options, huh.NewOption(sTitle, sTitle))
									}
									multiSelectForm := huh.NewForm(
										huh.NewGroup(
											huh.NewMultiSelect[string]().
												Title("Select 1-3 titles for A/B testing (first = uploaded to YouTube)").
												Description("Use space to select, enter to confirm").
												Options(options...).
												Value(&selectedTitles).
												Limit(3),
										),
									)
									aiSelectErr := multiSelectForm.Run()
									if aiSelectErr == nil && len(selectedTitles) > 0 {
										// Update title variables based on selection order
										title1 = ""
										title2 = ""
										title3 = ""
										for i, title := range selectedTitles {
											switch i {
											case 0:
												title1 = title
												fmt.Println(m.normalStyle.Render(fmt.Sprintf("✓ Title 1 (Uploaded): %s", title)))
											case 1:
												title2 = title
												fmt.Println(m.normalStyle.Render(fmt.Sprintf("  Title 2 (Variant): %s", title)))
											case 2:
												title3 = title
												fmt.Println(m.normalStyle.Render(fmt.Sprintf("  Title 3 (Variant): %s", title)))
											}
										}
										// Continue to show the form again with new values
									} else if aiSelectErr != nil && aiSelectErr != huh.ErrUserAborted {
										fmt.Fprintf(os.Stderr, "Error during AI title selection: %v\n", aiSelectErr)
									}
								} else {
									fmt.Println(m.normalStyle.Render("AI did not return any title suggestions."))
								}
							}
						}
					}
				case generalActionSkip:
					df.revertField(originalFieldValue)
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Skipped changes for '%s'.", df.name)))
					fieldSavedOrSkipped = true
				default:
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Unknown action for title field: %d", selectedAction)))
					fieldSavedOrSkipped = true
				}
			}
		} else if df.isDescriptionField {
			tempDescriptionValue := originalFieldValue.(string)
			fieldSavedOrSkipped := false
			for !fieldSavedOrSkipped {
				var selectedAction int = generalActionUnknown
				descriptionFieldItself := huh.NewText().Title(m.colorTitleString(constants.FieldTitleDescription, tempDescriptionValue)).Description(df.description).Lines(7).CharLimit(5000).Value(&tempDescriptionValue) // Ensure Lines(7)
				actionSelect := huh.NewSelect[int]().Title("Action for Description").Options(
					huh.NewOption("Save Description & Continue", generalActionSave),
					huh.NewOption("Ask AI for Suggestion", generalActionAskAI),
					huh.NewOption("Continue Without Saving Description", generalActionSkip),
				).Value(&selectedAction)
				descriptionGroup := huh.NewGroup(descriptionFieldItself, actionSelect)
				descriptionForm := huh.NewForm(descriptionGroup)
				formError = descriptionForm.Run()
				if formError != nil {
					if formError == huh.ErrUserAborted {
						fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Action for '%s' aborted by user.", df.name)))
						df.revertField(originalFieldValue)
						if fieldIdx == 0 {
							fmt.Println(m.normalStyle.Render(MessageDefinitionPhaseAborted))
							return originalVideoForThisCall, nil
						}
						fieldSavedOrSkipped = true
						continue
					}
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error in description form: %v", formError)))
					return videoToEdit, formError
				}
				switch selectedAction {
				case generalActionSave:
					df.updateAction(tempDescriptionValue)
					saveErr := yamlHelper.WriteVideo(videoToEdit, videoToEdit.Path) // Renamed err to saveErr
					if saveErr != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error saving changes for '%s': %v", df.name, saveErr)))
						df.revertField(originalFieldValue)
						return videoToEdit, saveErr
					}
					fieldSavedOrSkipped = true
				case generalActionAskAI:
					fmt.Println(m.normalStyle.Render("Attempting to get AI description suggestion..."))
					if videoToEdit.Gist == "" {
						fmt.Fprintf(os.Stderr, "Manuscript/Gist path is not defined. Cannot fetch content for AI.\n")
					} else {
						aiConfig, cfgErr := ai.GetAIConfig()
						if cfgErr != nil {
							fmt.Fprintf(os.Stderr, "Error getting AI config: %v\n", cfgErr)
						} else {
							manuscriptContent, readErr := m.videoService.GetVideoManuscript(videoToEdit.Name, videoToEdit.Category)
							if readErr != nil {
								fmt.Fprintf(os.Stderr, "Error reading manuscript: %v\n", readErr)
							} else {
								suggestedDescription, suggErr := ai.SuggestDescription(context.Background(), manuscriptContent, aiConfig)
								if suggErr != nil {
									fmt.Fprintf(os.Stderr, "Error suggesting description: %v\n", suggErr)
								} else if suggestedDescription != "" {
									fmt.Println(m.normalStyle.Render("AI suggested description received."))
									tempDescriptionValue = suggestedDescription
								} else {
									fmt.Println(m.normalStyle.Render("AI did not return any description suggestion."))
								}
							}
						}
					}
				case generalActionSkip:
					df.revertField(originalFieldValue)
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Skipped changes for '%s'.", df.name)))
					fieldSavedOrSkipped = true
				default:
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Unknown action for description field: %d", selectedAction)))
					fieldSavedOrSkipped = true
				}
			}
		} else if df.isTagsField { // Existing block for Tags field
			tempTagsValue := originalFieldValue.(string)
			fieldSavedOrSkipped := false

			for !fieldSavedOrSkipped {
				var selectedAction int = generalActionUnknown

				tagsFieldItself := huh.NewText().Title(m.colorTitleString(constants.FieldTitleTags, tempTagsValue)).Description(df.description).Lines(3).CharLimit(450).Value(&tempTagsValue) // Set Lines(3)
				actionSelect := huh.NewSelect[int]().Title("Action for Tags").Options(
					huh.NewOption("Save Tags & Continue", generalActionSave),
					huh.NewOption("Ask AI for Suggestion", generalActionAskAI),
					huh.NewOption("Continue Without Saving Tags", generalActionSkip),
				).Value(&selectedAction)

				tagsGroup := huh.NewGroup(tagsFieldItself, actionSelect)
				tagsForm := huh.NewForm(tagsGroup)
				formError = tagsForm.Run()

				if formError != nil {
					if formError == huh.ErrUserAborted {
						fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Action for '%s' aborted by user.", df.name)))
						df.revertField(originalFieldValue)
						if fieldIdx == 0 { // If first field, aborting means exiting this phase
							fmt.Println(m.normalStyle.Render(MessageDefinitionPhaseAborted))
							return originalVideoForThisCall, nil
						}
						fieldSavedOrSkipped = true // Mark as skipped to exit inner loop and go to next field
						continue                   // Continue the outer loop (next field)
					}
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error in tags form: %v", formError)))
					return videoToEdit, formError // Critical error, exit function
				}

				switch selectedAction {
				case generalActionSave:
					df.updateAction(tempTagsValue)
					saveErr := yamlHelper.WriteVideo(videoToEdit, videoToEdit.Path) // Renamed err to saveErr
					if saveErr != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error saving changes for '%s': %v", df.name, saveErr)))
						df.revertField(originalFieldValue) // Revert on save error
						return videoToEdit, saveErr        // Critical error
					}
					fieldSavedOrSkipped = true
				case generalActionAskAI:
					fmt.Println(m.normalStyle.Render("Attempting to get AI tags suggestion..."))
					if videoToEdit.Gist == "" {
						fmt.Fprintf(os.Stderr, "Manuscript/Gist path is not defined. Cannot fetch content for AI tags.\n")
					} else {
						aiConfig, cfgErr := ai.GetAIConfig()
						if cfgErr != nil {
							fmt.Fprintf(os.Stderr, "Error getting AI config: %v\n", cfgErr)
						} else {
							manuscriptPath := videoToEdit.Gist
							manuscriptContent, readErr := os.ReadFile(manuscriptPath)
							if readErr != nil {
								fmt.Fprintf(os.Stderr, "Error reading manuscript file %s: %v\n", manuscriptPath, readErr)
							} else {
								suggestedTagsString, suggErr := ai.SuggestTags(context.Background(), string(manuscriptContent), aiConfig)
								if suggErr != nil {
									fmt.Fprintf(os.Stderr, "Error suggesting tags: %v\n", suggErr)
								} else if suggestedTagsString != "" {
									fmt.Println(m.normalStyle.Render("AI suggested tags received."))
									tempTagsValue = suggestedTagsString // Update the temp value to show in the input field
								} else {
									fmt.Println(m.normalStyle.Render("AI did not return any tags suggestion."))
								}
							}
						}
					}
					// Loop continues to show the form again with the new tempTagsValue
				case generalActionSkip:
					df.revertField(originalFieldValue)
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Skipped changes for '%s'.", df.name)))
					fieldSavedOrSkipped = true
				default:
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Unknown action for tags field: %d", selectedAction)))
					fieldSavedOrSkipped = true // Treat as skip
				}
			}
		} else if df.isDescriptionTagsField { // New block for Description Tags field
			tempDescTagsValue := originalFieldValue.(string)
			fieldSavedOrSkipped := false

			for !fieldSavedOrSkipped {
				var selectedAction int = generalActionUnknown

				descTagsFieldItself := huh.NewText().Title(m.colorTitleString(constants.FieldTitleDescriptionTags, tempDescTagsValue)).Description(df.description).Lines(2).CharLimit(0).Value(&tempDescTagsValue) // Set Lines(2)
				actionSelect := huh.NewSelect[int]().Title("Action for Description Tags").Options(
					huh.NewOption("Save Description Tags & Continue", generalActionSave),
					huh.NewOption("Ask AI for Suggestion", generalActionAskAI),
					huh.NewOption("Continue Without Saving Description Tags", generalActionSkip),
				).Value(&selectedAction)

				descTagsGroup := huh.NewGroup(descTagsFieldItself, actionSelect)
				descTagsForm := huh.NewForm(descTagsGroup)
				formError = descTagsForm.Run()

				if formError != nil {
					if formError == huh.ErrUserAborted {
						fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Action for '%s' aborted by user.", df.name)))
						df.revertField(originalFieldValue)
						if fieldIdx == 0 { // If first field, aborting means exiting this phase
							fmt.Println(m.normalStyle.Render(MessageDefinitionPhaseAborted))
							return originalVideoForThisCall, nil
						}
						fieldSavedOrSkipped = true // Mark as skipped to exit inner loop and go to next field
						continue                   // Continue the outer loop (next field)
					}
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error in description tags form: %v", formError)))
					return videoToEdit, formError // Critical error, exit function
				}

				switch selectedAction {
				case generalActionSave:
					df.updateAction(tempDescTagsValue)
					saveErr := yamlHelper.WriteVideo(videoToEdit, videoToEdit.Path) // Renamed err to saveErr
					if saveErr != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error saving changes for '%s': %v", df.name, saveErr)))
						df.revertField(originalFieldValue) // Revert on save error
						return videoToEdit, saveErr        // Critical error
					}
					fieldSavedOrSkipped = true
				case generalActionAskAI:
					fmt.Println(m.normalStyle.Render("Attempting to get AI description tags suggestion..."))
					if videoToEdit.Gist == "" {
						fmt.Fprintf(os.Stderr, "Manuscript/Gist path is not defined. Cannot fetch content for AI description tags.\n")
					} else {
						aiConfig, cfgErr := ai.GetAIConfig() // Reuse existing GetAIConfig
						if cfgErr != nil {
							fmt.Fprintf(os.Stderr, "Error getting AI config: %v\n", cfgErr)
						} else {
							manuscriptPath := videoToEdit.Gist
							manuscriptContent, readErr := os.ReadFile(manuscriptPath)
							if readErr != nil {
								fmt.Fprintf(os.Stderr, "Error reading manuscript file %s: %v\n", manuscriptPath, readErr)
							} else {
								suggestedDescTags, suggErr := ai.SuggestDescriptionTags(context.Background(), string(manuscriptContent), aiConfig)
								if suggErr != nil {
									fmt.Fprintf(os.Stderr, "Error suggesting description tags: %v\n", suggErr)
								} else if suggestedDescTags != "" {
									fmt.Println(m.normalStyle.Render("AI suggested description tags received."))
									tempDescTagsValue = suggestedDescTags // Update the temp value to show in the input field
								} else {
									fmt.Println(m.normalStyle.Render("AI did not return any description tags suggestion."))
								}
							}
						}
					}
					// Loop continues to show the form again with the new tempDescTagsValue
				case generalActionSkip:
					df.revertField(originalFieldValue)
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Skipped changes for '%s'.", df.name)))
					fieldSavedOrSkipped = true
				default:
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Unknown action for description tags field: %d", selectedAction)))
					fieldSavedOrSkipped = true // Treat as skip
				}
			}
		} else if df.isTweetField { // New block for Tweet field
			tempTweetValue := originalFieldValue.(string)
			fieldSavedOrSkipped := false

			for !fieldSavedOrSkipped {
				var selectedAction int = generalActionUnknown

				tweetFieldItself := huh.NewText().Title(m.colorTitleString(constants.FieldTitleTweet, tempTweetValue)).Description(df.description).Lines(4).CharLimit(280).Value(&tempTweetValue) // Set Lines(4)
				actionSelect := huh.NewSelect[int]().Title("Action for Tweet").Options(
					huh.NewOption("Save Tweet & Continue", generalActionSave),
					huh.NewOption("Ask AI for Suggestions", generalActionAskAI),
					huh.NewOption("Continue Without Saving Tweet", generalActionSkip),
				).Value(&selectedAction)

				tweetGroup := huh.NewGroup(tweetFieldItself, actionSelect)
				tweetForm := huh.NewForm(tweetGroup)
				formError = tweetForm.Run()

				if formError != nil {
					if formError == huh.ErrUserAborted {
						fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Action for '%s' aborted by user.", df.name)))
						df.revertField(originalFieldValue)
						if fieldIdx == 0 {
							fmt.Println(m.normalStyle.Render(MessageDefinitionPhaseAborted))
							return originalVideoForThisCall, nil
						}
						fieldSavedOrSkipped = true
						continue
					}
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error in tweet form: %v", formError)))
					return videoToEdit, formError
				}

				switch selectedAction {
				case generalActionSave:
					df.updateAction(tempTweetValue)
					saveErr := yamlHelper.WriteVideo(videoToEdit, videoToEdit.Path) // Renamed err to saveErr
					if saveErr != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error saving changes for '%s': %v", df.name, saveErr)))
						df.revertField(originalFieldValue)
						return videoToEdit, saveErr
					}
					fieldSavedOrSkipped = true
				case generalActionAskAI:
					fmt.Println(m.normalStyle.Render("Attempting to get AI tweet suggestions..."))
					if videoToEdit.Gist == "" {
						fmt.Fprintf(os.Stderr, "Manuscript/Gist path is not defined. Cannot fetch content for AI tweets.\n")
					} else {
						aiConfig, cfgErr := ai.GetAIConfig()
						if cfgErr != nil {
							fmt.Fprintf(os.Stderr, "Error getting AI config: %v\n", cfgErr)
						} else {
							manuscriptPath := videoToEdit.Gist
							manuscriptContent, readErr := os.ReadFile(manuscriptPath)
							if readErr != nil {
								fmt.Fprintf(os.Stderr, "Error reading manuscript file %s: %v\n", manuscriptPath, readErr)
							} else {
								suggestedTweets, suggErr := ai.SuggestTweets(context.Background(), string(manuscriptContent), aiConfig)
								if suggErr != nil {
									fmt.Fprintf(os.Stderr, "Error suggesting tweets: %v\n", suggErr)
								} else if len(suggestedTweets) > 0 {
									var selectedAITweet string
									options := []huh.Option[string]{}
									for _, sTweet := range suggestedTweets {
										options = append(options, huh.NewOption(sTweet, sTweet))
									}
									aiSelectForm := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Title("Select an AI Suggested Tweet (or Esc to use current)").Options(options...).Value(&selectedAITweet)))
									aiSelectErr := aiSelectForm.Run()
									if aiSelectErr == nil && selectedAITweet != "" {
										fmt.Println(m.normalStyle.Render(fmt.Sprintf("AI Suggested tweet selected: %s", selectedAITweet)))
										tempTweetValue = selectedAITweet + "\n\n[YOUTUBE]"
									} else if aiSelectErr != nil && aiSelectErr != huh.ErrUserAborted {
										fmt.Fprintf(os.Stderr, "Error during AI tweet selection: %v\n", aiSelectErr)
									}
								} else {
									fmt.Println(m.normalStyle.Render("AI did not return any tweet suggestions."))
								}
							}
						}
					}
				case generalActionSkip:
					df.revertField(originalFieldValue)
					fmt.Println(m.normalStyle.Render(fmt.Sprintf("Skipped changes for '%s'.", df.name)))
					fieldSavedOrSkipped = true
				default:
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Unknown action for tweet field: %d", selectedAction)))
					fieldSavedOrSkipped = true
				}
			}
		} else if df.isAnimationsField { // New block for Animations field
			tempAnimationsValue := originalFieldValue.(string)
			fieldSavedOrSkipped := false

			const (
				animationActionSave           = 0
				animationActionGenerate       = 1
				animationActionSkip           = 2
				animationActionGenerateSimple = 3 // Option if timecodes are too complex
			)

			for !fieldSavedOrSkipped {
				var selectedAction int = generalActionUnknown

				animationsFieldItself := huh.NewText().
					Title(m.colorTitleString(constants.FieldTitleAnimationsScript, tempAnimationsValue)).
					Description(df.description).
					Lines(10). // More lines for animations
					CharLimit(10000).
					Value(&tempAnimationsValue)

				actionSelect := huh.NewSelect[int]().
					Title("Action for Animations").
					Options(
						huh.NewOption("Save Animations & Continue", animationActionSave).Selected(true),
						huh.NewOption("Generate from Gist (Animations & Timecodes)", animationActionGenerate),
						huh.NewOption("Continue Without Saving Animations", animationActionSkip),
					).
					Value(&selectedAction)

				group := huh.NewGroup(animationsFieldItself, actionSelect)
				form := huh.NewForm(group).WithTheme(cli.GetCustomHuhTheme()) // Calling the function

				if err := form.Run(); err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						fmt.Println(m.normalStyle.Render("Animations editing aborted.")) // Use m.normalStyle
						df.revertField(originalFieldValue)                               // Revert on abort
						fieldSavedOrSkipped = true
						// If aborting on the first field, return originalVideoForThisCall
						if fieldIdx == 0 {
							return originalVideoForThisCall, nil
						}
						continue
					}
					log.Printf("Error running animations form: %v", err)
					return videoToEdit, err // Or handle more gracefully
				}

				switch selectedAction {
				case animationActionSave:
					df.updateAction(tempAnimationsValue)
					if err := yamlHelper.WriteVideo(videoToEdit, videoToEdit.Path); err != nil {
						log.Printf("Error saving video data after Animations update: %v", err)
						// Potentially revert or offer retry
					}
					fieldSavedOrSkipped = true
				case animationActionGenerate:
					fsOps := filesystem.NewOperations()
					animLines, animSections, errGen := fsOps.GetAnimations(videoToEdit.Gist)
					if errGen != nil {
						fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error generating animations from Gist: %v", errGen)))
						// Stay in the loop, don't skip field
						continue
					}
					if len(animLines) == 0 {
						fmt.Println(m.normalStyle.Render("No animation cues (TODO: or ## Sections) found in Gist.")) // Use m.normalStyle
						tempAnimationsValue = ""                                                                     // Clear if nothing found
					} else {
						var sb strings.Builder
						for _, line := range animLines {
							sb.WriteString(fmt.Sprintf("- %s\n", line))
						}
						tempAnimationsValue = strings.TrimSpace(sb.String())
					}

					// Update timecodes as well, based on original logic
					if len(animSections) > 0 {
						var tcSb strings.Builder
						tcSb.WriteString("00:00 FIXME:") // Initial FIXME
						for _, section := range animSections {
							tcSb.WriteString(fmt.Sprintf("\nFIXME:FIXME %s", strings.TrimPrefix(section, "Section: ")))
						}
						videoToEdit.Timecodes = tcSb.String()
						// Notify user that timecodes were also updated implicitly
						fmt.Println(m.normalStyle.Render(fmt.Sprintf("Timecodes updated based on Gist sections. Original Timecodes were: [%s]", videoToEdit.Timecodes))) // Use m.normalStyle
					} else {
						// If no sections found, perhaps clear timecodes or leave them as is?
						// For now, let's clear it to indicate it's based on the new Gist parse.
						videoToEdit.Timecodes = ""                                                            // Or some default like "00:00 FIXME: No sections found in Gist"
						fmt.Println(m.normalStyle.Render("No sections found in Gist to generate timecodes.")) // Use m.normalStyle
					}
					// Loop back to show the generated animations to the user
				case animationActionSkip:
					if tempAnimationsValue != originalFieldValue.(string) {
						df.revertField(originalFieldValue)
						fmt.Println(m.normalStyle.Render("Animations reverted to original value.")) // Use m.normalStyle
					} else {
						fmt.Println(m.normalStyle.Render("Animations skipped, no changes made.")) // Use m.normalStyle
					}
					fieldSavedOrSkipped = true
				}
			}
		} else {
			var tempFieldValue interface{} = originalFieldValue
			var fieldInput huh.Field
			var saveThisField bool = true

			switch v := tempFieldValue.(type) {
			case string:
				currentStrVal := v
				fieldInput = huh.NewInput().Title(df.name).Description(df.description).Value(&currentStrVal)
				tempFieldValue = &currentStrVal
			case bool:
				currentBoolVal := v
				// if df.isThumbnailField {  // Keep this outer if, but remove prints
				// 	fmt.Printf("[DEBUG THUMBNAIL VAL] Start of bool case: originalFieldValue (v) = %v\n", v)
				// }
				fieldInput = huh.NewConfirm().Title(df.name).Description(df.description).Value(&currentBoolVal)
				tempFieldValue = &currentBoolVal
			default:
				return videoToEdit, fmt.Errorf("unsupported type for field '%s'", df.name)
			}

			fieldGroup := huh.NewGroup(
				fieldInput,
				huh.NewConfirm().
					Key("saveAction").
					Title(fmt.Sprintf("Finished with '%s'?", df.name)).
					Affirmative("Save & Next").
					Negative("Skip & Next").
					Value(&saveThisField),
			)
			fieldForm := huh.NewForm(fieldGroup)
			formError = fieldForm.Run()

			// if df.isThumbnailField {  // Keep this outer if, but remove prints
			// 	fmt.Printf("[DEBUG THUMBNAIL VAL] After form.Run(), currentBoolVal (pointed to by tempFieldValue) = %v\n", reflect.ValueOf(tempFieldValue).Elem().Bool())
			// }

			if formError != nil {
				if formError == huh.ErrUserAborted {
					fmt.Println(m.orangeStyle.Render(fmt.Sprintf("Action for '%s' aborted by user.", df.name)))
					df.revertField(originalFieldValue)
					if fieldIdx == 0 {
						fmt.Println(m.normalStyle.Render(MessageDefinitionPhaseAborted))
						return originalVideoForThisCall, nil
					}
					continue
				}
				fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error in form for '%s': %v", df.name, formError)))
				return videoToEdit, formError
			}

			if saveThisField {
				finalValue := reflect.ValueOf(tempFieldValue).Elem().Interface()
				// if df.isThumbnailField { // Keep this outer if, but remove prints
				// 	fmt.Printf("[DEBUG THUMBNAIL VAL] Inside saveThisField, finalValue = %v\n", finalValue)
				// }
				df.updateAction(finalValue)
				saveErr := yamlHelper.WriteVideo(videoToEdit, videoToEdit.Path) // Renamed err to saveErr
				if saveErr != nil {
					fmt.Println(m.errorStyle.Render(fmt.Sprintf("Error saving changes for '%s': %v", df.name, saveErr)))
					df.revertField(originalFieldValue)
					return videoToEdit, saveErr
				}
				if df.isThumbnailField && videoToEdit.RequestThumbnail && !initialRequestThumbnailForThisField {
					if settings.Email.Password != "" {
						fmt.Println(m.normalStyle.Render("RequestThumbnail is true, and was false. Sending email..."))
						emailService := notification.NewEmail(settings.Email.Password)
						if err := emailService.SendThumbnail(settings.Email.From, settings.Email.ThumbnailTo, videoToEdit); err != nil {
							fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to send thumbnail request email: %v", err)))
						} else {
							fmt.Println(m.confirmationStyle.Render("Thumbnail request email sent successfully."))
						}
					} else {
						fmt.Println(m.orangeStyle.Render("RequestThumbnail is true, but email app password is not configured. Skipping email."))
					}
				}
				// The problematic 'else' block that was here has been removed.
				// Reverting should only happen if 'saveThisField' is false.
			} else { // This 'else' is for when saveThisField is false (user clicked "Skip & Next")
				fmt.Println(m.normalStyle.Render(fmt.Sprintf("Skipped changes for '%s'.", df.name)))
				df.revertField(originalFieldValue)
			}
		}
	}

	fmt.Println(m.normalStyle.Render(MessageDefinitionPhaseComplete))
	return videoToEdit, nil
}
