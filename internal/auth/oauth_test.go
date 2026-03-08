package auth

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestTokenFileOperations(t *testing.T) {
	tempDir := t.TempDir()
	tokenPath := filepath.Join(tempDir, "token.json")

	testToken := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
	}

	// Manually write a token file
	f, err := os.Create(tokenPath)
	if err != nil {
		t.Fatalf("Failed to create token file: %v", err)
	}
	if err := json.NewEncoder(f).Encode(testToken); err != nil {
		f.Close()
		t.Fatalf("Failed to write token: %v", err)
	}
	f.Close()

	// Test TokenFromFile
	readToken, err := TokenFromFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read token file: %v", err)
	}
	if readToken.AccessToken != testToken.AccessToken {
		t.Errorf("Expected access token %s, got %s", testToken.AccessToken, readToken.AccessToken)
	}
	if readToken.RefreshToken != testToken.RefreshToken {
		t.Errorf("Expected refresh token %s, got %s", testToken.RefreshToken, readToken.RefreshToken)
	}

	// Test SaveToken with a new token
	newToken := &oauth2.Token{
		AccessToken:  "new-access-token",
		TokenType:    "Bearer",
		RefreshToken: "new-refresh-token",
	}

	os.Remove(tokenPath)

	if err := SaveToken(tokenPath, newToken); err != nil {
		t.Fatalf("SaveToken returned error: %v", err)
	}

	readNewToken, err := TokenFromFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read new token file: %v", err)
	}
	if readNewToken.AccessToken != newToken.AccessToken {
		t.Errorf("Expected new access token %s, got %s", newToken.AccessToken, readNewToken.AccessToken)
	}
	if readNewToken.RefreshToken != newToken.RefreshToken {
		t.Errorf("Expected new refresh token %s, got %s", newToken.RefreshToken, readNewToken.RefreshToken)
	}

	// Test error cases
	_, err = TokenFromFile("/nonexistent/path/token.json")
	if err == nil {
		t.Error("Expected error when reading non-existent token file, got nil")
	}
}

func TestOpenURL(t *testing.T) {
	originalExec := ExecCommandFunc
	defer func() {
		ExecCommandFunc = originalExec
	}()

	var capturedCommand string
	var capturedArgs []string

	ExecCommandFunc = func(command string, args ...string) *exec.Cmd {
		capturedCommand = command
		capturedArgs = args
		return originalExec("echo", "test")
	}

	testURL := "http://example.com"
	err := openURL(testURL)
	if err != nil {
		t.Fatalf("openURL returned error: %v", err)
	}

	switch runtime.GOOS {
	case "linux":
		if capturedCommand != "xdg-open" {
			t.Errorf("Expected 'xdg-open' command, got '%s'", capturedCommand)
		}
		if len(capturedArgs) != 1 || capturedArgs[0] != testURL {
			t.Errorf("Expected argument '%s', got '%v'", testURL, capturedArgs)
		}
	case "darwin":
		if capturedCommand != "open" {
			t.Errorf("Expected 'open' command, got '%s'", capturedCommand)
		}
		if len(capturedArgs) != 1 || capturedArgs[0] != testURL {
			t.Errorf("Expected argument '%s', got '%v'", testURL, capturedArgs)
		}
	case "windows":
		if capturedCommand != "rundll32" {
			t.Errorf("Expected 'rundll32' command, got '%s'", capturedCommand)
		}
	}
}

func TestTokenCacheFileWithName(t *testing.T) {
	tests := []struct {
		name         string
		tokenName    string
		wantContains string
	}{
		{
			name:         "default token name",
			tokenName:    "youtube-go.json",
			wantContains: "youtube-go.json",
		},
		{
			name:         "custom token name",
			tokenName:    "gdrive-go.json",
			wantContains: "gdrive-go.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := TokenCacheFileWithName(tt.tokenName)
			if err != nil {
				t.Errorf("TokenCacheFileWithName(%q) returned error: %v", tt.tokenName, err)
				return
			}
			if !strings.Contains(path, tt.wantContains) {
				t.Errorf("TokenCacheFileWithName(%q) = %q, want path containing %q", tt.tokenName, path, tt.wantContains)
			}
			if !strings.Contains(path, ".credentials") {
				t.Errorf("TokenCacheFileWithName(%q) = %q, want path containing .credentials", tt.tokenName, path)
			}
		})
	}
}

func TestGetClient_MissingCredentials(t *testing.T) {
	cfg := OAuthConfig{
		CredentialsFile: "/nonexistent/client_secret.json",
		TokenFileName:   "test-token.json",
		CallbackPort:    8099,
		Scopes:          []string{"https://www.googleapis.com/auth/drive.file"},
	}

	_, err := GetClient(t.Context(), cfg)
	if err == nil {
		t.Error("Expected error for missing credentials file, got nil")
	}
	if !strings.Contains(err.Error(), "unable to read client secret file") {
		t.Errorf("Expected 'unable to read client secret file' error, got: %v", err)
	}
}

func TestGetClient_InvalidCredentials(t *testing.T) {
	tempDir := t.TempDir()
	credFile := filepath.Join(tempDir, "client_secret.json")
	os.WriteFile(credFile, []byte("not json"), 0600)

	cfg := OAuthConfig{
		CredentialsFile: credFile,
		TokenFileName:   "test-token.json",
		CallbackPort:    8099,
		Scopes:          []string{"https://www.googleapis.com/auth/drive.file"},
	}

	_, err := GetClient(t.Context(), cfg)
	if err == nil {
		t.Error("Expected error for invalid credentials file, got nil")
	}
	if !strings.Contains(err.Error(), "unable to parse client secret file") {
		t.Errorf("Expected 'unable to parse client secret file' error, got: %v", err)
	}
}
