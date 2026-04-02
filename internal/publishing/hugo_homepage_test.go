package publishing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleHomepage = `+++
title = "DevOps Toolkit"
+++

# Latest Posts

<a href="/devops/old-post"><img src="/devops/old-post/thumbnail.jpg" style="width:50%; float:right; padding: 10px"></a>

## [Old Post](/devops/old-post)

Old intro text.

**[Full article >>](/devops/old-post)**

---

`

func setupHomepage(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	contentDir := filepath.Join(dir, "content")
	require.NoError(t, os.MkdirAll(contentDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte(content), 0644))
	return dir
}

func readHomepage(t *testing.T, basePath string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(basePath, "content", "_index.md"))
	require.NoError(t, err)
	return string(data)
}

func TestAddHomepageEntry(t *testing.T) {
	t.Run("inserts entry after header", func(t *testing.T) {
		basePath := setupHomepage(t, sampleHomepage)

		err := AddHomepageEntry(basePath, "devops", "new-post", "New Post Title", "New intro text.", true)
		require.NoError(t, err)

		result := readHomepage(t, basePath)
		assert.Contains(t, result, "## [New Post Title](/devops/new-post)")
		assert.Contains(t, result, `<a href="/devops/new-post"><img src="/devops/new-post/thumbnail.jpg"`)
		assert.Contains(t, result, "New intro text.")
		assert.Contains(t, result, `**[Full article >>](/devops/new-post)**`)

		// New entry should appear before old entry
		newIdx := strings.Index(result, "New Post Title")
		oldIdx := strings.Index(result, "Old Post")
		assert.Greater(t, oldIdx, newIdx, "new entry should come before old entry")
	})

	t.Run("error when header not found", func(t *testing.T) {
		basePath := setupHomepage(t, "no header here")
		err := AddHomepageEntry(basePath, "cat", "slug", "Title", "Intro", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not find")
	})

	t.Run("error when file missing", func(t *testing.T) {
		err := AddHomepageEntry("/nonexistent", "cat", "slug", "Title", "Intro", true)
		assert.Error(t, err)
	})
}

func TestTrimHomepageEntries(t *testing.T) {
	t.Run("trims to max entries", func(t *testing.T) {
		// Build a homepage with 12 entries
		var sb strings.Builder
		sb.WriteString("+++\ntitle = \"Test\"\n+++\n\n# Latest Posts\n\n")
		for i := 1; i <= 12; i++ {
			sb.WriteString("## Entry " + strings.Repeat("X", i) + "\n\nContent.\n\n---\n\n")
		}
		basePath := setupHomepage(t, sb.String())

		err := TrimHomepageEntries(basePath, 10)
		require.NoError(t, err)

		result := readHomepage(t, basePath)
		// Count "---" separators
		count := strings.Count(result, "\n---\n")
		assert.Equal(t, 10, count, "should have exactly 10 entries")
	})

	t.Run("does nothing when under limit", func(t *testing.T) {
		basePath := setupHomepage(t, sampleHomepage)
		err := TrimHomepageEntries(basePath, 10)
		require.NoError(t, err)

		result := readHomepage(t, basePath)
		assert.Contains(t, result, "Old Post")
	})

	t.Run("does nothing when no header", func(t *testing.T) {
		basePath := setupHomepage(t, "no header")
		err := TrimHomepageEntries(basePath, 10)
		assert.NoError(t, err)
	})

	t.Run("error when file missing", func(t *testing.T) {
		err := TrimHomepageEntries("/nonexistent", 10)
		assert.Error(t, err)
	})
}

func TestBuildHomepageEntry(t *testing.T) {
	t.Run("with thumbnail", func(t *testing.T) {
		entry := buildHomepageEntry("/devops/my-post", "My Post", "Intro text.", true)

		assert.Contains(t, entry, `<a href="/devops/my-post">`)
		assert.Contains(t, entry, `<img src="/devops/my-post/thumbnail.jpg"`)
		assert.Contains(t, entry, "## [My Post](/devops/my-post)")
		assert.Contains(t, entry, "Intro text.")
		assert.Contains(t, entry, `**[Full article >>](/devops/my-post)**`)
		assert.Contains(t, entry, "---")
	})

	t.Run("without thumbnail", func(t *testing.T) {
		entry := buildHomepageEntry("/devops/my-post", "My Post", "Intro text.", false)

		assert.NotContains(t, entry, "thumbnail.jpg")
		assert.Contains(t, entry, "## [My Post](/devops/my-post)")
		assert.Contains(t, entry, "Intro text.")
		assert.Contains(t, entry, `**[Full article >>](/devops/my-post)**`)
	})
}
