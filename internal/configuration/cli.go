package configuration

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// osExit is a variable to allow mocking os.Exit in tests
var osExit = os.Exit

var RootCmd = &cobra.Command{
	Use:   "youtube-release",
	Short: "youtube-release is a super fancy CLI for releasing YouTube videos.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

// requiredFlags collects flag names that must be set in CLI mode.
var requiredFlags []string

// markRequired records a flag as required (validated at run-time, not at init-time).
func markRequired(name string) {
	requiredFlags = append(requiredFlags, name)
}

type Settings struct {
	Email          SettingsEmail          `yaml:"email"`
	AI             SettingsAI             `yaml:"ai"`
	YouTube        SettingsYouTube        `yaml:"youtube"`
	Hugo           SettingsHugo           `yaml:"hugo"`
	Bluesky        SettingsBluesky        `yaml:"bluesky"`
	VideoDefaults  SettingsVideoDefaults  `yaml:"videoDefaults"`
	Slack          SettingsSlack          `yaml:"slack"`
	Timing         TimingConfig           `yaml:"timing"`
	Calendar       SettingsCalendar       `yaml:"calendar"`
	Shorts         ShortsConfig           `yaml:"shorts"`
API            SettingsAPI            `yaml:"api"`
	Git            SettingsGit            `yaml:"git"`
	GDrive         SettingsGDrive         `yaml:"gdrive"`
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

func init() {
	// Load settings from YAML file
	yamlFile, err := os.ReadFile("settings.yaml")
	if err == nil {
		if err := yaml.Unmarshal(yamlFile, &GlobalSettings); err != nil {
			fmt.Printf("Error parsing config file: %s\n", err)
		}
	}

	// Define command-line flags
	RootCmd.Flags().StringVar(&GlobalSettings.Email.From, "email-from", GlobalSettings.Email.From, "From which email to send messages. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Email.ThumbnailTo, "email-thumbnail-to", GlobalSettings.Email.ThumbnailTo, "To which email to send requests for thumbnails. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Email.EditTo, "email-edit-to", GlobalSettings.Email.EditTo, "To which email to send requests for edits. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Email.FinanceTo, "email-finance-to", GlobalSettings.Email.FinanceTo, "To which email to send emails related to finances. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Email.Password, "email-password", GlobalSettings.Email.Password, "Email server password. Environment variable `EMAIL_PASSWORD` is supported as well. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Provider, "ai-provider", GlobalSettings.AI.Provider, "AI provider (azure or anthropic). Environment variable `AI_PROVIDER` is supported as well. Defaults to anthropic.")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Azure.Endpoint, "ai-endpoint", GlobalSettings.AI.Azure.Endpoint, "AI endpoint. For Azure OpenAI. (required for azure)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Azure.Key, "ai-key", GlobalSettings.AI.Azure.Key, "AI key. Environment variable `AI_KEY` is supported as well. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Azure.Deployment, "ai-deployment", GlobalSettings.AI.Azure.Deployment, "AI Deployment. For Azure OpenAI. (required for azure)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Azure.APIVersion, "ai-api-version", GlobalSettings.AI.Azure.APIVersion, "Azure OpenAI API Version (e.g., 2023-05-15). Defaults to a common version if not set.")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Anthropic.Key, "anthropic-key", GlobalSettings.AI.Anthropic.Key, "Anthropic API key. Environment variable `ANTHROPIC_API_KEY` is supported as well. (required for anthropic)")
	RootCmd.Flags().StringVar(&GlobalSettings.AI.Anthropic.Model, "anthropic-model", GlobalSettings.AI.Anthropic.Model, "Anthropic model (e.g., claude-sonnet-4-20250514). Environment variable `ANTHROPIC_MODEL` is supported as well.")
	RootCmd.Flags().StringVar(&GlobalSettings.YouTube.APIKey, "youtube-api-key", GlobalSettings.YouTube.APIKey, "YouTube API key. Environment variable `YOUTUBE_API_KEY` is supported as well. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Hugo.Path, "hugo-path", GlobalSettings.Hugo.Path, "Path to the repo with Hugo posts. (required)")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.Identifier, "bluesky-identifier", GlobalSettings.Bluesky.Identifier, "Bluesky username/identifier (e.g., username.bsky.social)")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.Password, "bluesky-password", GlobalSettings.Bluesky.Password, "Bluesky password. Environment variable `BLUESKY_PASSWORD` is supported as well.")
	RootCmd.Flags().StringVar(&GlobalSettings.Bluesky.URL, "bluesky-url", GlobalSettings.Bluesky.URL, "Bluesky API URL")
	RootCmd.Flags().StringVar(&GlobalSettings.VideoDefaults.Language, "video-defaults-language", "", "Default language for videos (e.g., 'en', 'es')")
	RootCmd.Flags().StringVar(&GlobalSettings.VideoDefaults.AudioLanguage, "video-defaults-audio-language", "", "Default audio language for videos (e.g., 'en', 'es')")
	RootCmd.Flags().BoolVar(&GlobalSettings.Calendar.Disabled, "calendar-disabled", GlobalSettings.Calendar.Disabled, "Disable Google Calendar event creation after video upload")

	// Default Bluesky URL if not set by file or flag
	if GlobalSettings.Bluesky.URL == "" {
		GlobalSettings.Bluesky.URL = "https://bsky.social/xrpc"
	}

	// Added for PRD: Automated Video Language Setting
	// Default video language if not set by file or flag (flag default is 'en')
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

	// Calendar settings: enabled by default, set calendar.disabled: true to disable

	// Check required fields and environment variables
	if GlobalSettings.Email.From == "" {
		markRequired("email-from")
	}
	if GlobalSettings.Email.ThumbnailTo == "" {
		markRequired("email-thumbnail-to")
	}
	if GlobalSettings.Email.EditTo == "" {
		markRequired("email-edit-to")
	}
	if GlobalSettings.Email.FinanceTo == "" {
		markRequired("email-finance-to")
	}

	// Check environment variables
	if envPassword := os.Getenv("EMAIL_PASSWORD"); envPassword != "" {
		GlobalSettings.Email.Password = envPassword
	} else if GlobalSettings.Email.Password == "" {
		markRequired("email-password")
	}

	// Override AI provider from environment variable
	if envAIProvider := os.Getenv("AI_PROVIDER"); envAIProvider != "" {
		GlobalSettings.AI.Provider = envAIProvider
	}

	// Default to anthropic provider
	if GlobalSettings.AI.Provider == "" {
		GlobalSettings.AI.Provider = "anthropic"
	}

	// Provider-specific validation
	switch GlobalSettings.AI.Provider {
	case "azure":
		if GlobalSettings.AI.Azure.Endpoint == "" {
			markRequired("ai-endpoint")
		}

		if envAIKey := os.Getenv("AI_KEY"); envAIKey != "" {
			GlobalSettings.AI.Azure.Key = envAIKey
		} else if GlobalSettings.AI.Azure.Key == "" {
			markRequired("ai-key")
		}

		if GlobalSettings.AI.Azure.Deployment == "" {
			markRequired("ai-deployment")
		}

		// Default API version if not set
		if GlobalSettings.AI.Azure.APIVersion == "" {
			GlobalSettings.AI.Azure.APIVersion = "2023-05-15" // Defaulting to a common version
		}

	case "anthropic":
		if envAnthropicKey := os.Getenv("ANTHROPIC_API_KEY"); envAnthropicKey != "" {
			GlobalSettings.AI.Anthropic.Key = envAnthropicKey
		} else if GlobalSettings.AI.Anthropic.Key == "" {
			markRequired("anthropic-key")
		}

		if envModel := os.Getenv("ANTHROPIC_MODEL"); envModel != "" {
			GlobalSettings.AI.Anthropic.Model = envModel
		} else if GlobalSettings.AI.Anthropic.Model == "" {
			GlobalSettings.AI.Anthropic.Model = "claude-sonnet-4-20250514" // Default model
		}

	default:
		fmt.Printf("Unsupported AI provider: %s. Supported providers: azure, anthropic\n", GlobalSettings.AI.Provider)
		osExit(1)
	}

	if envYouTubeKey := os.Getenv("YOUTUBE_API_KEY"); envYouTubeKey != "" {
		GlobalSettings.YouTube.APIKey = envYouTubeKey
	} else if GlobalSettings.YouTube.APIKey == "" {
		markRequired("youtube-api-key")
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

	// Path is only required when repoURL is not set (local mode)
	if GlobalSettings.Hugo.Path == "" && GlobalSettings.Hugo.RepoURL == "" {
		markRequired("hugo-path")
	}

	// Bluesky validation
	if GlobalSettings.Bluesky.Identifier != "" {
		envBlueskyPassword := os.Getenv("BLUESKY_PASSWORD")
		if envBlueskyPassword != "" {
			GlobalSettings.Bluesky.Password = envBlueskyPassword
		} else if GlobalSettings.Bluesky.Password == "" {
			markRequired("bluesky-password")
		}
	} else if envBlueskyPassword := os.Getenv("BLUESKY_PASSWORD"); envBlueskyPassword != "" {
		GlobalSettings.Bluesky.Password = envBlueskyPassword
	}

	// Validate required flags only when running in CLI mode (not serve subcommand).
	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip validation for subcommands (e.g. serve)
		if cmd.Name() != RootCmd.Name() {
			return nil
		}
		for _, name := range requiredFlags {
			f := RootCmd.Flags().Lookup(name)
			if f == nil {
				continue
			}
			if !f.Changed {
				return fmt.Errorf("required flag \"%s\" not set", name)
			}
		}
		return nil
	}
}

func GetArgs() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing the CLI '%s'", err)
		osExit(1)
	}
}
