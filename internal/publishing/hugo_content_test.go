package publishing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractIntro(t *testing.T) {
	tests := []struct {
		name          string
		manuscript    string
		wantIntro     string
		wantBodyPart  string // substring expected in body
		introMissing  bool   // if true, expect empty intro
	}{
		{
			name: "standard intro section",
			manuscript: `# Title

## Intro

This is the intro paragraph.

It has multiple lines.

## Next Section

Body content here.`,
			wantIntro:    "This is the intro paragraph.\n\nIt has multiple lines.",
			wantBodyPart: "## Next Section",
		},
		{
			name: "intro at end of document",
			manuscript: `# Title

## Setup

Setup content.

## Intro

Final intro paragraph.`,
			wantIntro:    "Final intro paragraph.",
			wantBodyPart: "## Setup",
		},
		{
			name:         "no intro section",
			manuscript:   "# Title\n\n## Setup\n\nContent here.",
			wantIntro:    "",
			introMissing: true,
			wantBodyPart: "# Title",
		},
		{
			name: "intro with empty lines",
			manuscript: `## Intro

First paragraph.

Second paragraph.

## Body

Rest of content.`,
			wantIntro:    "First paragraph.\n\nSecond paragraph.",
			wantBodyPart: "## Body",
		},
		{
			name:         "empty manuscript",
			manuscript:   "",
			wantIntro:    "",
			introMissing: true,
		},
		{
			name: "intro section removes from body",
			manuscript: `## Intro

The intro text.

## Details

Details here.`,
			wantIntro:    "The intro text.",
			wantBodyPart: "## Details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intro, body := ExtractIntro(tt.manuscript)
			if tt.introMissing {
				assert.Empty(t, intro)
			} else {
				assert.Equal(t, tt.wantIntro, intro)
			}
			if tt.wantBodyPart != "" {
				assert.Contains(t, body, tt.wantBodyPart)
			}
			// Intro content should not appear in body
			if intro != "" {
				assert.NotContains(t, body, "## Intro")
			}
		})
	}
}

func TestRemoveTODOAndFIXMELines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes TODO lines",
			input: "line 1\nTODO: fix this\nline 3",
			want:  "line 1\nline 3",
		},
		{
			name:  "removes FIXME lines",
			input: "line 1\nFIXME: broken\nline 3",
			want:  "line 1\nline 3",
		},
		{
			name:  "removes indented TODO",
			input: "line 1\n  TODO: indented\nline 3",
			want:  "line 1\nline 3",
		},
		{
			name:  "removes indented FIXME",
			input: "line 1\n\tFIXME: tabbed\nline 3",
			want:  "line 1\nline 3",
		},
		{
			name:  "keeps lines with TODO in middle",
			input: "This is not a TODO: line\nTODO: this is",
			want:  "This is not a TODO: line",
		},
		{
			name:  "no removals needed",
			input: "line 1\nline 2\nline 3",
			want:  "line 1\nline 2\nline 3",
		},
		{
			name:  "removes multiple",
			input: "TODO: first\nkeep\nFIXME: second\nTODO: third",
			want:  "keep",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, RemoveTODOAndFIXMELines(tt.input))
		})
	}
}

func TestParseImageReferences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single image",
			input: "text ![alt](image.png) more",
			want:  []string{"image.png"},
		},
		{
			name:  "multiple images",
			input: "![](one.png)\n![caption](two.jpg)",
			want:  []string{"one.png", "two.jpg"},
		},
		{
			name:  "skips URLs",
			input: "![](https://example.com/img.png)\n![](local.png)",
			want:  []string{"local.png"},
		},
		{
			name:  "skips http URLs",
			input: "![](http://example.com/img.png)",
			want:  nil,
		},
		{
			name:  "no images",
			input: "just text content",
			want:  nil,
		},
		{
			name:  "image with path",
			input: "![](images/diagram.png)",
			want:  []string{"images/diagram.png"},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseImageReferences(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildHugoPost(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		result := BuildHugoPost("My Title", "2024-01-15T10:00", "vid123", "Intro text here.", "Body content.")
		assert.Contains(t, result, `title = "My Title"`)
		assert.Contains(t, result, "date = 2024-01-15T10:00:00+00:00")
		assert.Contains(t, result, "draft = false")
		assert.Contains(t, result, "Intro text here.")
		assert.Contains(t, result, "<!--more-->")
		assert.Contains(t, result, "{{< youtube vid123 >}}")
		assert.Contains(t, result, "Body content.")

		// Verify intro comes before <!--more-->
		introIdx := contentIndexOf(result, "Intro text here.")
		moreIdx := contentIndexOf(result, "<!--more-->")
		assert.Greater(t, moreIdx, introIdx, "intro should come before <!--more-->")
	})

	t.Run("without video ID", func(t *testing.T) {
		result := BuildHugoPost("Title", "2024-01-15T10:00", "", "Intro.", "Body.")
		assert.NotContains(t, result, "youtube")
		assert.NotContains(t, result, "FIXME")
	})

	t.Run("without intro", func(t *testing.T) {
		result := BuildHugoPost("Title", "2024-01-15T10:00", "vid1", "", "Body only.")
		assert.Contains(t, result, "<!--more-->")
		assert.Contains(t, result, "Body only.")
	})

	t.Run("no FIXME in output", func(t *testing.T) {
		result := BuildHugoPost("Title", "2024-01-15T10:00", "vid1", "Intro.", "Body.")
		assert.NotContains(t, result, "FIXME")
	})
}

func contentIndexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
