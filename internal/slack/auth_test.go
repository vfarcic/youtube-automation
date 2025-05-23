package slack

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewSlackAuth(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
		errText string // Expected error text if wantErr is true
	}{
		{
			name:    "valid bot token",
			token:   botTokenPrefix + strings.Repeat("a", 45), // xoxb- + 45 'a's
			wantErr: false,
		},
		{
			name:    "valid user token",
			token:   userTokenPrefix + strings.Repeat("b", 45), // xoxp- + 45 'b's
			wantErr: false,
		},
		{
			name:    "token too short (less than minTokenLength)",
			token:   "xoxb-short",
			wantErr: true,
			errText: "invalid Slack token: token too short",
		},
		{
			name:    "token with invalid prefix",
			token:   "xoxa-" + strings.Repeat("c", 45),
			wantErr: true,
			errText: "invalid Slack token: must start with '" + botTokenPrefix + "' or '" + userTokenPrefix + "'",
		},
		{
			name: "token with correct prefix but overall too short",
			// minTokenLength is 45, prefix is 5 chars, so needs 40 more. This has prefix + 30.
			token:   botTokenPrefix + strings.Repeat("d", 30),
			wantErr: true,
			errText: "invalid Slack token: token too short",
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errText: "invalid Slack token: token too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewSlackAuth(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSlackAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("NewSlackAuth() error = %v, want errText containing %q", err, tt.errText)
				}
			} else {
				if auth == nil {
					t.Errorf("NewSlackAuth() expected auth instance, got nil")
				} else if auth.Token != tt.token {
					t.Errorf("NewSlackAuth() token = %q, want %q", auth.Token, tt.token)
				}
			}
		})
	}
}

func TestSlackAuth_GetToken(t *testing.T) {
	expectedToken := "test-token"
	auth := &SlackAuth{Token: expectedToken}
	if got := auth.GetToken(); got != expectedToken {
		t.Errorf("SlackAuth.GetToken() = %v, want %v", got, expectedToken)
	}
}

func TestSlackAuth_ValidateToken(t *testing.T) {
	type mockResponse struct {
		status int
		body   string        // JSON string or malformed
		delay  time.Duration // To simulate timeouts
	}

	tests := []struct {
		name          string
		token         string        // Token to set in SlackAuth
		mockResp      *mockResponse // If nil, server is not started (for empty token test or network error simulation)
		wantErr       bool
		wantErrorText string // Substring to check in error message
	}{
		{
			name:  "valid token",
			token: "xoxb-valid-token",
			mockResp: &mockResponse{
				status: http.StatusOK,
				body:   `{"ok": true, "team": "T123", "user": "U456"}`,
			},
			wantErr: false,
		},
		{
			name:  "invalid token - invalid_auth",
			token: "xoxb-invalid-token",
			mockResp: &mockResponse{
				status: http.StatusOK,
				body:   `{"ok": false, "error": "invalid_auth"}`,
			},
			wantErr:       true,
			wantErrorText: "slack token validation failed: invalid_auth",
		},
		{
			name:  "invalid token - token_revoked with team/user",
			token: "xoxb-revoked-token",
			mockResp: &mockResponse{
				status: http.StatusOK,
				body:   `{"ok": false, "error": "token_revoked", "team": "TXYZ", "user": "UABC"}`,
			},
			wantErr:       true,
			wantErrorText: "slack token validation failed: token_revoked (Team: TXYZ, User: UABC)",
		},
		{
			name:  "http non-200 status",
			token: "xoxb-http-error-token",
			mockResp: &mockResponse{
				status: http.StatusServiceUnavailable,
				body:   `Service Unavailable`, // Not JSON
			},
			wantErr:       true,
			wantErrorText: "auth.test request failed with status: 503 Service Unavailable",
		},
		{
			name:  "malformed json response",
			token: "xoxb-malformed-json-token",
			mockResp: &mockResponse{
				status: http.StatusOK,
				body:   `{"ok": true, "error": "missing_quote_ทำให้_json_invalid`,
			},
			wantErr:       true,
			wantErrorText: "failed to decode auth.test response",
		},
		{
			name:  "empty token string",
			token: "", // Auth object will have empty token
			// mockResp is nil, no server needed
			wantErr:       true,
			wantErrorText: "token is empty",
		},
		{
			name:  "network error (server not running)",
			token: "xoxb-network-error-token",
			// mockResp is nil, but we'll try to make a request to a non-existent server by setting authTestURL temporarily
			wantErr:       true,
			wantErrorText: "failed to perform auth.test request", // This will vary based on OS, "connection refused" etc.
		},
		{
			name:  "request timeout",
			token: "xoxb-timeout-token",
			mockResp: &mockResponse{
				status: http.StatusOK,
				body:   `{"ok": true}`,
				delay:  (requestTimeout + 1*time.Second), // Delay longer than client timeout
			},
			wantErr:       true,
			wantErrorText: "failed to perform auth.test request", // Error will likely be context.DeadlineExceeded
		},
	}

	originalAuthTestURL := authTestURL // Save original for restoration

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.mockResp != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tt.mockResp.delay > 0 {
						time.Sleep(tt.mockResp.delay)
					}
					// Check auth header
					authHeader := r.Header.Get("Authorization")
					expectedHeader := "Bearer " + tt.token
					if authHeader != expectedHeader {
						t.Errorf("Expected Authorization header %q, got %q", expectedHeader, authHeader)
						// Still write the response to allow the test to continue if header check isn't the primary focus
					}
					w.WriteHeader(tt.mockResp.status)
					fmt.Fprintln(w, tt.mockResp.body)
				}))
				defer server.Close()
				authTestURL = server.URL // Point our client to the mock server
			} else {
				if tt.name == "network error (server not running)" {
					// For a true network error, point to an invalid address or a closed port
					// Using the httptest.Server mechanism with a nil mockResp is tricky for this specific case.
					// Instead, we'll rely on the default authTestURL which might not be running,
					// or ideally, use a specific non-routable address if possible.
					// For this test, we ensure no server is started.
					authTestURL = "http://localhost:12345" // A port that is likely not in use
				}
			}

			auth := &SlackAuth{Token: tt.token}
			err := auth.ValidateToken()

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || (tt.wantErrorText != "" && !strings.Contains(err.Error(), tt.wantErrorText)) {
					t.Errorf("ValidateToken() error = %q, want error containing %q", err, tt.wantErrorText)
				}
			}
			authTestURL = originalAuthTestURL // Restore for next test
		})
	}
}

func TestLoadTokenFromEnv(t *testing.T) {
	tests := []struct {
		name          string
		envVarName    string
		setEnvValue   *string // Pointer to distinguish not set (nil) from empty string ("")
		wantToken     string
		wantErr       bool
		wantErrorText string
	}{
		{
			name:        "env var set with valid token",
			envVarName:  "TEST_SLACK_TOKEN_VALID",
			setEnvValue: func() *string { s := "xoxb-env-token-123"; return &s }(),
			wantToken:   "xoxb-env-token-123",
			wantErr:     false,
		},
		{
			name:          "env var not set",
			envVarName:    "TEST_SLACK_TOKEN_UNSET",
			setEnvValue:   nil, // Do not set this env var
			wantErr:       true,
			wantErrorText: "Slack token not found in environment variable \"TEST_SLACK_TOKEN_UNSET\"",
		},
		{
			name:          "env var set to empty string",
			envVarName:    "TEST_SLACK_TOKEN_EMPTY",
			setEnvValue:   func() *string { s := ""; return &s }(),
			wantErr:       true,
			wantErrorText: "Slack token not found in environment variable \"TEST_SLACK_TOKEN_EMPTY\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Manage environment variable for the test
			if tt.setEnvValue != nil {
				t.Setenv(tt.envVarName, *tt.setEnvValue)
			} else {
				// Ensure it's not set from a previous run or environment
				// t.Setenv(tt.envVarName, "") // This sets it to empty, not unsets.
				// For robust unsetting, one might os.Unsetenv, but t.Setenv should handle new test runs cleanly.
				// If a test relies on it truly being absent, ensure no default is picked up.
				// The current logic of LoadTokenFromEnv correctly handles empty string as "not found".
			}

			gotToken, err := LoadTokenFromEnv(tt.envVarName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTokenFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || (tt.wantErrorText != "" && !strings.Contains(err.Error(), tt.wantErrorText)) {
					t.Errorf("LoadTokenFromEnv() error = %q, want error containing %q", err, tt.wantErrorText)
				}
			} else {
				if gotToken != tt.wantToken {
					t.Errorf("LoadTokenFromEnv() gotToken = %q, want %q", gotToken, tt.wantToken)
				}
			}
		})
	}
}

func TestLoadTokenFromConfig(t *testing.T) {
	validConfigContent := `{"slack_token": "xoxb-file-token-456"}`
	missingTokenFieldContent := `{"other_field": "some_value"}`
	emptyTokenFieldContent := `{"slack_token": ""}`
	malformedJSONContent := `{"slack_token": "xoxb-bad-json,`

	tests := []struct {
		name          string
		setupFile     func(t *testing.T) (configPath string, cleanup func()) // Returns path and cleanup func
		wantToken     string
		wantErr       bool
		wantErrorText string
	}{
		{
			name: "valid config file",
			setupFile: func(t *testing.T) (string, func()) {
				return createTestConfigFile(t, "valid.json", validConfigContent)
			},
			wantToken: "xoxb-file-token-456",
			wantErr:   false,
		},
		{
			name: "config file not found",
			setupFile: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.json"), func() {} // No file created, no cleanup needed
			},
			wantErr:       true,
			wantErrorText: "failed to read Slack config file",
		},
		{
			name: "malformed JSON in config file",
			setupFile: func(t *testing.T) (string, func()) {
				return createTestConfigFile(t, "malformed.json", malformedJSONContent)
			},
			wantErr:       true,
			wantErrorText: "failed to unmarshal Slack config",
		},
		{
			name: "slack_token field missing",
			setupFile: func(t *testing.T) (string, func()) {
				return createTestConfigFile(t, "missing_field.json", missingTokenFieldContent)
			},
			wantErr:       true,
			wantErrorText: "slack_token not found or empty",
		},
		{
			name: "slack_token field empty",
			setupFile: func(t *testing.T) (string, func()) {
				return createTestConfigFile(t, "empty_field.json", emptyTokenFieldContent)
			},
			wantErr:       true,
			wantErrorText: "slack_token not found or empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, cleanup := tt.setupFile(t)
			defer cleanup()

			gotToken, err := LoadTokenFromConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTokenFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || (tt.wantErrorText != "" && !strings.Contains(err.Error(), tt.wantErrorText)) {
					t.Errorf("LoadTokenFromConfig() error = %q, want error containing %q", err, tt.wantErrorText)
				}
			} else {
				if gotToken != tt.wantToken {
					t.Errorf("LoadTokenFromConfig() gotToken = %q, want %q", gotToken, tt.wantToken)
				}
			}
		})
	}
}

// Helper function to create a temporary config file for testing
func createTestConfigFile(t *testing.T, fileName string, content string) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, fileName)
	err := os.WriteFile(configPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config file %q: %v", configPath, err)
	}
	return configPath, func() { /* os.RemoveAll(tmpDir) is handled by t.TempDir() */ }
}
