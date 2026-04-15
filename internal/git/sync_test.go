package git

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecutor records commands and returns preconfigured results
type mockExecutor struct {
	calls   []mockCall
	results map[string]mockResult
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
	return &mockExecutor{
		results: make(map[string]mockResult),
	}
}

func (m *mockExecutor) setResult(key string, output []byte, err error) {
	m.results[key] = mockResult{Output: output, Err: err}
}

func (m *mockExecutor) Run(dir string, name string, args ...string) ([]byte, error) {
	m.calls = append(m.calls, mockCall{Dir: dir, Name: name, Args: args})

	// Build key from command args
	key := name
	if len(args) > 0 {
		key = name + " " + args[0]
	}

	if result, ok := m.results[key]; ok {
		return result.Output, result.Err
	}

	return []byte{}, nil
}

func (m *mockExecutor) findCall(namePrefix string) *mockCall {
	for i := range m.calls {
		key := m.calls[i].Name
		if len(m.calls[i].Args) > 0 {
			key += " " + m.calls[i].Args[0]
		}
		if key == namePrefix {
			return &m.calls[i]
		}
	}
	return nil
}

func TestInitialSync_ClonesWhenNoGitDir(t *testing.T) {
	dataDir := t.TempDir()
	// Remove the dir so clone creates it
	os.RemoveAll(dataDir)

	mock := newMockExecutor()
	sm := NewSyncManagerWithExecutor(
		"https://github.com/user/repo.git", "main", dataDir, "mytoken", mock,
	)

	err := sm.InitialSync()
	require.NoError(t, err)

	// Should have called git clone
	cloneCall := mock.findCall("git clone")
	require.NotNil(t, cloneCall, "should have called git clone")
	assert.Equal(t, ".", cloneCall.Dir)
	assert.Contains(t, cloneCall.Args, "--branch")
	assert.Contains(t, cloneCall.Args, "main")
	assert.Contains(t, cloneCall.Args, dataDir)
}

func TestInitialSync_PullsWhenGitDirExists(t *testing.T) {
	dataDir := t.TempDir()
	// Create .git directory to simulate existing clone
	require.NoError(t, os.MkdirAll(dataDir+"/.git", 0755))

	mock := newMockExecutor()
	sm := NewSyncManagerWithExecutor(
		"https://github.com/user/repo.git", "main", dataDir, "mytoken", mock,
	)

	err := sm.InitialSync()
	require.NoError(t, err)

	// Should have called git pull, not clone
	pullCall := mock.findCall("git pull")
	require.NotNil(t, pullCall, "should have called git pull")
	assert.Equal(t, dataDir, pullCall.Dir)

	cloneCall := mock.findCall("git clone")
	assert.Nil(t, cloneCall, "should NOT have called git clone")
}

func TestInitialSync_CloneFailure(t *testing.T) {
	dataDir := t.TempDir()
	os.RemoveAll(dataDir)

	mock := newMockExecutor()
	mock.setResult("git clone", []byte("fatal: repo not found"), fmt.Errorf("exit status 128"))

	sm := NewSyncManagerWithExecutor(
		"https://github.com/user/repo.git", "main", dataDir, "mytoken", mock,
	)

	err := sm.InitialSync()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")
	// Token should be sanitized in error output
	assert.NotContains(t, err.Error(), "mytoken")
}

func TestInitialSync_PullFailure(t *testing.T) {
	dataDir := t.TempDir()
	require.NoError(t, os.MkdirAll(dataDir+"/.git", 0755))

	mock := newMockExecutor()
	mock.setResult("git pull", []byte("error: cannot pull with rebase"), fmt.Errorf("exit status 1"))

	sm := NewSyncManagerWithExecutor(
		"https://github.com/user/repo.git", "main", dataDir, "", mock,
	)

	err := sm.InitialSync()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git pull failed")
}

func TestCommitAndPush_FullSequence(t *testing.T) {
	dataDir := t.TempDir()

	mock := newMockExecutor()
	// git status returns changes
	mock.setResult("git status", []byte("M index.yaml\n"), nil)

	sm := NewSyncManagerWithExecutor(
		"https://github.com/user/repo.git", "main", dataDir, "tok123", mock,
	)

	err := sm.CommitAndPush("test commit")
	require.NoError(t, err)

	// Verify command sequence: add (x3 for *.yaml, *.yml, *.md), status, commit, pull, push
	require.GreaterOrEqual(t, len(mock.calls), 7)
	assert.Equal(t, "add", mock.calls[0].Args[0])
	assert.Equal(t, "add", mock.calls[1].Args[0])
	assert.Equal(t, "add", mock.calls[2].Args[0])
	assert.Equal(t, "status", mock.calls[3].Args[0])
	assert.Equal(t, "commit", mock.calls[4].Args[0])
	assert.Equal(t, "pull", mock.calls[5].Args[0])
	assert.Equal(t, "push", mock.calls[6].Args[0])

	// Verify add patterns
	assert.Contains(t, mock.calls[0].Args, "*.yaml")
	assert.Contains(t, mock.calls[1].Args, "*.yml")
	assert.Contains(t, mock.calls[2].Args, "*.md")

	// Verify commit message
	commitCall := mock.calls[4]
	assert.Contains(t, commitCall.Args, "-m")
	assert.Contains(t, commitCall.Args, "test commit")
}

func TestCommitAndPush_SkipsWhenClean(t *testing.T) {
	dataDir := t.TempDir()

	mock := newMockExecutor()
	// git status returns empty (no changes)
	mock.setResult("git status", []byte(""), nil)

	sm := NewSyncManagerWithExecutor(
		"https://github.com/user/repo.git", "main", dataDir, "", mock,
	)

	err := sm.CommitAndPush("test commit")
	require.NoError(t, err)

	// Should only have add (x3) and status calls, no commit/push
	assert.Equal(t, 4, len(mock.calls))
	assert.Equal(t, "add", mock.calls[0].Args[0])
	assert.Equal(t, "add", mock.calls[1].Args[0])
	assert.Equal(t, "add", mock.calls[2].Args[0])
	assert.Equal(t, "status", mock.calls[3].Args[0])
}

func TestCommitAndPush_PushFailure(t *testing.T) {
	dataDir := t.TempDir()

	mock := newMockExecutor()
	mock.setResult("git status", []byte("M file.yaml\n"), nil)
	mock.setResult("git push", []byte("error: push rejected"), fmt.Errorf("exit status 1"))

	sm := NewSyncManagerWithExecutor(
		"https://github.com/user/repo.git", "main", dataDir, "", mock,
	)

	err := sm.CommitAndPush("test commit")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git push failed")
}

func TestSyncManagerAuthenticatedURL(t *testing.T) {
	// Verifies that SyncManager.authenticatedURL() delegates to the shared AuthenticatedURL
	sm := &SyncManager{repoURL: "https://github.com/user/repo.git", token: "ghp_abc123"}
	result := sm.authenticatedURL()
	assert.Equal(t, "https://x-access-token:ghp_abc123@github.com/user/repo.git", result)
}

func TestSanitizeOutput(t *testing.T) {
	assert.Equal(t, "url https://***@github.com", SanitizeOutput([]byte("url https://secret@github.com"), "secret"))
	assert.Equal(t, "no token here", SanitizeOutput([]byte("no token here"), ""))
	assert.Equal(t, "no token here", SanitizeOutput([]byte("no token here"), "missing"))
}

func TestNewSyncManager(t *testing.T) {
	sm := NewSyncManager("https://github.com/user/repo.git", "main", "./tmp", "tok")
	assert.NotNil(t, sm)
	assert.Equal(t, "https://github.com/user/repo.git", sm.repoURL)
	assert.Equal(t, "main", sm.branch)
	assert.Equal(t, "./tmp", sm.dataDir)
	assert.Equal(t, "tok", sm.token)
	assert.IsType(t, &DefaultExecutor{}, sm.executor)
}

// newPullTestSyncManager builds a SyncManager with a controllable clock for
// PullIfStale tests.
func newPullTestSyncManager(t *testing.T, mock *mockExecutor, clock *time.Time) *SyncManager {
	t.Helper()
	sm := NewSyncManagerWithExecutor(
		"https://github.com/user/repo.git", "main", t.TempDir(), "tok123", mock,
	)
	sm.now = func() time.Time { return *clock }
	return sm
}

func TestPullIfStale_PullsWhenCleanAndNotAhead(t *testing.T) {
	mock := newMockExecutor()
	mock.setResult("git status", []byte(""), nil)
	mock.setResult("git rev-list", []byte("0\n"), nil)

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	sm := newPullTestSyncManager(t, mock, &now)

	err := sm.PullIfStale(10 * time.Second)
	require.NoError(t, err)

	// Verify command sequence: status, rev-list, pull
	require.Len(t, mock.calls, 3)
	assert.Equal(t, "status", mock.calls[0].Args[0])
	assert.Equal(t, "rev-list", mock.calls[1].Args[0])
	assert.Equal(t, "pull", mock.calls[2].Args[0])
	assert.Contains(t, mock.calls[2].Args, "--rebase")
	assert.Contains(t, mock.calls[2].Args, "main")
}

func TestPullIfStale_SkipsWhenWithinThrottleWindow(t *testing.T) {
	mock := newMockExecutor()
	mock.setResult("git status", []byte(""), nil)
	mock.setResult("git rev-list", []byte("0\n"), nil)

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	sm := newPullTestSyncManager(t, mock, &now)

	// First call: pulls and updates lastAttempt
	require.NoError(t, sm.PullIfStale(10*time.Second))
	require.Len(t, mock.calls, 3)

	// Advance clock by less than the throttle window
	now = now.Add(5 * time.Second)
	require.NoError(t, sm.PullIfStale(10*time.Second))

	// No new git commands should have run
	assert.Len(t, mock.calls, 3, "second call within throttle window should be a no-op")
}

func TestPullIfStale_PullsAgainAfterThrottleWindow(t *testing.T) {
	mock := newMockExecutor()
	mock.setResult("git status", []byte(""), nil)
	mock.setResult("git rev-list", []byte("0\n"), nil)

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	sm := newPullTestSyncManager(t, mock, &now)

	require.NoError(t, sm.PullIfStale(10*time.Second))
	require.Len(t, mock.calls, 3)

	// Advance past the throttle window
	now = now.Add(11 * time.Second)
	require.NoError(t, sm.PullIfStale(10*time.Second))

	// Should have run status, rev-list, pull a second time
	assert.Len(t, mock.calls, 6)
}

func TestPullIfStale_SkipsWhenWorkingTreeDirty(t *testing.T) {
	mock := newMockExecutor()
	mock.setResult("git status", []byte("M index.yaml\n"), nil)

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	sm := newPullTestSyncManager(t, mock, &now)

	err := sm.PullIfStale(10 * time.Second)
	require.NoError(t, err, "dirty tree should be a silent skip, not an error")

	// Should only call status — no rev-list, no pull
	require.Len(t, mock.calls, 1)
	assert.Equal(t, "status", mock.calls[0].Args[0])

	// Throttle should still be armed so we don't hammer git status on every read
	now = now.Add(5 * time.Second)
	require.NoError(t, sm.PullIfStale(10*time.Second))
	assert.Len(t, mock.calls, 1, "skip should still update lastAttempt to throttle subsequent reads")
}

func TestPullIfStale_SkipsWhenAheadOfOrigin(t *testing.T) {
	mock := newMockExecutor()
	mock.setResult("git status", []byte(""), nil)
	mock.setResult("git rev-list", []byte("2\n"), nil)

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	sm := newPullTestSyncManager(t, mock, &now)

	err := sm.PullIfStale(10 * time.Second)
	require.NoError(t, err, "unpushed commits should be a silent skip")

	// status and rev-list ran; pull did not
	require.Len(t, mock.calls, 2)
	assert.Equal(t, "status", mock.calls[0].Args[0])
	assert.Equal(t, "rev-list", mock.calls[1].Args[0])
}

func TestPullIfStale_ReturnsErrorOnPullFailure(t *testing.T) {
	mock := newMockExecutor()
	mock.setResult("git status", []byte(""), nil)
	mock.setResult("git rev-list", []byte("0\n"), nil)
	mock.setResult("git pull", []byte("fatal: unable to access"), fmt.Errorf("exit status 128"))

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	sm := newPullTestSyncManager(t, mock, &now)

	err := sm.PullIfStale(10 * time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "git pull failed")
	// Token must not leak into surfaced error
	assert.NotContains(t, err.Error(), "tok123")

	// Even on failure, throttle must engage to avoid hammering on every read
	now = now.Add(5 * time.Second)
	require.NoError(t, sm.PullIfStale(10*time.Second))
	assert.Len(t, mock.calls, 3, "failed pull must still arm the throttle")
}

func TestPullIfStale_ReturnsErrorOnStatusFailure(t *testing.T) {
	mock := newMockExecutor()
	mock.setResult("git status", []byte("fatal: not a git repository"), fmt.Errorf("exit status 128"))

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	sm := newPullTestSyncManager(t, mock, &now)

	err := sm.PullIfStale(10 * time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "git status failed")
}
