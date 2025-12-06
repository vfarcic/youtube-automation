package yt_transcript_formatters

import (
	"encoding/json"

	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
)

type JSONTranscriptLine struct {
	Text     string  `json:"text"`
	Start    float64 `json:"start,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

type JSONTranscripts struct {
	LanguageCode *string              `json:"language_code"`
	Transcripts  []JSONTranscriptLine `json:"transcripts"`
}

type JSONFormatterOption func(*JSONFormatter)

type JSONFormatter struct {
	BaseFormatter
	PrettyPrint bool
}

func NewJSONFormatter(baseOptions ...FormatterOption) *JSONFormatter {
	f := &JSONFormatter{
		BaseFormatter: BaseFormatter{
			IncludeTimestamps:   true,
			IncludeLanguageCode: true,
		},
		PrettyPrint: false,
	}

	for _, opt := range baseOptions {
		opt(&f.BaseFormatter)
	}
	return f
}

func WithPrettyPrint(pretty bool) JSONFormatterOption {
	return func(f *JSONFormatter) {
		f.PrettyPrint = pretty
	}
}

func (f *JSONFormatter) Configure(options ...JSONFormatterOption) {
	for _, opt := range options {
		opt(f)
	}
}

func (f *JSONFormatter) Format(transcripts []yt_transcript_models.Transcript) (string, error) {
	jsonTranscripts := make([]JSONTranscripts, len(transcripts))

	for i, transcript := range transcripts {
		lines := make([]JSONTranscriptLine, len(transcript.Lines))
		for j, line := range transcript.Lines {
			if f.IncludeTimestamps {
				lines[j] = JSONTranscriptLine{
					Text:     line.Text,
					Start:    line.Start,
					Duration: line.Duration,
				}
			} else {
				lines[j] = JSONTranscriptLine{
					Text: line.Text,
				}
			}
		}

		jsonTranscripts[i] = JSONTranscripts{
			Transcripts: lines,
		}
		if f.IncludeLanguageCode {
			jsonTranscripts[i].LanguageCode = &transcript.LanguageCode
		}
	}

	var (
		bytes []byte
		err   error
	)

	if f.PrettyPrint {
		bytes, err = json.MarshalIndent(jsonTranscripts, "", "  ")
	} else {
		bytes, err = json.Marshal(jsonTranscripts)
	}

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
