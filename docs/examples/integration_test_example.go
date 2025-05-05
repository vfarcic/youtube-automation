package examples

import (
	"os"
	"path/filepath"
	"testing"
)

// VideoProcessor handles video processing workflows
type VideoProcessor struct {
	YouTubeAPI    YouTubeAPI
	EmailNotifier EmailNotifier
}

// EmailNotifier interface for sending notifications
type EmailNotifier interface {
	SendNotification(recipient, subject, body string) error
}

// MockEmailNotifier for testing
type MockEmailNotifier struct {
	SendCallCount int
	LastRecipient string
	LastSubject   string
	LastBody      string
	SendError     error
}

// SendNotification implements the EmailNotifier interface
func (m *MockEmailNotifier) SendNotification(recipient, subject, body string) error {
	m.SendCallCount++
	m.LastRecipient = recipient
	m.LastSubject = subject
	m.LastBody = body
	return m.SendError
}

// ProcessAndPublishVideo processes and publishes a video
func (p *VideoProcessor) ProcessAndPublishVideo(videoConfigPath, recipientEmail string) (string, error) {
	// Read video configuration
	_, err := os.ReadFile(videoConfigPath)
	if err != nil {
		return "", err
	}

	// Parse video configuration (simplified for example)
	title := "Example Video"
	description := "This is an example video"
	tags := []string{"example", "test"}
	category := "22"
	videoPath := filepath.Join(filepath.Dir(videoConfigPath), "video.mp4")
	thumbnailPath := filepath.Join(filepath.Dir(videoConfigPath), "thumbnail.jpg")

	// Create metadata
	metadata := VideoMetadata{
		Title:       title,
		Description: description,
		Tags:        tags,
		Category:    category,
	}

	// Upload to YouTube
	videoID, err := PublishVideo(p.YouTubeAPI, metadata, videoPath, thumbnailPath)
	if err != nil {
		return "", err
	}

	// Send notification email
	notificationSubject := "New Video Published: " + title
	notificationBody := "Your video has been published to YouTube with ID: " + videoID

	err = p.EmailNotifier.SendNotification(recipientEmail, notificationSubject, notificationBody)
	if err != nil {
		// Note: We still return the video ID even if notification fails
		return videoID, err
	}

	return videoID, nil
}

// TestProcessAndPublishVideo demonstrates an integration test
func TestProcessAndPublishVideo(t *testing.T) {
	// Setup test directory and files
	tempDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test video config file
	configPath := filepath.Join(tempDir, "video.yaml")
	err = os.WriteFile(configPath, []byte(`
title: Test Integration
description: Testing the integration between components
tags:
  - test
  - integration
category: 22
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create dummy video and thumbnail files
	videoPath := filepath.Join(tempDir, "video.mp4")
	err = os.WriteFile(videoPath, []byte("dummy video data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test video: %v", err)
	}

	thumbnailPath := filepath.Join(tempDir, "thumbnail.jpg")
	err = os.WriteFile(thumbnailPath, []byte("dummy thumbnail data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test thumbnail: %v", err)
	}

	// Create mocks for both components
	mockYouTube := NewMockYouTubeAPI()
	mockEmailer := &MockEmailNotifier{}

	// Create the processor with mocks
	processor := VideoProcessor{
		YouTubeAPI:    mockYouTube,
		EmailNotifier: mockEmailer,
	}

	// Call the integrated function
	videoID, err := processor.ProcessAndPublishVideo(configPath, "user@example.com")

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if videoID != "mock-video-id-123" {
		t.Errorf("Expected video ID 'mock-video-id-123', got: %s", videoID)
	}

	// Verify YouTube API was called
	if !mockYouTube.UploadCalled {
		t.Error("Expected YouTube upload to be called")
	}

	// Verify email notification was sent
	if mockEmailer.SendCallCount != 1 {
		t.Errorf("Expected 1 email notification, got: %d", mockEmailer.SendCallCount)
	}

	if mockEmailer.LastRecipient != "user@example.com" {
		t.Errorf("Expected email recipient 'user@example.com', got: %s", mockEmailer.LastRecipient)
	}

	// The test verifies that both components work together correctly
	// and the workflow executes from start to finish
}
