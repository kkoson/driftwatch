// Package scheduler drives periodic snapshot collection and drift detection
// for every configured provider.
package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/provider"
	"github.com/yourorg/driftwatch/internal/snapshot"
)

// Job holds the runtime state for a single provider polling loop.
type Job struct {
	Name     string
	Provider provider.Provider
	Detector *drift.Detector
	Interval time.Duration
}

// Scheduler owns a collection of Jobs and runs them concurrently.
type Scheduler struct {
	jobs []*Job
	wg   sync.WaitGroup
}

// New creates a Scheduler pre-loaded with the supplied jobs.
func New(jobs []*Job) *Scheduler {
	return &Scheduler{jobs: jobs}
}

// Run starts every job in its own goroutine and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	for _, j := range s.jobs {
		s.wg.Add(1)
		go s.runJob(ctx, j)
	}
	s.wg.Wait()
}

func (s *Scheduler) runJob(ctx context.Context, j *Job) {
	defer s.wg.Done()

	log.Printf("scheduler: starting job %q (interval=%s)", j.Name, j.Interval)
	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()

	// Collect immediately on startup, then on each tick.
	s.collect(ctx, j)

	for {
		select {
		case <-ticker.C:
			s.collect(ctx, j)
		case <-ctx.Done():
			log.Printf("scheduler: stopping job %q", j.Name)
			return
		}
	}
}

func (s *Scheduler) collect(ctx context.Context, j *Job) {
	snaps, err := j.Provider.Collect(ctx)
	if err != nil {
		log.Printf("scheduler: job %q collect error: %v", j.Name, err)
		return
	}
	for _, snap := range snaps {
		if err := j.Detector.Observe(ctx, snap); err != nil {
			log.Printf("scheduler: job %q observe error for %s: %v", j.Name, snap.ID, err)
		}
	}
	_ = snapshot.Snapshot{} // keep import live during compilation
}
