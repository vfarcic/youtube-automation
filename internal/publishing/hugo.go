package publishing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/gdrive"
	gitpkg "devopstoolkit/youtube-automation/internal/git"
	"devopstoolkit/youtube-automation/internal/storage"
)

// Hugo handles creating Hugo blog posts either locally or via GitHub PR.
type Hugo struct {
	repoURL  string
	branch   string
	token    string
	path     string // local override (CLI mode)
	executor gitpkg.CommandExecutor
	// httpClient is used for GitHub API calls; defaults to http.DefaultClient
	httpClient HTTPClient
}

// HTTPClient abstracts HTTP requests for testability.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHugo creates a Hugo instance from configuration.
// If cfg.RepoURL is set, Post() will use the PR workflow.
// Otherwise it uses local filesystem writes.
func NewHugo(cfg configuration.SettingsHugo) *Hugo {
	branch := cfg.Branch
	if branch == "" {
		branch = "main"
	}
	return &Hugo{
		repoURL:    cfg.RepoURL,
		branch:     branch,
		token:      cfg.Token,
		path:       cfg.Path,
		executor:   &gitpkg.DefaultExecutor{},
		httpClient: http.DefaultClient,
	}
}

// NewHugoWithDeps creates a Hugo instance with injected dependencies (for testing).
func NewHugoWithDeps(cfg configuration.SettingsHugo, executor gitpkg.CommandExecutor, httpClient HTTPClient) *Hugo {
	h := NewHugo(cfg)
	h.executor = executor
	h.httpClient = httpClient
	return h
}

// HugoPostOptions carries optional dependencies for enriching Hugo posts.
// Pass nil from CLI mode to gracefully skip Drive-dependent steps.
type HugoPostOptions struct {
	DriveService  gdrive.DriveService
	DriveFolderID string
}

func (r *Hugo) Post(video *storage.Video, opts *HugoPostOptions) (string, error) {
	if video.Gist == "N/A" {
		return "", nil
	}
	title := video.GetUploadTitle()
	post, err := r.getPost(video.Gist, title, video.Date, video.VideoId)
	if err != nil {
		return "", err
	}
	if r.repoURL != "" {
		return r.postViaPR(video, title, post, opts)
	}
	basePath := r.basePath()
	hugoPath, err := r.hugoFromMarkdown(video.Gist, title, post, basePath)
	if err != nil {
		return "", err
	}
	if err := r.enrichPostDir(context.Background(), video, title, filepath.Dir(hugoPath), basePath, opts); err != nil {
		fmt.Printf("Warning: post enrichment failed: %v\n", err)
	}
	return hugoPath, nil
}

// basePath returns the local Hugo repo path, falling back to the global config.
func (r *Hugo) basePath() string {
	if r.path != "" {
		return r.path
	}
	return configuration.GlobalSettings.Hugo.Path
}

// SanitizeTitle sanitizes a title for use as a directory name in Hugo
func SanitizeTitle(title string) string {
	postDir := title
	postDir = strings.ReplaceAll(postDir, "(", "")
	postDir = strings.ReplaceAll(postDir, ")", "")
	postDir = strings.ReplaceAll(postDir, ":", "")
	postDir = strings.ReplaceAll(postDir, "&", "")
	postDir = strings.ReplaceAll(postDir, "'", "")
	postDir = strings.ReplaceAll(postDir, "!", "")
	postDir = strings.ReplaceAll(postDir, "?", "")
	postDir = strings.ReplaceAll(postDir, ".", "")
	// Convert hyphens and slashes to spaces so Fields() collapses them with surrounding whitespace
	postDir = strings.ReplaceAll(postDir, "-", " ")
	postDir = strings.ReplaceAll(postDir, "/", " ")
	postDir = strings.ToLower(postDir)
	// Fields splits on any whitespace runs, then join with single hyphen
	postDir = strings.Join(strings.Fields(postDir), "-")
	return postDir
}

// GetCategoryFromFilePath extracts the category from a manuscript file path.
// It finds the "manuscript/" segment in the path and returns the directory
// immediately after it (e.g. "ai" from ".../manuscript/ai/foo.md").
// This works for both relative and absolute paths regardless of Hugo.Path.
func GetCategoryFromFilePath(filePath string) string {
	// Normalize separators
	normalized := filepath.ToSlash(filePath)
	// Find "manuscript/" in the path
	const marker = "manuscript/"
	idx := strings.LastIndex(normalized, marker)
	if idx >= 0 {
		after := normalized[idx+len(marker):]
		// Take only the first path segment (the category)
		parts := strings.SplitN(after, "/", 2)
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}
	// Fallback: directory name of the file
	return filepath.Base(filepath.Dir(filePath))
}

// ConstructHugoURL constructs the Hugo URL based on title and category without creating the post
func ConstructHugoURL(title, category string) string {
	sanitizedTitle := SanitizeTitle(title)
	return fmt.Sprintf("https://devopstoolkit.live/%s/%s", category, sanitizedTitle)
}

func (r *Hugo) hugoFromMarkdown(filePath, title, post, basePath string) (string, error) {
	// Convert the manuscript path to a content path
	relPath := GetCategoryFromFilePath(filePath)

	// Use filepath.Join for proper path construction
	categoryDir := filepath.Join(basePath, "content", relPath)

	// Sanitize the title for use as a directory name
	postDir := SanitizeTitle(title)

	// Create the full directory path using filepath.Join
	fullDir := filepath.Join(categoryDir, postDir)
	if err := os.MkdirAll(fullDir, os.FileMode(0755)); err != nil {
		return "", err
	}

	// Create the output file path using filepath.Join
	hugoPath := filepath.Join(fullDir, "_index.md")
	if err := os.WriteFile(hugoPath, []byte(post), 0644); err != nil {
		return "", err
	}
	return hugoPath, nil
}

func (r *Hugo) getPost(filePath, title, date, videoId string) (string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	manuscript := string(contentBytes)

	// Extract intro and clean up the manuscript
	intro, body := ExtractIntro(manuscript)
	intro = RemoveTODOAndFIXMELines(intro)
	body = RemoveTODOAndFIXMELines(body)

	return BuildHugoPost(title, date, videoId, intro, body), nil
}

// enrichPostDir performs post-creation enrichment: copies images, thumbnail,
// and updates the home page. Non-fatal errors are logged as warnings.
func (r *Hugo) enrichPostDir(ctx context.Context, video *storage.Video, title, postDir, basePath string, opts *HugoPostOptions) error {
	// Read manuscript for image references
	contentBytes, err := os.ReadFile(video.Gist)
	if err != nil {
		// Not fatal — manuscript might not be accessible in PR workflow
		return nil
	}

	imageFiles := ParseImageReferences(string(contentBytes))

	var driveService gdrive.DriveService
	var driveFolderID string
	if opts != nil {
		driveService = opts.DriveService
		driveFolderID = opts.DriveFolderID
	}

	// Copy images from Drive
	if err := CopyImagesFromDrive(ctx, driveService, video.Name, driveFolderID, postDir, imageFiles); err != nil {
		fmt.Printf("Warning: failed to copy images: %v\n", err)
	}

	// Copy thumbnail from Drive
	hasThumbnail := true
	if err := CopyThumbnailFromDrive(ctx, driveService, video.ThumbnailVariants, postDir); err != nil {
		fmt.Printf("Warning: failed to copy thumbnail: %v\n", err)
		hasThumbnail = false
	}

	// Update home page
	category := GetCategoryFromFilePath(video.Gist)
	slug := SanitizeTitle(title)
	intro, _ := ExtractIntro(string(contentBytes))
	intro = RemoveTODOAndFIXMELines(intro)

	if err := AddHomepageEntry(basePath, category, slug, title, intro, hasThumbnail); err != nil {
		fmt.Printf("Warning: failed to add homepage entry: %v\n", err)
	}

	if err := TrimHomepageEntries(basePath, 10); err != nil {
		fmt.Printf("Warning: failed to trim homepage entries: %v\n", err)
	}

	return nil
}

// postViaPR clones the Hugo repo, writes the post, pushes a branch, and creates a GitHub PR.
func (r *Hugo) postViaPR(video *storage.Video, title, post string, opts *HugoPostOptions) (string, error) {
	tmpDir, err := os.MkdirTemp("", "hugo-pr-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	authURL := gitpkg.AuthenticatedURL(r.repoURL, r.token)

	// Clone
	if output, err := r.executor.Run(".", "git", "clone", "--depth", "1", "--branch", r.branch, authURL, tmpDir); err != nil {
		return "", fmt.Errorf("git clone failed: %s: %w", gitpkg.SanitizeOutput(output, r.token), err)
	}

	// Configure git user for the commit
	if output, err := r.executor.Run(tmpDir, "git", "config", "user.email", "automation@devopstoolkit.live"); err != nil {
		return "", fmt.Errorf("git config email failed: %s: %w", string(output), err)
	}
	if output, err := r.executor.Run(tmpDir, "git", "config", "user.name", "YouTube Automation"); err != nil {
		return "", fmt.Errorf("git config name failed: %s: %w", string(output), err)
	}

	// Create branch
	branchName := "hugo-post/" + SanitizeTitle(title)
	if output, err := r.executor.Run(tmpDir, "git", "checkout", "-b", branchName); err != nil {
		return "", fmt.Errorf("git checkout -b failed: %s: %w", string(output), err)
	}

	// Write post file into the cloned repo
	hugoPath, err := r.hugoFromMarkdown(video.Gist, title, post, tmpDir)
	if err != nil {
		return "", fmt.Errorf("writing hugo post: %w", err)
	}

	// Enrich the post directory (images, thumbnail, homepage)
	if err := r.enrichPostDir(context.Background(), video, title, filepath.Dir(hugoPath), tmpDir, opts); err != nil {
		fmt.Printf("Warning: post enrichment failed: %v\n", err)
	}

	// Stage + commit
	if output, err := r.executor.Run(tmpDir, "git", "add", "-A"); err != nil {
		return "", fmt.Errorf("git add failed: %s: %w", string(output), err)
	}
	commitMsg := fmt.Sprintf("Add post: %s", title)
	if output, err := r.executor.Run(tmpDir, "git", "commit", "-m", commitMsg); err != nil {
		return "", fmt.Errorf("git commit failed: %s: %w", string(output), err)
	}

	// Push (force so retries can update an existing remote branch from a
	// previous attempt; --force-with-lease can't be used because we shallow
	// clone only the base branch and have no local ref for branchName)
	if output, err := r.executor.Run(tmpDir, "git", "push", "--force", authURL, branchName); err != nil {
		return "", fmt.Errorf("git push failed: %s: %w", gitpkg.SanitizeOutput(output, r.token), err)
	}

	// Create PR via GitHub API
	prURL, err := r.createPR(title, branchName)
	if err != nil {
		return "", fmt.Errorf("creating PR: %w", err)
	}

	return prURL, nil
}

// repoOwnerAndName extracts "owner" and "repo" from a GitHub HTTPS URL.
func repoOwnerAndName(repoURL string) (string, string, error) {
	// Expected format: https://github.com/{owner}/{repo}.git or https://github.com/{owner}/{repo}
	trimmed := strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("cannot parse owner/repo from URL: %s", repoURL)
	}
	repo := parts[len(parts)-1]
	owner := parts[len(parts)-2]
	return owner, repo, nil
}

// createPR creates a GitHub pull request and returns the PR URL.
func (r *Hugo) createPR(title, branchName string) (string, error) {
	owner, repo, err := repoOwnerAndName(r.repoURL)
	if err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	body := map[string]string{
		"title": fmt.Sprintf("Add post: %s", title),
		"head":  branchName,
		"base":  r.branch,
		"body":  fmt.Sprintf("Automated Hugo post for: %s", title),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshalling PR body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating PR request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("PR API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var prResp struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(respBody, &prResp); err != nil {
		return "", fmt.Errorf("parsing PR response: %w", err)
	}

	return prResp.HTMLURL, nil
}
