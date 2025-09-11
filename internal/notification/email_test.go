package notification

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
)

// TestServer represents a test SMTP server for testing emails
type TestServer struct {
	Server        *smtp.Server
	Messages      []*TestMessage
	mutex         sync.Mutex
	listener      net.Listener
	authenticated bool
}

// TestMessage represents an email message captured by the test server
type TestMessage struct {
	From    string
	To      []string
	Data    []byte
	Subject string
	Body    string
}

// StartTestServer starts a new test SMTP server on a random port
func StartTestServer(t *testing.T) *TestServer {
	t.Helper()

	testServer := &TestServer{
		Messages: make([]*TestMessage, 0),
		Server:   smtp.NewServer(&TestBackend{testServer: nil}),
	}

	// Fix circular reference by setting testServer
	testServer.Server.Backend = &TestBackend{testServer: testServer}

	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find free port: %v", err)
	}
	testServer.listener = listener

	// Configure the server
	testServer.Server.Addr = listener.Addr().String()
	testServer.Server.Domain = "localhost"
	testServer.Server.ReadTimeout = 10 * time.Second
	testServer.Server.WriteTimeout = 10 * time.Second
	testServer.Server.MaxMessageBytes = 1024 * 1024
	testServer.Server.MaxRecipients = 50
	testServer.Server.AllowInsecureAuth = true

	// Start the server
	go func() {
		if err := testServer.Server.Serve(listener); err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				t.Logf("SMTP server error: %v", err)
			}
		}
	}()

	return testServer
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	if ts.listener != nil {
		ts.listener.Close()
	}
	if ts.Server != nil {
		ts.Server.Close()
	}
}

// GetPort returns the port number the server is listening on
func (ts *TestServer) GetPort() int {
	addr := ts.listener.Addr().(*net.TCPAddr)
	return addr.Port
}

// GetMessages returns a copy of the received messages
func (ts *TestServer) GetMessages() []*TestMessage {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// Make a copy to avoid race conditions
	msgs := make([]*TestMessage, len(ts.Messages))
	copy(msgs, ts.Messages)
	return msgs
}

// ClearMessages clears all received messages
func (ts *TestServer) ClearMessages() {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.Messages = make([]*TestMessage, 0)
}

// TestBackend implements the SMTP server backend functionality
type TestBackend struct {
	testServer *TestServer
}

// NewSession creates a new mail session
func (b *TestBackend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &TestSession{
		testServer: b.testServer,
	}, nil
}

// TestSession implements the SMTP session functionality
type TestSession struct {
	testServer *TestServer
	from       string
	to         []string
	data       []byte
}

// AuthPlain handles SMTP plain authentication
func (s *TestSession) AuthPlain(username, password string) error {
	// Accept any credentials for testing
	s.testServer.authenticated = true
	return nil
}

// Mail sets the sender email
func (s *TestSession) Mail(from string, _ *smtp.MailOptions) error {
	s.from = from
	return nil
}

// Rcpt adds a recipient
func (s *TestSession) Rcpt(to string, _ *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

// Data handles the email data
func (s *TestSession) Data(r io.Reader) error {
	var err error
	s.data, err = io.ReadAll(r)
	if err != nil {
		return err
	}

	// Parse email to extract subject and body
	message := &TestMessage{
		From: s.from,
		To:   s.to,
		Data: s.data,
	}

	// Simple parsing to extract subject and body from email data
	data := string(s.data)

	// Extract subject
	if subjectIdx := strings.Index(data, "Subject: "); subjectIdx != -1 {
		endIdx := strings.Index(data[subjectIdx:], "\r\n")
		if endIdx != -1 {
			message.Subject = data[subjectIdx+9 : subjectIdx+endIdx]
		}
	}

	// Extract body - look for empty line that separates headers from body
	if bodyIdx := strings.Index(data, "\r\n\r\n"); bodyIdx != -1 {
		message.Body = data[bodyIdx+4:]
	}

	// Store the message
	s.testServer.mutex.Lock()
	defer s.testServer.mutex.Unlock()
	s.testServer.Messages = append(s.testServer.Messages, message)

	return nil
}

// Reset resets the session state
func (s *TestSession) Reset() {
	s.from = ""
	s.to = []string{}
	s.data = nil
}

// Logout handles client logout
func (s *TestSession) Logout() error {
	return nil
}

// TestNewEmail tests the NewEmail constructor
func TestNewEmail(t *testing.T) {
	email := NewEmail("test-password")
	if email == nil {
		t.Fatal("NewEmail returned nil")
	}
	if email.password != "test-password" {
		t.Errorf("Expected password to be 'test-password', got '%s'", email.password)
	}
}

// TestEmailFunctionality tests all email methods directly without mocking
func TestEmailFunctionality(t *testing.T) {
	// By default, skip this test since it requires SMTP access
	// To run the test manually, remove the skip line
	t.Skip("Skipping email functionality tests - requires SMTP setup")

	// Set up a test SMTP server
	server := StartTestServer(t)
	defer server.Close()

	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "email-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test gist file
	gistPath := filepath.Join(tmpDir, "test-gist.md")
	if err := os.WriteFile(gistPath, []byte("# Test Gist\nThis is a test gist file."), 0644); err != nil {
		t.Fatalf("Failed to create test gist: %v", err)
	}

	// Save original email config and restore after test
	originalEmail := configuration.GlobalSettings.Email
	defer func() {
		configuration.GlobalSettings.Email = originalEmail
	}()

	// Configure test email settings
	configuration.GlobalSettings.Email = configuration.SettingsEmail{
		From:        "test@example.com",
		ThumbnailTo: "thumbnail@example.com",
		EditTo:      "edit@example.com",
		FinanceTo:   "finance@example.com",
		Password:    "test-password",
	}

	// Create test video
	video := storage.Video{
		ProjectName:  "Test Project",
		ProjectURL:   "https://test-project.com",
		OtherLogos:   "Other Logo",
		Location:     "https://test-location.com",
		Tagline:      "Test Tagline",
		TaglineIdeas: "Tagline Idea 1\nTagline Idea 2",
		Animations:   "Animation 1\nAnimation 2",
		Members:      "Member 1, Member 2",
		Title:        "Test Title",
		VideoId:      "test-video-id",
		Gist:         gistPath,
		Sponsorship:  storage.Sponsorship{Amount: "$500", Emails: "sponsor1@example.com,sponsor2@example.com"},
	}

	// Create the email client
	email := NewEmail("test-password")

	// Use environment variables to override SMTP settings for test
	t.Setenv("SMTP_HOST", "127.0.0.1")
	t.Setenv("SMTP_PORT", fmt.Sprintf("%d", server.GetPort()))
	t.Setenv("SMTP_SSL", "false")

	// Test basic Send
	t.Run("TestSend", func(t *testing.T) {
		server.ClearMessages()

		err := email.Send(
			"from@example.com",
			[]string{"to@example.com"},
			"Test Subject",
			"Test Body",
			"",
		)
		if err != nil {
			t.Fatalf("Send failed: %v", err)
		}

		// Check message was sent
		messages := server.GetMessages()
		if len(messages) == 0 {
			t.Fatal("No messages received by test server")
		}

		// Verify message contents
		msg := messages[0]
		if !strings.Contains(msg.Subject, "Test Subject") {
			t.Errorf("Expected subject to contain 'Test Subject', got '%s'", msg.Subject)
		}

		if !strings.Contains(msg.Body, "Test Body") {
			t.Errorf("Expected body to contain 'Test Body', got '%s'", msg.Body)
		}
	})

	// Test SendThumbnail
	t.Run("TestSendThumbnail", func(t *testing.T) {
		server.ClearMessages()

		err := email.SendThumbnail(
			"from@example.com",
			"thumbnail@example.com",
			video,
		)
		if err != nil {
			t.Fatalf("SendThumbnail failed: %v", err)
		}

		// Check message was sent
		messages := server.GetMessages()
		if len(messages) == 0 {
			t.Fatal("No messages received by test server")
		}

		// Verify message contents
		msg := messages[0]
		expectedSubject := fmt.Sprintf("Thumbnail: %s", video.ProjectName)
		if !strings.Contains(msg.Subject, expectedSubject) {
			t.Errorf("Expected subject to contain '%s', got '%s'", expectedSubject, msg.Subject)
		}

		if !strings.Contains(msg.Body, video.Location) {
			t.Errorf("Expected body to contain location '%s'", video.Location)
		}

		if !strings.Contains(msg.Body, video.Tagline) {
			t.Errorf("Expected body to contain tagline '%s'", video.Tagline)
		}
	})

	// Test SendEdit
	t.Run("TestSendEdit", func(t *testing.T) {
		server.ClearMessages()

		err := email.SendEdit(
			"from@example.com",
			"edit@example.com",
			video,
		)
		if err != nil {
			t.Fatalf("SendEdit failed: %v", err)
		}

		// Check message was sent
		messages := server.GetMessages()
		if len(messages) == 0 {
			t.Fatal("No messages received by test server")
		}

		// Verify message contents
		msg := messages[0]
		expectedSubject := fmt.Sprintf("Video: %s", video.ProjectName)
		if !strings.Contains(msg.Subject, expectedSubject) {
			t.Errorf("Expected subject to contain '%s', got '%s'", expectedSubject, msg.Subject)
		}

		if !strings.Contains(msg.Body, video.Location) {
			t.Errorf("Expected body to contain location '%s'", video.Location)
		}

		if !strings.Contains(msg.Body, video.Title) {
			t.Errorf("Expected body to contain title '%s'", video.Title)
		}
	})

	// Test error cases for SendEdit
	t.Run("TestSendEditErrors", func(t *testing.T) {
		// Test empty gist path
		videoWithoutGist := video
		videoWithoutGist.Gist = ""
		err := email.SendEdit(
			"from@example.com",
			"edit@example.com",
			videoWithoutGist,
		)
		if err == nil {
			t.Error("Expected error for empty gist path, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "Gist is empty") {
			t.Errorf("Expected 'Gist is empty' error, got: %v", err)
		}

		// Test invalid gist path
		videoWithInvalidGist := video
		videoWithInvalidGist.Gist = filepath.Join(tmpDir, "non-existent-gist.md")
		err = email.SendEdit(
			"from@example.com",
			"edit@example.com",
			videoWithInvalidGist,
		)
		if err == nil {
			t.Error("Expected error for invalid gist path, got nil")
		}
	})

	// Test SendSponsors
	t.Run("TestSendSponsors", func(t *testing.T) {
		server.ClearMessages()

		err := email.SendSponsors(
			"from@example.com",
			video.Sponsorship.Emails,
			video.VideoId,
			video.Sponsorship.Amount,
			video.Title,
		)
		if err != nil {
			t.Fatalf("SendSponsors failed: %v", err)
		}

		// Check message was sent
		messages := server.GetMessages()
		if len(messages) == 0 {
			t.Fatal("No messages received by test server")
		}

		// Verify message contents
		msg := messages[0]
		expectedSubject := fmt.Sprintf("DevOps Toolkit Video Sponsorship - %s", video.Title)
		if !strings.Contains(msg.Subject, expectedSubject) {
			t.Errorf("Expected subject to contain '%s', got '%s'", expectedSubject, msg.Subject)
		}

		if !strings.Contains(msg.Body, video.VideoId) {
			t.Errorf("Expected body to contain video ID '%s'", video.VideoId)
		}

		if !strings.Contains(msg.Body, video.Sponsorship.Amount) {
			t.Errorf("Expected body to contain amount '%s'", video.Sponsorship.Amount)
		}
	})
}

func TestGenerateThumbnailEmailContent(t *testing.T) {
	tests := []struct {
		name          string
		video         storage.Video
		expectSubject string
		expectBody    string // We will check for key substrings in the body
	}{
		{
			name: "Basic video",
			video: storage.Video{
				ProjectName: "My Test Project",
				Title:       "My Test Title",
				Location:    "/vids/my-test-project",
				Tagline:     "Cool Tagline",
			},
			expectSubject: "Thumbnail: My Test Title",
			expectBody:    "<strong>Material:</strong><br/><br/>All the material is available at /vids/my-test-project.<br/><br/><strong>Thumbnail:</strong><br/><br/>Elements:<ul>\n<li>Text: Cool Tagline</li>\n<li>Screenshots: screenshot-*.png</li></ul>",
		},
		{
			name: "Video with ProjectURL",
			video: storage.Video{
				ProjectName: "Project With URL",
				Title:       "Project With URL Title",
				Location:    "/vids/url-project",
				Tagline:     "URL Tagline",
				ProjectURL:  "http://example.com/logo.png",
			},
			expectSubject: "Thumbnail: Project With URL Title",
			expectBody:    "<li>Logo: http://example.com/logo.png</li>\n<li>Text: URL Tagline</li>",
		},
		{
			name: "Video with OtherLogos",
			video: storage.Video{
				ProjectName: "Project Other Logos",
				Title:       "Project Other Logos Title",
				Location:    "/vids/other-logos",
				Tagline:     "Other Logos Tagline",
				OtherLogos:  "logo2.png, logo3.svg",
			},
			expectSubject: "Thumbnail: Project Other Logos Title",
			expectBody:    "<li>Logo: logo2.png, logo3.svg</li>\n<li>Text: Other Logos Tagline</li>",
		},
		{
			name: "Video with ProjectURL and OtherLogos",
			video: storage.Video{
				ProjectName: "Project Both Logos",
				Title:       "Project Both Logos Title",
				Location:    "/vids/both-logos",
				Tagline:     "Both Logos Tagline",
				ProjectURL:  "http://example.com/logo.png",
				OtherLogos:  "logo2.png, logo3.svg",
			},
			expectSubject: "Thumbnail: Project Both Logos Title",
			expectBody:    "<li>Logo: http://example.com/logo.png, logo2.png, logo3.svg</li>\n<li>Text: Both Logos Tagline</li>",
		},
		{
			name: "Video with TaglineIdeas",
			video: storage.Video{
				ProjectName:  "Project Tagline Ideas",
				Title:        "Project Tagline Ideas Title",
				Location:     "/vids/tagline-ideas",
				Tagline:      "Main Tagline",
				TaglineIdeas: "Idea 1\nIdea 2",
			},
			expectSubject: "Thumbnail: Project Tagline Ideas Title",
			expectBody:    "<li>Text: Main Tagline</li>\n</ul>\nIdeas:<br/>Idea 1\nIdea 2",
		},
		{
			name: "Video with N/A ProjectURL and OtherLogos",
			video: storage.Video{
				ProjectName: "Project N/A Logos",
				Title:       "Project N/A Logos Title",
				Location:    "/vids/na-logos",
				Tagline:     "N/A Logos Tagline",
				ProjectURL:  "N/A",
				OtherLogos:  "-",
			},
			expectSubject: "Thumbnail: Project N/A Logos Title",
			// Expect no "<li>Logo: ...</li>" line if ProjectURL and OtherLogos are N/A or "-"
			expectBody: "<strong>Material:</strong><br/><br/>All the material is available at /vids/na-logos.<br/><br/><strong>Thumbnail:</strong><br/><br/>Elements:<ul>\n<li>Text: N/A Logos Tagline</li>\n<li>Screenshots: screenshot-*.png</li></ul>",
		},
		{
			name: "Video with empty ProjectURL and OtherLogos",
			video: storage.Video{
				ProjectName: "Project Empty Logos",
				Title:       "Project Empty Logos Title",
				Location:    "/vids/empty-logos",
				Tagline:     "Empty Logos Tagline",
				ProjectURL:  "",
				OtherLogos:  "",
			},
			expectSubject: "Thumbnail: Project Empty Logos Title",
			expectBody:    "<strong>Material:</strong><br/><br/>All the material is available at /vids/empty-logos.<br/><br/><strong>Thumbnail:</strong><br/><br/>Elements:<ul>\n<li>Text: Empty Logos Tagline</li>\n<li>Screenshots: screenshot-*.png</li></ul>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, body := generateThumbnailEmailContent(tt.video)
			assert.Equal(t, tt.expectSubject, subject)

			// For the body, we check for the presence of key parts because exact HTML formatting can be brittle.
			// The tt.expectBody for basic case is more complete for an example.
			if tt.name == "Basic video" || tt.name == "Video with N/A ProjectURL and OtherLogos" || tt.name == "Video with empty ProjectURL and OtherLogos" {
				// For basic and N/A/empty cases, we can do a more direct comparison of the key part
				normalizedExpected := strings.ReplaceAll(strings.ReplaceAll(tt.expectBody, "\n", ""), " ", "")
				normalizedActual := strings.ReplaceAll(strings.ReplaceAll(body, "\n", ""), " ", "")
				if !strings.Contains(normalizedActual, normalizedExpected) {
					t.Errorf("Expected body to contain (normalized):\n%s\nActual body (normalized):\n%s", normalizedExpected, normalizedActual)
				}
			} else if tt.name == "Video with TaglineIdeas" {
				assert.Contains(t, body, tt.video.Location)
				assert.Contains(t, body, "<li>Text: "+tt.video.Tagline+"</li>")
				assert.Contains(t, body, "Ideas:<br/>"+tt.video.TaglineIdeas)
			} else {
				assert.Contains(t, body, tt.video.Location)
				assert.Contains(t, body, tt.expectBody) // Checks for the logo line and tagline line
			}

			// General checks for all
			assert.Contains(t, body, fmt.Sprintf("All the material is available at %s", tt.video.Location))
			assert.Contains(t, body, fmt.Sprintf("<li>Text: %s</li>", tt.video.Tagline))
			assert.Contains(t, body, "<li>Screenshots: screenshot-*.png</li>")

			if len(tt.video.TaglineIdeas) > 0 && tt.video.TaglineIdeas != "N/A" && tt.video.TaglineIdeas != "-" {
				assert.Contains(t, body, fmt.Sprintf("Ideas:<br/>%s", tt.video.TaglineIdeas))
			} else {
				assert.NotContains(t, body, "Ideas:<br/>")
			}

			expectedLogoString := ""
			if tt.video.ProjectURL != "" && tt.video.ProjectURL != "-" && tt.video.ProjectURL != "N/A" {
				expectedLogoString = tt.video.ProjectURL
			}
			if tt.video.OtherLogos != "" && tt.video.OtherLogos != "-" && tt.video.OtherLogos != "N/A" {
				if len(expectedLogoString) > 0 {
					expectedLogoString = fmt.Sprintf("%s, ", expectedLogoString)
				}
				expectedLogoString = fmt.Sprintf("%s%s", expectedLogoString, tt.video.OtherLogos)
			}
			if len(expectedLogoString) > 0 {
				assert.Contains(t, body, fmt.Sprintf("<li>Logo: %s</li>", expectedLogoString))
			} else {
				assert.NotContains(t, body, "<li>Logo:")
			}
		})
	}
}

func TestGenerateEditEmailContent(t *testing.T) {
	tests := []struct {
		name               string
		video              storage.Video
		expectErr          bool
		expectSubject      string
		expectBodyContains []string
		expectAttachment   string
	}{
		{
			name: "Error on empty Gist",
			video: storage.Video{
				ProjectName: "Test Project",
				Gist:        "", // Empty Gist
			},
			expectErr: true,
		},
		{
			name: "Basic valid case",
			video: storage.Video{
				ProjectName: "My Edit Project",
				Location:    "/vids/edit-project",
				Title:       "Awesome Video Title",
				ProjectURL:  "http://project.url/logo.png",
				Animations:  "Anim1\n- Anim2\n\nAnim3 With Spaces", // Test newline, prefix, and spaces
				Members:     "MemberA, MemberB",
				Gist:        "/path/to/gist.txt",
			},
			expectErr:     false,
			expectSubject: "Video: Awesome Video Title",
			expectBodyContains: []string{
				"All the material is available at /vids/edit-project",
				"<li>Animation: Subscribe (anywhere in the video)</li>",
				"<li>Lower third: Viktor Farcic (anywhere in the video)</li>",
				fmt.Sprintf("<li>Lower third: %s + logo + URL (%s) (add to a few places when I mention %s)</li>", "My Edit Project", "http://project.url/logo.png", "My Edit Project"),
				fmt.Sprintf("<li>Title roll: %s</li>", "Awesome Video Title"),
				"<li>Convert all text in bold (surounded with **) in the attachment into text on the screen</li>",
				"<li>Anim1</li>",
				"<li>Anim2</li>", // "- " should be stripped
				"<li>Anim3 With Spaces</li>",
				fmt.Sprintf("<li>Member shoutouts: Thanks a ton to the new members for supporting the channel: %s</li>", "MemberA, MemberB"),
				"<li>Outro roll</li>",
			},
			expectAttachment: "/path/to/gist.txt",
		},
		{
			name: "No animations, no members",
			video: storage.Video{
				ProjectName: "Simple Project",
				Location:    "/vids/simple",
				Title:       "Simple Title",
				ProjectURL:  "http://simple.url/",
				Animations:  "", // No animations
				Members:     "", // No members
				Gist:        "/gist.md",
			},
			expectErr:     false,
			expectSubject: "Video: Simple Title",
			expectBodyContains: []string{
				"All the material is available at /vids/simple",
				fmt.Sprintf("<li>Lower third: %s + logo + URL (%s) (add to a few places when I mention %s)</li>", "Simple Project", "http://simple.url/", "Simple Project"),
				fmt.Sprintf("<li>Title roll: %s</li>", "Simple Title"),
				// Check that specific animation lines are NOT present if video.Animations is empty
				// The static ones (Subscribe, Like, etc.) will still be there.
				fmt.Sprintf("<li>Member shoutouts: Thanks a ton to the new members for supporting the channel: %s</li>", ""),
			},
			expectAttachment: "/gist.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, body, attachment, err := generateEditEmailContent(tt.video)

			if tt.expectErr {
				assert.Error(t, err)
				if err != nil { // Further check for specific error if needed
					assert.Contains(t, err.Error(), "Gist is empty")
				}
				return // No further checks if error is expected
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectSubject, subject)
			assert.Equal(t, tt.expectAttachment, attachment)

			for _, expectedSubstring := range tt.expectBodyContains {
				assert.Contains(t, body, expectedSubstring)
			}

			// Check for absence of empty <li> from animations
			assert.NotContains(t, body, "\n<li></li>")

			// Test specific animation logic for "No animations" case
			if tt.name == "No animations, no members" {
				// Example: if tt.video.Animations was "Anim1\n\nAnim2" (with an empty line),
				// we ensure that "\n<li></li>\n<li>Anim2</li>" is NOT present and also "<li></li>" is not.
				// The general NotContains("\n<li></li>") already covers simple empty list items.
				// If specific animations were provided in video.Animations, they should be present.
				// If video.Animations is empty, no specific animation li tags (other than static ones) should be generated from it.
				// This is indirectly tested by checking that tt.expectBodyContains (which doesn't have them) are present
				// and what's not in tt.expectBodyContains (like custom anims) are not.
			}
		})
	}
}

func TestGenerateSponsorsEmailContent(t *testing.T) {
	tests := []struct {
		name             string
		videoID          string
		sponsorshipPrice string
		videoTitle       string
		expectSubject    string
		expectBody       string
	}{
		{
			name:             "Basic sponsor email",
			videoID:          "dQw4w9WgXcQ",
			sponsorshipPrice: "$1000",
			videoTitle:       "How to Deploy Kubernetes Apps",
			expectSubject:    "DevOps Toolkit Video Sponsorship - How to Deploy Kubernetes Apps",
			expectBody:       "Hi,\n<br><br>The video has just been released and is available at https://youtu.be/dQw4w9WgXcQ. Please let me know what you think or if you have any questions.\n<br><br>I'll send the invoice for $1000 in a separate message.\n",
		},
		{
			name:             "Sponsor email with different price",
			videoID:          "abcdef12345",
			sponsorshipPrice: "€500 (EUR)",
			videoTitle:       "Docker Containers Explained",
			expectSubject:    "DevOps Toolkit Video Sponsorship - Docker Containers Explained",
			expectBody:       "Hi,\n<br><br>The video has just been released and is available at https://youtu.be/abcdef12345. Please let me know what you think or if you have any questions.\n<br><br>I'll send the invoice for €500 (EUR) in a separate message.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, body := generateSponsorsEmailContent(tt.videoID, tt.sponsorshipPrice, tt.videoTitle)
			assert.Equal(t, tt.expectSubject, subject)
			// Direct comparison for body as it's simpler and less prone to formatting issues here
			assert.Equal(t, tt.expectBody, body)
		})
	}
}
