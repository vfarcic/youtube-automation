package publishing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
	gitpkg "devopstoolkit/youtube-automation/internal/git"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

// mockExecutor records git commands and returns preconfigured results.
type mockExecutor struct {
	calls   []mockCall
	results []mockResult // matched in order of calls; default is success
}

type mockCall struct {
	Dir  string
	Name string
	Args []string
}

type mockResult struct {
	Output []byte
	Err    error
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{}
}

// pushResult enqueues a result that will be returned for the next unmatched call.
func (m *mockExecutor) pushResult(output []byte, err error) {
	m.results = append(m.results, mockResult{Output: output, Err: err})
}

func (m *mockExecutor) Run(dir string, name string, args ...string) ([]byte, error) {
	idx := len(m.calls)
	m.calls = append(m.calls, mockCall{Dir: dir, Name: name, Args: args})
	if idx < len(m.results) {
		return m.results[idx].Output, m.results[idx].Err
	}
	return []byte{}, nil
}

// roundTripFunc implements HTTPClient for testing.
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// testVideo creates a storage.Video for testing with the given parameters.
func testVideo(gist, title, date, videoId string) *storage.Video {
	return &storage.Video{
		Gist:    gist,
		Titles:  []storage.TitleVariant{{Index: 1, Text: title}},
		Date:    date,
		VideoId: videoId,
		Name:    "test-video",
	}
}

// --- NewHugo tests ---

func TestNewHugo(t *testing.T) {
	t.Run("defaults branch to main", func(t *testing.T) {
		h := NewHugo(configuration.SettingsHugo{Path: "/tmp/hugo"})
		assert.Equal(t, "main", h.branch)
		assert.Equal(t, "/tmp/hugo", h.path)
	})

	t.Run("preserves configured branch", func(t *testing.T) {
		h := NewHugo(configuration.SettingsHugo{Branch: "develop"})
		assert.Equal(t, "develop", h.branch)
	})

	t.Run("stores repo URL and token", func(t *testing.T) {
		h := NewHugo(configuration.SettingsHugo{
			RepoURL: "https://github.com/user/repo.git",
			Token:   "ghp_abc",
		})
		assert.Equal(t, "https://github.com/user/repo.git", h.repoURL)
		assert.Equal(t, "ghp_abc", h.token)
	})
}

// --- local mode tests ---

func TestHugoFunctionErrors(t *testing.T) {
	tempDir := t.TempDir()

	originalSettings := configuration.GlobalSettings
	defer func() { configuration.GlobalSettings = originalSettings }()
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{Path: tempDir},
	}

	testContent := "Test content"
	hugo := NewHugo(configuration.SettingsHugo{Path: tempDir})

	t.Run("MkdirAll error", func(t *testing.T) {
		blockerFile := filepath.Join(tempDir, "content", "test-cat")
		require.NoError(t, os.MkdirAll(filepath.Dir(blockerFile), 0755))
		require.NoError(t, os.WriteFile(blockerFile, []byte("blocker"), 0644))

		_, err := hugo.hugoFromMarkdown(
			filepath.Join(tempDir, "manuscript", "test-cat", "file.md"),
			"Test Title",
			testContent,
			tempDir,
		)
		assert.Error(t, err)
	})

	t.Run("WriteFile error", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "content", "readonly")
		require.NoError(t, os.MkdirAll(testDir, 0755))

		parentDir := filepath.Join(testDir, "test-title")
		require.NoError(t, os.MkdirAll(parentDir, 0500))

		_, err := hugo.hugoFromMarkdown(
			filepath.Join(tempDir, "manuscript", "readonly", "file.md"),
			"Test Title",
			testContent,
			tempDir,
		)
		if err == nil {
			t.Log("WriteFile error test passed unexpectedly, possibly running with elevated permissions")
		}
	})
}

func TestHugoIntegration(t *testing.T) {
	tempDir := t.TempDir()

	manuscriptDir := filepath.Join(tempDir, "manuscript", "test-category")
	contentDir := filepath.Join(tempDir, "content", "test-category")
	require.NoError(t, os.MkdirAll(manuscriptDir, 0755))
	require.NoError(t, os.MkdirAll(contentDir, 0755))

	originalSettings := configuration.GlobalSettings
	defer func() { configuration.GlobalSettings = originalSettings }()
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{Path: tempDir},
	}

	complexContent := `# Complex Test Post

## Intro

This is the intro paragraph for testing.

## Code Block
` + "```go" + `
func main() {
    fmt.Println("Hello, Hugo!")
}
` + "```" + `

## List
- Item 1
- Item 2
- Item 3
`

	testFilePath := filepath.Join(manuscriptDir, "complex-post.md")
	require.NoError(t, os.WriteFile(testFilePath, []byte(complexContent), 0644))

	// Create homepage for enrichment
	homepageContent := "+++\ntitle = \"Test\"\n+++\n\n# Latest Posts\n\n---\n\n"
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "content", "_index.md"), []byte(homepageContent), 0644))

	hugo := NewHugo(configuration.SettingsHugo{Path: tempDir})

	t.Run("Post with regular title", func(t *testing.T) {
		video := testVideo(testFilePath, "Test Hugo Post", "2023-05-15T12:00", "testVideoId123")
		hugoPath, err := hugo.Post(video, nil)
		require.NoError(t, err)

		expectedPath := filepath.Join(tempDir, "content", "test-category", "test-hugo-post", "_index.md")
		assert.Equal(t, expectedPath, hugoPath)

		content, err := os.ReadFile(hugoPath)
		require.NoError(t, err)

		for _, expected := range []string{
			`title = "Test Hugo Post"`,
			"date = 2023-05-15T12:00:00+00:00",
			"draft = false",
			"{{< youtube testVideoId123 >}}",
			"# Complex Test Post",
		} {
			assert.Contains(t, string(content), expected)
		}
		// Intro should be above <!--more-->
		assert.Contains(t, string(content), "This is the intro paragraph for testing.")
		// No FIXME in output
		assert.NotContains(t, string(content), "FIXME")
	})

	t.Run("Post with special characters in title", func(t *testing.T) {
		video := testVideo(testFilePath, "Test: Hugo & Post (Special) Characters!'", "2023-05-15T12:00", "anotherIdAbc")
		hugoPath, err := hugo.Post(video, nil)
		require.NoError(t, err)

		dirName := filepath.Base(filepath.Dir(hugoPath))
		assert.False(t, strings.ContainsAny(dirName, ":&()!'"))
	})

	t.Run("Post with N/A gist", func(t *testing.T) {
		video := testVideo("N/A", "Test Title", "2023-05-15T12:00", "")
		hugoPath, err := hugo.Post(video, nil)
		assert.NoError(t, err)
		assert.Empty(t, hugoPath)
	})

	t.Run("Post with non-existent file", func(t *testing.T) {
		video := testVideo(filepath.Join(manuscriptDir, "non-existent.md"), "Test Title", "2023-05-15T12:00", "")
		_, err := hugo.Post(video, nil)
		assert.Error(t, err)
	})

	t.Run("Post without VideoID", func(t *testing.T) {
		video := testVideo(testFilePath, "Test Post No Video ID", "2023-05-17T10:00", "")
		hugoPath, err := hugo.Post(video, nil)
		require.NoError(t, err)

		content, err := os.ReadFile(hugoPath)
		require.NoError(t, err)
		// No FIXME placeholder — without videoId, youtube shortcode is simply omitted
		assert.NotContains(t, string(content), "FIXME")
	})

	t.Run("Post with VideoID", func(t *testing.T) {
		video := testVideo(testFilePath, "Test Post With Video ID", "2023-05-18T10:00", "actualVideoId12345")
		hugoPath, err := hugo.Post(video, nil)
		require.NoError(t, err)

		content, err := os.ReadFile(hugoPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "{{< youtube actualVideoId12345 >}}")
	})

	t.Run("Post with question mark in title", func(t *testing.T) {
		video := testVideo(testFilePath, "What is Go? A Test Post", "2023-05-16T10:00", "whatIsGoVideo789")
		hugoPath, err := hugo.Post(video, nil)
		require.NoError(t, err)

		assert.False(t, strings.Contains(hugoPath, "?"))
		expectedPath := filepath.Join(tempDir, "content", "test-category", "what-is-go-a-test-post", "_index.md")
		assert.Equal(t, expectedPath, hugoPath)
	})

	t.Run("Post with directory creation error", func(t *testing.T) {
		blockerPath := filepath.Join(tempDir, "content", "test-category", "test-blocked-dir")
		require.NoError(t, os.MkdirAll(filepath.Dir(blockerPath), 0755))
		require.NoError(t, os.WriteFile(blockerPath, []byte("blocker"), 0644))

		video := testVideo(testFilePath, "Test Blocked Dir", "2023-05-15T12:00", "blockedVideoId")
		_, err := hugo.Post(video, nil)
		assert.Error(t, err)
	})

	t.Run("Post removes TODO and FIXME from manuscript", func(t *testing.T) {
		contentWithTodos := `## Intro

Intro text.

## Body

Real content here.
TODO: remove this later
FIXME: broken thing
More real content.
`
		todoFilePath := filepath.Join(manuscriptDir, "todo-post.md")
		require.NoError(t, os.WriteFile(todoFilePath, []byte(contentWithTodos), 0644))

		video := testVideo(todoFilePath, "Post With Todos", "2023-05-19T10:00", "vid1")
		hugoPath, err := hugo.Post(video, nil)
		require.NoError(t, err)

		content, err := os.ReadFile(hugoPath)
		require.NoError(t, err)
		assert.NotContains(t, string(content), "TODO:")
		assert.NotContains(t, string(content), "FIXME:")
		assert.Contains(t, string(content), "Real content here.")
		assert.Contains(t, string(content), "More real content.")
	})

	t.Run("Post extracts intro as excerpt", func(t *testing.T) {
		introContent := `## Intro

This is the excerpt.

## Main Content

Body goes here.
`
		introFilePath := filepath.Join(manuscriptDir, "intro-post.md")
		require.NoError(t, os.WriteFile(introFilePath, []byte(introContent), 0644))

		video := testVideo(introFilePath, "Intro Excerpt Post", "2023-05-20T10:00", "vid2")
		hugoPath, err := hugo.Post(video, nil)
		require.NoError(t, err)

		content, err := os.ReadFile(hugoPath)
		require.NoError(t, err)
		s := string(content)

		// Verify intro is above <!--more-->
		introIdx := strings.Index(s, "This is the excerpt.")
		moreIdx := strings.Index(s, "<!--more-->")
		assert.NotEqual(t, -1, introIdx)
		assert.NotEqual(t, -1, moreIdx)
		assert.Less(t, introIdx, moreIdx, "intro should come before <!--more-->")

		// Body should be below <!--more-->
		bodyIdx := strings.Index(s, "Body goes here.")
		assert.Greater(t, bodyIdx, moreIdx, "body should come after <!--more-->")
	})
}

func TestSanitizeTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple title", "Hello World", "hello-world"},
		{"title with dash surrounded by spaces", "Stop Designing UIs for AI - Let the LLM Decide What You See", "stop-designing-uis-for-ai-let-the-llm-decide-what-you-see"},
		{"title with multiple dashes", "Part 1 -- The Beginning", "part-1-the-beginning"},
		{"title with special characters", "What is Go? A Test & Post!", "what-is-go-a-test-post"},
		{"title with parentheses and colons", "Kubernetes (K8s): The Basics", "kubernetes-k8s-the-basics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SanitizeTitle(tt.input))
		})
	}
}

// --- PR workflow tests ---

func TestPostViaPR_Success(t *testing.T) {
	tempDir := t.TempDir()

	// Create manuscript file
	manuscriptDir := filepath.Join(tempDir, "manuscript", "devops")
	require.NoError(t, os.MkdirAll(manuscriptDir, 0755))
	gistFile := filepath.Join(manuscriptDir, "my-post.md")
	require.NoError(t, os.WriteFile(gistFile, []byte("# My Post\nContent here."), 0644))

	originalSettings := configuration.GlobalSettings
	defer func() { configuration.GlobalSettings = originalSettings }()
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{Path: tempDir},
	}

	mock := newMockExecutor()
	// The PR workflow runs: clone, config email, config name, checkout -b, add, commit, push = 7 git calls
	// All succeed by default (no pushed results needed except we need enough for 7 calls)

	httpClient := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", req.Method)
		assert.Contains(t, req.URL.String(), "/repos/user/hugo-repo/pulls")
		assert.Equal(t, "Bearer ghp_test", req.Header.Get("Authorization"))

		var body map[string]string
		json.NewDecoder(req.Body).Decode(&body)
		assert.Equal(t, "hugo-post/my-test-title", body["head"])
		assert.Equal(t, "main", body["base"])

		respBody := `{"html_url": "https://github.com/user/hugo-repo/pull/42"}`
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(bytes.NewBufferString(respBody)),
		}, nil
	})

	hugo := NewHugoWithDeps(configuration.SettingsHugo{
		RepoURL: "https://github.com/user/hugo-repo.git",
		Branch:  "main",
		Token:   "ghp_test",
	}, mock, httpClient)

	video := testVideo(gistFile, "My Test Title", "2024-01-15T10:00", "videoABC")
	prURL, err := hugo.Post(video, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/user/hugo-repo/pull/42", prURL)

	// Verify git commands
	require.GreaterOrEqual(t, len(mock.calls), 7)
	assert.Equal(t, "clone", mock.calls[0].Args[0])
	assert.Equal(t, "config", mock.calls[1].Args[0])
	assert.Equal(t, "config", mock.calls[2].Args[0])
	assert.Equal(t, "checkout", mock.calls[3].Args[0])
	assert.Contains(t, mock.calls[3].Args, "hugo-post/my-test-title")
	assert.Equal(t, "add", mock.calls[4].Args[0])
	assert.Equal(t, "commit", mock.calls[5].Args[0])
	assert.Contains(t, mock.calls[5].Args, "Add post: My Test Title")
	assert.Equal(t, "push", mock.calls[6].Args[0])
}

func TestPostViaPR_CloneFailure(t *testing.T) {
	tempDir := t.TempDir()
	manuscriptDir := filepath.Join(tempDir, "manuscript", "devops")
	require.NoError(t, os.MkdirAll(manuscriptDir, 0755))
	gistFile := filepath.Join(manuscriptDir, "post.md")
	require.NoError(t, os.WriteFile(gistFile, []byte("content"), 0644))

	originalSettings := configuration.GlobalSettings
	defer func() { configuration.GlobalSettings = originalSettings }()
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{Path: tempDir},
	}

	mock := newMockExecutor()
	mock.pushResult([]byte("fatal: repo not found"), fmt.Errorf("exit 128"))

	hugo := NewHugoWithDeps(configuration.SettingsHugo{
		RepoURL: "https://github.com/user/repo.git",
		Token:   "tok",
	}, mock, nil)

	video := testVideo(gistFile, "Title", "2024-01-15T10:00", "vid")
	_, err := hugo.Post(video, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")
	// Token must be sanitized
	assert.NotContains(t, err.Error(), "tok")
}

func TestPostViaPR_PushFailure(t *testing.T) {
	tempDir := t.TempDir()
	manuscriptDir := filepath.Join(tempDir, "manuscript", "devops")
	require.NoError(t, os.MkdirAll(manuscriptDir, 0755))
	gistFile := filepath.Join(manuscriptDir, "post.md")
	require.NoError(t, os.WriteFile(gistFile, []byte("content"), 0644))

	originalSettings := configuration.GlobalSettings
	defer func() { configuration.GlobalSettings = originalSettings }()
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{Path: tempDir},
	}

	mock := newMockExecutor()
	// clone, config email, config name, checkout, add, commit succeed (6 calls)
	for i := 0; i < 6; i++ {
		mock.pushResult(nil, nil)
	}
	// push fails
	mock.pushResult([]byte("error: push rejected"), fmt.Errorf("exit 1"))

	hugo := NewHugoWithDeps(configuration.SettingsHugo{
		RepoURL: "https://github.com/user/repo.git",
		Token:   "secret",
	}, mock, nil)

	video := testVideo(gistFile, "Title", "2024-01-15T10:00", "vid")
	_, err := hugo.Post(video, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "git push failed")
	assert.NotContains(t, err.Error(), "secret")
}

func TestPostViaPR_PRCreationFailure(t *testing.T) {
	tempDir := t.TempDir()
	manuscriptDir := filepath.Join(tempDir, "manuscript", "devops")
	require.NoError(t, os.MkdirAll(manuscriptDir, 0755))
	gistFile := filepath.Join(manuscriptDir, "post.md")
	require.NoError(t, os.WriteFile(gistFile, []byte("content"), 0644))

	originalSettings := configuration.GlobalSettings
	defer func() { configuration.GlobalSettings = originalSettings }()
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{Path: tempDir},
	}

	mock := newMockExecutor()

	httpClient := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Validation Failed"}`)),
		}, nil
	})

	hugo := NewHugoWithDeps(configuration.SettingsHugo{
		RepoURL: "https://github.com/user/repo.git",
		Token:   "tok",
	}, mock, httpClient)

	video := testVideo(gistFile, "Title", "2024-01-15T10:00", "vid")
	_, err := hugo.Post(video, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub API returned 422")
}

func TestPostViaPR_SkipsForNA(t *testing.T) {
	hugo := NewHugo(configuration.SettingsHugo{
		RepoURL: "https://github.com/user/repo.git",
		Token:   "tok",
	})

	video := testVideo("N/A", "Title", "2024-01-15T10:00", "vid")
	result, err := hugo.Post(video, nil)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestRepoOwnerAndName(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{"https with .git", "https://github.com/vfarcic/devopstoolkit-live.git", "vfarcic", "devopstoolkit-live", false},
		{"https without .git", "https://github.com/vfarcic/devopstoolkit-live", "vfarcic", "devopstoolkit-live", false},
		{"too short", "repo", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := repoOwnerAndName(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
				assert.Equal(t, tt.wantRepo, repo)
			}
		})
	}
}

func TestPostLocalMode_UsesNewHugo(t *testing.T) {
	// Verify that NewHugo in local mode (no repoURL) behaves like the old &Hugo{}
	tempDir := t.TempDir()
	manuscriptDir := filepath.Join(tempDir, "manuscript", "cat")
	require.NoError(t, os.MkdirAll(manuscriptDir, 0755))
	gistFile := filepath.Join(manuscriptDir, "post.md")
	require.NoError(t, os.WriteFile(gistFile, []byte("# Hello"), 0644))

	// Create homepage
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "content"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "content", "_index.md"),
		[]byte("+++\ntitle = \"Test\"\n+++\n\n# Latest Posts\n\n---\n\n"), 0644))

	originalSettings := configuration.GlobalSettings
	defer func() { configuration.GlobalSettings = originalSettings }()
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{Path: tempDir},
	}

	hugo := NewHugo(configuration.SettingsHugo{Path: tempDir})
	video := testVideo(gistFile, "Local Post", "2024-06-01T08:00", "vid123")
	result, err := hugo.Post(video, nil)
	require.NoError(t, err)

	expectedPath := filepath.Join(tempDir, "content", "cat", "local-post", "_index.md")
	assert.Equal(t, expectedPath, result)
	assert.FileExists(t, result)
}

// Ensure the DefaultExecutor satisfies the interface used by Hugo.
func TestDefaultExecutorSatisfiesInterface(t *testing.T) {
	var _ gitpkg.CommandExecutor = &gitpkg.DefaultExecutor{}
}

func TestGetCategoryFromFilePath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"relative path", "manuscript/ai/knowledge-to-ai.md", "ai"},
		{"absolute path", "/data/tmp/manuscript/ai/knowledge-to-ai.md", "ai"},
		{"absolute path devops", "/data/tmp/manuscript/devops/some-video.md", "devops"},
		{"nested manuscript", "/some/deep/path/manuscript/k8s/video.md", "k8s"},
		{"fallback no manuscript", "some/other/path/video.md", "path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCategoryFromFilePath(tt.filePath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemoveTODOFromIntro(t *testing.T) {
	manuscript := `## Intro

This is the intro.

TODO: remove this line

More intro text.

## Section

Body content.
TODO: body todo
`
	intro, body := ExtractIntro(manuscript)
	intro = RemoveTODOAndFIXMELines(intro)
	body = RemoveTODOAndFIXMELines(body)

	assert.NotContains(t, intro, "TODO:")
	assert.NotContains(t, body, "TODO:")
	assert.Contains(t, intro, "This is the intro.")
	assert.Contains(t, body, "Body content.")
}
