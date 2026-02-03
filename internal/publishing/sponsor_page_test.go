package publishing

import (
	"os"
	"path/filepath"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
)

func TestGetSponsorPagePath(t *testing.T) {
	// Save and restore original setting
	originalPath := configuration.GlobalSettings.Hugo.Path
	defer func() { configuration.GlobalSettings.Hugo.Path = originalPath }()

	tests := []struct {
		name     string
		hugoPath string
		want     string
	}{
		{
			name:     "standard path",
			hugoPath: "/home/user/hugo-site",
			want:     "/home/user/hugo-site/content/sponsor/_index.md",
		},
		{
			name:     "path with trailing slash",
			hugoPath: "/var/www/blog/",
			want:     "/var/www/blog/content/sponsor/_index.md",
		},
		{
			name:     "empty path",
			hugoPath: "",
			want:     "content/sponsor/_index.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configuration.GlobalSettings.Hugo.Path = tt.hugoPath
			got := GetSponsorPagePath()
			if got != tt.want {
				t.Errorf("GetSponsorPagePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdateContentBetweenMarkers(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		newSection string
		want       string
	}{
		{
			name: "markers exist - replace content",
			content: `# Sponsor Page

Some intro text.

<!-- SPONSOR_ANALYTICS_START -->
Old analytics content here
<!-- SPONSOR_ANALYTICS_END -->

Footer content.`,
			newSection: `<!-- SPONSOR_ANALYTICS_START -->
New analytics content
<!-- SPONSOR_ANALYTICS_END -->`,
			want: `# Sponsor Page

Some intro text.

<!-- SPONSOR_ANALYTICS_START -->
New analytics content
<!-- SPONSOR_ANALYTICS_END -->

Footer content.`,
		},
		{
			name: "markers don't exist - append to end",
			content: `# Sponsor Page

Some content here.`,
			newSection: `<!-- SPONSOR_ANALYTICS_START -->
Analytics content
<!-- SPONSOR_ANALYTICS_END -->`,
			want: `# Sponsor Page

Some content here.

<!-- SPONSOR_ANALYTICS_START -->
Analytics content
<!-- SPONSOR_ANALYTICS_END -->`,
		},
		{
			name:       "empty content - just return new section",
			content:    "",
			newSection: "<!-- SPONSOR_ANALYTICS_START -->\nContent\n<!-- SPONSOR_ANALYTICS_END -->",
			want:       "<!-- SPONSOR_ANALYTICS_START -->\nContent\n<!-- SPONSOR_ANALYTICS_END -->",
		},
		{
			name: "only start marker exists - append to end",
			content: `# Page
<!-- SPONSOR_ANALYTICS_START -->
Some incomplete content`,
			newSection: "<!-- SPONSOR_ANALYTICS_START -->\nNew\n<!-- SPONSOR_ANALYTICS_END -->",
			want: `# Page
<!-- SPONSOR_ANALYTICS_START -->
Some incomplete content

<!-- SPONSOR_ANALYTICS_START -->
New
<!-- SPONSOR_ANALYTICS_END -->`,
		},
		{
			name: "only end marker exists - append to end",
			content: `# Page
Some content
<!-- SPONSOR_ANALYTICS_END -->`,
			newSection: "<!-- SPONSOR_ANALYTICS_START -->\nNew\n<!-- SPONSOR_ANALYTICS_END -->",
			want: `# Page
Some content
<!-- SPONSOR_ANALYTICS_END -->

<!-- SPONSOR_ANALYTICS_START -->
New
<!-- SPONSOR_ANALYTICS_END -->`,
		},
		{
			name: "markers in wrong order - append to end",
			content: `# Page
<!-- SPONSOR_ANALYTICS_END -->
Some content
<!-- SPONSOR_ANALYTICS_START -->`,
			newSection: "<!-- SPONSOR_ANALYTICS_START -->\nNew\n<!-- SPONSOR_ANALYTICS_END -->",
			want: `# Page
<!-- SPONSOR_ANALYTICS_END -->
Some content
<!-- SPONSOR_ANALYTICS_START -->

<!-- SPONSOR_ANALYTICS_START -->
New
<!-- SPONSOR_ANALYTICS_END -->`,
		},
		{
			name: "content without trailing newline - adds newline before append",
			content: `# Page
No trailing newline`,
			newSection: "<!-- SPONSOR_ANALYTICS_START -->\nContent\n<!-- SPONSOR_ANALYTICS_END -->",
			want: `# Page
No trailing newline

<!-- SPONSOR_ANALYTICS_START -->
Content
<!-- SPONSOR_ANALYTICS_END -->`,
		},
		{
			name: "content with trailing newline - preserves format",
			content: `# Page
Has trailing newline
`,
			newSection: "<!-- SPONSOR_ANALYTICS_START -->\nContent\n<!-- SPONSOR_ANALYTICS_END -->",
			want: `# Page
Has trailing newline

<!-- SPONSOR_ANALYTICS_START -->
Content
<!-- SPONSOR_ANALYTICS_END -->`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateContentBetweenMarkers(tt.content, SponsorAnalyticsStartMarker, SponsorAnalyticsEndMarker, tt.newSection)
			if got != tt.want {
				t.Errorf("updateContentBetweenMarkers() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestReadSponsorPage(t *testing.T) {
	// Save and restore original setting
	originalPath := configuration.GlobalSettings.Hugo.Path
	defer func() { configuration.GlobalSettings.Hugo.Path = originalPath }()

	t.Run("file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		sponsorDir := filepath.Join(tmpDir, "content", "sponsor")
		if err := os.MkdirAll(sponsorDir, 0755); err != nil {
			t.Fatalf("failed to create sponsor dir: %v", err)
		}

		expectedContent := "# Sponsor Page\n\nContent here."
		sponsorFile := filepath.Join(sponsorDir, "_index.md")
		if err := os.WriteFile(sponsorFile, []byte(expectedContent), 0644); err != nil {
			t.Fatalf("failed to write sponsor file: %v", err)
		}

		configuration.GlobalSettings.Hugo.Path = tmpDir

		got, err := ReadSponsorPage()
		if err != nil {
			t.Errorf("ReadSponsorPage() error = %v, want nil", err)
		}
		if got != expectedContent {
			t.Errorf("ReadSponsorPage() = %q, want %q", got, expectedContent)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configuration.GlobalSettings.Hugo.Path = tmpDir

		_, err := ReadSponsorPage()
		if err == nil {
			t.Error("ReadSponsorPage() error = nil, want error")
		}
	})
}

func TestUpdateSponsorPageAnalytics(t *testing.T) {
	// Save and restore original setting
	originalPath := configuration.GlobalSettings.Hugo.Path
	defer func() { configuration.GlobalSettings.Hugo.Path = originalPath }()

	t.Run("update existing markers", func(t *testing.T) {
		tmpDir := t.TempDir()
		sponsorDir := filepath.Join(tmpDir, "content", "sponsor")
		if err := os.MkdirAll(sponsorDir, 0755); err != nil {
			t.Fatalf("failed to create sponsor dir: %v", err)
		}

		initialContent := `# Sponsor Page

Intro text.

<!-- SPONSOR_ANALYTICS_START -->
Old content
<!-- SPONSOR_ANALYTICS_END -->

Footer.`
		sponsorFile := filepath.Join(sponsorDir, "_index.md")
		if err := os.WriteFile(sponsorFile, []byte(initialContent), 0644); err != nil {
			t.Fatalf("failed to write sponsor file: %v", err)
		}

		configuration.GlobalSettings.Hugo.Path = tmpDir

		newSection := `<!-- SPONSOR_ANALYTICS_START -->
## Channel Analytics
New content here
<!-- SPONSOR_ANALYTICS_END -->`

		err := UpdateSponsorPageAnalytics(newSection)
		if err != nil {
			t.Fatalf("UpdateSponsorPageAnalytics() error = %v", err)
		}

		// Read back and verify
		content, err := os.ReadFile(sponsorFile)
		if err != nil {
			t.Fatalf("failed to read sponsor file: %v", err)
		}

		expected := `# Sponsor Page

Intro text.

<!-- SPONSOR_ANALYTICS_START -->
## Channel Analytics
New content here
<!-- SPONSOR_ANALYTICS_END -->

Footer.`
		if string(content) != expected {
			t.Errorf("File content =\n%q\nwant:\n%q", string(content), expected)
		}
	})

	t.Run("append when no markers exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		sponsorDir := filepath.Join(tmpDir, "content", "sponsor")
		if err := os.MkdirAll(sponsorDir, 0755); err != nil {
			t.Fatalf("failed to create sponsor dir: %v", err)
		}

		initialContent := `# Sponsor Page

Some content here.`
		sponsorFile := filepath.Join(sponsorDir, "_index.md")
		if err := os.WriteFile(sponsorFile, []byte(initialContent), 0644); err != nil {
			t.Fatalf("failed to write sponsor file: %v", err)
		}

		configuration.GlobalSettings.Hugo.Path = tmpDir

		newSection := `<!-- SPONSOR_ANALYTICS_START -->
Analytics
<!-- SPONSOR_ANALYTICS_END -->`

		err := UpdateSponsorPageAnalytics(newSection)
		if err != nil {
			t.Fatalf("UpdateSponsorPageAnalytics() error = %v", err)
		}

		// Read back and verify
		content, err := os.ReadFile(sponsorFile)
		if err != nil {
			t.Fatalf("failed to read sponsor file: %v", err)
		}

		expected := `# Sponsor Page

Some content here.

<!-- SPONSOR_ANALYTICS_START -->
Analytics
<!-- SPONSOR_ANALYTICS_END -->`
		if string(content) != expected {
			t.Errorf("File content =\n%q\nwant:\n%q", string(content), expected)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configuration.GlobalSettings.Hugo.Path = tmpDir

		err := UpdateSponsorPageAnalytics("new section")
		if err == nil {
			t.Error("UpdateSponsorPageAnalytics() error = nil, want error")
		}
	})
}
