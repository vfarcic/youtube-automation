package api

import (
	"context"

	"devopstoolkit/youtube-automation/internal/ai"
)

// AIService abstracts AI content generation for testability.
type AIService interface {
	SuggestTitles(ctx context.Context, manuscript string, dataDir string) ([]string, error)
	SuggestDescription(ctx context.Context, manuscript string) (string, error)
	SuggestTags(ctx context.Context, manuscript string) (string, error)
	SuggestTweets(ctx context.Context, manuscript string) ([]string, error)
	SuggestDescriptionTags(ctx context.Context, manuscript string) (string, error)
	AnalyzeShorts(ctx context.Context, manuscript string) ([]ai.ShortCandidate, error)
	GenerateThumbnailVariations(ctx context.Context, imagePath string) (ai.VariationPrompts, error)
	TranslateVideoMetadata(ctx context.Context, input ai.VideoMetadataInput, targetLanguage string) (*ai.VideoMetadataOutput, error)
	GenerateAMAContent(ctx context.Context, transcript string) (*ai.AMAContent, error)
	GenerateAMATitle(ctx context.Context, transcript string) (string, error)
	GenerateAMADescription(ctx context.Context, transcript string) (string, error)
	GenerateAMATimecodes(ctx context.Context, transcript string) (string, error)
	SuggestIllustrations(ctx context.Context, manuscript, tagline string) ([]string, error)
}

// DefaultAIService delegates to the ai package functions.
type DefaultAIService struct{}

func (d *DefaultAIService) SuggestTitles(ctx context.Context, manuscript string, dataDir string) ([]string, error) {
	return ai.SuggestTitles(ctx, manuscript, dataDir)
}

func (d *DefaultAIService) SuggestDescription(ctx context.Context, manuscript string) (string, error) {
	return ai.SuggestDescription(ctx, manuscript)
}

func (d *DefaultAIService) SuggestTags(ctx context.Context, manuscript string) (string, error) {
	return ai.SuggestTags(ctx, manuscript)
}

func (d *DefaultAIService) SuggestTweets(ctx context.Context, manuscript string) ([]string, error) {
	return ai.SuggestTweets(ctx, manuscript)
}

func (d *DefaultAIService) SuggestDescriptionTags(ctx context.Context, manuscript string) (string, error) {
	return ai.SuggestDescriptionTags(ctx, manuscript)
}

func (d *DefaultAIService) AnalyzeShorts(ctx context.Context, manuscript string) ([]ai.ShortCandidate, error) {
	return ai.AnalyzeShortsFromManuscript(ctx, manuscript)
}

func (d *DefaultAIService) GenerateThumbnailVariations(ctx context.Context, imagePath string) (ai.VariationPrompts, error) {
	return ai.GenerateThumbnailVariations(ctx, imagePath)
}

func (d *DefaultAIService) TranslateVideoMetadata(ctx context.Context, input ai.VideoMetadataInput, targetLanguage string) (*ai.VideoMetadataOutput, error) {
	return ai.TranslateVideoMetadata(ctx, input, targetLanguage)
}

func (d *DefaultAIService) GenerateAMAContent(ctx context.Context, transcript string) (*ai.AMAContent, error) {
	return ai.GenerateAMAContent(ctx, transcript)
}

func (d *DefaultAIService) GenerateAMATitle(ctx context.Context, transcript string) (string, error) {
	return ai.GenerateAMATitle(ctx, transcript)
}

func (d *DefaultAIService) GenerateAMADescription(ctx context.Context, transcript string) (string, error) {
	return ai.GenerateAMADescription(ctx, transcript)
}

func (d *DefaultAIService) GenerateAMATimecodes(ctx context.Context, transcript string) (string, error) {
	return ai.GenerateAMATimecodes(ctx, transcript)
}

func (d *DefaultAIService) SuggestIllustrations(ctx context.Context, manuscript, tagline string) ([]string, error) {
	return ai.SuggestIllustrations(ctx, manuscript, tagline)
}
