package thumbnail

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"unicode"
)

// ErrEmptySubject is returned by BuildPhotoRealisticPrompt when the subject
// is empty after sanitization. Callers should treat it as "skip this variant"
// rather than render a malformed prompt.
var ErrEmptySubject = errors.New("photo-realistic subject is empty after sanitization")

// BackgroundColor defines a background color with its allowed text colors.
type BackgroundColor struct {
	Name       string
	Hex        string
	TextColors []TextColor
}

// TextColor defines a text color option.
type TextColor struct {
	Name string
	Hex  string
}

// PersonPlacement defines where the person appears and which direction they face.
type PersonPlacement struct {
	Description   string
	FaceDirection string // "left" or "right"
}

// ChannelPalette holds all available background colors.
var ChannelPalette = []BackgroundColor{
	{
		Name: "Orange",
		Hex:  "#FF8800",
		TextColors: []TextColor{
			{Name: "White", Hex: "#FFFFFF"},
			{Name: "Dark charcoal", Hex: "#1D242C"},
		},
	},
	{
		Name: "Sky blue",
		Hex:  "#59C5F3",
		TextColors: []TextColor{
			{Name: "White", Hex: "#FFFFFF"},
			{Name: "Dark charcoal", Hex: "#1D242C"},
		},
	},
	{
		Name: "Yellow/gold",
		Hex:  "#FFD800",
		TextColors: []TextColor{
			{Name: "White", Hex: "#FFFFFF"},
			{Name: "Dark charcoal", Hex: "#1D242C"},
		},
	},
	{
		Name: "Dark charcoal",
		Hex:  "#1D242C",
		TextColors: []TextColor{
			{Name: "White", Hex: "#FFFFFF"},
			{Name: "Orange", Hex: "#FF8800"},
			{Name: "Sky blue", Hex: "#59C5F3"},
		},
	},
	{
		Name: "Deep purple",
		Hex:  "#6A0DAD",
		TextColors: []TextColor{
			{Name: "White", Hex: "#FFFFFF"},
			{Name: "Yellow/gold", Hex: "#FFD800"},
			{Name: "Lime green", Hex: "#AAFF00"},
		},
	},
}

// PersonPlacements holds all available placement options with face direction logic.
var PersonPlacements = []PersonPlacement{
	{Description: "right side of the frame, slightly overlapping the right edge", FaceDirection: "left"},
	{Description: "center-right, angled slightly toward the viewer", FaceDirection: "left"},
	{Description: "right third, turned at a three-quarter view facing left", FaceDirection: "left"},
	{Description: "left side of the frame, slightly overlapping the left edge", FaceDirection: "right"},
	{Description: "center-left, angled slightly toward the viewer", FaceDirection: "right"},
	{Description: "left third, turned at a three-quarter view facing right", FaceDirection: "right"},
}

// PromptConfig holds the randomly selected parameters for a prompt.
type PromptConfig struct {
	Background   BackgroundColor
	TextColor    TextColor
	Placement    PersonPlacement
	Tagline      string
	Illustration string // empty means no illustration
}

// PhotoRealisticPromptConfig holds parameters for the photo-realistic
// thumbnail variant (no B&W treatment, no text overlay, contextual subject
// rendered realistically).
type PhotoRealisticPromptConfig struct {
	Placement PersonPlacement
	Subject   string
}

const maxPromptInputLen = 200

// sanitizationIterationCap bounds the fixed-point sanitization loop so that
// adversarial inputs (e.g., deeply nested injection patterns) cannot cause
// the function to do more than this many passes.
const sanitizationIterationCap = 8

// injectionRegexps are the prompt-injection patterns SanitizePromptInput
// scrubs from user-provided text. They are applied in order, repeatedly,
// until the output stops changing or sanitizationIterationCap is reached.
//
// All patterns use the (?i) flag (case-insensitive) and are anchored on \b
// (word boundary) so benign words are NOT clobbered. For example:
//   - "filesystem:" and "system administrator" are preserved (no \b match
//     against "system:" because it requires a colon after the word).
//   - "developer experience" is preserved (no colon).
//   - "ignore the noise" is preserved (no "previous"/"above" anchor after).
//   - "previous version" is preserved (no override verb before "previous").
//
// Each regex matches a category of attack:
//
//	role-tag prefixes        — system:, assistant:, user:, developer:,
//	                           human:, model: (LLM role spoofing)
//	instruction-override     — ignore/disregard/forget/override/discard/skip
//	                           + optional quantifiers (all/the/every/of) +
//	                           previous/above/prior/earlier
//	"forget everything"      — forget + everything|all (no anchor needed)
//	LLM special markers      — <| and |> (e.g. <|endoftext|>, <|im_start|>)
//
// Extend this list cautiously: each new pattern must be anchored against
// false positives and accompanied by a benign-input regression test.
var injectionRegexps = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(system|assistant|user|developer|human|model)\s*:`),
	regexp.MustCompile(`(?i)\b(ignore|disregard|forget|override|discard|skip)(\s+(all|the|every|of))*\s+(previous|above|prior|earlier)\b`),
	regexp.MustCompile(`(?i)\bforget\s+(everything|all)\b`),
	regexp.MustCompile(`<\|`),
	regexp.MustCompile(`\|>`),
}

// SanitizePromptInput scrubs user-provided text before it flows into a prompt:
//   - drops invalid UTF-8 byte sequences,
//   - strips Unicode control (Cc) and format (Cf) characters — covers ASCII
//     control bytes, C1 controls, zero-width characters (ZWSP/ZWJ/ZWNJ/BOM),
//     and bidi-override marks (LRO/RLO/PDF/LRE/RLE/LRI/RLI/FSI/PDI), all of
//     which are standard prompt-injection vectors,
//   - removes known injection patterns (see injectionRegexps), iterating to
//     a fixed point so repeated occurrences and patterns exposed by earlier
//     removals are all caught,
//   - collapses whitespace and truncates to a safe length.
//
// The function strips offending content rather than rejecting it, matching the
// permissive policy used elsewhere in the package.
func SanitizePromptInput(s string) string {
	// Drop invalid UTF-8 byte sequences entirely.
	s = strings.ToValidUTF8(s, "")

	// Strip Cc (control) + Cf (format) characters. Cc covers ASCII control
	// bytes (\x00-\x1f, \x7f) and C1 controls; Cf covers zero-width chars
	// and bidi-override marks that are invisible in rendered text but can
	// alter how downstream models interpret the prompt.
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.In(r, unicode.Cc, unicode.Cf) {
			continue
		}
		b.WriteRune(r)
	}
	s = b.String()

	// Apply injection-pattern regexps to a fixed point. ReplaceAllString
	// already catches every non-overlapping match in a single pass, so
	// repeats like "ignore previous ignore previous ignore previous" fall
	// in the first iteration. The loop exists to catch patterns that are
	// only exposed AFTER an earlier replacement (e.g., "ignignore previousore
	// previous" → after one pass "ign previousore previous" → second
	// pass strips the newly-aligned "previous" — defense in depth).
	for range sanitizationIterationCap {
		before := s
		for _, re := range injectionRegexps {
			s = re.ReplaceAllString(s, " ")
		}
		if s == before {
			break
		}
	}

	// Collapse multiple spaces that may result from removals.
	s = strings.Join(strings.Fields(s), " ")

	// Truncate to maximum allowed length.
	if len(s) > maxPromptInputLen {
		s = s[:maxPromptInputLen]
	}

	return strings.TrimSpace(s)
}

// RandSource allows injecting a random source for deterministic testing.
type RandSource interface {
	Intn(n int) int
}

// defaultRandSource wraps the global rand functions.
type defaultRandSource struct{}

func (d defaultRandSource) Intn(n int) int {
	return rand.Intn(n)
}

// BuildPromptConfig selects random colors and placement for a thumbnail prompt.
// User-provided tagline and illustration text are sanitized to prevent prompt injection.
func BuildPromptConfig(tagline, illustration string, rng RandSource) PromptConfig {
	if rng == nil {
		rng = defaultRandSource{}
	}

	tagline = SanitizePromptInput(tagline)
	illustration = SanitizePromptInput(illustration)

	bg := ChannelPalette[rng.Intn(len(ChannelPalette))]
	tc := bg.TextColors[rng.Intn(len(bg.TextColors))]
	placement := PersonPlacements[rng.Intn(len(PersonPlacements))]

	return PromptConfig{
		Background:   bg,
		TextColor:    tc,
		Placement:    placement,
		Tagline:      tagline,
		Illustration: illustration,
	}
}

// BuildPrompt generates the full prompt text from a PromptConfig.
//
// Defensively sanitizes user-controlled fields (Tagline, Illustration) so that
// callers which construct PromptConfig directly — bypassing BuildPromptConfig
// — still get prompt-injection protection at the point closest to the prompt.
func BuildPrompt(cfg PromptConfig) string {
	cfg.Tagline = SanitizePromptInput(cfg.Tagline)
	cfg.Illustration = SanitizePromptInput(cfg.Illustration)

	bgLabel := fmt.Sprintf("%s (%s)", cfg.Background.Name, cfg.Background.Hex)
	textColorLabel := fmt.Sprintf("%s (%s)", cfg.TextColor.Name, cfg.TextColor.Hex)

	textLayoutDesc := buildTextLayoutDescription(cfg.Tagline, cfg.Placement)

	var sb strings.Builder

	sb.WriteString(`Create a YouTube thumbnail using the photo I've attached. The thumbnail must feature ME (the person in the attached photo) — do NOT generate a different person. Use my actual photo as the base.`)
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf(`**Background:** Solid flat %s. No gradients, no patterns — pure solid color.`, bgLabel))
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf(
		`**My photo:** Take the photo of me that I attached and convert it to high-contrast black and white — like a photocopy or stencil art. Keep my pose exactly as it is in the photo. Show me waist-up, positioned on the %s, overlapping the text. My body must extend to the bottom edge of the image — anchor me to the bottom so I look grounded, not floating. I must be facing/looking toward the %s side of the frame (toward the text, away from my placement). If the source photo has me facing the other way, horizontally mirror (flip left-to-right) the photo so I face the correct direction. If there is a microphone visible in the photo (handheld, on a stand, boom mic, or lapel), remove it completely from the image. The text is the main element, I am secondary.`,
		cfg.Placement.Description, cfg.Placement.FaceDirection))
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf(
		`**Text:** Display EXACTLY: "%s" — no extra, missing, or altered letters. Use a bold condensed sans-serif font (extremely narrow/tall letterforms like Impact). %s. The text must be ENORMOUS — letters should touch or nearly touch the top and bottom edges of the image. The text arrangement should fill the ENTIRE canvas so very little background color is visible. Arrange as an asymmetric poster layout: %s. Zero letter spacing. The text sits BEHIND me — it is fine if my figure partially overlaps and covers some letters, as long as the text remains readable. ALL letters must be fully visible within the image — nothing cut off at edges.`,
		cfg.Tagline, textColorLabel, textLayoutDesc))
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf(`**Layering order (back to front):**
1. Solid %s background
2. Massive "%s" text (filling 85-90%% of frame, very little background visible)
3. Threshold-processed photo of me, overlapping the text`, bgLabel, cfg.Tagline))
	sb.WriteString("\n\n")

	if cfg.Illustration != "" {
		sb.WriteString(fmt.Sprintf(
			`**Illustration:** Add a %s as a 3D-rendered element, placed to complement the composition without overwhelming the text or person. It should be secondary to both.`,
			cfg.Illustration))
		sb.WriteString("\n\n")
	}

	sb.WriteString(fmt.Sprintf(`**Rules:**
- You MUST use my attached photo — do NOT generate a different person
- Text must say EXACTLY "%s" — check spelling carefully, no extra or missing letters
- My photo must be high-contrast black and white (stencil/photocopy style)
- ALL text must be fully visible — no letters cut off at image edges
- Text is the dominant element; I am secondary
- Background is a single solid color, no gradients`, cfg.Tagline))

	return sb.String()
}

// BuildPhotoRealisticPromptConfig selects a random placement for the
// photo-realistic variant and sanitizes the subject input. Returns a config
// with an empty Subject if the sanitized subject is empty — callers should
// treat that as "skip this variant".
func BuildPhotoRealisticPromptConfig(subject string, rng RandSource) PhotoRealisticPromptConfig {
	if rng == nil {
		rng = defaultRandSource{}
	}

	subject = SanitizePromptInput(subject)
	placement := PersonPlacements[rng.Intn(len(PersonPlacements))]

	return PhotoRealisticPromptConfig{
		Placement: placement,
		Subject:   subject,
	}
}

// BuildPhotoRealisticPrompt generates the prompt text for the photo-realistic
// variant: creator photo rendered photo-realistically (no B&W/stencil), a
// contextual subject also rendered photo-realistically, and no text overlay.
//
// Defensively sanitizes cfg.Subject so callers that construct
// PhotoRealisticPromptConfig directly — bypassing BuildPhotoRealisticPromptConfig
// — still get prompt-injection protection at the point closest to the prompt.
//
// Returns ErrEmptySubject when the subject is empty after sanitization. The
// orchestrator skips this variant on a sentinel-error return rather than
// producing a malformed prompt.
func BuildPhotoRealisticPrompt(cfg PhotoRealisticPromptConfig) (string, error) {
	cfg.Subject = SanitizePromptInput(cfg.Subject)
	if cfg.Subject == "" {
		return "", ErrEmptySubject
	}

	var sb strings.Builder

	sb.WriteString(`Create a YouTube thumbnail using the photo I've attached. The thumbnail must feature ME (the person in the attached photo) — do NOT generate a different person. Use my actual photo as the base.`)
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf(
		`**My photo:** Render me in PHOTO-REALISTIC style — keep the natural colors, skin tones, lighting, and detail of the original photograph. Do NOT apply any threshold, stencil, photocopy, or black-and-white treatment. Do NOT posterize or flatten the image. Show me waist-up, positioned on the %s. My body must extend to the bottom edge of the image — anchor me to the bottom so I look grounded, not floating. I must be facing/looking toward the %s side of the frame. If the source photo has me facing the other way, horizontally mirror (flip left-to-right) the photo so I face the correct direction. If there is a microphone visible in the photo (handheld, on a stand, boom mic, or lapel), remove it completely from the image.`,
		cfg.Placement.Description, cfg.Placement.FaceDirection))
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf(
		`**Contextual subject:** Include %s as a PHOTO-REALISTIC element — rendered with natural lighting, realistic textures, materials, and lifelike detail. NOT flat, NOT a cartoon, NOT line art, NOT a stylized icon. Integrate the subject naturally into the composition alongside me so it reads as part of the same photograph.`,
		cfg.Subject))
	sb.WriteString("\n\n")

	sb.WriteString(`**Background:** A photo-realistic environment that complements the subject and the creator — with depth, realistic lighting, and natural detail. Do NOT use a solid flat color block. Do NOT use a stencil or graphic-poster aesthetic.`)
	sb.WriteString("\n\n")

	sb.WriteString(`**No text:** Do NOT render any text, tagline, title, caption, watermark, logo lettering, or written words anywhere in the image. Zero text of any kind. If a sign, label, screen, or surface in the scene would naturally contain text, leave it blank. If you are about to add text, stop and remove it.`)
	sb.WriteString("\n\n")

	sb.WriteString(`**Rules:**
- You MUST use my attached photo — do NOT generate a different person
- My photo MUST be photo-realistic — NO threshold, NO stencil, NO black-and-white treatment, NO posterization
- The contextual subject MUST be photo-realistic — NO flat illustration, NO cartoon, NO line art, NO stylized icon
- ZERO text in the image — no tagline, no title, no captions, no watermark, no letters of any kind`)

	return sb.String(), nil
}

// buildTextLayoutDescription creates a word-by-word positioning description
// based on the tagline and person placement.
func buildTextLayoutDescription(tagline string, placement PersonPlacement) string {
	words := strings.Fields(tagline)
	if len(words) == 0 {
		return "fill the canvas with the text"
	}

	// Determine which side has the text (opposite of person)
	textSide := "left"
	if strings.Contains(placement.Description, "left") {
		textSide = "right"
	}

	oppositeSide := "right"
	if textSide == "right" {
		oppositeSide = "left"
	}

	if len(words) == 1 {
		return fmt.Sprintf(`"%s" fills the entire canvas at maximum size`, words[0])
	}

	if len(words) == 2 {
		return fmt.Sprintf(
			`"%s" rotated 90° counterclockwise running down the %s edge, "%s" horizontal and massive filling the center-%s`,
			words[0], oppositeSide, words[1], textSide)
	}

	// For 3+ words, distribute across the canvas
	var parts []string
	for i, w := range words {
		switch {
		case i == 0:
			parts = append(parts, fmt.Sprintf(`"%s" rotated 90° counterclockwise running down the %s edge`, w, oppositeSide))
		case i == len(words)-1:
			parts = append(parts, fmt.Sprintf(`"%s" horizontal across the bottom`, w))
		default:
			parts = append(parts, fmt.Sprintf(`"%s" horizontal and massive in the center`, w))
		}
	}

	return strings.Join(parts, ", ")
}
