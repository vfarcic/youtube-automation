package scheduler

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeJobRunner struct {
	mu      sync.Mutex
	calls   int32
	lastCtx context.Context
	result  AMAJobResult
	block   chan struct{} // when non-nil, Run blocks until close or ctx cancel
	started chan struct{} // when non-nil, signaled once on first Run entry
}

func (f *fakeJobRunner) Run(ctx context.Context) AMAJobResult {
	atomic.AddInt32(&f.calls, 1)
	f.mu.Lock()
	f.lastCtx = ctx
	block := f.block
	started := f.started
	f.mu.Unlock()

	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
		}
	}
	return f.result
}

func (f *fakeJobRunner) callCount() int32 {
	return atomic.LoadInt32(&f.calls)
}

type fakeNotifier struct {
	mu      sync.Mutex
	calls   []notifyCall
	err     error
	errOnce bool // if true, only the first call returns err
}

type notifyCall struct {
	result  AMAJobResult
	emailTo string
}

func (f *fakeNotifier) Notify(result AMAJobResult, emailTo string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, notifyCall{result: result, emailTo: emailTo})
	err := f.err
	if f.errOnce {
		f.err = nil
	}
	return err
}

func (f *fakeNotifier) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func (f *fakeNotifier) snapshot() []notifyCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]notifyCall, len(f.calls))
	copy(out, f.calls)
	return out
}

func newScheduler(schedule string, job jobRunner, notifier resultNotifier) *Scheduler {
	return &Scheduler{
		Schedule:   schedule,
		PlaylistID: "PL-test",
		EmailTo:    "ops@example.com",
		Job:        job,
		Notifier:   notifier,
	}
}

// waitFor polls the predicate until it returns true or the timeout expires.
func waitFor(t *testing.T, timeout time.Duration, pred func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pred() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return pred()
}

func TestScheduler_Start_RegistersTrigger(t *testing.T) {
	s := newScheduler("0 10 * * *", &fakeJobRunner{}, &fakeNotifier{})
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	defer s.Stop(context.Background())

	if s.cron == nil {
		t.Fatalf("cron not initialized after Start")
	}
	entries := s.cron.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 cron entry, got %d", len(entries))
	}
	if !s.started {
		t.Fatalf("expected started=true after Start")
	}
}

func TestScheduler_Start_InvalidSchedule(t *testing.T) {
	s := newScheduler("not a cron expression", &fakeJobRunner{}, &fakeNotifier{})
	err := s.Start(context.Background())
	if err == nil {
		t.Fatalf("expected error for invalid schedule, got nil")
	}
	if s.started {
		t.Fatalf("scheduler should not be marked started after invalid schedule")
	}
}

func TestScheduler_Start_TwiceIsNoop(t *testing.T) {
	s := newScheduler("@every 1h", &fakeJobRunner{}, &fakeNotifier{})
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("first Start error: %v", err)
	}
	defer s.Stop(context.Background())
	firstCron := s.cron

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("second Start error: %v", err)
	}
	if s.cron != firstCron {
		t.Fatalf("second Start replaced cron instance — expected no-op")
	}
}

func TestScheduler_Stop_BeforeStart(t *testing.T) {
	s := newScheduler("@every 1h", &fakeJobRunner{}, &fakeNotifier{})
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("Stop before Start returned error: %v", err)
	}
}

func TestScheduler_Stop_Twice(t *testing.T) {
	s := newScheduler("@every 1h", &fakeJobRunner{}, &fakeNotifier{})
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("first Stop error: %v", err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop error: %v", err)
	}
	if s.started {
		t.Fatalf("scheduler still marked started after Stop")
	}
}

func TestScheduler_Stop_AfterStartCleanly(t *testing.T) {
	job := &fakeJobRunner{result: AMAJobResult{Outcome: OutcomeSkipped, VideoID: "v1"}}
	notifier := &fakeNotifier{}
	s := newScheduler("@every 50ms", job, notifier)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Stop(stopCtx); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if s.started {
		t.Fatalf("scheduler still marked started after clean Stop")
	}
}

func TestScheduler_Tick_CallsJobThenNotifier(t *testing.T) {
	wantResult := AMAJobResult{Outcome: OutcomeProcessed, VideoID: "abc", VideoURL: "https://youtu.be/abc"}
	job := &fakeJobRunner{result: wantResult}
	notifier := &fakeNotifier{}
	s := newScheduler("@every 50ms", job, notifier)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer s.Stop(context.Background())

	if !waitFor(t, 1*time.Second, func() bool {
		return notifier.callCount() >= 1
	}) {
		t.Fatalf("notifier was not called within deadline; job calls=%d, notify calls=%d",
			job.callCount(), notifier.callCount())
	}

	if job.callCount() == 0 {
		t.Fatalf("expected job to be called, got 0 calls")
	}

	calls := notifier.snapshot()
	first := calls[0]
	if first.emailTo != "ops@example.com" {
		t.Errorf("notifier emailTo = %q, want %q", first.emailTo, "ops@example.com")
	}
	if first.result.Outcome != wantResult.Outcome ||
		first.result.VideoID != wantResult.VideoID ||
		first.result.VideoURL != wantResult.VideoURL {
		t.Errorf("notifier received result %+v, want %+v", first.result, wantResult)
	}
}

func TestScheduler_Tick_NotifierErrorDoesNotCrashLoop(t *testing.T) {
	job := &fakeJobRunner{result: AMAJobResult{Outcome: OutcomeProcessed, VideoID: "abc"}}
	notifier := &fakeNotifier{err: errors.New("smtp boom")}
	s := newScheduler("@every 50ms", job, notifier)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer s.Stop(context.Background())

	if !waitFor(t, 2*time.Second, func() bool {
		return notifier.callCount() >= 2
	}) {
		t.Fatalf("expected at least 2 notifier calls (loop should continue after error); got %d",
			notifier.callCount())
	}
	if job.callCount() < 2 {
		t.Fatalf("expected at least 2 job calls, got %d", job.callCount())
	}
}

func TestScheduler_Tick_PassesCancellableContext(t *testing.T) {
	block := make(chan struct{})
	started := make(chan struct{}, 1)
	job := &fakeJobRunner{
		result:  AMAJobResult{Outcome: OutcomeProcessed, VideoID: "abc"},
		block:   block,
		started: started,
	}
	notifier := &fakeNotifier{}
	s := newScheduler("@every 50ms", job, notifier)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		_ = s.Stop(context.Background())
		t.Fatalf("job never started running")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	stopErr := s.Stop(stopCtx)
	if stopErr != nil {
		t.Fatalf("Stop returned error after cancel: %v", stopErr)
	}

	job.mu.Lock()
	ctx := job.lastCtx
	job.mu.Unlock()
	if ctx == nil {
		t.Fatalf("job did not record its context")
	}
	if ctx.Err() == nil {
		t.Errorf("expected job's context to be cancelled after Stop, but ctx.Err() == nil")
	}
}

func TestScheduler_Stop_RespectsCtxDeadline(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	started := make(chan struct{}, 1)

	hung := &hangingJob{started: started, release: release}
	notifier := &fakeNotifier{}
	s := newScheduler("@every 30ms", hung, notifier)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		_ = s.Stop(context.Background())
		t.Fatalf("hanging job never started")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := s.Stop(stopCtx)
	if err == nil {
		t.Fatalf("expected ctx deadline error from Stop, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestSlogCronLogger_ErrorDoesNotPanic(t *testing.T) {
	// SkipIfStillRunning only invokes the Info path, but the cron.Logger
	// interface still requires Error. Exercise it directly to confirm the
	// args-prepending logic doesn't panic on either zero or N kv pairs.
	var l slogCronLogger
	l.Error(errors.New("boom"), "no extras")
	l.Error(errors.New("boom"), "with extras", "k1", "v1", "k2", 42)
}

// hangingJob ignores ctx cancellation so we can test that Stop respects its
// own deadline when an in-flight run will not return promptly.
type hangingJob struct {
	started chan struct{}
	release chan struct{} // closed by the test cleanup to let the goroutine exit
}

func (h *hangingJob) Run(ctx context.Context) AMAJobResult {
	select {
	case h.started <- struct{}{}:
	default:
	}
	<-h.release
	return AMAJobResult{Outcome: OutcomeProcessed, VideoID: "hung"}
}

// panicOnceJob panics on its first Run() invocation, then behaves normally so
// later ticks can verify the loop survived the panic.
type panicOnceJob struct {
	calls int32
}

func (p *panicOnceJob) Run(ctx context.Context) AMAJobResult {
	if atomic.AddInt32(&p.calls, 1) == 1 {
		panic("intentional test panic in Job.Run")
	}
	return AMAJobResult{Outcome: OutcomeProcessed, VideoID: "after-panic"}
}

func (p *panicOnceJob) callCount() int32 {
	return atomic.LoadInt32(&p.calls)
}

// panicOnceNotifier panics on its first Notify() call, then succeeds.
type panicOnceNotifier struct {
	mu    sync.Mutex
	calls int
}

func (p *panicOnceNotifier) Notify(result AMAJobResult, emailTo string) error {
	p.mu.Lock()
	p.calls++
	n := p.calls
	p.mu.Unlock()
	if n == 1 {
		panic("intentional test panic in Notifier.Notify")
	}
	return nil
}

func (p *panicOnceNotifier) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func TestScheduler_Tick_PanicDoesNotCrashLoop(t *testing.T) {
	job := &panicOnceJob{}
	notifier := &fakeNotifier{}
	s := newScheduler("@every 50ms", job, notifier)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer s.Stop(context.Background())

	// First tick panics inside Job.Run; second tick should still fire and
	// reach the notifier — proves the panic was recovered and the cron loop
	// is intact.
	if !waitFor(t, 2*time.Second, func() bool {
		return job.callCount() >= 2 && notifier.callCount() >= 1
	}) {
		t.Fatalf("expected loop to survive panic; job calls=%d, notifier calls=%d",
			job.callCount(), notifier.callCount())
	}
}

func TestScheduler_Tick_NotifierPanicDoesNotCrashLoop(t *testing.T) {
	job := &fakeJobRunner{result: AMAJobResult{Outcome: OutcomeProcessed, VideoID: "v"}}
	notifier := &panicOnceNotifier{}
	s := newScheduler("@every 50ms", job, notifier)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer s.Stop(context.Background())

	// First tick: job runs, notifier panics — recovered.
	// Second tick: job runs, notifier returns nil. Both happen.
	if !waitFor(t, 2*time.Second, func() bool {
		return notifier.callCount() >= 2 && job.callCount() >= 2
	}) {
		t.Fatalf("expected loop to survive notifier panic; job calls=%d, notifier calls=%d",
			job.callCount(), notifier.callCount())
	}
}

// slowJob's Run blocks until `duration` elapses or ctx is cancelled, counting
// completions. Used to demonstrate that SkipIfStillRunning drops overlapping
// ticks — while one Run is in flight, subsequent scheduled ticks are dropped
// rather than running concurrently.
type slowJob struct {
	duration  time.Duration
	completed int32
}

func (s *slowJob) Run(ctx context.Context) AMAJobResult {
	select {
	case <-time.After(s.duration):
	case <-ctx.Done():
	}
	atomic.AddInt32(&s.completed, 1)
	return AMAJobResult{Outcome: OutcomeProcessed, VideoID: "slow"}
}

func (s *slowJob) completedCount() int32 {
	return atomic.LoadInt32(&s.completed)
}

func TestScheduler_OverlappingRuns_AreSkipped(t *testing.T) {
	// robfig/cron's @every parser rounds sub-second durations up to 1s, so
	// we use the smallest supported tick (1s) and a 3s blocking job. With
	// SkipIfStillRunning, only the first tick executes — the next 2-3 ticks
	// fire while Run is in flight and are dropped. Without the wrapper, each
	// tick would spawn its own concurrent Run goroutine.
	job := &slowJob{duration: 3 * time.Second}
	notifier := &fakeNotifier{}
	s := newScheduler("@every 1s", job, notifier)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	// 3-4 nominal ticks fire during this window (at ~1s, 2s, 3s).
	time.Sleep(3500 * time.Millisecond)

	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Stop(stopCtx); err != nil {
		t.Fatalf("Stop error: %v", err)
	}

	completed := job.completedCount()
	notifyCalls := int32(notifier.callCount())

	// With SkipIfStillRunning: ~1 completion (the first tick's Run, cut
	// short by Stop's ctx cancellation). Without it: 3 concurrent Runs
	// would have started and all completed when Stop cancelled their ctx.
	if completed >= 3 {
		t.Errorf("expected SkipIfStillRunning to drop overlapping ticks; "+
			"got %d completed Run() calls, want roughly 1-2", completed)
	}
	if completed == 0 {
		t.Errorf("expected at least 1 completed Run() call, got 0")
	}
	// Notifier is called once per completed Run, never more.
	if notifyCalls > completed {
		t.Errorf("notifier called %d times, exceeds completed Run count %d",
			notifyCalls, completed)
	}
}
