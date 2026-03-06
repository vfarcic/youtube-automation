package publishing

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"devopstoolkit/youtube-automation/internal/auth"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const channelID = "UCfz8x0lVzJpb_dgWm9kPVrw"

// OAuthConfig holds OAuth configuration for a YouTube channel
type OAuthConfig struct {
	CredentialsFile string // Path to client_secret file
	TokenFileName   string // Name of token cache file (stored in ~/.credentials/)
	CallbackPort    int    // Port for OAuth callback
}

// DefaultOAuthConfig returns the config for the main English channel
func DefaultOAuthConfig() OAuthConfig {
	return OAuthConfig{
		CredentialsFile: "client_secret.json",
		TokenFileName:   "youtube-go.json",
		CallbackPort:    8090,
	}
}

// SpanishOAuthConfig returns the config for the Spanish channel from settings
func SpanishOAuthConfig() OAuthConfig {
	cfg := configuration.GlobalSettings.SpanishChannel
	credFile := cfg.CredentialsFile
	if credFile == "" {
		credFile = "client_secret_spanish.json"
	}
	tokenFile := cfg.TokenFile
	if tokenFile == "" {
		tokenFile = "youtube-go-spanish.json"
	}
	port := cfg.CallbackPort
	if port == 0 {
		port = 8091
	}
	return OAuthConfig{
		CredentialsFile: credFile,
		TokenFileName:   tokenFile,
		CallbackPort:    port,
	}
}

// youtubeScopes defines all OAuth2 scopes required for YouTube operations.
// These scopes are requested during the initial authentication and cached.
// All scopes must be included upfront to avoid re-authentication issues.
var youtubeScopes = []string{
	youtube.YoutubeUploadScope,                             // Upload videos and thumbnails
	youtube.YoutubeReadonlyScope,                           // Read video metadata
	youtube.YoutubeForceSslScope,                           // Access captions (list and download)
	"https://www.googleapis.com/auth/yt-analytics.readonly", // Access analytics data
}

// getClient uses a Context to retrieve a Token and generate a Client.
// It uses the centralized youtubeScopes for all YouTube operations.
func getClient(ctx context.Context) *http.Client {
	return getClientWithConfig(ctx, DefaultOAuthConfig())
}

// getClientWithConfig uses a Context and OAuthConfig to retrieve a Token and generate a Client.
// This allows different channels to use different credentials and callback ports.
func getClientWithConfig(ctx context.Context, oauthCfg OAuthConfig) *http.Client {
	authCfg := auth.OAuthConfig{
		CredentialsFile: oauthCfg.CredentialsFile,
		TokenFileName:   oauthCfg.TokenFileName,
		CallbackPort:    oauthCfg.CallbackPort,
		Scopes:          youtubeScopes,
	}
	client, err := auth.GetClient(ctx, authCfg)
	if err != nil {
		log.Fatalf("OAuth failed: %v", err)
	}
	return client
}

// GetSpanishChannelClient returns an authenticated HTTP client for the Spanish YouTube channel.
// It uses separate credentials and token cache from the main English channel.
func GetSpanishChannelClient(ctx context.Context) *http.Client {
	return getClientWithConfig(ctx, SpanishOAuthConfig())
}

// GetSpanishChannelID returns the configured Spanish channel ID.
func GetSpanishChannelID() string {
	return configuration.GlobalSettings.SpanishChannel.ChannelID
}

func UploadVideo(video *storage.Video) string {
	if video.UploadVideo == "" {
		log.Fatalf("You must provide a filename of a video file to upload")
		return ""
	}
	if video.Thumbnail == "" {
		log.Fatalf("You must provide a thumbnail of the video file to upload")
		return ""
	}
	client := getClient(context.Background())

	// FIXME: Remove the comment
	// service, err := youtube.New(client)
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	// service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}
	timecodes := ""
	if len(video.Timecodes) > 0 && video.Timecodes != "N/A" {
		timecodes = fmt.Sprintf("▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬\n%s", video.Timecodes)
	}

	// Construct Hugo URL from title and category for video description
	hugoURL := ""
	if video.GetUploadTitle() != "" && video.Gist != "" {
		category := GetCategoryFromFilePath(video.Gist)
		hugoURL = ConstructHugoURL(video.GetUploadTitle(), category)
	}

	// Build sponsor section if both name and URL are available and not "N/A"
	sponsorSection := ""
	if video.Sponsorship.Name != "" && video.Sponsorship.Name != "N/A" &&
	   video.Sponsorship.URL != "" && video.Sponsorship.URL != "N/A" {
		sponsorSection = fmt.Sprintf(`▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
Sponsor: %s
🔗 %s
▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬

`, video.Sponsorship.Name, video.Sponsorship.URL)
	}

	description := fmt.Sprintf(`%s

%s%s

Consider joining the channel: https://www.youtube.com/c/devopstoolkit/join

▬▬▬▬▬▬ 🔗 Additional Info 🔗 ▬▬▬▬▬▬
%s
▬▬▬▬▬▬ 💰 Sponsorships 💰 ▬▬▬▬▬▬
If you are interested in sponsoring this channel, please visit https://devopstoolkit.live/sponsor for more information. Alternatively, feel free to contact me over Twitter or LinkedIn (see below).

▬▬▬▬▬▬ 👋 Contact me 👋 ▬▬▬▬▬▬
➡ BlueSky: https://vfarcic.bsky.social
➡ LinkedIn: https://www.linkedin.com/in/viktorfarcic/

▬▬▬▬▬▬ 🚀 Other Channels 🚀 ▬▬▬▬▬▬
🎤 Podcast: https://www.devopsparadox.com/
💬 Live streams: https://www.youtube.com/c/DevOpsParadox

%s
`, video.Description, sponsorSection, video.DescriptionTags, GetAdditionalInfo(hugoURL, video.ProjectName, video.ProjectURL, video.RelatedVideos), timecodes)

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       video.GetUploadTitle(),
			Description: description,
			CategoryId:  "28",
			ChannelId:   channelID,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "private",
			PublishAt:     video.Date,
		},
		// MonetizationDetails: &youtube.VideoMonetizationDetails{
		// 	Access: &youtube.AccessPolicy{
		// 		Allowed: true,
		// 	},
		// },
	}
	// The API returns a 400 Bad Request response if tags is an empty string.
	if strings.Trim(video.Tags, "") != "" {
		upload.Snippet.Tags = strings.Split(video.Tags, ",")
	}

	// Determine languages to set
	finalDefaultLanguage := video.Language
	if finalDefaultLanguage == "" {
		finalDefaultLanguage = configuration.GlobalSettings.VideoDefaults.Language // Guaranteed non-empty by cli.go
	}

	finalDefaultAudioLanguage := video.AudioLanguage
	if finalDefaultAudioLanguage == "" {
		finalDefaultAudioLanguage = configuration.GlobalSettings.VideoDefaults.AudioLanguage // Guaranteed non-empty by cli.go
	}

	upload.Snippet.DefaultLanguage = finalDefaultLanguage
	upload.Snippet.DefaultAudioLanguage = finalDefaultAudioLanguage

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	file, err := os.Open(video.UploadVideo)
	if err != nil {
		log.Fatalf("Error opening %v: %v", video.UploadVideo, err)
	}

	response, err := call.Media(file).Do()
	file.Close()
	if err != nil {
		log.Fatalf("Error getting response from YouTube during insert: %v", err)
	}
	fmt.Printf("Upload successful! Video ID: %v\n", response.Id)

	// Save the applied languages back to the video struct
	video.AppliedLanguage = finalDefaultLanguage
	video.AppliedAudioLanguage = finalDefaultAudioLanguage
	log.Printf("DEBUG: Language %s and Audio Language %s stored in video struct for video ID %s", video.AppliedLanguage, video.AppliedAudioLanguage, response.Id)

	return response.Id
}

// GetAdditionalInfoFromPath converts a Hugo path to URL and calls GetAdditionalInfo
// This maintains backward compatibility for existing callers that pass Hugo paths
func GetAdditionalInfoFromPath(hugoPath, projectName, projectURL, relatedVideosRaw string) string {
	hugoURL := ""
	if len(hugoPath) > 0 {
		hugoPage := strings.ReplaceAll(hugoPath, "../", "")
		hugoPage = strings.ReplaceAll(hugoPage, "devopstoolkit-live/content/", "")
		hugoPage = strings.ReplaceAll(hugoPage, "/_index.md", "")
		hugoURL = fmt.Sprintf("https://devopstoolkit.live/%s", hugoPage)
	}
	return GetAdditionalInfo(hugoURL, projectName, projectURL, relatedVideosRaw)
}

func GetAdditionalInfo(hugoURL, projectName, projectURL, relatedVideosRaw string) string {
	relatedVideos := ""
	relatedVideosArray := strings.Split(relatedVideosRaw, "\n")
	for i := range relatedVideosArray {
		relatedVideosArray[i] = strings.TrimSpace(relatedVideosArray[i])
	}
	for i := range relatedVideosArray {
		if len(relatedVideosArray[i]) > 0 && relatedVideosArray[i] != "N/A" {
			relatedVideos = fmt.Sprintf("%s🎬 %s\n", relatedVideos, relatedVideosArray[i])
		}
	}
	gist := ""
	if len(hugoURL) > 0 {
		gist = fmt.Sprintf("➡ Transcript and commands: %s\n", hugoURL)
	}
	projectInfo := ""
	if projectName != "N/A" && projectURL != "N/A" {
		projectInfo = fmt.Sprintf("🔗 %s: %s\n", projectName, projectURL)
	}
	return fmt.Sprintf("%s%s%s", gist, projectInfo, relatedVideos)
}


func UploadThumbnail(videoId string, thumbnailPath string) error {
	client := getClient(context.Background())

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}
	file, err := os.Open(thumbnailPath)
	if err != nil {
		return err
	}
	defer file.Close()
	call := service.Thumbnails.Set(videoId)
	response, err := call.Media(file).Do()
	if err != nil {
		return err
	}
	fmt.Printf("Thumbnail uploaded, URL: %s\n", response.Items[0].Default.Url)
	return nil
}

func GetYouTubeURL(videoId string) string {
	return fmt.Sprintf("https://youtu.be/%s", videoId)
}

// videoUpdateDoer defines an interface for the Do() method of a video update call.
type videoUpdateDoer interface {
	Do(opts ...googleapi.CallOption) (*youtube.Video, error)
}

// videoServiceUpdater defines an interface for the Update() method of a video service.
type videoServiceUpdater interface {
	Update(part []string, video *youtube.Video) videoUpdateDoer
}

// youtubeServiceAdapter adapts *youtube.Service to the videoServiceUpdater interface.
type youtubeServiceAdapter struct {
	service *youtube.Service
}

// Update calls the underlying YouTube service's Videos.Update method.
func (a *youtubeServiceAdapter) Update(part []string, video *youtube.Video) videoUpdateDoer {
	return a.service.Videos.Update(part, video)
}

func updateVideoLanguage(updater videoServiceUpdater, videoID string, languageCode string, audioLanguageCode string) error {
	// Determine final language codes with fallbacks
	finalLangCode := languageCode
	if finalLangCode == "" {
		finalLangCode = configuration.GlobalSettings.VideoDefaults.Language // Guaranteed non-empty by cli.go
	}

	finalAudioLangCode := audioLanguageCode
	if finalAudioLangCode == "" {
		finalAudioLangCode = configuration.GlobalSettings.VideoDefaults.AudioLanguage // Guaranteed non-empty by cli.go
	}

	updateVideo := &youtube.Video{
		Id: videoID,
		Snippet: &youtube.VideoSnippet{
			DefaultLanguage:      finalLangCode,
			DefaultAudioLanguage: finalAudioLangCode,
		},
	}

	updateCall := updater.Update([]string{"snippet"}, updateVideo)
	_, err := updateCall.Do()
	return err
}

// UploadShort uploads a YouTube Short with scheduled publishing.
// The short's description includes a link back to the main video.
//
// Parameters:
//   - filePath: Path to the short video file
//   - short: Short metadata (title, scheduled date)
//   - mainVideoID: YouTube ID of the main video to link to
//
// Returns:
//   - string: The YouTube video ID of the uploaded short
//   - error: Any error that occurred during upload
func UploadShort(filePath string, short storage.Short, mainVideoID string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path is required")
	}
	if short.Title == "" {
		return "", fmt.Errorf("short title is required")
	}
	if short.ScheduledDate == "" {
		return "", fmt.Errorf("scheduled date is required")
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("video file does not exist: %s", filePath)
	}

	client := getClient(context.Background())
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("error creating YouTube client: %w", err)
	}

	// Build description with link to main video
	description := BuildShortDescription(short.Title, mainVideoID)

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       short.Title,
			Description: description,
			CategoryId:  "28", // Science & Technology
			ChannelId:   channelID,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "private",
			PublishAt:     short.ScheduledDate,
		},
	}

	// Set default language
	defaultLanguage := configuration.GlobalSettings.VideoDefaults.Language
	if defaultLanguage != "" {
		upload.Snippet.DefaultLanguage = defaultLanguage
	}
	defaultAudioLanguage := configuration.GlobalSettings.VideoDefaults.AudioLanguage
	if defaultAudioLanguage != "" {
		upload.Snippet.DefaultAudioLanguage = defaultAudioLanguage
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening video file %s: %w", filePath, err)
	}
	defer file.Close()

	response, err := call.Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("error uploading short to YouTube: %w", err)
	}

	fmt.Printf("Short uploaded successfully! Video ID: %v\n", response.Id)
	return response.Id, nil
}

// BuildShortDescription creates the description for a YouTube Short
// with a link back to the main video.
func BuildShortDescription(title string, mainVideoID string) string {
	mainVideoURL := ""
	if mainVideoID != "" {
		mainVideoURL = fmt.Sprintf("\nWatch the full video: %s\n", GetYouTubeURL(mainVideoID))
	}
	return fmt.Sprintf("%s%s\n#Shorts", title, mainVideoURL)
}
