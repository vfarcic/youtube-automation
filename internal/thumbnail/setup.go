package thumbnail

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"
)

const (
	// DefaultStoreTTL is the default time-to-live for generated images in the store.
	DefaultStoreTTL = 10 * time.Minute

	// DefaultCleanupInterval is the default interval between store cleanup runs.
	DefaultCleanupInterval = 1 * time.Minute
)

// ProviderEnvKeys maps provider names to their API key environment variables.
var ProviderEnvKeys = map[string]string{
	"gemini":    "GEMINI_API_KEY",
	"gpt-image": "OPENAI_API_KEY",
}

// EnvLookupFunc is a function that looks up an environment variable.
// Defaults to os.Getenv but can be overridden for testing.
type EnvLookupFunc func(string) string

// CreateProviders creates ImageGenerator instances from the thumbnail generation config.
// Providers whose API key env var is not set are skipped with a warning log.
func CreateProviders(cfg configuration.SettingsThumbnailGeneration, envLookup EnvLookupFunc) []ImageGenerator {
	if envLookup == nil {
		envLookup = os.Getenv
	}

	var generators []ImageGenerator

	for _, p := range cfg.Providers {
		envKey, known := ProviderEnvKeys[p.Name]
		if !known {
			slog.Warn("unknown thumbnail provider, skipping", "provider", p.Name)
			continue
		}

		apiKey := envLookup(envKey)
		if apiKey == "" {
			slog.Warn("thumbnail provider API key not set, skipping", "provider", p.Name, "envVar", envKey)
			continue
		}

		gen, err := createProvider(p.Name, p.Model, apiKey)
		if err != nil {
			slog.Warn("failed to create thumbnail provider", "provider", p.Name, "error", err)
			continue
		}

		generators = append(generators, gen)
		slog.Info("thumbnail provider configured", "provider", p.Name, "model", p.Model)
	}

	return generators
}

func createProvider(name, model, apiKey string) (ImageGenerator, error) {
	switch name {
	case "gemini":
		return NewGeminiClient(apiKey, model, nil)
	case "gpt-image":
		return NewGPTImageClient(apiKey, model, nil)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

// StartCleanupLoop starts a background goroutine that periodically cleans up
// expired images from the store. It runs until the stop channel is closed.
func StartCleanupLoop(store *GeneratedImageStore, interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				removed := store.Cleanup()
				if removed > 0 {
					slog.Info("thumbnail store cleanup", "removed", removed)
				}
			case <-stop:
				return
			}
		}
	}()
}
