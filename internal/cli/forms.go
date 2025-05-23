package cli

import (
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GetCreateVideoFields returns the form fields for creating a new video
func GetCreateVideoFields(name, category *string, save *bool) ([]huh.Field, error) {
	categories, err := GetCategories()
	if err != nil {
		return nil, err
	}
	return []huh.Field{
		huh.NewInput().Prompt("Name: ").Value(name),
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
