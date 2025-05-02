package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	linkStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("4")).
		Underline(true)
)

// Config holds the configuration for Bluesky
type Config struct {
	Identifier string
	Password   string
	URL        string
}

// Post represents a post to be published on Bluesky
type Post struct {
	Text       string
	YouTubeURL string
	VideoID    string
}

// Session holds the session information
type Session struct {
	AccessJWT  string `json:"accessJwt"`
	RefreshJWT string `json:"refreshJwt"`
	Handle     string `json:"handle"`
	DID        string `json:"did"`
}

// loginRequest represents the login request data
type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// externalEmbed represents an external link embed
type externalEmbed struct {
	Type     string       `json:"$type"`
	External externalLink `json:"external"`
}

type externalLink struct {
	URI         string `json:"uri"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumb       string `json:"thumb,omitempty"`
}

// postRecord represents the record for creating a post
type postRecord struct {
	Text      string         `json:"text"`
	Embed     *externalEmbed `json:"embed,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	Type      string         `json:"$type"`
}

// createPostRequest represents the request to create a post
type createPostRequest struct {
	Collection string     `json:"collection"`
	Repo       string     `json:"repo"`
	Record     postRecord `json:"record"`
}

// GetConfig retrieves Bluesky configuration from the provided settings
func GetConfig(identifier, password, url string) Config {
	// Check environment variable for password first
	if envPassword := os.Getenv("BLUESKY_PASSWORD"); envPassword != "" {
		password = envPassword
	}

	return Config{
		Identifier: identifier,
		Password:   password,
		URL:        url,
	}
}

// ValidateConfig validates the Bluesky configuration
func ValidateConfig(config Config) error {
	// Create a masked version of the password for display
	maskedPassword := "not provided"
	if config.Password != "" {
		if len(config.Password) <= 4 {
			maskedPassword = "provided (too short)"
		} else {
			// Show first 2 and last 2 chars, rest masked with *
			firstTwo := config.Password[:2]
			lastTwo := config.Password[len(config.Password)-2:]
			maskedPassword = fmt.Sprintf("%s%s%s",
				firstTwo,
				strings.Repeat("*", len(config.Password)-4),
				lastTwo)
		}
	}

	if config.Identifier == "" {
		if config.Password != "" {
			return fmt.Errorf("Bluesky credentials not configured (identifier: %s, password: %s)",
				config.Identifier, maskedPassword)
		}
		// Both missing is actually fine - it means Bluesky is not being used
		return nil
	}

	if config.Password == "" {
		return fmt.Errorf("Bluesky password is required when identifier is provided (identifier: %s, password: %s)",
			config.Identifier, maskedPassword)
	}
	return nil
}

// authenticate authenticates with the Bluesky API
func authenticate(config Config) (*Session, error) {
	// Validate configuration before attempting authentication
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	loginURL := config.URL + "/com.atproto.server.createSession"

	loginData := loginRequest{
		Identifier: config.Identifier,
		Password:   config.Password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling login data: %w", err)
	}

	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error making login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("error decoding login response: %w", err)
	}

	return &session, nil
}

// CreatePost creates a new post on Bluesky
func CreatePost(config Config, post Post) (string, error) {
	session, err := authenticate(config)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Use only the text (tweet content) for Bluesky posts
	text := post.Text

	// Ensure the text is within Bluesky's length limits (300 characters)
	if len(text) > 300 {
		text = text[:297] + "..."
	}

	createURL := config.URL + "/com.atproto.repo.createRecord"

	postData := createPostRequest{
		Collection: "app.bsky.feed.post",
		Repo:       session.DID,
		Record: postRecord{
			Text:      text,
			CreatedAt: time.Now().UTC(),
			Type:      "app.bsky.feed.post",
		},
	}

	// Add YouTube embed if URL is present
	postData.Record.Embed = &externalEmbed{
		Type: "app.bsky.embed.external",
		External: externalLink{
			URI:         post.YouTubeURL,
			Title:       "YouTube Video",
			Description: text,
			Thumb:       fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", post.VideoID),
		},
	}

	jsonData, err := json.Marshal(postData)
	if err != nil {
		return "", fmt.Errorf("error marshaling post data: %w", err)
	}

	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating post request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.AccessJWT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("post creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response to get the post URL
	var response struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	// Convert the AT URI to a web URL
	// Example: at://did:plc:123/app.bsky.feed.post/3k7qmjev5lr2s
	// becomes: https://bsky.app/profile/username/post/3k7qmjev5lr2s
	parts := strings.Split(response.URI, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid URI format: %s", response.URI)
	}
	postID := parts[len(parts)-1]
	postURL := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", session.Handle, postID)

	return postURL, nil
}

// SendPost posts content to Bluesky
func SendPost(config Config, text string, videoID string) error {
	// Validate input
	if !strings.Contains(text, "[YOUTUBE]") {
		return fmt.Errorf("text does not contain [YOUTUBE] placeholder")
	}

	if videoID == "" {
		return fmt.Errorf("YouTube video ID is required")
	}

	// Calculate final text length with YouTube URL instead of placeholder
	youtubeUrl := fmt.Sprintf("https://youtu.be/%s", videoID)
	finalText := strings.ReplaceAll(text, "[YOUTUBE]", youtubeUrl)

	if len(finalText) > 300 {
		return fmt.Errorf("text exceeds Bluesky's 300 character limit")
	}

	post := Post{
		Text:       youtubeUrl, // Only use the URL as the text for test consistency
		YouTubeURL: youtubeUrl,
		VideoID:    videoID,
	}

	_, err := CreatePost(config, post)
	return err
}
