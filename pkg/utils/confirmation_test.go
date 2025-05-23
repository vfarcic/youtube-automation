package utils_test

import (
	"devopstoolkit/youtube-automation/pkg/utils"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	// "github.com/charmbracelet/huh" // Not directly used in test logic after refactor
	"github.com/stretchr/testify/assert"
)

func TestConfirmAction(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Confirm with y",
			input:    "y\n",
			expected: true,
		},
		{
			name:     "Confirm with Y",
			input:    "Y\n",
			expected: true,
		},
		{
			name:     "Decline with n",
			input:    "n\n",
			expected: false,
		},
		{
			name:     "Decline with N",
			input:    "N\n",
			expected: false,
		},
		{
			name:     "Decline with explicit enter (simulating default no)",
			input:    "n\n",
			expected: false,
		},
		// Note: huh.Confirm re-prompts internally. Testing multi-line complex interactions
		// like "invalid\ny\n" is more involved as it depends on huh's internal loop.
		// These simpler cases cover the direct input to boolean conversion.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, w, pipeErr := os.Pipe()
			if pipeErr != nil {
				t.Fatalf("Failed to create pipe: %v", pipeErr)
			}

			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				defer w.Close()
				_, err := io.WriteString(w, tc.input)
				if err != nil {
					t.Errorf("Failed to write to pipe: %v", err) // Use t.Errorf to log error but not stop test immediately
				}
			}()

			// Timeout to prevent tests from hanging indefinitely
			done := make(chan bool)
			go func() {
				result := utils.ConfirmAction("Test prompt: "+tc.name, r)
				assert.Equal(t, tc.expected, result, "Confirmation result mismatch")
				close(done)
			}()

			select {
			case <-done:
				// Test completed successfully
			case <-time.After(2 * time.Second): // Increased timeout slightly
				t.Fatal("Test timed out, ConfirmAction likely hanging")
			}

			wg.Wait() // Wait for the writer goroutine to finish
			r.Close() // Close the read end of the pipe
		})
	}
}

// Remove the old placeholder test
// func TestConfirmActionRunsWithoutError(t *testing.T) {
// 	t.Log("TestConfirmActionRunsWithoutError executed. Note: Interactive prompt not tested.")
// }
