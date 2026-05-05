package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/scheduler"
	"github.com/yourorg/driftwatch/internal/snapshot"
)

// stubProvider counts how many times Collect is called.
type stubProvider struct {
	calls atomic.Int32
	snaps []*snapshot.Snapshot
}

func (p *stubProvider) Collect(_ context.Context) ([]*snapshot.Snapshot, error) {
	p.calls.Add(1)
	return p.snaps, nil
}

func newJob(t *testing.T, p *stubProvider, interval time.Duration) *scheduler.Job {
	t.Helper()
	store := snapshot.NewStore()
	sink := alert.NewLogSink(nil)
	fanout := alert.NewFanout(sink)
	detector := drift.NewDetector(store, fanout)
	return &scheduler.Job{
		Name:     "test-job",
		Provider: p,
		Detector: detector,
		Interval: interval,
	}
}

// TestRun_CallsCollectImmediately verifies the first collection happens before
// the first tick fires.
func TestRun_CallsCollectImmediately(t *testing.T) {
	p := &stubProvider{}
	job := newJob(t, p, 10*time.Second) // long interval – tick must not fire

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sched := scheduler.New([]*scheduler.Job{job})
	sched.Run(ctx)

	if got := p.calls.Load(); got < 1 {
		t.Fatalf("expected at least 1 Collect call, got %d", got)
	}
}

// TestRun_TicksMultipleTimes verifies repeated collection over several ticks.
func TestRun_TicksMultipleTimes(t *testing.T) {
	p := &stubProvider{}
	job := newJob(t, p, 20*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	sched := scheduler.New([]*scheduler.Job{job})
	sched.Run(ctx)

	// Expect initial call + at least 3 ticks within 120 ms.
	if got := p.calls.Load(); got < 4 {
		t.Fatalf("expected >=4 Collect calls, got %d", got)
	}
}

// TestRun_StopsOnContextCancel verifies the scheduler exits promptly.
func TestRun_StopsOnContextCancel(t *testing.T) {
	p := &stubProvider{}
	job := newJob(t, p, 10*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	sched := scheduler.New([]*scheduler.Job{job})

	done := make(chan struct{})
	go func() {
		sched.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}
