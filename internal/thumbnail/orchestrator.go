package thumbnail

import (
	"context"
	"fmt"
	"sync"
)

// GeneratedImage holds a generated thumbnail image with its metadata.
type GeneratedImage struct {
	Provider string // e.g. "gemini", "gpt-image"
	Style    string // "with illustration" or "without illustration"
	Data     []byte // raw image bytes
}

// GenerateRequest holds the parameters for a thumbnail generation run.
type GenerateRequest struct {
	PromptWithIllustration    string
	PromptWithoutIllustration string
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

// GenerateThumbnails runs all providers concurrently, each generating two
// thumbnails (with and without illustration). Individual provider failures
// are collected but do not block other providers.
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

// runProvider generates both with-illustration and without-illustration
// thumbnails for a single provider concurrently.
func runProvider(ctx context.Context, gen ImageGenerator, req GenerateRequest) providerResult {
	type styleResult struct {
		style string
		data  []byte
		err   error
	}

	ch := make(chan styleResult, 2)

	go func() {
		data, err := gen.GenerateImage(ctx, req.PromptWithIllustration, req.Photos)
		ch <- styleResult{style: "with illustration", data: data, err: err}
	}()

	go func() {
		data, err := gen.GenerateImage(ctx, req.PromptWithoutIllustration, req.Photos)
		ch <- styleResult{style: "without illustration", data: data, err: err}
	}()

	res := providerResult{provider: gen.Name()}
	for range 2 {
		sr := <-ch
		if sr.err != nil {
			// Record error but continue collecting the other style's result.
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
