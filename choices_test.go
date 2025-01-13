package main

import (
	"testing"

	"github.com/charmbracelet/huh"
)

func TestChoices_getCategories(t *testing.T) {
	choices := &Choices{}
	categories, err := choices.getCategories()
	if err != nil {
		t.Errorf("Error occurred while getting categories: %v", err)
	}

	expectedCategories := []huh.Option[string]{
		huh.NewOption("Category 01", "category-01"),
		huh.NewOption("Category 02", "category-02"),
		huh.NewOption("Category 03", "category-03"),
		huh.NewOption("Category 04", "category-04"),
		huh.NewOption("Category 05", "category-05"),
	}

	if len(categories) != len(expectedCategories) {
		t.Errorf("Expected %d categories, but got %d", len(expectedCategories), len(categories))
	}

	for i, category := range categories {
		if category != expectedCategories[i] {
			t.Errorf("Expected category %v, but got %v", expectedCategories[i], category)
		}
	}
}

func TestChoices_getCreateVideoFields(t *testing.T) {
	choices := &Choices{}
	var name, category string
	var save bool
	fields, err := choices.getCreateVideoFields(&name, &category, &save)
	if err != nil {
		t.Errorf("Error occurred while getting fields: %v", err)
	}
	expectedFieldsNum := 3
	if len(fields) != expectedFieldsNum {
		t.Errorf("Expected %d categories, but got %d", expectedFieldsNum, len(fields))
	}
}

func TestChoices_getIndexOptions(t *testing.T) {
	choices := &Choices{}
	indexOptions := choices.getIndexOptions()
	expectedIndexOptions := []huh.Option[int]{
		huh.NewOption("Create Video", indexCreateVideo),
		huh.NewOption("List Videos", indexListVideos),
		huh.NewOption("Exit", actionReturn),
	}
	if len(indexOptions) != len(expectedIndexOptions) {
		t.Errorf("Expected %d index options, but got %d", len(expectedIndexOptions), len(indexOptions))
	}
	for i, indexOption := range indexOptions {
		if indexOption != expectedIndexOptions[i] {
			t.Errorf("Expected index option %v, but got %v", expectedIndexOptions[i], indexOption)
		}
	}
}
func TestChoices_getActionOptions(t *testing.T) {
	choices := &Choices{}
	actionOptions := choices.getActionOptions()
	expectedActionOptions := []huh.Option[int]{
		huh.NewOption("Edit", actionEdit),
		huh.NewOption("Delete", actionDelete),
		huh.NewOption("Return", actionReturn),
	}
	if len(actionOptions) != len(expectedActionOptions) {
		t.Errorf("Expected %d action options, but got %d", len(expectedActionOptions), len(actionOptions))
	}
	for i, actionOption := range actionOptions {
		if actionOption != expectedActionOptions[i] {
			t.Errorf("Expected action option %v, but got %v", expectedActionOptions[i], actionOption)
		}
	}
}
