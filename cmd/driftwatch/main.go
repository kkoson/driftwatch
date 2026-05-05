// Package main is the entry point for the driftwatch daemon.
// It wires together configuration, providers, drift detection,
// alerting, and scheduling into a running process.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/config"
	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/provider"
	"github.com/yourorg/driftwatch/internal/scheduler"
	"github.com/yourorg/driftwatch/internal/snapshot"

	// Register built-in providers via their init() functions.
	_ "github.com/yourorg/driftwatch/internal/provider/aws"
	_ "github.com/yourorg/driftwatch/internal/provider/azure"
	_ "github.com/yourorg/driftwatch/internal/provider/gcp"
	_ "github.com/yourorg/driftwatch/internal/provider/mock"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "driftwatch: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := flag.String("config", "driftwatch.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting driftwatch",
		"provider", cfg.Provider.Type,
		"poll_interval", cfg.PollInterval,
	)

	// Build the configured provider.
	p, err := provider.Build(cfg.Provider.Type, cfg.Provider.Options)
	if err != nil {
		return fmt.Errorf("build provider %q: %w", cfg.Provider.Type, err)
	}

	// Set up alert sinks.
	sinks := []alert.Sink{
		alert.NewLogSink(logger),
	}
	if cfg.Webhook.URL != "" {
		sinks = append(sinks, alert.NewWebhookSink(cfg.Webhook.URL, cfg.Webhook.Timeout))
		slog.Info("webhook alerting enabled", "url", cfg.Webhook.URL)
	}
	fanout := alert.NewFanout(sinks...)

	// Shared snapshot store and detector.
	store := snapshot.NewStore()
	detector := drift.NewDetector(store, fanout)

	// Build the scheduler job: collect snapshots then observe each for drift.
	job := func(ctx context.Context) error {
		snaps, err := p.Collect(ctx)
		if err != nil {
			return fmt.Errorf("collect: %w", err)
		}
		for _, s := range snaps {
			if obsErr := detector.Observe(ctx, s); obsErr != nil {
				slog.Warn("observe error", "resource", s.ResourceID, "err", obsErr)
			}
		}
		return nil
	}

	sched := scheduler.New(cfg.PollInterval, job)

	// Run until SIGINT / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("driftwatch running — press Ctrl+C to stop")
	if err := sched.Run(ctx); err != nil {
		return fmt.Errorf("scheduler: %w", err)
	}

	slog.Info("driftwatch stopped")
	return nil
}
