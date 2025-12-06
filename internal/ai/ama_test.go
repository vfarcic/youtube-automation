package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const sampleTranscript = `1
00:00:00,000 --> 00:00:05,000
[Music playing]

2
00:00:05,000 --> 00:00:15,000
Welcome everyone to this week's AMA session.

3
00:02:15,000 --> 00:02:30,000
First question from John: How do you handle secrets in GitOps workflows?

4
00:02:30,000 --> 00:03:45,000
Great question. There are several approaches to handling secrets in GitOps.

5
00:08:42,000 --> 00:09:00,000
Next question: What's your opinion on Kubernetes vs Nomad for small teams?

6
00:09:00,000 --> 00:12:30,000
Both are great tools. Kubernetes has more ecosystem support but Nomad is simpler.

7
00:15:30,000 --> 00:16:00,000
Another question about multi-cluster management best practices.
`

func TestGenerateAMATitle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		transcript        string
		mockResponse      string
		mockError         error
		wantTitle         string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:         "Successful title generation",
			transcript:   sampleTranscript,
			mockResponse: "DevOps Q&A: GitOps Secrets, Kubernetes vs Nomad, and Multi-Cluster",
			wantTitle:    "DevOps Q&A: GitOps Secrets, Kubernetes vs Nomad, and Multi-Cluster",
			wantErr:      false,
		},
		{
			name:         "Title with quotes - should be trimmed",
			transcript:   sampleTranscript,
			mockResponse: `"DevOps Q&A: GitOps and Kubernetes"`,
			wantTitle:    "DevOps Q&A: GitOps and Kubernetes",
			wantErr:      false,
		},
		{
			name:              "Empty transcript",
			transcript:        "",
			mockResponse:      "",
			wantTitle:         "",
			wantErr:           true,
			expectedErrSubstr: "transcript is empty",
		},
		{
			name:              "Whitespace-only transcript",
			transcript:        "   \n\t  ",
			mockResponse:      "",
			wantTitle:         "",
			wantErr:           true,
			expectedErrSubstr: "transcript is empty",
		},
		{
			name:              "AI returns empty response",
			transcript:        sampleTranscript,
			mockResponse:      "",
			wantTitle:         "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for AMA title",
		},
		{
			name:              "AI generation error",
			transcript:        sampleTranscript,
			mockError:         fmt.Errorf("AI service unavailable"),
			wantTitle:         "",
			wantErr:           true,
			expectedErrSubstr: "AI service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			gotTitle, err := GenerateAMATitle(ctx, tt.transcript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateAMATitle() error = nil, wantErr = true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("GenerateAMATitle() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateAMATitle() unexpected error = %v", err)
					return
				}
				if gotTitle != tt.wantTitle {
					t.Errorf("GenerateAMATitle() = %q, want %q", gotTitle, tt.wantTitle)
				}
			}
		})
	}
}

func TestGenerateAMATimecodes(t *testing.T) {
	ctx := context.Background()

	expectedTimecodes := `00:00 Intro (skip to first question)
02:15 How do you handle secrets in GitOps?
08:42 Kubernetes vs Nomad for small teams
15:30 Multi-cluster management best practices`

	tests := []struct {
		name              string
		transcript        string
		mockResponse      string
		mockError         error
		wantTimecodes     string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:          "Successful timecodes generation",
			transcript:    sampleTranscript,
			mockResponse:  expectedTimecodes,
			wantTimecodes: expectedTimecodes,
			wantErr:       false,
		},
		{
			name:              "Empty transcript",
			transcript:        "",
			mockResponse:      "",
			wantTimecodes:     "",
			wantErr:           true,
			expectedErrSubstr: "transcript is empty",
		},
		{
			name:              "AI returns empty response",
			transcript:        sampleTranscript,
			mockResponse:      "",
			wantTimecodes:     "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for AMA timecodes",
		},
		{
			name:              "AI generation error",
			transcript:        sampleTranscript,
			mockError:         fmt.Errorf("AI timeout"),
			wantTimecodes:     "",
			wantErr:           true,
			expectedErrSubstr: "AI timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			gotTimecodes, err := GenerateAMATimecodes(ctx, tt.transcript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateAMATimecodes() error = nil, wantErr = true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("GenerateAMATimecodes() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateAMATimecodes() unexpected error = %v", err)
					return
				}
				if gotTimecodes != tt.wantTimecodes {
					t.Errorf("GenerateAMATimecodes() = %q, want %q", gotTimecodes, tt.wantTimecodes)
				}
			}
		})
	}
}

func TestGenerateAMADescription(t *testing.T) {
	ctx := context.Background()

	expectedDescription := "In this AMA session, we covered GitOps secrets management, compared Kubernetes and Nomad for small teams, and discussed multi-cluster best practices."

	tests := []struct {
		name              string
		transcript        string
		mockResponse      string
		mockError         error
		wantDescription   string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:            "Successful description generation",
			transcript:      sampleTranscript,
			mockResponse:    expectedDescription,
			wantDescription: expectedDescription,
			wantErr:         false,
		},
		{
			name:            "Description with extra whitespace",
			transcript:      sampleTranscript,
			mockResponse:    "  " + expectedDescription + "  \n",
			wantDescription: expectedDescription,
			wantErr:         false,
		},
		{
			name:              "Empty transcript",
			transcript:        "",
			mockResponse:      "",
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "transcript is empty",
		},
		{
			name:              "AI returns empty response",
			transcript:        sampleTranscript,
			mockResponse:      "   ",
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for AMA description",
		},
		{
			name:              "AI generation error",
			transcript:        sampleTranscript,
			mockError:         fmt.Errorf("rate limit exceeded"),
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			gotDescription, err := GenerateAMADescription(ctx, tt.transcript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateAMADescription() error = nil, wantErr = true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("GenerateAMADescription() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateAMADescription() unexpected error = %v", err)
					return
				}
				if gotDescription != tt.wantDescription {
					t.Errorf("GenerateAMADescription() = %q, want %q", gotDescription, tt.wantDescription)
				}
			}
		})
	}
}

func TestGenerateAMATags(t *testing.T) {
	ctx := context.Background()

	expectedTags := "AMA, Q&A, livestream, GitOps, Kubernetes, Nomad, secrets management, multi-cluster, DevOps"

	tests := []struct {
		name              string
		transcript        string
		mockResponse      string
		mockError         error
		wantTags          string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:         "Successful tags generation",
			transcript:   sampleTranscript,
			mockResponse: expectedTags,
			wantTags:     expectedTags,
			wantErr:      false,
		},
		{
			name:         "Tags exceeding 450 chars - intelligent truncation",
			transcript:   sampleTranscript,
			mockResponse: strings.Repeat("tag,", 150), // 600 chars
			wantTags:     strings.Repeat("tag,", 111) + "tag", // truncated at last comma before 450 (447 chars)
			wantErr:      false,
		},
		{
			name:         "Tags exceeding 450 chars - hard truncation (no comma)",
			transcript:   sampleTranscript,
			mockResponse: strings.Repeat("a", 500),
			wantTags:     strings.Repeat("a", 450),
			wantErr:      false,
		},
		{
			name:              "Empty transcript",
			transcript:        "",
			mockResponse:      "",
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "transcript is empty",
		},
		{
			name:              "AI returns empty response",
			transcript:        sampleTranscript,
			mockResponse:      "",
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for AMA tags",
		},
		{
			name:              "AI generation error",
			transcript:        sampleTranscript,
			mockError:         fmt.Errorf("connection refused"),
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			gotTags, err := GenerateAMATags(ctx, tt.transcript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateAMATags() error = nil, wantErr = true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("GenerateAMATags() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateAMATags() unexpected error = %v", err)
					return
				}
				// For truncation tests, check length constraint
				if len(gotTags) > 450 {
					t.Errorf("GenerateAMATags() length = %d, want <= 450", len(gotTags))
				}
				if gotTags != tt.wantTags {
					t.Errorf("GenerateAMATags() = %q, want %q", gotTags, tt.wantTags)
				}
			}
		})
	}
}

func TestGenerateAMAContent(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		transcript        string
		mockResponses     []string // title, timecodes, description, tags
		mockError         error
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:       "Successful full content generation",
			transcript: sampleTranscript,
			mockResponses: []string{
				"DevOps Q&A: GitOps, Kubernetes, and Multi-Cluster",
				"00:00 Intro\n02:15 GitOps secrets",
				"This AMA covered GitOps and Kubernetes topics.",
				"AMA, Q&A, GitOps, Kubernetes",
			},
			wantErr: false,
		},
		{
			name:              "Empty transcript",
			transcript:        "",
			mockResponses:     nil,
			wantErr:           true,
			expectedErrSubstr: "transcript is empty",
		},
		{
			name:              "AI error on first call (title)",
			transcript:        sampleTranscript,
			mockError:         fmt.Errorf("API error"),
			wantErr:           true,
			expectedErrSubstr: "failed to generate title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				// Return a provider that cycles through responses
				return &sequentialMockProvider{
					responses: tt.mockResponses,
					callCount: &callCount,
					err:       tt.mockError,
				}, nil
			}

			gotContent, err := GenerateAMAContent(ctx, tt.transcript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateAMAContent() error = nil, wantErr = true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("GenerateAMAContent() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateAMAContent() unexpected error = %v", err)
					return
				}
				if gotContent == nil {
					t.Errorf("GenerateAMAContent() returned nil content")
					return
				}
				if gotContent.Title == "" {
					t.Errorf("GenerateAMAContent() Title is empty")
				}
				if gotContent.Timecodes == "" {
					t.Errorf("GenerateAMAContent() Timecodes is empty")
				}
				if gotContent.Description == "" {
					t.Errorf("GenerateAMAContent() Description is empty")
				}
				if gotContent.Tags == "" {
					t.Errorf("GenerateAMAContent() Tags is empty")
				}
			}
		})
	}
}

// sequentialMockProvider returns different responses for each call
type sequentialMockProvider struct {
	responses []string
	callCount *int
	err       error
}

func (s *sequentialMockProvider) GenerateContent(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	if s.responses == nil || *s.callCount >= len(s.responses) {
		return "", fmt.Errorf("no more mock responses")
	}
	response := s.responses[*s.callCount]
	*s.callCount++
	return response, nil
}

func TestAMATemplateExecution(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		transcript string
	}{
		{
			name:       "Normal transcript",
			transcript: sampleTranscript,
		},
		{
			name:       "Transcript with special characters",
			transcript: "Question: How do you use <brackets> & \"quotes\" in YAML?",
		},
		{
			name:       "Very long transcript",
			transcript: strings.Repeat("This is a long transcript segment. ", 500),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: "Test response",
				err:      nil,
			}

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			// Test that templates execute without error
			_, err := GenerateAMATitle(ctx, tt.transcript)
			if err != nil {
				t.Errorf("GenerateAMATitle() template error = %v", err)
			}

			_, err = GenerateAMATimecodes(ctx, tt.transcript)
			if err != nil {
				t.Errorf("GenerateAMATimecodes() template error = %v", err)
			}
		})
	}
}

func TestGetAIProviderError(t *testing.T) {
	ctx := context.Background()

	originalGetAIProvider := GetAIProvider
	defer func() { GetAIProvider = originalGetAIProvider }()

	GetAIProvider = func() (AIProvider, error) {
		return nil, fmt.Errorf("provider configuration error")
	}

	_, err := GenerateAMATitle(ctx, sampleTranscript)
	if err == nil || !strings.Contains(err.Error(), "failed to create AI provider") {
		t.Errorf("GenerateAMATitle() should fail with provider error, got: %v", err)
	}

	_, err = GenerateAMATimecodes(ctx, sampleTranscript)
	if err == nil || !strings.Contains(err.Error(), "failed to create AI provider") {
		t.Errorf("GenerateAMATimecodes() should fail with provider error, got: %v", err)
	}

	_, err = GenerateAMADescription(ctx, sampleTranscript)
	if err == nil || !strings.Contains(err.Error(), "failed to create AI provider") {
		t.Errorf("GenerateAMADescription() should fail with provider error, got: %v", err)
	}

	_, err = GenerateAMATags(ctx, sampleTranscript)
	if err == nil || !strings.Contains(err.Error(), "failed to create AI provider") {
		t.Errorf("GenerateAMATags() should fail with provider error, got: %v", err)
	}
}
