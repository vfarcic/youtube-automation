package cli

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestGetCreateVideoFields(t *testing.T) {
	// Create a temporary manuscript directory for testing
	tempDir, err := os.MkdirTemp("", "manuscript-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test category directories
	categories := []string{"test-category-1", "test-category-2"}
	for _, category := range categories {
		categoryDir := filepath.Join(tempDir, category)
		if err := os.MkdirAll(categoryDir, 0755); err != nil {
			t.Fatalf("Failed to create category dir %s: %v", categoryDir, err)
		}
	}

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create manuscript directory in temp location
	manuscriptDir := filepath.Join(tempDir, "manuscript")
	if err := os.MkdirAll(manuscriptDir, 0755); err != nil {
		t.Fatalf("Failed to create manuscript dir: %v", err)
	}

	// Create test categories in manuscript directory
	for _, category := range categories {
		categoryDir := filepath.Join(manuscriptDir, category)
		if err := os.MkdirAll(categoryDir, 0755); err != nil {
			t.Fatalf("Failed to create manuscript category dir %s: %v", categoryDir, err)
		}
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	var name, category, date string
	save := true

	fields, err := GetCreateVideoFields(&name, &category, &date, &save)
	if err != nil {
		t.Fatalf("GetCreateVideoFields failed: %v", err)
	}

	if len(fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(fields))
	}

	// Note: huh fields may not expose specific field type information easily
	// We can verify that we got the right number of fields and that the function returns successfully
	// More detailed testing would require deeper integration with the huh library
}

func TestGetCategories(t *testing.T) {
	// Create a temporary manuscript directory for testing
	tempDir, err := os.MkdirTemp("", "manuscript-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create manuscript directory in temp location
	manuscriptDir := filepath.Join(tempDir, "manuscript")
	if err := os.MkdirAll(manuscriptDir, 0755); err != nil {
		t.Fatalf("Failed to create manuscript dir: %v", err)
	}

	// Create test category directories
	categories := []string{"test-category-1", "test-category-2", "another-test"}
	for _, category := range categories {
		categoryDir := filepath.Join(manuscriptDir, category)
		if err := os.MkdirAll(categoryDir, 0755); err != nil {
			t.Fatalf("Failed to create category dir %s: %v", categoryDir, err)
		}
	}

	// Create a file (should be ignored)
	testFile := filepath.Join(manuscriptDir, "test-file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	options, err := GetCategories()
	if err != nil {
		t.Fatalf("GetCategories failed: %v", err)
	}

	if len(options) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(options))
	}

	// Extract titles and values for comparison
	var actualTitles []string
	var actualValues []string
	for _, option := range options {
		actualTitles = append(actualTitles, option.Key)
		actualValues = append(actualValues, option.Value)
	}

	// Sort for consistent comparison (since GetCategories doesn't guarantee order)
	sort.Strings(actualTitles)
	sort.Strings(actualValues)

	expectedTitles := []string{"Another Test", "Test Category 1", "Test Category 2"}
	expectedValues := []string{"another-test", "test-category-1", "test-category-2"}
	sort.Strings(expectedTitles)
	sort.Strings(expectedValues)

	for i, expectedTitle := range expectedTitles {
		if i < len(actualTitles) {
			if actualTitles[i] != expectedTitle {
				t.Errorf("Expected category title %s, got %s at index %d", expectedTitle, actualTitles[i], i)
			}
		}
	}

	for i, expectedValue := range expectedValues {
		if i < len(actualValues) {
			if actualValues[i] != expectedValue {
				t.Errorf("Expected category value %s, got %s at index %d", expectedValue, actualValues[i], i)
			}
		}
	}
}

func TestGetActionOptions(t *testing.T) {
	options := GetActionOptions()

	if len(options) != 4 {
		t.Errorf("Expected 4 action options, got %d", len(options))
	}

	expectedActions := []string{"Edit", "Delete", "Move Video", "Return"}
	for i, option := range options {
		if i < len(expectedActions) {
			if option.Key != expectedActions[i] {
				t.Errorf("Expected action %s, got %s", expectedActions[i], option.Key)
			}
		}
	}
}
