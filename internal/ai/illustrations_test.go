package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestSuggestIllustrations(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		manuscript        string
		tagline           string
		mockResponse      string
		mockError         error
		providerError     error
		wantCount         int
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:         "success with tagline",
			manuscript:   "This video covers Kubernetes security best practices for production clusters.",
			tagline:      "Secure Your Clusters",
			mockResponse: `["A fortress protecting server racks", "Shield icons surrounding a Kubernetes wheel", "A cracked padlock being repaired"]`,
			wantCount:    3,
			wantErr:      false,
		},
		{
			name:         "success without tagline",
			manuscript:   "This video covers CI/CD pipeline automation with GitHub Actions.",
			tagline:      "",
			mockResponse: `["Conveyor belt assembling code blocks", "Robot arms welding pipeline segments", "Gears turning inside a GitHub logo"]`,
			wantCount:    3,
			wantErr:      false,
		},
		{
			name:         "success with markdown code fence",
			manuscript:   "This video is about monitoring with Prometheus and Grafana.",
			tagline:      "Monitor Everything",
			mockResponse: "```json\n[\"Dashboard with rising graphs\", \"Eye watching server metrics\"]\n```",
			wantCount:    2,
			wantErr:      false,
		},
		{
			name:         "four suggestions",
			manuscript:   "A comprehensive guide to GitOps workflows.",
			tagline:      "GitOps Done Right",
			mockResponse: `["Git branch tree growing leaves", "Arrows flowing from repo to cluster", "Robot merging pull requests", "Cloud with git icons raining down"]`,
			wantCount:    4,
			wantErr:      false,
		},
		{
			name:              "empty manuscript",
			manuscript:        "",
			tagline:           "Some Tagline",
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "whitespace-only manuscript",
			manuscript:        "   \n\t  ",
			tagline:           "Some Tagline",
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "AI provider error",
			manuscript:        "Valid manuscript content here.",
			tagline:           "Tagline",
			mockError:         fmt.Errorf("rate limit exceeded"),
			wantErr:           true,
			expectedErrSubstr: "AI illustration suggestion failed",
		},
		{
			name:              "AI returns invalid JSON",
			manuscript:        "Valid manuscript content.",
			tagline:           "Tagline",
			mockResponse:      "Here are some ideas: fire, water, earth",
			wantErr:           true,
			expectedErrSubstr: "failed to parse JSON response",
		},
		{
			name:              "AI returns empty array",
			manuscript:        "Valid manuscript content.",
			tagline:           "Tagline",
			mockResponse:      `[]`,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty list of illustrations",
		},
		{
			name:              "provider creation error",
			manuscript:        "Valid manuscript content.",
			tagline:           "Tagline",
			providerError:     fmt.Errorf("no API key configured"),
			wantErr:           true,
			expectedErrSubstr: "failed to create AI provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			if tt.providerError != nil {
				GetAIProvider = func() (AIProvider, error) {
					return nil, tt.providerError
				}
			} else {
				mock := &MockProvider{
					response: tt.mockResponse,
					err:      tt.mockError,
				}
				GetAIProvider = func() (AIProvider, error) {
					return mock, nil
				}
			}

			got, err := SuggestIllustrations(ctx, tt.manuscript, tt.tagline)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestIllustrations() error = nil, wantErr true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("SuggestIllustrations() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("SuggestIllustrations() unexpected error = %v", err)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("SuggestIllustrations() returned %d illustrations, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestSuggestIllustrations_PromptContainsInputs(t *testing.T) {
	originalGetAIProvider := GetAIProvider
	defer func() { GetAIProvider = originalGetAIProvider }()

	mock := &MockProvider{
		response: `["Illustration idea one", "Illustration idea two", "Illustration idea three"]`,
	}
	GetAIProvider = func() (AIProvider, error) {
		return mock, nil
	}

	manuscript := "A unique manuscript about serverless computing on edge devices."
	tagline := "Edge Computing Unleashed"

	_, err := SuggestIllustrations(context.Background(), manuscript, tagline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(mock.lastPrompt, manuscript) {
		t.Error("prompt does not contain manuscript content")
	}
	if !strings.Contains(mock.lastPrompt, tagline) {
		t.Error("prompt does not contain tagline")
	}
}

func TestParseIllustrationsResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "plain JSON array",
			input:     `["idea one", "idea two", "idea three"]`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "JSON with code fence",
			input:     "```json\n[\"a\", \"b\"]\n```",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "JSON with plain code fence",
			input:     "```\n[\"a\", \"b\", \"c\"]\n```",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "JSON with explanatory text before",
			input:     "Here are some illustration ideas for your thumbnail:\n\n[\"Robot painting\", \"Cloud cityscape\"]",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "JSON with explanatory text before and after",
			input:     "Based on your manuscript, I suggest:\n\n[\"Fortress protecting servers\", \"Shield icons\"]\n\nThese ideas convey security visually.",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "markdown-wrapped JSON with explanatory text",
			input:     "Here are my suggestions:\n\n```json\n[\"idea one\", \"idea two\"]\n```\n\nLet me know if you need more.",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "plain code fence with surrounding text",
			input:     "Suggestions:\n```\n[\"a\", \"b\", \"c\"]\n```\nHope these help!",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:    "not JSON",
			input:   "just some text",
			wantErr: true,
		},
		{
			name:    "empty array",
			input:   `[]`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "empty array in code fence",
			input:   "```json\n[]\n```",
			wantErr: true,
		},
		{
			name:    "no JSON array in mixed text",
			input:   "Here are some ideas: fire, water, earth. They represent elements.",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIllustrationsResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("parseIllustrationsResponse() error = nil, wantErr true")
				}
				return
			}
			if err != nil {
				t.Errorf("parseIllustrationsResponse() unexpected error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("parseIllustrationsResponse() returned %d items, want %d", len(got), tt.wantCount)
			}
		})
	}
}
