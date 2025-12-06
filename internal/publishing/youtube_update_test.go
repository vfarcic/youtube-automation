package publishing

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple comma-separated tags",
			input:    "kubernetes,docker,devops",
			expected: []string{"kubernetes", "docker", "devops"},
		},
		{
			name:     "tags with spaces",
			input:    " kubernetes , docker , devops ",
			expected: []string{"kubernetes", "docker", "devops"},
		},
		{
			name:     "single tag",
			input:    "kubernetes",
			expected: []string{"kubernetes"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "tags with empty entries",
			input:    "kubernetes,,docker,,devops",
			expected: []string{"kubernetes", "docker", "devops"},
		},
		{
			name:     "only commas",
			input:    ",,,",
			expected: []string{},
		},
		{
			name:     "tags with special characters",
			input:    "CI/CD,GitOps,cloud-native",
			expected: []string{"CI/CD", "GitOps", "cloud-native"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseTags(%q) returned %d tags, expected %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("parseTags(%q)[%d] = %q, expected %q", tt.input, i, tag, tt.expected[i])
				}
			}
		})
	}
}

func TestExtractBoilerplate(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "description with boilerplate",
			description: "This is my video description.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nConsider joining the channel: https://youtube.com/...\n\n▬▬▬▬▬▬ 🔗 Additional Info 🔗 ▬▬▬▬▬▬\nMore info here",
			expected:    "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nConsider joining the channel: https://youtube.com/...\n\n▬▬▬▬▬▬ 🔗 Additional Info 🔗 ▬▬▬▬▬▬\nMore info here",
		},
		{
			name:        "description without boilerplate",
			description: "This is just a simple description without any boilerplate.",
			expected:    "",
		},
		{
			name:        "empty description",
			description: "",
			expected:    "",
		},
		{
			name:        "description with boilerplate and timecodes",
			description: "My description.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate content\n\n▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬\n00:00 Intro\n02:00 Question 1",
			expected:    "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate content",
		},
		{
			name:        "boilerplate at start",
			description: "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nOnly boilerplate",
			expected:    "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nOnly boilerplate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBoilerplate(tt.description)
			if result != tt.expected {
				t.Errorf("extractBoilerplate() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestBuildAMADescription(t *testing.T) {
	tests := []struct {
		name               string
		newDescription     string
		currentDescription string
		timecodes          string
		shouldContain      []string
		shouldNotContain   []string
	}{
		{
			name:               "new description with boilerplate and timecodes",
			newDescription:     "This is my new AMA description about Kubernetes.",
			currentDescription: "Old description.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nConsider joining: https://youtube.com/...",
			timecodes:          "00:00 Intro\n02:30 Question about GitOps",
			shouldContain: []string{
				"This is my new AMA description about Kubernetes.",
				"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬",
				"Consider joining: https://youtube.com/...",
				"▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬",
				"00:00 Intro",
				"02:30 Question about GitOps",
			},
			shouldNotContain: []string{
				"Old description.",
			},
		},
		{
			name:               "only new description, no current boilerplate",
			newDescription:     "Brand new description.",
			currentDescription: "",
			timecodes:          "00:00 Start\n01:00 Topic 1",
			shouldContain: []string{
				"Brand new description.",
				"▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬",
				"00:00 Start",
			},
			shouldNotContain: []string{},
		},
		{
			name:               "empty new description preserves boilerplate",
			newDescription:     "",
			currentDescription: "Some text.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate here",
			timecodes:          "00:00 Intro",
			shouldContain: []string{
				"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬",
				"Boilerplate here",
				"▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬",
				"00:00 Intro",
			},
			shouldNotContain: []string{
				"Some text.",
			},
		},
		{
			name:               "no timecodes provided",
			newDescription:     "My description.",
			currentDescription: "Old.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate",
			timecodes:          "",
			shouldContain: []string{
				"My description.",
				"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬",
				"Boilerplate",
			},
			shouldNotContain: []string{
				"▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬",
			},
		},
		{
			name:               "replaces existing timecodes",
			newDescription:     "New content.",
			currentDescription: "Old.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate\n\n▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬\n00:00 Old intro",
			timecodes:          "00:00 New intro\n05:00 New question",
			shouldContain: []string{
				"New content.",
				"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬",
				"Boilerplate",
				"00:00 New intro",
				"05:00 New question",
			},
			shouldNotContain: []string{
				"Old.",
				"00:00 Old intro",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAMADescription(tt.newDescription, tt.currentDescription, tt.timecodes)

			for _, s := range tt.shouldContain {
				if !contains(result, s) {
					t.Errorf("buildAMADescription() should contain %q, but got:\n%s", s, result)
				}
			}

			for _, s := range tt.shouldNotContain {
				if contains(result, s) {
					t.Errorf("buildAMADescription() should NOT contain %q, but got:\n%s", s, result)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestBuildAMADescriptionOrder(t *testing.T) {
	newDesc := "New description here."
	currentDesc := "Old.\n\n▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\nBoilerplate content"
	timecodes := "00:00 Intro"

	result := buildAMADescription(newDesc, currentDesc, timecodes)

	// Verify order: new description comes before boilerplate, timecodes come last
	newDescIdx := indexOf(result, "New description here.")
	boilerplateIdx := indexOf(result, "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬")
	timecodesIdx := indexOf(result, "▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬")

	if newDescIdx == -1 || boilerplateIdx == -1 || timecodesIdx == -1 {
		t.Fatalf("Missing expected content in result:\n%s", result)
	}

	if newDescIdx >= boilerplateIdx {
		t.Errorf("New description should come before boilerplate. newDesc at %d, boilerplate at %d", newDescIdx, boilerplateIdx)
	}

	if boilerplateIdx >= timecodesIdx {
		t.Errorf("Boilerplate should come before timecodes. boilerplate at %d, timecodes at %d", boilerplateIdx, timecodesIdx)
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestGetVideoMetadataEmptyID(t *testing.T) {
	_, err := GetVideoMetadata("")
	if err == nil {
		t.Error("GetVideoMetadata should return error for empty video ID")
	}
	expectedErr := "video ID cannot be empty"
	if err.Error() != expectedErr {
		t.Errorf("GetVideoMetadata error = %q, expected %q", err.Error(), expectedErr)
	}
}

func TestUpdateAMAVideoEmptyID(t *testing.T) {
	err := UpdateAMAVideo("", "title", "desc", "tags", "timecodes")
	if err == nil {
		t.Error("UpdateAMAVideo should return error for empty video ID")
	}
	expectedErr := "video ID cannot be empty"
	if err.Error() != expectedErr {
		t.Errorf("UpdateAMAVideo error = %q, expected %q", err.Error(), expectedErr)
	}
}
