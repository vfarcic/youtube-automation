package configuration

import (
	"os"
	"testing"
)

func TestBlueskyConfigFromEnv(t *testing.T) {
	// Save original values
	originalIdentifier := os.Getenv("BLUESKY_IDENTIFIER")
	originalPassword := os.Getenv("BLUESKY_PASSWORD")
	defer func() {
		if originalIdentifier != "" {
			os.Setenv("BLUESKY_IDENTIFIER", originalIdentifier)
		} else {
			os.Unsetenv("BLUESKY_IDENTIFIER")
		}
		if originalPassword != "" {
			os.Setenv("BLUESKY_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("BLUESKY_PASSWORD")
		}
	}()

	tests := []struct {
		name           string
		identifier     string
		password       string
		wantIdentifier string
		wantPassword   string
	}{
		{
			name:           "both from env",
			identifier:     "test.bsky.social",
			password:       "testpass123",
			wantIdentifier: "test.bsky.social",
			wantPassword:   "testpass123",
		},
		{
			name:           "identifier only",
			identifier:     "user.bsky.social",
			password:       "",
			wantIdentifier: "user.bsky.social",
			wantPassword:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset GlobalSettings
			GlobalSettings = Settings{}

			if tt.identifier != "" {
				os.Setenv("BLUESKY_IDENTIFIER", tt.identifier)
			} else {
				os.Unsetenv("BLUESKY_IDENTIFIER")
			}

			if tt.password != "" {
				os.Setenv("BLUESKY_PASSWORD", tt.password)
			} else {
				os.Unsetenv("BLUESKY_PASSWORD")
			}

			if err := InitGlobalSettings(); err != nil {
				t.Fatalf("InitGlobalSettings() error = %v", err)
			}

			if GlobalSettings.Bluesky.Identifier != tt.wantIdentifier {
				t.Errorf("Identifier = %q, want %q", GlobalSettings.Bluesky.Identifier, tt.wantIdentifier)
			}

			if GlobalSettings.Bluesky.Password != tt.wantPassword {
				t.Errorf("Password = %q, want %q", GlobalSettings.Bluesky.Password, tt.wantPassword)
			}
		})
	}
}
