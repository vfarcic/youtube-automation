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
	Text          string
	YouTubeURL    string
	VideoID       string
	ThumbnailPath string
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

// blobLink represents the reference to an uploaded blob
type blobLink struct {
	Link string `json:"$link"`
}

// blobRef represents a reference to an uploaded image blob
type blobRef struct {
	Type     string   `json:"$type"`
	Ref      blobLink `json:"ref"`
	MimeType string   `json:"mimeType"`
	Size     int64    `json:"size"`
}

// externalEmbed represents an external link embed
type externalEmbed struct {
	Type     string       `json:"$type"`
	External externalLink `json:"external"`
}

type externalLink struct {
	URI         string   `json:"uri"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Thumb       *blobRef `json:"thumb,omitempty"`
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

// uploadBlobResponse represents the response from uploading a blob
type uploadBlobResponse struct {
	Blob blobRef `json:"blob"`
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

// uploadThumbnail uploads an image from a URL to Bluesky and returns a blob reference
func uploadThumbnail(config Config, session *Session, thumbnailPath string) (*blobRef, error) {
	// 1. Read the image from the local file path
	file, err := os.Open(thumbnailPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open thumbnail file %s: %w", thumbnailPath, err)
	}
	defer file.Close()

	// Read the image data into memory
	imageData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read thumbnail image data: %w", err)
	}

	// 2. Detect the MIME type (use http.DetectContentType for simplicity)
	mimeType := http.DetectContentType(imageData)
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("downloaded file is not a recognized image type: %s", mimeType)
	}

	// 3. Send a POST request to uploadBlob endpoint
	uploadURL := config.URL + "/com.atproto.repo.uploadBlob"

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	// 4. Set headers
	req.Header.Set("Content-Type", mimeType)
	req.Header.Set("Authorization", "Bearer "+session.AccessJWT)

	// Execute the request
	client := &http.Client{}
	uploadResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute upload request: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(uploadResp.Body)
		return nil, fmt.Errorf("thumbnail upload failed with status %d: %s", uploadResp.StatusCode, string(bodyBytes))
	}

	// 5. Parse the JSON response
	var blobResp uploadBlobResponse
	if err := json.NewDecoder(uploadResp.Body).Decode(&blobResp); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	// 6. Return the blobRef struct (make sure it's a pointer)
	return &blobResp.Blob, nil
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

	// Prepare embed data
	var embed *externalEmbed
	if post.YouTubeURL != "" {
		embed = &externalEmbed{
			Type: "app.bsky.embed.external",
			External: externalLink{
				URI:         post.YouTubeURL,
				Title:       "YouTube Video",
				Description: text,
			},
		}

		// Construct thumbnail URL and attempt upload
		if post.VideoID != "" {
			if post.ThumbnailPath != "" {
				thumbBlob, err := uploadThumbnail(config, session, post.ThumbnailPath)
				if err != nil {
					// Log warning but continue without thumbnail
					fmt.Printf("Warning: Failed to upload Bluesky thumbnail from path %s: %v\n", post.ThumbnailPath, err)
				} else {
					embed.External.Thumb = thumbBlob
				}
			} else {
				fmt.Printf("Warning: No local thumbnail path provided for video %s, skipping Bluesky thumbnail.\n", post.VideoID)
			}
		}
	}
	postData.Record.Embed = embed

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
func SendPost(config Config, text string, videoID string, thumbnailPath string) error {
	// Validate input
	if !strings.Contains(text, "[YOUTUBE]") {
		return fmt.Errorf("text does not contain [YOUTUBE] placeholder")
	}

	if videoID == "" {
		return fmt.Errorf("YouTube video ID is required")
	}

	if thumbnailPath == "" {
		return fmt.Errorf("Thumbnail path is required for Bluesky post")
	}

	// Calculate final text length with YouTube URL instead of placeholder
	youtubeUrl := fmt.Sprintf("https://youtu.be/%s", videoID)
	finalText := strings.ReplaceAll(text, "[YOUTUBE]", youtubeUrl)

	if len(finalText) > 300 {
		return fmt.Errorf("text exceeds Bluesky's 300 character limit")
	}

	post := Post{
		Text:          finalText,
		YouTubeURL:    youtubeUrl,
		VideoID:       videoID,
		ThumbnailPath: thumbnailPath,
	}

	postURL, err := CreatePost(config, post)
	if err != nil {
		return err
	}

	// Print the URL to the Bluesky post
	fmt.Println("Posted to Bluesky:", postURL)

	return nil
}
