package ai

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// MockProvider implements AIProvider for testing
type MockProvider struct {
	response   string
	err        error
	lastPrompt string
}

func (m *MockProvider) GenerateContent(ctx context.Context, prompt string, maxTokens int) (string, error) {
	m.lastPrompt = prompt
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

// setupTitlesTestDir creates a temp directory with a titles.md file and changes to it.
// Returns a cleanup function that restores the original working directory.
func setupTitlesTestDir(t *testing.T) func() {
	t.Helper()
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working dir: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	// Write the default template as titles.md
	if err := os.WriteFile("titles.md", []byte(defaultTitlesTemplate), 0644); err != nil {
		t.Fatalf("Failed to write titles.md: %v", err)
	}
	return func() {
		os.Chdir(originalWd)
	}
}

func TestSuggestTitles(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video title generation. It needs multiple suggestions."

	tests := []struct {
		name              string
		mockResponse      string
		mockError         error
		manuscript        string
		wantTitles        []string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:         "Successful titles suggestion - direct array",
			mockResponse: `["AI Title 1", "Catchy Title 2"]`,
			manuscript:   validManuscript,
			wantTitles:   []string{"AI Title 1", "Catchy Title 2"},
			wantErr:      false,
		},
		{
			name:         "Successful titles suggestion - with JSON code fence",
			mockResponse: "```json\n[\"Fenced Title A\"]\n```",
			manuscript:   validManuscript,
			wantTitles:   []string{"Fenced Title A"},
			wantErr:      false,
		},
		{
			name:              "AI returns empty response for titles",
			mockResponse:      "",
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for titles",
		},
		{
			name:              "AI returns malformed JSON for titles",
			mockResponse:      `["Unterminated title array]`,
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "failed to parse JSON response",
		},
		{
			name:              "AI generation error for titles",
			mockError:         fmt.Errorf("AI titles service unavailable"),
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI titles service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTitlesTestDir(t)
			defer cleanup()

			mock := &MockProvider{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			// Store original GetAIProvider function
			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			// Mock the GetAIProvider function
			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			gotTitles, err := SuggestTitles(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestTitles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("SuggestTitles() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("SuggestTitles() unexpected error = %v", err)
					return
				}
			}

			if !reflect.DeepEqual(gotTitles, tt.wantTitles) {
				t.Errorf("SuggestTitles() gotTitles = %v, want %v", gotTitles, tt.wantTitles)
			}
		})
	}
}

func TestGetAIConfig(t *testing.T) {
	// Test the compatibility function
	config, err := GetAIConfig()
	if err != nil {
		t.Errorf("GetAIConfig() unexpected error = %v", err)
	}
	if config == nil {
		t.Errorf("GetAIConfig() returned nil config")
	}
}

func TestTitlesTemplateExecution(t *testing.T) {
	// Test that the template executes correctly with various manuscript inputs
	ctx := context.Background()

	tests := []struct {
		name       string
		manuscript string
		wantErr    bool
	}{
		{
			name:       "Normal manuscript",
			manuscript: "This is a test manuscript",
			wantErr:    false,
		},
		{
			name:       "Manuscript with special characters",
			manuscript: "Test <script>alert('xss')</script> & \"quotes\" 'single'",
			wantErr:    false,
		},
		{
			name:       "Empty manuscript",
			manuscript: "",
			wantErr:    false,
		},
		{
			name:       "Very long manuscript",
			manuscript: strings.Repeat("A", 10000),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTitlesTestDir(t)
			defer cleanup()

			mock := &MockProvider{
				response: `["Test Title 1", "Test Title 2"]`,
				err:      nil,
			}

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			titles, err := SuggestTitles(ctx, tt.manuscript)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && len(titles) != 2 {
				t.Errorf("Expected 2 titles, got %d", len(titles))
			}
		})
	}
}

func TestLoadTitlesTemplate(t *testing.T) {
	tests := []struct {
		name         string
		fileContent  string
		fileNotExist bool
		wantErr      bool
		errContains  string
	}{
		{
			name:        "loads titles.md from working directory",
			fileContent: "Custom template with {{.ManuscriptContent}}",
			wantErr:     false,
		},
		{
			name:         "returns error with instructions when file missing",
			fileNotExist: true,
			wantErr:      true,
			errContains:  "titles.md not found",
		},
		{
			name:         "error includes default template content",
			fileNotExist: true,
			wantErr:      true,
			errContains:  "Analyze → Titles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			originalWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working dir: %v", err)
			}
			defer os.Chdir(originalWd)

			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to chdir: %v", err)
			}

			if !tt.fileNotExist {
				if err := os.WriteFile("titles.md", []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("Failed to write titles.md: %v", err)
				}
			}

			got, err := LoadTitlesTemplate()

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error %q should contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if got != tt.fileContent {
					t.Errorf("Got %q, want %q", got, tt.fileContent)
				}
			}
		})
	}
}

