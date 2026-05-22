package ai

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates/photo_realistic_subject.md
var photoRealisticSubjectTemplate string

// maxPhotoRealisticSubjectLen caps the length of the parsed subject before it
// is returned to callers. The template asks for under 80 characters; this is
// a defensive ceiling to keep prompt-builder input bounded.
const maxPhotoRealisticSubjectLen = 200

type photoRealisticSubjectTemplateData struct {
	Manuscript string
}

// SuggestPhotoRealisticSubject asks the configured AI provider for ONE
// concrete, photographable noun phrase derived from the video manuscript —
// suitable for rendering as the contextual subject in the photo-realistic
// thumbnail variant (PRD 401).
//
// Returns the subject string on success. Returns a non-nil error when:
//   - the manuscript is empty,
//   - the AI provider fails,
//   - the response cannot be parsed into a non-empty single-line subject.
//
// Callers should treat the error as "skip the photo-realistic variant" —
// generation of the other two B&W variants must not be blocked.
func SuggestPhotoRealisticSubject(ctx context.Context, manuscript string) (string, error) {
	if strings.TrimSpace(manuscript) == "" {
		return "", fmt.Errorf("manuscript content is empty")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	tmpl, err := template.New("photo_realistic_subject").Parse(photoRealisticSubjectTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse photo-realistic subject template: %w", err)
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, photoRealisticSubjectTemplateData{Manuscript: manuscript}); err != nil {
		return "", fmt.Errorf("failed to execute photo-realistic subject template: %w", err)
	}

	responseContent, err := provider.GenerateContent(ctx, promptBuf.String(), 128)
	if err != nil {
		return "", fmt.Errorf("AI photo-realistic subject suggestion failed: %w", err)
	}

	return parsePhotoRealisticSubjectResponse(responseContent)
}

// preambleLineRegexps match common "preamble label" lines that AI models
// emit despite the template's "no preamble" rule. Lines that match any of
// these (or that end with ':') are discarded before we pick the answer.
//
// The patterns are deliberately anchored on a leading verb/phrase ("Here",
// "Sure", "Of course", "The subject is") so they will NOT clobber a
// legitimate noun phrase that happens to mention those words later in the
// sentence (e.g., "a small white rabbit posed for the photo here").
var preambleLineRegexps = []*regexp.Regexp{
	// "Here is the subject", "Here you go", "Here's the answer", "Here are some options".
	regexp.MustCompile(`(?i)^here\b`),
	// "The subject is", "The answer would be", "The noun phrase is", "The photo-realistic subject is".
	regexp.MustCompile(`(?i)^the (subject|answer|noun phrase|photo[- ]?realistic subject)\b`),
	// "Sure", "Sure!", "Sure, here it is", "Sure thing".
	regexp.MustCompile(`(?i)^sure\b`),
	// "Certainly", "Certainly!", "Certainly, ...".
	regexp.MustCompile(`(?i)^certainly\b`),
	// "Of course", "Of course!", "Of course, ...".
	regexp.MustCompile(`(?i)^of course\b`),
}

// isPreambleLine reports whether a line looks like an AI preamble label
// rather than the actual subject. Lines ending with ':' are treated as
// labels (e.g., "Here is the subject:", "Subject:"); lines matching any
// preambleLineRegexps are also dropped.
func isPreambleLine(line string) bool {
	if strings.HasSuffix(line, ":") {
		return true
	}
	for _, re := range preambleLineRegexps {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// parsePhotoRealisticSubjectResponse extracts a clean single-line subject
// from the AI response, tolerating common formatting slips the template
// forbids (markdown code fences, wrapping quotes, preamble labels,
// trailing explanation).
//
// Strategy: strip the wrapping fence (if any), split into lines, normalize
// each (drop quotes/backticks, trim), discard empties and preamble labels,
// then return the LAST remaining candidate. The answer typically appears
// after any preamble explanation; when the response is a single line, first
// equals last so the simple case is unaffected.
func parsePhotoRealisticSubjectResponse(text string) (string, error) {
	s := strings.TrimSpace(text)
	if s == "" {
		return "", fmt.Errorf("AI returned empty photo-realistic subject")
	}

	// Strip a wrapping fenced code block if present.
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		// Optional language tag on the opening fence line — drop it.
		if idx := strings.Index(s, "\n"); idx >= 0 {
			s = s[idx+1:]
		}
		if idx := strings.Index(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}

	// Collect candidate lines: non-empty, after quote/backtick stripping,
	// excluding preamble labels.
	var candidates []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, "`")
		line = strings.TrimSpace(line)
		line = strings.Trim(line, `"'`)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if isPreambleLine(line) {
			continue
		}
		candidates = append(candidates, line)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("AI returned empty photo-realistic subject after parsing")
	}

	// Last non-preamble line is the answer. When the AI prefaces an answer
	// with explanation, the answer comes last; when only the answer is
	// returned, first == last.
	subject := candidates[len(candidates)-1]

	// Cap length defensively.
	if len(subject) > maxPhotoRealisticSubjectLen {
		subject = subject[:maxPhotoRealisticSubjectLen]
	}

	return subject, nil
}
