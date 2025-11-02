package cli

import (
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GetCreateVideoFields returns the form fields for creating a new video
var GetCreateVideoFields = func(name, category, date *string, save *bool) ([]huh.Field, error) {
	categories, err := GetCategories()
	if err != nil {
		return nil, err
	}

	// Set default date to January 1st of next year at 16:00 UTC if not provided
	if *date == "" {
		currentYear := time.Now().UTC().Year()
		defaultDate := time.Date(currentYear+1, time.January, 1, 16, 0, 0, 0, time.UTC)
		*date = defaultDate.Format("2006-01-02T15:04")
	}

	return []huh.Field{
		huh.NewInput().Prompt("Name: ").Value(name),
		huh.NewInput().Title("Publish Date (YYYY-MM-DDTHH:MM)").Value(date),
		huh.NewSelect[string]().Title("Category").Options(categories...).Value(category),
		huh.NewConfirm().Affirmative("Save").Negative("Cancel").Value(save),
	}, nil
}

// GetCategories returns the available video categories from the manuscript directory
func GetCategories() ([]huh.Option[string], error) {
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

// GetActionOptions returns the action menu options
func GetActionOptions() []huh.Option[int] {
	const (
		actionEdit = iota
		actionDelete
		actionMoveFiles
	)
	const actionReturn = 99

	return []huh.Option[int]{
		huh.NewOption("Edit", actionEdit),
		huh.NewOption("Delete", actionDelete),
		huh.NewOption("Move Video", actionMoveFiles),
		huh.NewOption("Return", actionReturn),
	}
}
