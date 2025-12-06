package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/horiagug/youtube-transcript-api-go/internal/repository"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
	"golang.org/x/net/html"
)

type TranscriptService interface {
	GetTranscripts(videoID string, langauges []string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error)
	GetTranscriptsWithContext(ctx context.Context, videoID string, langauges []string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error)
}

type transcriptService struct {
	fetcher repository.HTMLFetcherType
}

type transcriptResult struct {
	transcript yt_transcript_models.Transcript
	err        error
}

func NewTranscriptService(fetcher repository.HTMLFetcherType) *transcriptService {
	return &transcriptService{
		fetcher: fetcher,
	}
}

func (t transcriptService) GetTranscripts(videoID string, languages []string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error) {
	return t.GetTranscriptsWithContext(context.Background(), videoID, languages, preserve_formatting)
}

func (t transcriptService) GetTranscriptsWithContext(ctx context.Context, videoID string, languages []string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error) {
	videoID = sanitizeVideoId(videoID)

	trascript_data, err := t.extractTranscriptList(ctx, videoID)
	if err != nil {
		return []yt_transcript_models.Transcript{}, fmt.Errorf("failed to extract list of transcripts: %w", err)
	}

	transcripts, err := t.getTranscriptsForLanguage(languages, *trascript_data.Transcripts)
	if err != nil {
		return []yt_transcript_models.Transcript{}, fmt.Errorf("failed to get transcript: %w", err)
	}

	return t.processCaptionTracksWithContext(ctx, videoID, transcripts, trascript_data.Title, preserve_formatting)
}

func (t *transcriptService) processCaptionTracks(video_id string, captionTracks []yt_transcript_models.CaptionTrack, title string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error) {
	return t.processCaptionTracksWithContext(context.Background(), video_id, captionTracks, title, preserve_formatting)
}

func (t *transcriptService) processCaptionTracksWithContext(ctx context.Context, video_id string, captionTracks []yt_transcript_models.CaptionTrack, title string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error) {
	resultChan := make(chan transcriptResult, len(captionTracks))
	var wg sync.WaitGroup

	// Pre-allocate results slice with known capacity
	results := make([]yt_transcript_models.Transcript, 0, len(captionTracks))

	for _, transcript := range captionTracks {
		wg.Add(1)
		go func(tr yt_transcript_models.CaptionTrack) {
			defer wg.Done()

			is_generated := false
			if tr.Kind != nil && *tr.Kind == "asr" {
				is_generated = true
			}

			lines, err := t.getTranscriptFromTrackWithContext(ctx, tr, preserve_formatting)
			if err != nil {
				resultChan <- transcriptResult{err: fmt.Errorf("error getting transcript from track: %w", err)}
				return
			}

			result := yt_transcript_models.Transcript{
				VideoID:        video_id,
				VideoTitle:     title,
				Language:       tr.Name.SimpleText,
				LanguageCode:   tr.LanguageCode,
				IsGenerated:    is_generated,
				IsTranslatable: tr.IsTranslatable,
				Lines:          lines,
			}

			resultChan <- transcriptResult{transcript: result}
		}(transcript)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("Error processing transcript: %v\n", result.err)
			return results, result.err
		}
		results = append(results, result.transcript)
	}
	return results, nil
}

// Pre-compiled regex for API key extraction
var innerTubeApiKeyRegex = regexp.MustCompile(`"INNERTUBE_API_KEY":\s*"([a-zA-Z0-9_-]+)"`)

func extractInnerTubeApiKey(htmlContent string) string {
	// Search for the pattern in the HTML content
	match := innerTubeApiKeyRegex.FindStringSubmatch(htmlContent)
	if len(match) == 2 {
		return match[1]
	}

	return ""
}

func extractInnertubeVideoDetails(data map[string]interface{}) (*yt_transcript_models.InnertubeData, error) {
	// Extract captions section directly
	captions, ok := data["captions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("captions not found in response")
	}

	renderer, ok := captions["playerCaptionsTracklistRenderer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("playerCaptionsTracklistRenderer not found")
	}

	// Extract caption tracks
	var captionTracks []yt_transcript_models.CaptionTrack
	if tracks, ok := renderer["captionTracks"].([]interface{}); ok {
		captionTracks = make([]yt_transcript_models.CaptionTrack, 0, len(tracks))
		for _, track := range tracks {
			if trackMap, ok := track.(map[string]interface{}); ok {
				captionTrack := yt_transcript_models.CaptionTrack{}

				if baseUrl, ok := trackMap["baseUrl"].(string); ok {
					captionTrack.BaseUrl = baseUrl
				}

				if langCode, ok := trackMap["languageCode"].(string); ok {
					captionTrack.LanguageCode = langCode
				}

				if name, ok := trackMap["name"].(map[string]interface{}); ok {
					if simpleText, ok := name["simpleText"].(string); ok {
						captionTrack.Name = yt_transcript_models.LanguageName{SimpleText: simpleText}
					}
				}

				if kind, ok := trackMap["kind"].(string); ok {
					captionTrack.Kind = &kind
				}

				if isTranslatable, ok := trackMap["isTranslatable"].(bool); ok {
					captionTrack.IsTranslatable = isTranslatable
				}

				captionTracks = append(captionTracks, captionTrack)
			}
		}
	}

	// Extract translation languages if available
	var translationLanguages *[]yt_transcript_models.LanguageData
	if transLangs, ok := renderer["translationLanguages"].([]interface{}); ok {
		langs := make([]yt_transcript_models.LanguageData, 0, len(transLangs))
		for _, lang := range transLangs {
			if langMap, ok := lang.(map[string]interface{}); ok {
				langData := yt_transcript_models.LanguageData{}

				if langCode, ok := langMap["languageCode"].(string); ok {
					langData.LanguageCode = langCode
				}

				if langName, ok := langMap["languageName"].(map[string]interface{}); ok {
					if simpleText, ok := langName["simpleText"].(string); ok {
						langData.Language = yt_transcript_models.LanguageName{SimpleText: simpleText}
					}
				}

				langs = append(langs, langData)
			}
		}
		if len(langs) > 0 {
			translationLanguages = &langs
		}
	}

	transcriptData := &yt_transcript_models.TranscriptData{
		CaptionTracks:        captionTracks,
		TranslationLanguages: translationLanguages,
	}

	return &yt_transcript_models.InnertubeData{
		Captions: yt_transcript_models.CaptionsDetails{
			PlayerCaptionsTracklistRenderer: transcriptData,
		},
	}, nil
}

func extractTitle(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Printf("Error fetching the title")
		return ""
	}

	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				title = n.FirstChild.Data
				return
			}
		}
		for c := n.FirstChild; c != nil && title == ""; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return title
}

func (t *transcriptService) extractTranscriptList(ctx context.Context, video_id string) (*yt_transcript_models.VideoTranscriptData, error) {
	html, err := t.fetcher.FetchVideo(video_id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video page: %w", err)
	}

	body := string(html)

	title := extractTitle(body)

	innertube_api_key := extractInnerTubeApiKey(body)

	innertube_data, err := t.fetcher.FetchInnertubeData(ctx, video_id, innertube_api_key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video page: %w", err)
	}

	// Directly extract data without unnecessary marshal/unmarshal
	videoDetails, err := extractInnertubeVideoDetails(innertube_data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract video details: %w", err)
	}

	if videoDetails.Captions.PlayerCaptionsTracklistRenderer == nil {
		return nil, fmt.Errorf("playerCaptionsTracklistRenderer not found")
	}

	transcripts := videoDetails.Captions.PlayerCaptionsTracklistRenderer

	return &yt_transcript_models.VideoTranscriptData{Transcripts: transcripts, Title: title}, nil
}

func (s transcriptService) getTranscriptsForLanguage(languages []string, transcripts yt_transcript_models.TranscriptData) ([]yt_transcript_models.CaptionTrack, error) {
	if len(languages) == 0 {
		return transcripts.CaptionTracks, nil
	}

	// Pre-allocate with capacity hint based on language count
	caption_tracks := make([]yt_transcript_models.CaptionTrack, 0, len(languages))

	for _, lang := range languages {
		for _, track := range transcripts.CaptionTracks {
			if track.LanguageCode == lang {
				caption_tracks = append(caption_tracks, track)
			}
		}
	}

	if len(caption_tracks) == 0 {
		return []yt_transcript_models.CaptionTrack{}, fmt.Errorf("no transcript found for languages %s", languages)
	}

	return caption_tracks, nil
}

func (s transcriptService) getTranscriptFromTrackWithContext(ctx context.Context, track yt_transcript_models.CaptionTrack, preserve_formatting bool) ([]yt_transcript_models.TranscriptLine, error) {
	url := strings.Replace(track.BaseUrl, "&fmt=srv3", "", -1)
	body, err := s.fetcher.FetchWithContext(ctx, url, nil)
	if err != nil {
		return []yt_transcript_models.TranscriptLine{}, fmt.Errorf("failed to fetch transcript: %w", err)
	}

	parser := repository.NewTranscriptParser(preserve_formatting)

	transcript, err := parser.Parse(string(body))
	if err != nil {
		return []yt_transcript_models.TranscriptLine{}, fmt.Errorf("failed to parse transcript: %w", err)
	}
	return transcript, nil
}

func sanitizeVideoId(videoID string) string {
	if strings.HasPrefix(videoID, "http://") || strings.HasPrefix(videoID, "https://") || strings.HasPrefix(videoID, "www.") {
		if strings.Contains(videoID, "youtube.com") {
			u, err := url.Parse(videoID)
			if err != nil {
				fmt.Println("Error parsing URL")
				return videoID
			}
			return u.Query().Get("v")
		} else if strings.Contains(videoID, "youtu.be") {
			u, err := url.Parse(videoID)
			if err != nil {
				fmt.Println("Error parsing URL")
				return videoID
			}
			// For youtu.be, the video ID is in the path
			return strings.TrimPrefix(u.Path, "/")
		}
		fmt.Println("Warning: this doesn't look like a youtube video, we'll still try to process it.")
	}
	return videoID
}
