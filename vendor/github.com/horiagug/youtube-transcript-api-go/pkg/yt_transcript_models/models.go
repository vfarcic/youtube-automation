package yt_transcript_models

type Transcript struct {
	VideoID        string
	VideoTitle     string
	Language       string
	LanguageCode   string
	IsGenerated    bool
	IsTranslatable bool
	Lines          []TranscriptLine
}

type TranscriptLine struct {
	Text     string  `json:"text"`
	Start    float64 `json:"start"`
	Duration float64 `json:"duration"`
}

type TranscriptList struct {
	VideoID       string
	CaptionTracks []CaptionTrack
}

type LanguageName struct {
	SimpleText string `json:"simpleText"`
}

type LanguageData struct {
	Language     LanguageName `json:"languageName"`
	LanguageCode string       `json:"languageCode"`
}

type CaptionTrack struct {
	Kind           *string      `json:"kind,omitempty"`
	LanguageCode   string       `json:"languageCode"`
	BaseUrl        string       `json:"baseUrl"`
	Name           LanguageName `json:"name"`
	IsTranslatable bool         `json:"isTranslatable"`
}

type TranscriptData struct {
	CaptionTracks        []CaptionTrack  `json:"captionTracks"`
	TranslationLanguages *[]LanguageData `json:"translationLanguages,omitempty"`
}

type VideoDetails struct {
	PlayerCaptionsTracklistRenderer *TranscriptData `json:"playerCaptionsTracklistRenderer"`
	Title                           string          `json:"title"`
}

type CaptionsDetails struct {
	PlayerCaptionsTracklistRenderer *TranscriptData `json:"playerCaptionsTracklistRenderer"`
}

type VideoTranscriptData struct {
	Transcripts *TranscriptData
	Title       string
}

type InnertubeData struct {
	Captions CaptionsDetails `json:"captions"`
}
