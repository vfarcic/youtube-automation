package configuration

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadTimingRecommendations reads timing recommendations from the given settings file path.
// Returns an empty slice if no recommendations exist or if file doesn't exist.
func LoadTimingRecommendations(settingsPath string) ([]TimingRecommendation, error) {
	yamlFile, err := os.ReadFile(settingsPath)
	if err != nil {
		// If file doesn't exist, return empty slice (graceful handling)
		if os.IsNotExist(err) {
			return []TimingRecommendation{}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", settingsPath, err)
	}

	var settings Settings
	if err := yaml.Unmarshal(yamlFile, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings.yaml: %w", err)
	}

	// Return empty slice if recommendations is nil
	if settings.Timing.Recommendations == nil {
		return []TimingRecommendation{}, nil
	}

	return settings.Timing.Recommendations, nil
}

// SaveTimingRecommendations writes timing recommendations to the given settings file path.
// Preserves all other settings while updating only the timing section.
func SaveTimingRecommendations(settingsPath string, recommendations []TimingRecommendation) error {
	// Read existing settings
	yamlFile, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", settingsPath, err)
	}

	var settings Settings
	if err := yaml.Unmarshal(yamlFile, &settings); err != nil {
		return fmt.Errorf("failed to parse %s: %w", settingsPath, err)
	}

	// Update timing recommendations
	settings.Timing.Recommendations = recommendations

	// Write back to file
	yamlData, err := yaml.Marshal(&settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", settingsPath, err)
	}

	return nil
}

type Settings struct {
	Email               SettingsEmail               `yaml:"email"`
	AI                  SettingsAI                  `yaml:"ai"`
	YouTube             SettingsYouTube             `yaml:"youtube"`
	Hugo                SettingsHugo                `yaml:"hugo"`
	Bluesky             SettingsBluesky             `yaml:"bluesky"`
	VideoDefaults       SettingsVideoDefaults       `yaml:"videoDefaults"`
	Slack               SettingsSlack               `yaml:"slack"`
	Timing              TimingConfig                `yaml:"timing"`
	Calendar            SettingsCalendar            `yaml:"calendar"`
	Shorts              ShortsConfig                `yaml:"shorts"`
	API                 SettingsAPI                 `yaml:"api"`
	Git                 SettingsGit                 `yaml:"git"`
	GDrive              SettingsGDrive              `yaml:"gdrive"`
	ThumbnailGeneration SettingsThumbnailGeneration `yaml:"thumbnailGeneration"`
}

// SettingsThumbnailGeneration holds configuration for AI-powered thumbnail generation
type SettingsThumbnailGeneration struct {
	PhotoDir  string                     `yaml:"photoDir"`
	Providers []SettingsThumbnailProvider `yaml:"providers"`
}

// SettingsThumbnailProvider defines an image generation provider and its model
type SettingsThumbnailProvider struct {
	Name  string `yaml:"name"`
	Model string `yaml:"model"`
}

type SettingsAPI struct {
	Token string `yaml:"token"`
}

// SettingsGit holds git sync configuration for the data repository
type SettingsGit struct {
	RepoURL string `yaml:"repoURL"`
	Branch  string `yaml:"branch"`
	Token   string `yaml:"token"`
}

// SettingsGDrive holds Google Drive configuration
type SettingsGDrive struct {
	CredentialsFile string `yaml:"credentialsFile"` // Path to client_secret JSON file
	TokenFile       string `yaml:"tokenFile"`       // Token cache filename (default: gdrive-go.json)
	CallbackPort    int    `yaml:"callbackPort"`    // OAuth callback port (default: 8092)
	FolderID        string `yaml:"folderId"`        // Root Drive folder ID for uploads (optional)
}

type SettingsEmail struct {
	From        string `yaml:"from"`
	ThumbnailTo string `yaml:"thumbnailTo"`
	EditTo      string `yaml:"editTo"`
	FinanceTo   string `yaml:"financeTo"`
	Password    string `yaml:"password"`
}

type SettingsHugo struct {
	Path    string `yaml:"path"`    // Local path (CLI mode)
	RepoURL string `yaml:"repoURL"` // GitHub repo URL for PR workflow (e.g., https://github.com/user/repo.git)
	Branch  string `yaml:"branch"`  // Base branch (default: "main")
	Token   string `yaml:"token"`   // GitHub token for clone + PR creation
}

type SettingsAI struct {
	Provider  string              `yaml:"provider"`
	Azure     SettingsAzureAI     `yaml:"azure"`
	Anthropic SettingsAnthropicAI `yaml:"anthropic"`
}

type SettingsAzureAI struct {
	Key        string `yaml:"key"`
	Endpoint   string `yaml:"endpoint"`
	Deployment string `yaml:"deployment"`
	APIVersion string `yaml:"apiVersion,omitempty"`
}

type SettingsAnthropicAI struct {
	Key   string `yaml:"key"`
	Model string `yaml:"model"`
}

type SettingsYouTube struct {
	APIKey    string `yaml:"apiKey"`
	ChannelId string `yaml:"channelId"`
}

type SettingsBluesky struct {
	Identifier string `yaml:"identifier"`
	Password   string `yaml:"password"`
	URL        string `yaml:"url"`
}

type SettingsVideoDefaults struct {
	Language      string `yaml:"language"`
	AudioLanguage string `yaml:"audioLanguage"`
}

type SettingsSlack struct {
	TargetChannelIDs []string `yaml:"targetChannelIDs"`
}

// SettingsCalendar holds Google Calendar integration settings
type SettingsCalendar struct {
	Disabled bool `yaml:"disabled"` // Set to true to disable calendar event creation (enabled by default)
}

// TimingRecommendation represents a single timing recommendation
// for video publishing based on audience behavior and performance data
type TimingRecommendation struct {
	Day       string `yaml:"day" json:"day"`             // "Monday", "Tuesday", etc.
	Time      string `yaml:"time" json:"time"`           // "16:00", "09:00", etc. (UTC)
	Reasoning string `yaml:"reasoning" json:"reasoning"` // Why this slot recommended
}

// TimingConfig holds timing recommendations for video publishing
type TimingConfig struct {
	Recommendations []TimingRecommendation `yaml:"recommendations" json:"recommendations"`
}

// ShortsConfig holds configuration for YouTube Shorts identification and scheduling
type ShortsConfig struct {
	MaxWords       int `yaml:"maxWords" json:"maxWords"`             // Maximum word count for a Short segment (default: 150)
	CandidateCount int `yaml:"candidateCount" json:"candidateCount"` // Number of Short candidates to generate (default: 10)
}

var GlobalSettings Settings

// InitGlobalSettings loads configuration from settings.yaml and environment variables.
// It should be called once at application startup.
func InitGlobalSettings() error {
	// Load settings from YAML file
	settingsFile := "settings.yaml"
	if envPath := os.Getenv("SETTINGS_FILE"); envPath != "" {
		settingsFile = envPath
	}
	yamlFile, err := os.ReadFile(settingsFile)
	if err == nil {
		if err := yaml.Unmarshal(yamlFile, &GlobalSettings); err != nil {
			return fmt.Errorf("error parsing settings.yaml: %w", err)
		}
	}

	// Default Bluesky URL if not set by file
	if GlobalSettings.Bluesky.URL == "" {
		GlobalSettings.Bluesky.URL = "https://bsky.social/xrpc"
	}

	// Default video language if not set by file
	if GlobalSettings.VideoDefaults.Language == "" {
		GlobalSettings.VideoDefaults.Language = "en"
	}
	if GlobalSettings.VideoDefaults.AudioLanguage == "" {
		GlobalSettings.VideoDefaults.AudioLanguage = "en"
	}

	// Default Shorts settings
	if GlobalSettings.Shorts.MaxWords == 0 {
		GlobalSettings.Shorts.MaxWords = 150
	}
	if GlobalSettings.Shorts.CandidateCount == 0 {
		GlobalSettings.Shorts.CandidateCount = 10
	}

	// Git sync defaults — env vars override settings.yaml
	if envGitRepo := os.Getenv("GIT_REPO_URL"); envGitRepo != "" {
		GlobalSettings.Git.RepoURL = envGitRepo
	}
	if envGitBranch := os.Getenv("GIT_BRANCH"); envGitBranch != "" {
		GlobalSettings.Git.Branch = envGitBranch
	}
	if GlobalSettings.Git.Branch == "" {
		GlobalSettings.Git.Branch = "main"
	}
	if envGitToken := os.Getenv("GIT_TOKEN"); envGitToken != "" {
		GlobalSettings.Git.Token = envGitToken
	}

	// Google Drive — env vars override settings.yaml
	if envGDriveCredentials := os.Getenv("GDRIVE_CREDENTIALS_FILE"); envGDriveCredentials != "" {
		GlobalSettings.GDrive.CredentialsFile = envGDriveCredentials
	}
	if envGDriveToken := os.Getenv("GDRIVE_TOKEN_FILE"); envGDriveToken != "" {
		GlobalSettings.GDrive.TokenFile = envGDriveToken
	}
	if envGDriveFolderID := os.Getenv("GDRIVE_FOLDER_ID"); envGDriveFolderID != "" {
		GlobalSettings.GDrive.FolderID = envGDriveFolderID
	}

	// Email environment variables
	if envFrom := os.Getenv("EMAIL_FROM"); envFrom != "" {
		GlobalSettings.Email.From = envFrom
	}
	if envThumbnailTo := os.Getenv("EMAIL_THUMBNAIL_TO"); envThumbnailTo != "" {
		GlobalSettings.Email.ThumbnailTo = envThumbnailTo
	}
	if envEditTo := os.Getenv("EMAIL_EDIT_TO"); envEditTo != "" {
		GlobalSettings.Email.EditTo = envEditTo
	}
	if envFinanceTo := os.Getenv("EMAIL_FINANCE_TO"); envFinanceTo != "" {
		GlobalSettings.Email.FinanceTo = envFinanceTo
	}
	if envPassword := os.Getenv("EMAIL_PASSWORD"); envPassword != "" {
		GlobalSettings.Email.Password = envPassword
	}

	// Override AI provider from environment variable
	if envAIProvider := os.Getenv("AI_PROVIDER"); envAIProvider != "" {
		GlobalSettings.AI.Provider = envAIProvider
	}

	// Default to anthropic provider
	if GlobalSettings.AI.Provider == "" {
		GlobalSettings.AI.Provider = "anthropic"
	}

	// Provider-specific validation and environment variables
	switch GlobalSettings.AI.Provider {
	case "azure":
		if envAIKey := os.Getenv("AI_KEY"); envAIKey != "" {
			GlobalSettings.AI.Azure.Key = envAIKey
		}

		// Default API version if not set
		if GlobalSettings.AI.Azure.APIVersion == "" {
			GlobalSettings.AI.Azure.APIVersion = "2023-05-15"
		}

	case "anthropic":
		if envAnthropicKey := os.Getenv("ANTHROPIC_API_KEY"); envAnthropicKey != "" {
			GlobalSettings.AI.Anthropic.Key = envAnthropicKey
		}

		if envModel := os.Getenv("ANTHROPIC_MODEL"); envModel != "" {
			GlobalSettings.AI.Anthropic.Model = envModel
		} else if GlobalSettings.AI.Anthropic.Model == "" {
			GlobalSettings.AI.Anthropic.Model = "claude-sonnet-4-20250514"
		}

	default:
		return fmt.Errorf("unsupported AI provider: %s (supported: azure, anthropic)", GlobalSettings.AI.Provider)
	}

	if envYouTubeKey := os.Getenv("YOUTUBE_API_KEY"); envYouTubeKey != "" {
		GlobalSettings.YouTube.APIKey = envYouTubeKey
	}
	if envChannelID := os.Getenv("YOUTUBE_CHANNEL_ID"); envChannelID != "" {
		GlobalSettings.YouTube.ChannelId = envChannelID
	}

	// Hugo settings: env vars override settings.yaml
	if envHugoRepo := os.Getenv("HUGO_REPO_URL"); envHugoRepo != "" {
		GlobalSettings.Hugo.RepoURL = envHugoRepo
	}
	if envHugoBranch := os.Getenv("HUGO_BRANCH"); envHugoBranch != "" {
		GlobalSettings.Hugo.Branch = envHugoBranch
	}
	if GlobalSettings.Hugo.Branch == "" {
		GlobalSettings.Hugo.Branch = "main"
	}
	if envGitHubToken := os.Getenv("GITHUB_TOKEN"); envGitHubToken != "" && GlobalSettings.Hugo.Token == "" {
		GlobalSettings.Hugo.Token = envGitHubToken
	}

	// Bluesky: env vars override settings.yaml
	if envBlueskyIdentifier := os.Getenv("BLUESKY_IDENTIFIER"); envBlueskyIdentifier != "" {
		GlobalSettings.Bluesky.Identifier = envBlueskyIdentifier
	}
	if envBlueskyPassword := os.Getenv("BLUESKY_PASSWORD"); envBlueskyPassword != "" {
		GlobalSettings.Bluesky.Password = envBlueskyPassword
	}

	// Slack channel IDs: env var overrides settings.yaml (comma-separated list)
	if envSlackChannels := os.Getenv("SLACK_CHANNEL_IDS"); envSlackChannels != "" {
		channels := strings.Split(envSlackChannels, ",")
		GlobalSettings.Slack.TargetChannelIDs = make([]string, 0, len(channels))
		for _, ch := range channels {
			if trimmed := strings.TrimSpace(ch); trimmed != "" {
				GlobalSettings.Slack.TargetChannelIDs = append(GlobalSettings.Slack.TargetChannelIDs, trimmed)
			}
		}
	}

	return nil
}
