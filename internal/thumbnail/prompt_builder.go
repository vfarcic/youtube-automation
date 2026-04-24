package thumbnail

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

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
	Background    BackgroundColor
	TextColor     TextColor
	Placement     PersonPlacement
	Tagline       string
	Illustration  string // empty means no illustration
}

const maxPromptInputLen = 200

// controlCharsRegexp matches control characters (except space) that should be stripped from prompt inputs.
var controlCharsRegexp = regexp.MustCompile(`[\x00-\x1f\x7f]`)

// injectionPatterns are phrases commonly used in prompt injection attempts.
var injectionPatterns = []string{
	"ignore previous",
	"ignore above",
	"disregard previous",
	"disregard above",
	"forget previous",
	"forget above",
	"system:",
	"assistant:",
	"user:",
	"<|",
	"|>",
}

// SanitizePromptInput strips control characters, prompt injection patterns,
// and truncates user-provided text to a safe length.
func SanitizePromptInput(s string) string {
	// Strip control characters.
	s = controlCharsRegexp.ReplaceAllString(s, "")

	// Remove known injection patterns (case-insensitive).
	lower := strings.ToLower(s)
	for _, pat := range injectionPatterns {
		if idx := strings.Index(lower, pat); idx != -1 {
			s = s[:idx] + s[idx+len(pat):]
			lower = strings.ToLower(s)
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
func BuildPrompt(cfg PromptConfig) string {
	bgLabel := fmt.Sprintf("%s (%s)", cfg.Background.Name, cfg.Background.Hex)
	textColorLabel := fmt.Sprintf("%s (%s)", cfg.TextColor.Name, cfg.TextColor.Hex)

	textLayoutDesc := buildTextLayoutDescription(cfg.Tagline, cfg.Placement)

	var sb strings.Builder

	sb.WriteString(`Create a YouTube thumbnail using the photo I've attached. The thumbnail must feature ME (the person in the attached photo) — do NOT generate a different person. Use my actual photo as the base.`)
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf(`**Background:** Solid flat %s. No gradients, no patterns — pure solid color.`, bgLabel))
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf(
		`**My photo:** Take the photo of me that I attached and convert it to high-contrast black and white — like a photocopy or stencil art. Keep my pose exactly as it is in the photo. Show me waist-up, positioned on the %s, overlapping the text. My body must extend to the bottom edge of the image — anchor me to the bottom so I look grounded, not floating. I must be facing/looking toward the %s side of the frame (toward the text, away from my placement). If the source photo has me facing the other way, horizontally mirror (flip left-to-right) the photo so I face the correct direction. The text is the main element, I am secondary.`,
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
