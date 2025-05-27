package markdown

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestApplyHighlightsInGist(t *testing.T) {
	originalReadFile := osReadFile
	originalWriteFile := osWriteFile
	defer func() {
		osReadFile = originalReadFile
		osWriteFile = originalWriteFile
	}()

	tests := []struct {
		name              string
		mockReadContent   string
		mockReadError     error
		mockWriteError    error
		highlightsToApply []string
		expectedContent   string
		wantErr           bool
		expectedErrSubstr string
		checkWriteContent bool   // Flag to indicate if we should capture and check written content
		capturedWriteData string // To store data passed to writeFile
	}{
		{
			name:              "Successful highlighting of multiple phrases",
			mockReadContent:   "This is a test sentence with important keywords and phrases.",
			highlightsToApply: []string{"important keywords", "phrases"},
			expectedContent:   "This is a test sentence with **important keywords** and **phrases**.",
			wantErr:           false,
			checkWriteContent: true,
		},
		{
			name:              "No highlights provided",
			mockReadContent:   "Content should remain unchanged.",
			highlightsToApply: []string{},
			expectedContent:   "Content should remain unchanged.",
			wantErr:           false,
			checkWriteContent: true,
		},
		{
			name:              "Empty string in highlights",
			mockReadContent:   "Content with an empty highlight request.",
			highlightsToApply: []string{"Content", "", "highlight"},
			expectedContent:   "**Content** with an empty **highlight** request.",
			wantErr:           false,
			checkWriteContent: true,
		},
		{
			name:              "Highlight phrase not found",
			mockReadContent:   "This content does not contain the target.",
			highlightsToApply: []string{"nonexistent"},
			expectedContent:   "This content does not contain the target.",
			wantErr:           false,
			checkWriteContent: true,
		},
		{
			name:              "Error reading gist file",
			mockReadError:     fmt.Errorf("mock read error"),
			highlightsToApply: []string{"any"},
			wantErr:           true,
			expectedErrSubstr: "failed to read gist file",
		},
		{
			name:              "Error writing gist file",
			mockReadContent:   "Some content to attempt writing.",
			mockWriteError:    fmt.Errorf("mock write error"),
			highlightsToApply: []string{"content"},
			wantErr:           true,
			expectedErrSubstr: "failed to write modified content to gist file",
			checkWriteContent: false, // Don't check content if write fails, but write will be attempted
		},
		{
			name:              "Cleanup of **** to **",
			mockReadContent:   "Replace this and ****that****.", // Simulates a scenario where direct replacement might create ****
			highlightsToApply: []string{"this"},                 // If 'this' was replaced and became '**this**', then if 'that' was already bold.
			expectedContent:   "Replace **this** and **that**.", // The test content actually simulates it better: ****word**** -> **word**.
			wantErr:           false,
			checkWriteContent: true,
		},
		{
			name:              "Specific **** cleanup test",
			mockReadContent:   "This has ****bolded**** text.",
			highlightsToApply: []string{"This"}, // Highlighting "This" won't cause quadruple, but the ReplaceAll for **** should fix existing ones.
			expectedContent:   "**This** has **bolded** text.",
			wantErr:           false,
			checkWriteContent: true,
		},
		{
			name:              "Substring already bold",
			mockReadContent:   "This is a **bold** statement with bold.",
			highlightsToApply: []string{"bold"},
			expectedContent:   "This is a ****bold**** statement with **bold**.", // Current behavior: simple replacement leads to ****
			// After **** cleanup: "This is a **bold** statement with **bold**."
			wantErr:           false,
			checkWriteContent: true,
		},
		{
			name:              "Highlighting a phrase containing an already bolded word",
			mockReadContent:   "This is a **very** important test.",
			highlightsToApply: []string{"a **very** important"},
			expectedContent:   "This is **a **very** important** test.", // Shows that existing markdown within a phrase is preserved
			wantErr:           false,
			checkWriteContent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osReadFile = func(name string) ([]byte, error) {
				if tt.mockReadError != nil {
					return nil, tt.mockReadError
				}
				return []byte(tt.mockReadContent), nil
			}

			// Assign to a variable that can be captured in the closure
			var capturedData string
			osWriteFile = func(name string, data []byte, perm os.FileMode) error {
				capturedData = string(data) // Capture data
				return tt.mockWriteError
			}

			err := ApplyHighlightsInGist("fake/gist/path.md", tt.highlightsToApply)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ApplyHighlightsInGist() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("ApplyHighlightsInGist() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("ApplyHighlightsInGist() unexpected error = %v", err)
					return
				}
			}

			if tt.checkWriteContent && !tt.wantErr {
				// The **** to ** replacement happens *after* all individual phrase replacements.
				// So, for the "Substring already bold" case, the intermediate `content` might have ****
				// which then gets cleaned up. The final capturedData should reflect the final state.
				finalExpected := tt.expectedContent
				if tt.name == "Substring already bold" { // Special case for this test due to ReplaceAll order
					finalExpected = strings.ReplaceAll(tt.expectedContent, "****bold****", "**bold**")
				}
				if capturedData != finalExpected {
					t.Errorf("ApplyHighlightsInGist() captured write data = %q, want %q", capturedData, finalExpected)
				}
			}
		})
	}
}
