package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
)

func TestAnalyzeShortsFromManuscript(t *testing.T) {
	ctx := context.Background()
	validManuscript := `Kubernetes is not about containers. It's about declaring what you want and letting the system figure out how to get there. That's the real power of declarative configuration.

Many teams overcomplicate their deployments. You don't need a service mesh on day one. Start simple, add complexity only when you have a real problem to solve.

The biggest mistake I see is treating infrastructure as code without treating it as software. You need tests, you need reviews, you need the same discipline you apply to application code.`

	// Store original settings and restore after test
	originalMaxWords := configuration.GlobalSettings.Shorts.MaxWords
	originalCandidateCount := configuration.GlobalSettings.Shorts.CandidateCount
	defer func() {
		configuration.GlobalSettings.Shorts.MaxWords = originalMaxWords
		configuration.GlobalSettings.Shorts.CandidateCount = originalCandidateCount
	}()

	// Set test configuration
	configuration.GlobalSettings.Shorts.MaxWords = 150
	configuration.GlobalSettings.Shorts.CandidateCount = 3

	tests := []struct {
		name              string
		mockResponse      string
		mockError         error
		manuscript        string
		wantCandidates    int
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name: "Successful analysis with valid candidates",
			mockResponse: `[
				{"id": "short1", "title": "The Real Power of K8s", "text": "Kubernetes is not about containers. It's about declaring what you want.", "rationale": "Strong opening statement"},
				{"id": "short2", "title": "Start Simple", "text": "You don't need a service mesh on day one.", "rationale": "Practical advice"},
				{"id": "short3", "title": "IaC Discipline", "text": "Treat infrastructure as code like software.", "rationale": "Key insight"}
			]`,
			manuscript:     validManuscript,
			wantCandidates: 3,
			wantErr:        false,
		},
		{
			name:         "Successful analysis with JSON in code fence",
			mockResponse: "```json\n[{\"id\": \"short1\", \"title\": \"Test\", \"text\": \"Test text here.\", \"rationale\": \"Good\"}]\n```",
			manuscript:   validManuscript,
			wantCandidates: 1,
			wantErr:      false,
		},
		{
			name:              "Empty manuscript",
			mockResponse:      "",
			manuscript:        "",
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "Whitespace-only manuscript",
			mockResponse:      "",
			manuscript:        "   \n\t  ",
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "AI returns empty response",
			mockResponse:      "",
			manuscript:        validManuscript,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response",
		},
		{
			name:              "AI returns malformed JSON",
			mockResponse:      `[{"id": "short1", "title": "Test"`,
			manuscript:        validManuscript,
			wantErr:           true,
			expectedErrSubstr: "failed to parse shorts candidates",
		},
		{
			name:              "AI returns empty array",
			mockResponse:      `[]`,
			manuscript:        validManuscript,
			wantErr:           true,
			expectedErrSubstr: "AI returned no short candidates",
		},
		{
			name:              "AI generation error",
			mockError:         fmt.Errorf("AI service unavailable"),
			manuscript:        validManuscript,
			wantErr:           true,
			expectedErrSubstr: "AI service unavailable",
		},
		{
			name:              "Candidate missing ID",
			mockResponse:      `[{"id": "", "title": "Test", "text": "Some text.", "rationale": "Good"}]`,
			manuscript:        validManuscript,
			wantErr:           true,
			expectedErrSubstr: "has empty ID",
		},
		{
			name:              "Candidate missing title",
			mockResponse:      `[{"id": "short1", "title": "", "text": "Some text.", "rationale": "Good"}]`,
			manuscript:        validManuscript,
			wantErr:           true,
			expectedErrSubstr: "has empty title",
		},
		{
			name:              "Candidate missing text",
			mockResponse:      `[{"id": "short1", "title": "Test", "text": "", "rationale": "Good"}]`,
			manuscript:        validManuscript,
			wantErr:           true,
			expectedErrSubstr: "has empty text",
		},
		{
			name:              "Candidate exceeds word limit",
			mockResponse:      `[{"id": "short1", "title": "Test", "text": "` + strings.Repeat("word ", 200) + `", "rationale": "Too long"}]`,
			manuscript:        validManuscript,
			wantErr:           true,
			expectedErrSubstr: "exceeds word limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			candidates, err := AnalyzeShortsFromManuscript(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("AnalyzeShortsFromManuscript() error = nil, wantErr = true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("AnalyzeShortsFromManuscript() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("AnalyzeShortsFromManuscript() unexpected error = %v", err)
				return
			}

			if len(candidates) != tt.wantCandidates {
				t.Errorf("AnalyzeShortsFromManuscript() got %d candidates, want %d", len(candidates), tt.wantCandidates)
			}
		})
	}
}

func TestAnalyzeShortsFromManuscript_ProviderError(t *testing.T) {
	ctx := context.Background()

	// Store original GetAIProvider function
	originalGetAIProvider := GetAIProvider
	defer func() { GetAIProvider = originalGetAIProvider }()

	// Mock GetAIProvider to return an error
	GetAIProvider = func() (AIProvider, error) {
		return nil, fmt.Errorf("failed to initialize provider")
	}

	_, err := AnalyzeShortsFromManuscript(ctx, "Some manuscript content")
	if err == nil {
		t.Error("Expected error when provider fails to initialize")
	}
	if !strings.Contains(err.Error(), "failed to create AI provider") {
		t.Errorf("Expected 'failed to create AI provider' error, got: %v", err)
	}
}

func TestValidateShortCandidates(t *testing.T) {
	tests := []struct {
		name       string
		candidates []ShortCandidate
		maxWords   int
		wantErr    bool
		errSubstr  string
	}{
		{
			name: "Valid candidates",
			candidates: []ShortCandidate{
				{ID: "short1", Title: "Test", Text: "Some valid text here.", Rationale: "Good"},
				{ID: "short2", Title: "Another", Text: "More text.", Rationale: "Also good"},
			},
			maxWords: 150,
			wantErr:  false,
		},
		{
			name:       "Empty candidates slice",
			candidates: []ShortCandidate{},
			maxWords:   150,
			wantErr:    true,
			errSubstr:  "no short candidates",
		},
		{
			name:       "Nil candidates slice",
			candidates: nil,
			maxWords:   150,
			wantErr:    true,
			errSubstr:  "no short candidates",
		},
		{
			name: "Candidate at word limit",
			candidates: []ShortCandidate{
				{ID: "short1", Title: "Test", Text: "one two three four five", Rationale: "Exactly 5 words"},
			},
			maxWords: 5,
			wantErr:  false,
		},
		{
			name: "Candidate over word limit by one",
			candidates: []ShortCandidate{
				{ID: "short1", Title: "Test", Text: "one two three four five six", Rationale: "6 words"},
			},
			maxWords:  5,
			wantErr:   true,
			errSubstr: "exceeds word limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateShortCandidates(tt.candidates, tt.maxWords)
			if tt.wantErr {
				if err == nil {
					t.Error("validateShortCandidates() expected error, got nil")
				} else if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("validateShortCandidates() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("validateShortCandidates() unexpected error = %v", err)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "Simple sentence",
			text: "Hello world",
			want: 2,
		},
		{
			name: "Multiple spaces",
			text: "Hello    world   test",
			want: 3,
		},
		{
			name: "Empty string",
			text: "",
			want: 0,
		},
		{
			name: "Whitespace only",
			text: "   \t\n  ",
			want: 0,
		},
		{
			name: "Single word",
			text: "Kubernetes",
			want: 1,
		},
		{
			name: "Paragraph with punctuation",
			text: "Hello, world! This is a test.",
			want: 6,
		},
		{
			name: "Text with newlines",
			text: "Line one\nLine two\nLine three",
			want: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountWords(tt.text)
			if got != tt.want {
				t.Errorf("CountWords(%q) = %d, want %d", tt.text, got, tt.want)
			}
		})
	}
}

func TestShortCandidate_Fields(t *testing.T) {
	// Test that ShortCandidate struct has expected fields
	candidate := ShortCandidate{
		ID:        "short1",
		Title:     "Test Title",
		Text:      "This is the text segment.",
		Rationale: "This is a good short because...",
	}

	if candidate.ID != "short1" {
		t.Errorf("ID = %q, want %q", candidate.ID, "short1")
	}
	if candidate.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", candidate.Title, "Test Title")
	}
	if candidate.Text != "This is the text segment." {
		t.Errorf("Text = %q, want %q", candidate.Text, "This is the text segment.")
	}
	if candidate.Rationale != "This is a good short because..." {
		t.Errorf("Rationale = %q, want %q", candidate.Rationale, "This is a good short because...")
	}
}

func TestShortsTemplateData(t *testing.T) {
	// Test that template data struct has expected fields
	data := shortsTemplateData{
		ManuscriptContent: "Test content",
		MaxWords:          150,
		CandidateCount:    10,
	}

	if data.ManuscriptContent != "Test content" {
		t.Errorf("ManuscriptContent = %q, want %q", data.ManuscriptContent, "Test content")
	}
	if data.MaxWords != 150 {
		t.Errorf("MaxWords = %d, want %d", data.MaxWords, 150)
	}
	if data.CandidateCount != 10 {
		t.Errorf("CandidateCount = %d, want %d", data.CandidateCount, 10)
	}
}
