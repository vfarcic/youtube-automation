package thumbnail

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockGenerator is a test double for ImageGenerator.
type mockGenerator struct {
	name      string
	genFunc   func(ctx context.Context, prompt string, photos [][]byte) ([]byte, error)
	callCount atomic.Int32
}

func (m *mockGenerator) Name() string { return m.name }

func (m *mockGenerator) GenerateImage(ctx context.Context, prompt string, photos [][]byte) ([]byte, error) {
	m.callCount.Add(1)
	return m.genFunc(ctx, prompt, photos)
}

func TestGenerateThumbnails(t *testing.T) {
	photos := [][]byte{{0xFF, 0xD8}}

	tests := []struct {
		name           string
		providers      []ImageGenerator
		req            GenerateRequest
		wantImageCount int
		wantErrCount   int
	}{
		{
			name:      "no providers returns nil",
			providers: nil,
			req: GenerateRequest{
				PromptWithIllustration:    "with",
				PromptWithoutIllustration: "without",
				Photos:                    photos,
			},
			wantImageCount: 0,
			wantErrCount:   0,
		},
		{
			name: "single provider success",
			providers: []ImageGenerator{
				&mockGenerator{
					name: "mock-provider",
					genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
						return []byte("image-data"), nil
					},
				},
			},
			req: GenerateRequest{
				PromptWithIllustration:    "with illustration",
				PromptWithoutIllustration: "without illustration",
				Photos:                    photos,
			},
			wantImageCount: 2,
			wantErrCount:   0,
		},
		{
			name: "multiple providers all succeed",
			providers: []ImageGenerator{
				&mockGenerator{
					name: "provider-a",
					genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
						return []byte("a-data"), nil
					},
				},
				&mockGenerator{
					name: "provider-b",
					genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
						return []byte("b-data"), nil
					},
				},
			},
			req: GenerateRequest{
				PromptWithIllustration:    "with",
				PromptWithoutIllustration: "without",
				Photos:                    photos,
			},
			wantImageCount: 4, // 2 per provider
			wantErrCount:   0,
		},
		{
			name: "one provider fails completely",
			providers: []ImageGenerator{
				&mockGenerator{
					name: "good-provider",
					genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
						return []byte("good-data"), nil
					},
				},
				&mockGenerator{
					name: "bad-provider",
					genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
						return nil, errors.New("API error")
					},
				},
			},
			req: GenerateRequest{
				PromptWithIllustration:    "with",
				PromptWithoutIllustration: "without",
				Photos:                    photos,
			},
			wantImageCount: 2, // only good provider
			wantErrCount:   1, // bad provider error
		},
		{
			name: "all providers fail",
			providers: []ImageGenerator{
				&mockGenerator{
					name: "fail-a",
					genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
						return nil, errors.New("error a")
					},
				},
				&mockGenerator{
					name: "fail-b",
					genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
						return nil, errors.New("error b")
					},
				},
			},
			req: GenerateRequest{
				PromptWithIllustration:    "with",
				PromptWithoutIllustration: "without",
				Photos:                    photos,
			},
			wantImageCount: 0,
			wantErrCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			images, errs := GenerateThumbnails(context.Background(), tt.providers, tt.req)

			if len(images) != tt.wantImageCount {
				t.Errorf("got %d images, want %d", len(images), tt.wantImageCount)
			}
			if len(errs) != tt.wantErrCount {
				t.Errorf("got %d errors, want %d: %v", len(errs), tt.wantErrCount, errs)
			}
		})
	}
}

func TestGenerateThumbnails_ProvidersConcurrent(t *testing.T) {
	// Verify that providers are called concurrently by checking they all start
	// before any completes.
	started := make(chan struct{}, 3)
	proceed := make(chan struct{})

	makeProvider := func(name string) ImageGenerator {
		return &mockGenerator{
			name: name,
			genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
				started <- struct{}{}
				<-proceed
				return []byte(name + "-data"), nil
			},
		}
	}

	providers := []ImageGenerator{
		makeProvider("p1"),
		makeProvider("p2"),
		makeProvider("p3"),
	}

	req := GenerateRequest{
		PromptWithIllustration:    "with",
		PromptWithoutIllustration: "without",
		Photos:                    [][]byte{{0xFF}},
	}

	done := make(chan struct{})
	var images []GeneratedImage
	var errs []error

	go func() {
		images, errs = GenerateThumbnails(context.Background(), providers, req)
		close(done)
	}()

	// Wait for at least 3 goroutines to have started (one per provider starts
	// both style goroutines, so at least 3 total style goroutines signal).
	// This proves concurrency.
	for range 3 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for concurrent provider starts")
		}
	}

	// Let all proceed.
	close(proceed)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for GenerateThumbnails to complete")
	}

	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	// 3 providers × 2 styles = 6 images
	if len(images) != 6 {
		t.Errorf("got %d images, want 6", len(images))
	}
}

func TestGenerateThumbnails_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	provider := &mockGenerator{
		name: "ctx-provider",
		genFunc: func(ctx context.Context, _ string, _ [][]byte) ([]byte, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return []byte("data"), nil
			}
		},
	}

	req := GenerateRequest{
		PromptWithIllustration:    "with",
		PromptWithoutIllustration: "without",
		Photos:                    [][]byte{{0xFF}},
	}

	images, errs := GenerateThumbnails(ctx, []ImageGenerator{provider}, req)

	// With a canceled context, the provider should return errors.
	// Both styles may fail, resulting in 0 images and 1 aggregated error.
	if len(images) != 0 {
		t.Errorf("expected 0 images with canceled context, got %d", len(images))
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestGenerateThumbnails_PartialStyleFailure(t *testing.T) {
	// One style succeeds, one fails within the same provider.
	callNum := atomic.Int32{}
	provider := &mockGenerator{
		name: "partial-provider",
		genFunc: func(_ context.Context, prompt string, _ [][]byte) ([]byte, error) {
			n := callNum.Add(1)
			if n == 1 {
				return nil, errors.New("style error")
			}
			return []byte("ok-data"), nil
		},
	}

	req := GenerateRequest{
		PromptWithIllustration:    "with",
		PromptWithoutIllustration: "without",
		Photos:                    [][]byte{{0xFF}},
	}

	images, errs := GenerateThumbnails(context.Background(), []ImageGenerator{provider}, req)

	// Should get 1 image (the successful style) and 1 error (the failed style).
	if len(images) != 1 {
		t.Errorf("got %d images, want 1", len(images))
	}
	if len(errs) != 1 {
		t.Errorf("got %d errors, want 1: %v", len(errs), errs)
	}
}

func TestGenerateThumbnails_ImageMetadata(t *testing.T) {
	provider := &mockGenerator{
		name: "test-provider",
		genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
			return []byte("img"), nil
		},
	}

	req := GenerateRequest{
		PromptWithIllustration:    "with prompt",
		PromptWithoutIllustration: "without prompt",
		Photos:                    [][]byte{{0xFF}},
	}

	images, errs := GenerateThumbnails(context.Background(), []ImageGenerator{provider}, req)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(images) != 2 {
		t.Fatalf("got %d images, want 2", len(images))
	}

	// Sort for deterministic assertions.
	sort.Slice(images, func(i, j int) bool {
		return images[i].Style < images[j].Style
	})

	for _, img := range images {
		if img.Provider != "test-provider" {
			t.Errorf("image provider = %q, want %q", img.Provider, "test-provider")
		}
		if img.Style != "with illustration" && img.Style != "without illustration" {
			t.Errorf("unexpected style: %q", img.Style)
		}
		if len(img.Data) == 0 {
			t.Error("image data is empty")
		}
	}
}

func TestGenerateThumbnails_ConcurrencyLimit(t *testing.T) {
	// Create more providers than the concurrency limit to verify the semaphore
	// prevents unbounded goroutine spawning.
	const numProviders = 20
	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	providers := make([]ImageGenerator, numProviders)
	for i := range numProviders {
		providers[i] = &mockGenerator{
			name: "provider-" + string(rune('a'+i)),
			genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
				cur := concurrent.Add(1)
				// Track the peak concurrency observed.
				for {
					old := maxConcurrent.Load()
					if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
						break
					}
				}
				time.Sleep(10 * time.Millisecond) // hold the slot briefly
				concurrent.Add(-1)
				return []byte("data"), nil
			},
		}
	}

	req := GenerateRequest{
		PromptWithIllustration:    "with",
		PromptWithoutIllustration: "without",
		Photos:                    [][]byte{{0xFF}},
	}

	images, errs := GenerateThumbnails(context.Background(), providers, req)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if len(images) != numProviders*2 {
		t.Errorf("got %d images, want %d", len(images), numProviders*2)
	}

	peak := maxConcurrent.Load()
	// Each provider spawns 2 style goroutines inside runProvider, so the peak
	// concurrent GenFunc calls can be up to 2×maxConcurrentProviders. We just
	// verify the provider-level semaphore kept it bounded.
	if peak > int32(maxConcurrentProviders*2) {
		t.Errorf("peak concurrent calls = %d, want <= %d (semaphore not working)", peak, maxConcurrentProviders*2)
	}
}

func TestGenerateThumbnails_PhotoRealistic(t *testing.T) {
	photos := [][]byte{{0xFF, 0xD8}}

	tests := []struct {
		name           string
		req            GenerateRequest
		wantImageCount int
		wantStyles     map[string]int // style -> expected count across all providers
	}{
		{
			name: "photo-realistic prompt produces third image per provider",
			req: GenerateRequest{
				PromptWithIllustration:    "with",
				PromptWithoutIllustration: "without",
				PromptPhotoRealistic:      "photo-real",
				Photos:                    photos,
			},
			wantImageCount: 3,
			wantStyles: map[string]int{
				StyleWithIllustration:    1,
				StyleWithoutIllustration: 1,
				StylePhotoRealistic:      1,
			},
		},
		{
			name: "empty photo-realistic prompt skips the third image",
			req: GenerateRequest{
				PromptWithIllustration:    "with",
				PromptWithoutIllustration: "without",
				PromptPhotoRealistic:      "",
				Photos:                    photos,
			},
			wantImageCount: 2,
			wantStyles: map[string]int{
				StyleWithIllustration:    1,
				StyleWithoutIllustration: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockGenerator{
				name: "mock-provider",
				genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
					return []byte("img"), nil
				},
			}

			images, errs := GenerateThumbnails(context.Background(), []ImageGenerator{provider}, tt.req)
			if len(errs) != 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
			if len(images) != tt.wantImageCount {
				t.Fatalf("got %d images, want %d (styles: %+v)", len(images), tt.wantImageCount, images)
			}

			gotStyles := map[string]int{}
			for _, img := range images {
				gotStyles[img.Style]++
			}
			for style, want := range tt.wantStyles {
				if gotStyles[style] != want {
					t.Errorf("style %q count = %d, want %d", style, gotStyles[style], want)
				}
			}
		})
	}
}

func TestGenerateThumbnails_PhotoRealistic_MultipleProviders(t *testing.T) {
	// Three providers × three styles = nine images total.
	photos := [][]byte{{0xFF, 0xD8}}

	providers := []ImageGenerator{
		&mockGenerator{name: "provider-a", genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
			return []byte("a"), nil
		}},
		&mockGenerator{name: "provider-b", genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
			return []byte("b"), nil
		}},
		&mockGenerator{name: "provider-c", genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
			return []byte("c"), nil
		}},
	}

	req := GenerateRequest{
		PromptWithIllustration:    "with",
		PromptWithoutIllustration: "without",
		PromptPhotoRealistic:      "photo-real",
		Photos:                    photos,
	}

	images, errs := GenerateThumbnails(context.Background(), providers, req)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(images) != 9 {
		t.Fatalf("got %d images, want 9", len(images))
	}

	// Each provider must produce exactly one image of each style.
	type key struct{ provider, style string }
	counts := map[key]int{}
	for _, img := range images {
		counts[key{img.Provider, img.Style}]++
	}
	expectedProviders := []string{"provider-a", "provider-b", "provider-c"}
	expectedStyles := []string{StyleWithIllustration, StyleWithoutIllustration, StylePhotoRealistic}
	for _, p := range expectedProviders {
		for _, s := range expectedStyles {
			if counts[key{p, s}] != 1 {
				t.Errorf("provider %s style %s count = %d, want 1", p, s, counts[key{p, s}])
			}
		}
	}
}

// TestGenerateThumbnails_PhotoRealistic_PromptForwarded asserts the exact
// PromptPhotoRealistic string is forwarded to the provider for the
// photorealistic style, not one of the B&W prompts.
func TestGenerateThumbnails_PhotoRealistic_PromptForwarded(t *testing.T) {
	var mu sync.Mutex
	seen := map[string]bool{}

	provider := &mockGenerator{
		name: "mock-provider",
		genFunc: func(_ context.Context, prompt string, _ [][]byte) ([]byte, error) {
			mu.Lock()
			seen[prompt] = true
			mu.Unlock()
			return []byte("data"), nil
		},
	}

	req := GenerateRequest{
		PromptWithIllustration:    "PROMPT-WITH",
		PromptWithoutIllustration: "PROMPT-WITHOUT",
		PromptPhotoRealistic:      "PROMPT-PHOTOREAL",
		Photos:                    [][]byte{{0xFF}},
	}

	images, errs := GenerateThumbnails(context.Background(), []ImageGenerator{provider}, req)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(images) != 3 {
		t.Fatalf("got %d images, want 3", len(images))
	}

	mu.Lock()
	defer mu.Unlock()
	for _, want := range []string{"PROMPT-WITH", "PROMPT-WITHOUT", "PROMPT-PHOTOREAL"} {
		if !seen[want] {
			t.Errorf("provider was not called with prompt %q (saw: %v)", want, seen)
		}
	}
}

// TestGenerateThumbnails_PhotoRealistic_Concurrent verifies the third style
// runs concurrently with the existing two (no regression in latency).
func TestGenerateThumbnails_PhotoRealistic_Concurrent(t *testing.T) {
	started := make(chan struct{}, 3)
	proceed := make(chan struct{})

	provider := &mockGenerator{
		name: "concurrent-provider",
		genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
			started <- struct{}{}
			<-proceed
			return []byte("data"), nil
		},
	}

	req := GenerateRequest{
		PromptWithIllustration:    "with",
		PromptWithoutIllustration: "without",
		PromptPhotoRealistic:      "photo-real",
		Photos:                    [][]byte{{0xFF}},
	}

	done := make(chan struct{})
	var images []GeneratedImage
	var errs []error
	go func() {
		images, errs = GenerateThumbnails(context.Background(), []ImageGenerator{provider}, req)
		close(done)
	}()

	// All three style goroutines must start before any can complete.
	for range 3 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for all three style goroutines to start concurrently")
		}
	}

	close(proceed)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for GenerateThumbnails to complete")
	}

	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if len(images) != 3 {
		t.Errorf("got %d images, want 3", len(images))
	}
}

// TestGenerateThumbnails_PhotoRealistic_PartialFailure verifies that a
// failure in one style does not block the other two.
func TestGenerateThumbnails_PhotoRealistic_PartialFailure(t *testing.T) {
	provider := &mockGenerator{
		name: "partial-provider",
		genFunc: func(_ context.Context, prompt string, _ [][]byte) ([]byte, error) {
			if prompt == "PROMPT-PHOTOREAL" {
				return nil, errors.New("photoreal failure")
			}
			return []byte("ok"), nil
		},
	}

	req := GenerateRequest{
		PromptWithIllustration:    "PROMPT-WITH",
		PromptWithoutIllustration: "PROMPT-WITHOUT",
		PromptPhotoRealistic:      "PROMPT-PHOTOREAL",
		Photos:                    [][]byte{{0xFF}},
	}

	images, errs := GenerateThumbnails(context.Background(), []ImageGenerator{provider}, req)

	if len(images) != 2 {
		t.Errorf("got %d images, want 2 (only photoreal should fail)", len(images))
	}
	if len(errs) != 1 {
		t.Errorf("got %d errors, want 1", len(errs))
	}

	for _, img := range images {
		if img.Style == StylePhotoRealistic {
			t.Errorf("photo-realistic image should not be present when its provider call failed")
		}
	}
}

func TestGenerateThumbnails_AllProvidersCalled(t *testing.T) {
	providers := make([]*mockGenerator, 3)
	for i := range providers {
		providers[i] = &mockGenerator{
			name: "provider-" + string(rune('a'+i)),
			genFunc: func(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
				return []byte("data"), nil
			},
		}
	}

	ifaces := make([]ImageGenerator, len(providers))
	for i, p := range providers {
		ifaces[i] = p
	}

	req := GenerateRequest{
		PromptWithIllustration:    "with",
		PromptWithoutIllustration: "without",
		Photos:                    [][]byte{{0xFF}},
	}

	GenerateThumbnails(context.Background(), ifaces, req)

	for _, p := range providers {
		count := p.callCount.Load()
		if count != 2 {
			t.Errorf("provider %s called %d times, want 2", p.name, count)
		}
	}
}
