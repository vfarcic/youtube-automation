package api

import (
	"context"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/publishing"
)

// AnalyzeService abstracts the title analysis pipeline for testability.
type AnalyzeService interface {
	LoadVideosWithABData(indexPath, dataDir, manuscriptDir string) ([]ai.VideoABData, error)
	GetVideoAnalyticsForLastYear(ctx context.Context) ([]publishing.VideoAnalytics, error)
	EnrichWithFirstWeekMetrics(ctx context.Context, analytics []publishing.VideoAnalytics) ([]publishing.VideoAnalytics, error)
	EnrichWithAnalytics(videos []ai.VideoABData, analytics []publishing.VideoAnalytics) []ai.VideoABData
	AnalyzeTitles(ctx context.Context, videos []ai.VideoABData, baseDir string) (ai.TitleAnalysisResult, string, error)
}

// DefaultAnalyzeService delegates to the ai and publishing package functions.
type DefaultAnalyzeService struct{}

func (d *DefaultAnalyzeService) LoadVideosWithABData(indexPath, dataDir, manuscriptDir string) ([]ai.VideoABData, error) {
	return ai.LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
}

func (d *DefaultAnalyzeService) GetVideoAnalyticsForLastYear(ctx context.Context) ([]publishing.VideoAnalytics, error) {
	return publishing.GetVideoAnalyticsForLastYear(ctx)
}

func (d *DefaultAnalyzeService) EnrichWithFirstWeekMetrics(ctx context.Context, analytics []publishing.VideoAnalytics) ([]publishing.VideoAnalytics, error) {
	return publishing.EnrichWithFirstWeekMetrics(ctx, analytics)
}

func (d *DefaultAnalyzeService) EnrichWithAnalytics(videos []ai.VideoABData, analytics []publishing.VideoAnalytics) []ai.VideoABData {
	return ai.EnrichWithAnalytics(videos, analytics)
}

func (d *DefaultAnalyzeService) AnalyzeTitles(ctx context.Context, videos []ai.VideoABData, baseDir string) (ai.TitleAnalysisResult, string, error) {
	return ai.AnalyzeTitles(ctx, videos, baseDir)
}

// GitSyncService abstracts git commit+push for testability.
type GitSyncService interface {
	CommitAndPush(message string) error
}
