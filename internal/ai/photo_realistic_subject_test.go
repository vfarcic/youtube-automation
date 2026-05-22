package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestSuggestPhotoRealisticSubject(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		manuscript        string
		mockResponse      string
		mockError         error
		providerError     error
		want              string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:         "success — plain noun phrase",
			manuscript:   "This video covers Kubernetes security best practices for production clusters.",
			mockResponse: "a fortified server rack guarded by holographic shields",
			want:         "a fortified server rack guarded by holographic shields",
			wantErr:      false,
		},
		{
			name:         "success — wrapped in double quotes",
			manuscript:   "Build a serverless data pipeline.",
			mockResponse: `"a small white rabbit holding a code review checklist"`,
			want:         "a small white rabbit holding a code review checklist",
			wantErr:      false,
		},
		{
			name:         "success — wrapped in single quotes",
			manuscript:   "Manage feature flags in production.",
			mockResponse: `'a polished brass key on a velvet cushion'`,
			want:         "a polished brass key on a velvet cushion",
			wantErr:      false,
		},
		{
			name:         "success — wrapped in markdown code fence",
			manuscript:   "Build a Rust web service.",
			mockResponse: "```\na cast iron gear factory machine\n```",
			want:         "a cast iron gear factory machine",
			wantErr:      false,
		},
		{
			name:         "success — markdown fence with language tag",
			manuscript:   "Build a Rust web service.",
			mockResponse: "```text\na rust-colored bicycle by a workshop window\n```",
			want:         "a rust-colored bicycle by a workshop window",
			wantErr:      false,
		},
		{
			name:         "success — surrounding whitespace trimmed",
			manuscript:   "A short video.",
			mockResponse: "   \n\n  a vintage ship's wheel polished in brass  \n\n  ",
			want:         "a vintage ship's wheel polished in brass",
			wantErr:      false,
		},
		{
			name:         "success — preamble label skipped, answer line returned",
			manuscript:   "A video about AI safety.",
			mockResponse: "Here is the subject:\n\na blue robot holding a clipboard",
			want:         "a blue robot holding a clipboard",
			wantErr:      false,
			// The AI template forbids preambles, but models add them anyway
			// ("Here is the subject:", "Sure, here you go:"). The parser
			// strips lines that look like preamble labels and returns the
			// last non-preamble line — the answer.
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
			name:              "AI provider error (rate limit)",
			manuscript:        "Valid manuscript content here.",
			mockError:         fmt.Errorf("rate limit exceeded"),
			wantErr:           true,
			expectedErrSubstr: "AI photo-realistic subject suggestion failed",
		},
		{
			name:              "AI returns empty string",
			manuscript:        "Valid manuscript content.",
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "AI returned empty photo-realistic subject",
		},
		{
			name:              "AI returns whitespace only",
			manuscript:        "Valid manuscript content.",
			mockResponse:      "   \n\t \n  ",
			wantErr:           true,
			expectedErrSubstr: "AI returned empty photo-realistic subject",
		},
		{
			name:              "AI returns only markdown fence",
			manuscript:        "Valid manuscript content.",
			mockResponse:      "```\n\n```",
			wantErr:           true,
			expectedErrSubstr: "AI returned empty photo-realistic subject",
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

			got, err := SuggestPhotoRealisticSubject(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestPhotoRealisticSubject() error = nil, want error")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("SuggestPhotoRealisticSubject() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
				if got != "" {
					t.Errorf("SuggestPhotoRealisticSubject() = %q, want empty on error", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("SuggestPhotoRealisticSubject() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("SuggestPhotoRealisticSubject() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSuggestPhotoRealisticSubject_PromptContainsManuscript locks in that the
// manuscript is passed through to the AI provider (so future template edits
// can't quietly drop it).
func TestSuggestPhotoRealisticSubject_PromptContainsManuscript(t *testing.T) {
	originalGetAIProvider := GetAIProvider
	defer func() { GetAIProvider = originalGetAIProvider }()

	mock := &MockProvider{response: "a robot arm holding a wrench"}
	GetAIProvider = func() (AIProvider, error) { return mock, nil }

	manuscript := "A unique manuscript about distributed systems consensus algorithms."

	if _, err := SuggestPhotoRealisticSubject(context.Background(), manuscript); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(mock.lastPrompt, manuscript) {
		t.Errorf("prompt does not contain manuscript content; prompt=%s", mock.lastPrompt)
	}
}

// TestSuggestPhotoRealisticSubject_PromptForbidsAbstractContent locks in the
// key prompt constraint — the template instructs the model NOT to return
// abstract concepts. Editing the template to drop this guidance would
// regress prompt quality silently; this test catches that.
func TestSuggestPhotoRealisticSubject_PromptForbidsAbstractContent(t *testing.T) {
	originalGetAIProvider := GetAIProvider
	defer func() { GetAIProvider = originalGetAIProvider }()

	mock := &MockProvider{response: "a server"}
	GetAIProvider = func() (AIProvider, error) { return mock, nil }

	if _, err := SuggestPhotoRealisticSubject(context.Background(), "a video"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lower := strings.ToLower(mock.lastPrompt)
	required := []string{"concrete", "photographable", "abstract"}
	for _, s := range required {
		if !strings.Contains(lower, s) {
			t.Errorf("template missing required guidance %q; prompt=\n%s", s, mock.lastPrompt)
		}
	}
}

func TestParsePhotoRealisticSubjectResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain", input: "a small white rabbit", want: "a small white rabbit"},
		{name: "double-quoted", input: `"a small white rabbit"`, want: "a small white rabbit"},
		{name: "single-quoted", input: `'a small white rabbit'`, want: "a small white rabbit"},
		{name: "leading and trailing whitespace", input: "  a server rack  ", want: "a server rack"},
		{name: "markdown fence no lang", input: "```\na server rack\n```", want: "a server rack"},
		{name: "markdown fence with lang", input: "```text\na server rack\n```", want: "a server rack"},
		{name: "multi-line: last non-empty non-preamble line wins", input: "\n\nfirst candidate line\nthe final answer line", want: "the final answer line"},
		{name: "very long truncated", input: strings.Repeat("x", 300), want: strings.Repeat("x", 200)},
		{name: "empty", input: "", wantErr: true},
		{name: "whitespace only", input: " \n \t ", wantErr: true},
		{name: "fence with nothing inside", input: "```\n```", wantErr: true},

		// --- Reviewer-required preamble cases ---
		{
			name:  "preamble: 'Here is the subject:' then answer",
			input: "Here is the subject:\na small white rabbit",
			want:  "a small white rabbit",
		},
		{
			name:  "preamble: 'The photo-realistic subject is:' then answer",
			input: "The photo-realistic subject is:\na server rack with blinking lights",
			want:  "a server rack with blinking lights",
		},
		{
			name:  "preamble: 'Sure, here you go:' then answer",
			input: "Sure, here you go:\na coffee cup on a desk",
			want:  "a coffee cup on a desk",
		},
		{
			name:  "preamble: 'Of course! The answer is:' then answer",
			input: "Of course! The answer is:\na robot waving",
			want:  "a robot waving",
		},
		{
			name:  "single-line response without preamble (regression)",
			input: "a small white rabbit",
			want:  "a small white rabbit",
		},
		{
			name:  "4+ line response: preamble, blank, explanation, blank, answer",
			input: "Here is the photo-realistic subject:\n\nThis works because the topic centers on retro computing.\n\na vintage typewriter on a wooden desk",
			want:  "a vintage typewriter on a wooden desk",
		},
		{
			name:    "only a preamble line — no usable subject",
			input:   "Here is the subject:",
			wantErr: true,
		},
		{
			name:    "preamble followed by another preamble — no usable subject",
			input:   "Sure!\nHere is the answer:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePhotoRealisticSubjectResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parsePhotoRealisticSubjectResponse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
