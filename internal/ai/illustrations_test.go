package ai

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSuggestTaglineAndIllustrations(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		manuscript        string
		mockResponse      string
		mockError         error
		providerError     error
		wantTaglines      int
		wantIllustrations int
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:              "success",
			manuscript:        "This video covers Kubernetes security best practices for production clusters.",
			mockResponse:      `{"taglines": ["Secure Your Clusters", "Lock It Down", "Zero Trust Now"], "illustrations": ["A fortress protecting server racks", "Shield icons surrounding a Kubernetes wheel", "A cracked padlock being repaired"]}`,
			wantTaglines:      3,
			wantIllustrations: 3,
			wantErr:           false,
		},
		{
			name:              "success with markdown code fence",
			manuscript:        "This video is about monitoring with Prometheus and Grafana.",
			mockResponse:      "```json\n{\"taglines\": [\"Monitor Everything\", \"See It All\"], \"illustrations\": [\"Dashboard with rising graphs\", \"Eye watching server metrics\"]}\n```",
			wantTaglines:      2,
			wantIllustrations: 2,
			wantErr:           false,
		},
		{
			name:              "four illustrations",
			manuscript:        "A comprehensive guide to GitOps workflows.",
			mockResponse:      `{"taglines": ["GitOps Done Right", "Automate All"], "illustrations": ["Git branch tree growing leaves", "Arrows flowing from repo to cluster", "Robot merging pull requests", "Cloud with git icons raining down"]}`,
			wantTaglines:      2,
			wantIllustrations: 4,
			wantErr:           false,
		},
		{
			name:              "empty manuscript",
			manuscript:        "",
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "whitespace-only manuscript",
			manuscript:        "   \n\t  ",
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "AI provider error",
			manuscript:        "Valid manuscript content here.",
			mockError:         fmt.Errorf("rate limit exceeded"),
			wantErr:           true,
			expectedErrSubstr: "AI tagline and illustration suggestion failed",
		},
		{
			name:              "AI returns invalid JSON",
			manuscript:        "Valid manuscript content.",
			mockResponse:      "Here are some ideas: fire, water, earth",
			wantErr:           true,
			expectedErrSubstr: "failed to parse JSON response",
		},
		{
			name:              "AI returns empty taglines",
			manuscript:        "Valid manuscript content.",
			mockResponse:      `{"taglines": [], "illustrations": ["idea one"]}`,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty list of taglines",
		},
		{
			name:              "AI returns empty illustrations",
			manuscript:        "Valid manuscript content.",
			mockResponse:      `{"taglines": ["Tag One"], "illustrations": []}`,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty list of illustrations",
		},
		{
			name:              "provider creation error",
			manuscript:        "Valid manuscript content.",
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

			got, err := SuggestTaglineAndIllustrations(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestTaglineAndIllustrations() error = nil, wantErr true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("SuggestTaglineAndIllustrations() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("SuggestTaglineAndIllustrations() unexpected error = %v", err)
				return
			}

			if len(got.Taglines) != tt.wantTaglines {
				t.Errorf("SuggestTaglineAndIllustrations() returned %d taglines, want %d", len(got.Taglines), tt.wantTaglines)
			}
			if len(got.Illustrations) != tt.wantIllustrations {
				t.Errorf("SuggestTaglineAndIllustrations() returned %d illustrations, want %d", len(got.Illustrations), tt.wantIllustrations)
			}
		})
	}
}

func TestSuggestTaglineAndIllustrations_PromptContainsManuscript(t *testing.T) {
	originalGetAIProvider := GetAIProvider
	defer func() { GetAIProvider = originalGetAIProvider }()

	mock := &MockProvider{
		response: `{"taglines": ["Edge Computing"], "illustrations": ["Illustration idea one", "Illustration idea two", "Illustration idea three"]}`,
	}
	GetAIProvider = func() (AIProvider, error) {
		return mock, nil
	}

	manuscript := "A unique manuscript about serverless computing on edge devices."

	_, err := SuggestTaglineAndIllustrations(context.Background(), manuscript)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(mock.lastPrompt, manuscript) {
		t.Error("prompt does not contain manuscript content")
	}
}

func TestParseTaglineAndIllustrationsResponse(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantTaglines      int
		wantIllustrations int
		wantErr           bool
	}{
		{
			name:              "plain JSON object",
			input:             `{"taglines": ["Tag One", "Tag Two"], "illustrations": ["idea one", "idea two", "idea three"]}`,
			wantTaglines:      2,
			wantIllustrations: 3,
			wantErr:           false,
		},
		{
			name:              "JSON with code fence",
			input:             "```json\n{\"taglines\": [\"A\", \"B\"], \"illustrations\": [\"a\", \"b\"]}\n```",
			wantTaglines:      2,
			wantIllustrations: 2,
			wantErr:           false,
		},
		{
			name:              "JSON with plain code fence",
			input:             "```\n{\"taglines\": [\"A\"], \"illustrations\": [\"a\", \"b\", \"c\"]}\n```",
			wantTaglines:      1,
			wantIllustrations: 3,
			wantErr:           false,
		},
		{
			name:              "JSON with explanatory text before",
			input:             "Here are some suggestions:\n\n{\"taglines\": [\"Robot\", \"Cloud\"], \"illustrations\": [\"Robot painting\", \"Cloud cityscape\"]}",
			wantTaglines:      2,
			wantIllustrations: 2,
			wantErr:           false,
		},
		{
			name:              "markdown-wrapped JSON with explanatory text",
			input:             "Here are my suggestions:\n\n```json\n{\"taglines\": [\"One\"], \"illustrations\": [\"idea one\", \"idea two\"]}\n```\n\nLet me know if you need more.",
			wantTaglines:      1,
			wantIllustrations: 2,
			wantErr:           false,
		},
		{
			name:    "not JSON",
			input:   "just some text",
			wantErr: true,
		},
		{
			name:    "empty taglines",
			input:   `{"taglines": [], "illustrations": ["a"]}`,
			wantErr: true,
		},
		{
			name:    "empty illustrations",
			input:   `{"taglines": ["a"], "illustrations": []}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTaglineAndIllustrationsResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("parseTaglineAndIllustrationsResponse() error = nil, wantErr true")
				}
				return
			}
			if err != nil {
				t.Errorf("parseTaglineAndIllustrationsResponse() unexpected error = %v", err)
				return
			}
			if len(got.Taglines) != tt.wantTaglines {
				t.Errorf("parseTaglineAndIllustrationsResponse() returned %d taglines, want %d", len(got.Taglines), tt.wantTaglines)
			}
			if len(got.Illustrations) != tt.wantIllustrations {
				t.Errorf("parseTaglineAndIllustrationsResponse() returned %d illustrations, want %d", len(got.Illustrations), tt.wantIllustrations)
			}
		})
	}
}

func TestParseTaglineAndIllustrationsResponse_UppercasesTaglines(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantTaglines     []string
		wantIllustrations []string
	}{
		{
			name:              "all lowercase",
			input:             `{"taglines": ["big idea"], "illustrations": ["a tree"]}`,
			wantTaglines:      []string{"BIG IDEA"},
			wantIllustrations: []string{"a tree"},
		},
		{
			name:              "mixed case",
			input:             `{"taglines": ["Big Idea", "Lock It Down"], "illustrations": ["A Tree", "A Cloud"]}`,
			wantTaglines:      []string{"BIG IDEA", "LOCK IT DOWN"},
			wantIllustrations: []string{"A Tree", "A Cloud"},
		},
		{
			name:              "already uppercase stays uppercase",
			input:             `{"taglines": ["BIG IDEA"], "illustrations": ["a tree"]}`,
			wantTaglines:      []string{"BIG IDEA"},
			wantIllustrations: []string{"a tree"},
		},
		{
			name:              "with numbers",
			input:             `{"taglines": ["k8s 101", "Top 10"], "illustrations": ["a chart"]}`,
			wantTaglines:      []string{"K8S 101", "TOP 10"},
			wantIllustrations: []string{"a chart"},
		},
		{
			name:              "with special chars",
			input:             `{"taglines": ["it's epic!", "do-or-die"], "illustrations": ["a sword"]}`,
			wantTaglines:      []string{"IT'S EPIC!", "DO-OR-DIE"},
			wantIllustrations: []string{"a sword"},
		},
		{
			name:              "with surrounding whitespace",
			input:             `{"taglines": ["  Big Idea  ", "\tLock\t"], "illustrations": ["A Tree"]}`,
			wantTaglines:      []string{"BIG IDEA", "LOCK"},
			wantIllustrations: []string{"A Tree"},
		},
		{
			name:              "illustrations with mixed case are not uppercased",
			input:             `{"taglines": ["go fast"], "illustrations": ["A Crumbling Server Rack On Fire", "Cloud icons raining down"]}`,
			wantTaglines:      []string{"GO FAST"},
			wantIllustrations: []string{"A Crumbling Server Rack On Fire", "Cloud icons raining down"},
		},
		{
			name:              "via markdown code fence",
			input:             "```json\n{\"taglines\": [\"big idea\"], \"illustrations\": [\"A Tree\"]}\n```",
			wantTaglines:      []string{"BIG IDEA"},
			wantIllustrations: []string{"A Tree"},
		},
		{
			name:              "via mixed-text fallback parse",
			input:             "Here are some suggestions:\n\n{\"taglines\": [\"big idea\"], \"illustrations\": [\"A Tree\"]}\n\nLet me know.",
			wantTaglines:      []string{"BIG IDEA"},
			wantIllustrations: []string{"A Tree"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTaglineAndIllustrationsResponse(tt.input)
			if err != nil {
				t.Fatalf("parseTaglineAndIllustrationsResponse() unexpected error = %v", err)
			}
			if !reflect.DeepEqual(got.Taglines, tt.wantTaglines) {
				t.Errorf("Taglines = %v, want %v", got.Taglines, tt.wantTaglines)
			}
			if !reflect.DeepEqual(got.Illustrations, tt.wantIllustrations) {
				t.Errorf("Illustrations = %v, want %v", got.Illustrations, tt.wantIllustrations)
			}
		})
	}
}
