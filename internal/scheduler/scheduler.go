package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"

	"github.com/robfig/cron/v3"
)

// jobRunner is the minimal AMA-job surface the Scheduler depends on. *AMAJob
// satisfies it directly in production; tests inject a fake.
type jobRunner interface {
	Run(ctx context.Context) AMAJobResult
}

// resultNotifier is the minimal notifier surface the Scheduler depends on.
// *AMANotifier satisfies it directly in production; tests inject a fake.
type resultNotifier interface {
	Notify(result AMAJobResult, emailTo string) error
}

// Scheduler triggers the AMA job on a cron schedule and forwards each result
// to the notifier. Configuration values are injected at construction time;
// wiring from settings.yaml is the responsibility of Milestone 5.
type Scheduler struct {
	Schedule   string
	PlaylistID string
	EmailTo    string
	Job        jobRunner
	Notifier   resultNotifier

	mu      sync.Mutex
	cron    *cron.Cron
	cancel  context.CancelFunc
	runCtx  context.Context
	started bool
}

// Start parses the schedule, registers the AMA tick, and starts the cron loop
// in its own goroutine. Returns an error if the schedule is invalid; nil
// otherwise. Calling Start on an already-started Scheduler is a no-op.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return nil
	}

	c := cron.New()
	chain := cron.NewChain(cron.SkipIfStillRunning(slogCronLogger{}))
	wrapped := chain.Then(cron.FuncJob(s.tick))
	if _, err := c.AddJob(s.Schedule, wrapped); err != nil {
		return fmt.Errorf("scheduler: invalid schedule %q: %w", s.Schedule, err)
	}

	s.runCtx, s.cancel = context.WithCancel(ctx)
	s.cron = c
	s.started = true
	c.Start()

	slog.Info("AMA scheduler started", "schedule", s.Schedule, "playlistID", s.PlaylistID)
	return nil
}

// IsRunning reports whether the scheduler has been Started and not yet
// Stopped. It is safe to call concurrently with Start/Stop.
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

// Stop halts the cron loop and waits for in-flight runs to finish or for the
// supplied context to expire. It is idempotent: calling it before Start, or
// twice in a row, is safe and returns nil.
func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = false
	c := s.cron
	cancel := s.cancel
	s.cron = nil
	s.cancel = nil
	s.runCtx = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	stopCtx := c.Stop()
	select {
	case <-stopCtx.Done():
		slog.Info("AMA scheduler stopped")
		return nil
	case <-ctx.Done():
		slog.Warn("AMA scheduler stop deadline exceeded; in-flight job(s) may still be running",
			"err", ctx.Err(),
		)
		return ctx.Err()
	}
}

// tick is the cron callback. It runs the AMA job, then forwards the result to
// the notifier. Notifier errors are logged and swallowed so a transient email
// failure cannot crash the scheduler loop. A panic from Job.Run() or
// Notifier.Notify() is recovered here so a single bad tick cannot crash the
// process — an unrecovered panic in a goroutine takes the whole program down.
func (s *Scheduler) tick() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("AMA scheduler tick panicked",
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()

	s.mu.Lock()
	ctx := s.runCtx
	job := s.Job
	notifier := s.Notifier
	emailTo := s.EmailTo
	s.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	result := job.Run(ctx)
	slog.Info("AMA scheduler tick completed",
		"outcome", string(result.Outcome),
		"videoID", result.VideoID,
	)

	if err := notifier.Notify(result, emailTo); err != nil {
		slog.Error("AMA scheduler: notification failed",
			"err", err,
			"outcome", string(result.Outcome),
			"videoID", result.VideoID,
		)
	}
}

// slogCronLogger adapts robfig/cron's Logger interface to slog so chain
// wrappers (e.g. SkipIfStillRunning) emit through the standard logging path.
type slogCronLogger struct{}

func (slogCronLogger) Info(msg string, keysAndValues ...any) {
	slog.Info("AMA scheduler: "+msg, keysAndValues...)
}

func (slogCronLogger) Error(err error, msg string, keysAndValues ...any) {
	args := make([]any, 0, len(keysAndValues)+2)
	args = append(args, "err", err)
	args = append(args, keysAndValues...)
	slog.Error("AMA scheduler: "+msg, args...)
}
