package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// CommandExecutor abstracts command execution for testability
type CommandExecutor interface {
	Run(dir string, name string, args ...string) ([]byte, error)
}

// DefaultExecutor runs commands via os/exec
type DefaultExecutor struct{}

// Run executes a command in the given directory
func (e *DefaultExecutor) Run(dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	return cmd.CombinedOutput()
}

// SyncManager handles git clone, pull, commit, and push operations
type SyncManager struct {
	repoURL      string
	branch       string
	dataDir      string
	token        string
	mu           sync.Mutex
	executor     CommandExecutor
	now          func() time.Time
	lastAttempt  time.Time
}

// NewSyncManager creates a new SyncManager with the default command executor
func NewSyncManager(repoURL, branch, dataDir, token string) *SyncManager {
	return &SyncManager{
		repoURL:  repoURL,
		branch:   branch,
		dataDir:  dataDir,
		token:    token,
		executor: &DefaultExecutor{},
		now:      time.Now,
	}
}

// NewSyncManagerWithExecutor creates a new SyncManager with a custom command executor (for testing)
func NewSyncManagerWithExecutor(repoURL, branch, dataDir, token string, executor CommandExecutor) *SyncManager {
	return &SyncManager{
		repoURL:  repoURL,
		branch:   branch,
		dataDir:  dataDir,
		token:    token,
		executor: executor,
		now:      time.Now,
	}
}

// InitialSync clones the repo if no .git directory exists, otherwise pulls with rebase
func (s *SyncManager) InitialSync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	gitDir := s.dataDir + "/.git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Clone
		authURL := s.authenticatedURL()
		output, cloneErr := s.executor.Run(".", "git", "clone", "--branch", s.branch, authURL, s.dataDir)
		if cloneErr != nil {
			return fmt.Errorf("git clone failed: %s: %w", SanitizeOutput(output, s.token), cloneErr)
		}
		return nil
	}

	// Pull with rebase
	output, err := s.executor.Run(s.dataDir, "git", "pull", "--rebase", "origin", s.branch)
	if err != nil {
		return fmt.Errorf("git pull failed: %s: %w", SanitizeOutput(output, s.token), err)
	}
	return nil
}

// CommitAndPush stages all changes, commits with the given message, and pushes
func (s *SyncManager) CommitAndPush(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stage only data files (YAML + Markdown) to avoid accidentally committing
	// large temporary files (videos, images) that may exist in the working tree.
	for _, pattern := range []string{"*.yaml", "*.yml", "*.md"} {
		if output, err := s.executor.Run(s.dataDir, "git", "add", "--all", "--", pattern); err != nil {
			return fmt.Errorf("git add %s failed: %s: %w", pattern, SanitizeOutput(output, s.token), err)
		}
	}

	// Check if there are changes to commit
	statusOutput, err := s.executor.Run(s.dataDir, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("git status failed: %s: %w", SanitizeOutput(statusOutput, s.token), err)
	}
	if strings.TrimSpace(string(statusOutput)) == "" {
		return nil // Nothing to commit
	}

	// Commit
	if output, err := s.executor.Run(s.dataDir, "git", "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit failed: %s: %w", SanitizeOutput(output, s.token), err)
	}

	// Pull with rebase before push
	if output, err := s.executor.Run(s.dataDir, "git", "pull", "--rebase", "origin", s.branch); err != nil {
		return fmt.Errorf("git pull --rebase failed: %s: %w", SanitizeOutput(output, s.token), err)
	}

	// Push
	authURL := s.authenticatedURL()
	if output, err := s.executor.Run(s.dataDir, "git", "push", authURL, s.branch); err != nil {
		return fmt.Errorf("git push failed: %s: %w", SanitizeOutput(output, s.token), err)
	}

	return nil
}

// PullIfStale runs `git pull --rebase` if at least maxAge has elapsed since the
// last attempt. It is safe for concurrent reads and never races with
// CommitAndPush (they share the same mutex).
//
// The pull is skipped (without error) when:
//   - The previous attempt was less than maxAge ago.
//   - The working tree has uncommitted changes.
//   - There are local commits that have not been pushed to origin yet.
//
// Skips return nil because they are expected steady-state conditions, not
// failures. Actual git failures (network, conflict, etc.) are returned so
// callers can surface them to the user.
func (s *SyncManager) PullIfStale(maxAge time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lastAttempt.IsZero() && s.now().Sub(s.lastAttempt) < maxAge {
		return nil
	}

	// Skip if working tree is dirty — pull --rebase could fail or surprise the user.
	statusOutput, err := s.executor.Run(s.dataDir, "git", "status", "--porcelain")
	if err != nil {
		s.lastAttempt = s.now()
		return fmt.Errorf("git status failed: %s: %w", SanitizeOutput(statusOutput, s.token), err)
	}
	if strings.TrimSpace(string(statusOutput)) != "" {
		s.lastAttempt = s.now()
		return nil
	}

	// Skip if local has unpushed commits — let CommitAndPush handle it.
	aheadOutput, err := s.executor.Run(s.dataDir, "git", "rev-list", "--count", "@{u}..HEAD")
	if err != nil {
		s.lastAttempt = s.now()
		return fmt.Errorf("git rev-list failed: %s: %w", SanitizeOutput(aheadOutput, s.token), err)
	}
	if strings.TrimSpace(string(aheadOutput)) != "0" {
		s.lastAttempt = s.now()
		return nil
	}

	// Always update lastAttempt so a transient failure does not get retried on
	// every request within the throttle window.
	s.lastAttempt = s.now()
	output, err := s.executor.Run(s.dataDir, "git", "pull", "--rebase", "origin", s.branch)
	if err != nil {
		return fmt.Errorf("git pull failed: %s: %w", SanitizeOutput(output, s.token), err)
	}
	return nil
}

// authenticatedURL injects the token into the HTTPS URL
func (s *SyncManager) authenticatedURL() string {
	return AuthenticatedURL(s.repoURL, s.token)
}

