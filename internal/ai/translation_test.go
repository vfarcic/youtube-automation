package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestTranslateVideoMetadata(t *testing.T) {
	ctx := context.Background()

	validInput := VideoMetadataInput{
		Title:       "Why Kubernetes Is Taking Over the World",
		Description: "In this video, we explore Kubernetes and its ecosystem.",
		Tags:        "Kubernetes, DevOps, Cloud Native, Containers",
		Timecodes:   "0:00 Introduction\n2:30 What is Kubernetes\n5:00 Demo",
	}

	validOutput := `{
		"title": "Por qué Kubernetes está conquistando el mundo",
		"description": "En este video, exploramos Kubernetes y su ecosistema.",
		"tags": "Kubernetes, DevOps, Cloud Native, Contenedores",
		"timecodes": "0:00 Introducción\n2:30 Qué es Kubernetes\n5:00 Demo"
	}`

	tests := []struct {
		name              string
		input             VideoMetadataInput
		targetLanguage    string
		mockResponse      string
		mockError         error
		wantErr           bool
		expectedErrSubstr string
		validateOutput    func(*testing.T, *VideoMetadataOutput)
	}{
		{
			name:           "Successful translation",
			input:          validInput,
			targetLanguage: "Spanish",
			mockResponse:   validOutput,
			wantErr:        false,
			validateOutput: func(t *testing.T, output *VideoMetadataOutput) {
				if output.Title == "" {
					t.Error("Expected non-empty translated title")
				}
				if output.Description == "" {
					t.Error("Expected non-empty translated description")
				}
				if output.Tags == "" {
					t.Error("Expected non-empty translated tags")
				}
				if output.Timecodes == "" {
					t.Error("Expected non-empty translated timecodes")
				}
			},
		},
		{
			name:           "Translation with JSON code fence",
			input:          validInput,
			targetLanguage: "Spanish",
			mockResponse:   "```json\n" + validOutput + "\n```",
			wantErr:        false,
			validateOutput: func(t *testing.T, output *VideoMetadataOutput) {
				if output.Title == "" {
					t.Error("Expected non-empty translated title after stripping code fence")
				}
			},
		},
		{
			name:           "Translation with plain code fence",
			input:          validInput,
			targetLanguage: "Spanish",
			mockResponse:   "```\n" + validOutput + "\n```",
			wantErr:        false,
			validateOutput: func(t *testing.T, output *VideoMetadataOutput) {
				if output.Title == "" {
					t.Error("Expected non-empty translated title after stripping plain code fence")
				}
			},
		},
		{
			name:              "Empty target language",
			input:             validInput,
			targetLanguage:    "",
			mockResponse:      validOutput,
			wantErr:           true,
			expectedErrSubstr: "target language is required",
		},
		{
			name:              "Empty input fields",
			input:             VideoMetadataInput{},
			targetLanguage:    "Spanish",
			mockResponse:      validOutput,
			wantErr:           true,
			expectedErrSubstr: "at least one field",
		},
		{
			name: "Partial input - only title",
			input: VideoMetadataInput{
				Title: "Test Title",
			},
			targetLanguage: "Spanish",
			mockResponse:   `{"title": "Título de prueba", "description": "", "tags": "", "timecodes": ""}`,
			wantErr:        false,
			validateOutput: func(t *testing.T, output *VideoMetadataOutput) {
				if output.Title == "" {
					t.Error("Expected non-empty translated title")
				}
			},
		},
		{
			name: "Partial input - only short titles",
			input: VideoMetadataInput{
				ShortTitles: []string{"Short 1", "Short 2"},
			},
			targetLanguage: "Spanish",
			mockResponse:   `{"title": "", "description": "", "tags": "", "timecodes": "", "shortTitles": ["Corto 1", "Corto 2"]}`,
			wantErr:        false,
			validateOutput: func(t *testing.T, output *VideoMetadataOutput) {
				if len(output.ShortTitles) != 2 {
					t.Errorf("Expected 2 short titles, got %d", len(output.ShortTitles))
				}
			},
		},
		{
			name:              "AI returns empty response",
			input:             validInput,
			targetLanguage:    "Spanish",
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response",
		},
		{
			name:              "AI returns invalid JSON",
			input:             validInput,
			targetLanguage:    "Spanish",
			mockResponse:      `{"title": "incomplete`,
			wantErr:           true,
			expectedErrSubstr: "failed to parse JSON response",
		},
		{
			name:              "AI provider error",
			input:             validInput,
			targetLanguage:    "Spanish",
			mockError:         fmt.Errorf("AI service unavailable"),
			wantErr:           true,
			expectedErrSubstr: "AI service unavailable",
		},
		{
			name: "Translation with short titles",
			input: VideoMetadataInput{
				Title:       "Main Video Title",
				Description: "Video description",
				Tags:        "tag1, tag2",
				Timecodes:   "0:00 Intro",
				ShortTitles: []string{"Short 1: The First", "Short 2: The Second", "Short 3: The Third"},
			},
			targetLanguage: "Spanish",
			mockResponse: `{
				"title": "Título del Video Principal",
				"description": "Descripción del video",
				"tags": "etiqueta1, etiqueta2",
				"timecodes": "0:00 Introducción",
				"shortTitles": ["Short 1: El Primero", "Short 2: El Segundo", "Short 3: El Tercero"]
			}`,
			wantErr: false,
			validateOutput: func(t *testing.T, output *VideoMetadataOutput) {
				if len(output.ShortTitles) != 3 {
					t.Errorf("Expected 3 short titles, got %d", len(output.ShortTitles))
				}
				if len(output.ShortTitles) > 0 && output.ShortTitles[0] == "" {
					t.Error("Expected non-empty first short title")
				}
			},
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

			output, err := TranslateVideoMetadata(ctx, tt.input, tt.targetLanguage)

			if tt.wantErr {
				if err == nil {
					t.Errorf("TranslateVideoMetadata() error = nil, wantErr = true")
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("TranslateVideoMetadata() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("TranslateVideoMetadata() unexpected error = %v", err)
				return
			}

			if output == nil {
				t.Error("TranslateVideoMetadata() returned nil output")
				return
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}

func TestTranslateVideoMetadataWithSpecialCharacters(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		input        VideoMetadataInput
		mockResponse string
		wantErr      bool
	}{
		{
			name: "Input with quotes",
			input: VideoMetadataInput{
				Title:       `Why "Kubernetes" Matters`,
				Description: `He said "this is important"`,
				Tags:        "Kubernetes, DevOps",
				Timecodes:   "0:00 Introduction",
			},
			mockResponse: `{"title": "Por qué \"Kubernetes\" importa", "description": "Él dijo \"esto es importante\"", "tags": "Kubernetes, DevOps", "timecodes": "0:00 Introducción"}`,
			wantErr:      false,
		},
		{
			name: "Input with newlines in description",
			input: VideoMetadataInput{
				Title:       "Test Video",
				Description: "Line 1\nLine 2\nLine 3",
				Tags:        "test",
				Timecodes:   "0:00 Start\n1:00 Middle\n2:00 End",
			},
			mockResponse: `{"title": "Video de prueba", "description": "Línea 1\nLínea 2\nLínea 3", "tags": "prueba", "timecodes": "0:00 Inicio\n1:00 Medio\n2:00 Final"}`,
			wantErr:      false,
		},
		{
			name: "Input with backslashes",
			input: VideoMetadataInput{
				Title:       `Path: C:\Users\test`,
				Description: "Test description",
				Tags:        "test",
				Timecodes:   "",
			},
			mockResponse: `{"title": "Ruta: C:\\Users\\test", "description": "Descripción de prueba", "tags": "prueba", "timecodes": ""}`,
			wantErr:      false,
		},
		{
			name: "Input with URLs",
			input: VideoMetadataInput{
				Title:       "Check out this tool",
				Description: "Visit https://kubernetes.io for more info",
				Tags:        "Kubernetes",
				Timecodes:   "",
			},
			mockResponse: `{"title": "Mira esta herramienta", "description": "Visita https://kubernetes.io para más información", "tags": "Kubernetes", "timecodes": ""}`,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: tt.mockResponse,
				err:      nil,
			}

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			output, err := TranslateVideoMetadata(ctx, tt.input, "Spanish")

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if output == nil {
				t.Error("Expected non-nil output")
			}
		})
	}
}

func TestStripCodeFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No code fence",
			input:    `{"title": "test"}`,
			expected: `{"title": "test"}`,
		},
		{
			name:     "JSON code fence with newlines",
			input:    "```json\n{\"title\": \"test\"}\n```",
			expected: `{"title": "test"}`,
		},
		{
			name:     "JSON code fence without newlines",
			input:    "```json{\"title\": \"test\"}```",
			expected: `{"title": "test"}`,
		},
		{
			name:     "Plain code fence with newlines",
			input:    "```\n{\"title\": \"test\"}\n```",
			expected: `{"title": "test"}`,
		},
		{
			name:     "Plain code fence without newlines",
			input:    "```{\"title\": \"test\"}```",
			expected: `{"title": "test"}`,
		},
		{
			name:     "With leading/trailing whitespace",
			input:    "  ```json\n{\"title\": \"test\"}\n```  ",
			expected: `{"title": "test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripCodeFences(tt.input)
			if result != tt.expected {
				t.Errorf("stripCodeFences() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTranslateVideoMetadataProviderError(t *testing.T) {
	ctx := context.Background()

	input := VideoMetadataInput{
		Title: "Test Title",
	}

	originalGetAIProvider := GetAIProvider
	defer func() { GetAIProvider = originalGetAIProvider }()

	// Mock GetAIProvider to return an error
	GetAIProvider = func() (AIProvider, error) {
		return nil, fmt.Errorf("failed to initialize provider")
	}

	_, err := TranslateVideoMetadata(ctx, input, "Spanish")

	if err == nil {
		t.Error("Expected error when provider fails to initialize")
		return
	}

	if !strings.Contains(err.Error(), "failed to create AI provider") {
		t.Errorf("Expected error about provider creation, got: %v", err)
	}
}
