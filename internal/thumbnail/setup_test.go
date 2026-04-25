package thumbnail

import (
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"
)

func TestCreateProviders(t *testing.T) {
	tests := []struct {
		name       string
		cfg        configuration.SettingsThumbnailGeneration
		envVars    map[string]string
		wantCount  int
		wantNames  []string
	}{
		{
			name: "both providers configured with keys",
			cfg: configuration.SettingsThumbnailGeneration{
				Providers: []configuration.SettingsThumbnailProvider{
					{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
					{Name: "gpt-image", Model: "gpt-image-1"},
				},
			},
			envVars: map[string]string{
				"GEMINI_API_KEY": "test-gemini-key",
				"OPENAI_API_KEY": "test-openai-key",
			},
			wantCount: 2,
			wantNames: []string{"gemini", "gpt-image"},
		},
		{
			name: "gemini only with key",
			cfg: configuration.SettingsThumbnailGeneration{
				Providers: []configuration.SettingsThumbnailProvider{
					{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
					{Name: "gpt-image", Model: "gpt-image-1"},
				},
			},
			envVars: map[string]string{
				"GEMINI_API_KEY": "test-gemini-key",
			},
			wantCount: 1,
			wantNames: []string{"gemini"},
		},
		{
			name: "gpt-image only with key",
			cfg: configuration.SettingsThumbnailGeneration{
				Providers: []configuration.SettingsThumbnailProvider{
					{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
					{Name: "gpt-image", Model: "gpt-image-1"},
				},
			},
			envVars: map[string]string{
				"OPENAI_API_KEY": "test-openai-key",
			},
			wantCount: 1,
			wantNames: []string{"gpt-image"},
		},
		{
			name: "no API keys set",
			cfg: configuration.SettingsThumbnailGeneration{
				Providers: []configuration.SettingsThumbnailProvider{
					{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
					{Name: "gpt-image", Model: "gpt-image-1"},
				},
			},
			envVars:   map[string]string{},
			wantCount: 0,
			wantNames: nil,
		},
		{
			name:      "empty providers list",
			cfg:       configuration.SettingsThumbnailGeneration{},
			envVars:   map[string]string{},
			wantCount: 0,
			wantNames: nil,
		},
		{
			name: "unknown provider skipped",
			cfg: configuration.SettingsThumbnailGeneration{
				Providers: []configuration.SettingsThumbnailProvider{
					{Name: "unknown-provider", Model: "some-model"},
					{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
				},
			},
			envVars: map[string]string{
				"GEMINI_API_KEY": "test-gemini-key",
			},
			wantCount: 1,
			wantNames: []string{"gemini"},
		},
		{
			name: "single provider configured",
			cfg: configuration.SettingsThumbnailGeneration{
				Providers: []configuration.SettingsThumbnailProvider{
					{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
				},
			},
			envVars: map[string]string{
				"GEMINI_API_KEY": "test-gemini-key",
			},
			wantCount: 1,
			wantNames: []string{"gemini"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envLookup := func(key string) string {
				return tt.envVars[key]
			}

			generators := CreateProviders(tt.cfg, envLookup)

			if len(generators) != tt.wantCount {
				t.Errorf("got %d generators, want %d", len(generators), tt.wantCount)
				return
			}

			for i, gen := range generators {
				if gen.Name() != tt.wantNames[i] {
					t.Errorf("generator[%d].Name() = %q, want %q", i, gen.Name(), tt.wantNames[i])
				}
			}
		})
	}
}

func TestCreateProviders_NilEnvLookup(t *testing.T) {
	// With nil envLookup, it falls back to os.Getenv which won't have test keys
	cfg := configuration.SettingsThumbnailGeneration{
		Providers: []configuration.SettingsThumbnailProvider{
			{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
		},
	}

	generators := CreateProviders(cfg, nil)

	// Should get 0 generators since GEMINI_API_KEY is not set in test env
	if len(generators) != 0 {
		t.Errorf("expected 0 generators with nil envLookup (no real env vars), got %d", len(generators))
	}
}

func TestStartCleanupLoop(t *testing.T) {
	store := NewGeneratedImageStore(50 * time.Millisecond)

	// Inject a controllable clock
	now := time.Now()
	store.nowFunc = func() time.Time { return now }

	// Add an image that will be expired
	_, err := store.Add(GeneratedImage{
		Provider: "test",
		Style:    "test",
		Data:     []byte("test-data"),
	})
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	if store.Len() != 1 {
		t.Fatalf("store.Len() = %d, want 1", store.Len())
	}

	// Advance time past TTL
	now = now.Add(100 * time.Millisecond)

	stop := make(chan struct{})
	StartCleanupLoop(store, 20*time.Millisecond, stop)

	// Wait for cleanup to run
	time.Sleep(80 * time.Millisecond)

	close(stop)

	if store.Len() != 0 {
		t.Errorf("store.Len() = %d after cleanup, want 0", store.Len())
	}
}

func TestStartCleanupLoop_StopsOnClose(t *testing.T) {
	store := NewGeneratedImageStore(time.Hour)
	stop := make(chan struct{})

	StartCleanupLoop(store, 10*time.Millisecond, stop)

	// Close immediately — the goroutine should exit cleanly
	close(stop)

	// Give goroutine time to stop
	time.Sleep(30 * time.Millisecond)

	// No assertion needed — this test verifies no goroutine leak/panic
}

func TestCreateProvider_UnknownReturnsError(t *testing.T) {
	gen, err := createProvider("unknown-provider", "some-model", "some-key")
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
	if gen != nil {
		t.Errorf("expected nil generator for unknown provider, got %v", gen)
	}
	if !strings.Contains(err.Error(), "unknown provider: unknown-provider") {
		t.Errorf("error = %q, want it to contain 'unknown provider: unknown-provider'", err.Error())
	}
}

func TestProviderEnvKeys(t *testing.T) {
	expected := map[string]string{
		"gemini":    "GEMINI_API_KEY",
		"gpt-image": "OPENAI_API_KEY",
	}

	for provider, envKey := range expected {
		got, ok := ProviderEnvKeys[provider]
		if !ok {
			t.Errorf("ProviderEnvKeys missing key for %q", provider)
			continue
		}
		if got != envKey {
			t.Errorf("ProviderEnvKeys[%q] = %q, want %q", provider, got, envKey)
		}
	}

	if len(ProviderEnvKeys) != len(expected) {
		t.Errorf("ProviderEnvKeys has %d entries, want %d", len(ProviderEnvKeys), len(expected))
	}
}
