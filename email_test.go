package main

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

	"devopstoolkitseries/youtube-automation/internal/configuration"
	"devopstoolkitseries/youtube-automation/internal/storage"

	"github.com/emersion/go-smtp"
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
		if !strings.Contains(msg.Subject, "DevOps Toolkit Video Sponsorship") {
			t.Errorf("Expected subject to contain 'DevOps Toolkit Video Sponsorship', got '%s'", msg.Subject)
		}

		if !strings.Contains(msg.Body, video.VideoId) {
			t.Errorf("Expected body to contain video ID '%s'", video.VideoId)
		}

		if !strings.Contains(msg.Body, video.Sponsorship.Amount) {
			t.Errorf("Expected body to contain amount '%s'", video.Sponsorship.Amount)
		}
	})
}
