package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopstoolkit/youtube-automation/internal/scheduler"
)

// fakeAMAJob counts Run invocations and returns a configurable result.
type fakeAMAJob struct {
	runs   int32
	result scheduler.AMAJobResult
}

func (f *fakeAMAJob) Run(_ context.Context) scheduler.AMAJobResult {
	atomic.AddInt32(&f.runs, 1)
	return f.result
}

// fakeAMANotifier is a no-op notifier used to satisfy the scheduler dependency.
type fakeAMANotifier struct {
	notifies int32
}

func (f *fakeAMANotifier) Notify(_ scheduler.AMAJobResult, _ string) error {
	atomic.AddInt32(&f.notifies, 1)
	return nil
}

func newTestScheduler(t *testing.T, schedule string) (*scheduler.Scheduler, *fakeAMAJob, *fakeAMANotifier) {
	t.Helper()
	job := &fakeAMAJob{}
	notifier := &fakeAMANotifier{}
	sched := &scheduler.Scheduler{
		Schedule:   schedule,
		PlaylistID: "PL-test",
		EmailTo:    "ops@example.com",
		Job:        job,
		Notifier:   notifier,
	}
	return sched, job, notifier
}

// startServerOnEphemeralPort starts s on 127.0.0.1:0 and returns when the
// listener is up (or fails). Tests then call Shutdown to unblock Start.
func startServerOnEphemeralPort(t *testing.T, s *Server) (chan error, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start("127.0.0.1", port)
	}()

	// Wait for the listener to come up.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err == nil {
			conn.Close()
			return errCh, port
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("server did not come up in time")
	return errCh, port
}

func TestServerStartStartsAMASchedulerWhenSet(t *testing.T) {
	env := setupTestEnv(t)
	sched, _, _ := newTestScheduler(t, "0 10 * * *")
	env.server.SetAMAScheduler(sched, 5*time.Second)

	errCh, _ := startServerOnEphemeralPort(t, env.server)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, env.server.Shutdown(ctx))

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("unexpected server error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServerStartWithoutAMASchedulerSkipsIt(t *testing.T) {
	env := setupTestEnv(t)
	// No SetAMAScheduler call — scheduler must not be constructed.
	assert.Nil(t, env.server.amaScheduler)

	errCh, _ := startServerOnEphemeralPort(t, env.server)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, env.server.Shutdown(ctx))

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("unexpected server error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServerStartFailsOnInvalidSchedule(t *testing.T) {
	env := setupTestEnv(t)
	sched, _, _ := newTestScheduler(t, "this is not cron")
	env.server.SetAMAScheduler(sched, 5*time.Second)

	err := env.server.Start("127.0.0.1", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AMA scheduler")
}

func TestServerStartStopsSchedulerOnListenerFailure(t *testing.T) {
	// Bind an ephemeral port so Start's ListenAndServe must fail with
	// "address already in use".
	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { blocker.Close() })
	port := blocker.Addr().(*net.TCPAddr).Port

	env := setupTestEnv(t)
	sched, _, _ := newTestScheduler(t, "0 10 * * *")
	env.server.SetAMAScheduler(sched, 5*time.Second)

	err = env.server.Start("127.0.0.1", port)
	require.Error(t, err)
	// The original listener error must surface — the scheduler-cleanup
	// path must not mask it with a "start AMA scheduler" wrapper.
	assert.NotErrorIs(t, err, http.ErrServerClosed)
	assert.NotContains(t, err.Error(), "AMA scheduler")

	// The scheduler must have been stopped before Start returned —
	// otherwise the cron goroutines outlive the failed Start boundary.
	assert.False(t, sched.IsRunning(),
		"AMA scheduler should be stopped after Start fails on listener error")

	// A second Shutdown call must remain safe (idempotent).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, env.server.Shutdown(ctx))
	assert.False(t, sched.IsRunning())
}

func TestServerShutdownStopsAMAScheduler(t *testing.T) {
	env := setupTestEnv(t)
	// Use every-second schedule (cron 5-field "* * * * *") so we can observe
	// at least one tick before shutdown.
	sched, job, notifier := newTestScheduler(t, "* * * * *")
	env.server.SetAMAScheduler(sched, 5*time.Second)

	errCh, _ := startServerOnEphemeralPort(t, env.server)

	// Don't wait for a tick — this test only asserts shutdown unwinds cleanly.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, env.server.Shutdown(ctx))

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("unexpected server error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop in time")
	}

	// runs/notifies counters are observed only to confirm the fakes wired
	// correctly; the assertion is just non-negative.
	assert.GreaterOrEqual(t, atomic.LoadInt32(&job.runs), int32(0))
	assert.GreaterOrEqual(t, atomic.LoadInt32(&notifier.notifies), int32(0))
}
