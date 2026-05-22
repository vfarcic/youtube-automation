package thumbnail

import (
	"context"
	"fmt"
	"sync"
)

// Style labels persisted on GeneratedImage.Style.
const (
	StyleWithIllustration    = "with illustration"
	StyleWithoutIllustration = "without illustration"
	StylePhotoRealistic      = "photorealistic"
)

// GeneratedImage holds a generated thumbnail image with its metadata.
type GeneratedImage struct {
	Provider string // e.g. "gemini", "gpt-image"
	Style    string // e.g. "with illustration", "without illustration", "photorealistic"
	Data     []byte // raw image bytes
}

// GenerateRequest holds the parameters for a thumbnail generation run.
// PromptPhotoRealistic is optional: when empty, the orchestrator skips the
// photo-realistic variant and only produces the two B&W variants.
type GenerateRequest struct {
	PromptWithIllustration    string
	PromptWithoutIllustration string
	PromptPhotoRealistic      string
	Photos                    [][]byte
}

// providerResult collects one provider's outputs or error.
type providerResult struct {
	provider string
	images   []GeneratedImage
	err      error
}

// maxConcurrentProviders limits the number of provider goroutines that can
// run simultaneously, guarding against misconfiguration with many providers.
const maxConcurrentProviders = 10

// GenerateThumbnails runs all providers concurrently, each generating one
// thumbnail per configured style (B&W with/without illustration always; the
// photo-realistic variant additionally when PromptPhotoRealistic is set).
// Individual provider failures are collected but do not block other providers.
// Concurrency is capped at maxConcurrentProviders goroutines.
// Returns all successfully generated images and any per-provider errors.
func GenerateThumbnails(ctx context.Context, providers []ImageGenerator, req GenerateRequest) ([]GeneratedImage, []error) {
	if len(providers) == 0 {
		return nil, nil
	}

	resultsCh := make(chan providerResult, len(providers))
	sem := make(chan struct{}, maxConcurrentProviders)
	var wg sync.WaitGroup

	for _, gen := range providers {
		wg.Add(1)
		go func(g ImageGenerator) {
			defer wg.Done()
			sem <- struct{}{} // acquire semaphore slot
			defer func() { <-sem }()
			res := runProvider(ctx, g, req)
			resultsCh <- res
		}(gen)
	}

	// Close channel once all goroutines finish.
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var images []GeneratedImage
	var errs []error

	for res := range resultsCh {
		if res.err != nil {
			errs = append(errs, fmt.Errorf("provider %s: %w", res.provider, res.err))
		}
		images = append(images, res.images...)
	}

	return images, errs
}

// runProvider generates one thumbnail per configured style for a single
// provider, dispatching all style goroutines concurrently. The B&W
// with/without-illustration styles always run; the photo-realistic style
// runs only when PromptPhotoRealistic is set on the request.
func runProvider(ctx context.Context, gen ImageGenerator, req GenerateRequest) providerResult {
	type styleSpec struct {
		style  string
		prompt string
	}
	type styleResult struct {
		style string
		data  []byte
		err   error
	}

	specs := []styleSpec{
		{style: StyleWithIllustration, prompt: req.PromptWithIllustration},
		{style: StyleWithoutIllustration, prompt: req.PromptWithoutIllustration},
	}
	if req.PromptPhotoRealistic != "" {
		specs = append(specs, styleSpec{style: StylePhotoRealistic, prompt: req.PromptPhotoRealistic})
	}

	ch := make(chan styleResult, len(specs))
	for _, spec := range specs {
		go func(s styleSpec) {
			data, err := gen.GenerateImage(ctx, s.prompt, req.Photos)
			ch <- styleResult{style: s.style, data: data, err: err}
		}(spec)
	}

	res := providerResult{provider: gen.Name()}
	for range len(specs) {
		sr := <-ch
		if sr.err != nil {
			// Record error but continue collecting other styles' results.
			if res.err == nil {
				res.err = fmt.Errorf("%s: %w", sr.style, sr.err)
			} else {
				res.err = fmt.Errorf("%v; %s: %w", res.err, sr.style, sr.err)
			}
			continue
		}
		res.images = append(res.images, GeneratedImage{
			Provider: gen.Name(),
			Style:    sr.style,
			Data:     sr.data,
		})
	}

	return res
}
