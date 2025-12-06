package yt_transcript

import (
	"context"
	"fmt"
	"time"

	"github.com/horiagug/youtube-transcript-api-go/internal/repository"
	"github.com/horiagug/youtube-transcript-api-go/internal/service"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_formatters"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
)

type YtTranscriptClient struct {
	transcriptService service.TranscriptService
	Timeout           int
	Formatter         yt_transcript_formatters.Formatter
}

var preserve_formatting_default = false

func NewClient(options ...Option) *YtTranscriptClient {

	formatter := yt_transcript_formatters.NewJSONFormatter()
	formatter.Configure(yt_transcript_formatters.WithPrettyPrint(true))

	client := &YtTranscriptClient{
		Timeout:   30,
		Formatter: formatter,
	}

	for _, opt := range options {
		opt(client)
	}

	if client.transcriptService == nil {
		fetcher := repository.NewHTMLFetcher()
		client.transcriptService = service.NewTranscriptService(fetcher)
	}

	return client
}

func (c *YtTranscriptClient) GetFormattedTranscripts(videoID string, languages []string, preserve_formatting bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Second)
	defer cancel()

	transcripts, err := c.transcriptService.GetTranscriptsWithContext(ctx, videoID, languages, preserve_formatting)
	if err != nil {
		return "", err
	}

	if len(transcripts) == 0 {
		return "", fmt.Errorf("No transcripts found")
	}

	return c.Formatter.Format(transcripts)
}

func (c *YtTranscriptClient) GetTranscripts(videoID string, languages []string) ([]yt_transcript_models.Transcript, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Second)
	defer cancel()

	transcripts, err := c.transcriptService.GetTranscriptsWithContext(ctx, videoID, languages, true)
	if err != nil {
		return []yt_transcript_models.Transcript{}, err
	}

	return transcripts, nil
}
