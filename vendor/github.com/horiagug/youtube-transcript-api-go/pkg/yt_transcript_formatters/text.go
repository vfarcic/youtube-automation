package yt_transcript_formatters

import (
	"fmt"
	"strings"

	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
)

type TextFormatter struct {
	BaseFormatter
}

func NewTextFormatter(options ...FormatterOption) *TextFormatter {
	f := &TextFormatter{
		BaseFormatter: BaseFormatter{
			IncludeTimestamps:   true,
			IncludeLanguageCode: true,
		},
	}

	for _, opt := range options {
		opt(&f.BaseFormatter)
	}

	return f
}

func (t *TextFormatter) Format(transcripts []yt_transcript_models.Transcript) (string, error) {

	var (
		text strings.Builder
		err  error
	)

	for i, transcript := range transcripts {
		if t.IncludeLanguageCode {
			var language string
			if transcript.Language != "" {
				language = transcript.Language
			} else if transcript.LanguageCode != "" {
				language = transcript.LanguageCode
			}

			if language != "" {
				_, err = text.WriteString(fmt.Sprintf("Language: %s\n", language))
			}

			if err != nil {
				return "", err
			}
		}

		for _, line := range transcript.Lines {
			if t.IncludeTimestamps {
				_, err = text.WriteString(fmt.Sprintf("%f: %s\n", line.Start, line.Text))
			} else {
				_, err = text.WriteString(line.Text + "\n")
			}
		}

		if len(transcripts) > 1 && i < len(transcripts)-1 {
			text.WriteString("\n")
		}
	}

	if err != nil {
		return "", err
	}

	return text.String(), nil
}
