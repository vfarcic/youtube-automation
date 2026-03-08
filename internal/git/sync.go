package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	repoURL  string
	branch   string
	dataDir  string
	token    string
	mu       sync.Mutex
	executor CommandExecutor
}

// NewSyncManager creates a new SyncManager with the default command executor
func NewSyncManager(repoURL, branch, dataDir, token string) *SyncManager {
	return &SyncManager{
		repoURL:  repoURL,
		branch:   branch,
		dataDir:  dataDir,
		token:    token,
		executor: &DefaultExecutor{},
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

	// Stage all changes
	if output, err := s.executor.Run(s.dataDir, "git", "add", "-A"); err != nil {
		return fmt.Errorf("git add failed: %s: %w", SanitizeOutput(output, s.token), err)
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

// authenticatedURL injects the token into the HTTPS URL
func (s *SyncManager) authenticatedURL() string {
	return AuthenticatedURL(s.repoURL, s.token)
}

