// Package alert provides types and interfaces for emitting drift alerts.
package alert

import (
	"fmt"
	"time"
)

// Severity indicates how critical a drift event is.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Alert represents a detected drift event for a single resource.
type Alert struct {
	ResourceID   string
	Provider     string
	ChangedKeys  []string
	Severity     Severity
	DetectedAt   time.Time
	Message      string
}

// String returns a human-readable representation of the alert.
func (a Alert) String() string {
	return fmt.Sprintf(
		"[%s] drift detected on %s/%s at %s: changed keys=%v",
		a.Severity,
		a.Provider,
		a.ResourceID,
		a.DetectedAt.Format(time.RFC3339),
		a.ChangedKeys,
	)
}

// Sink is the interface that alert destinations must implement.
type Sink interface {
	Send(a Alert) error
}

// LogSink writes alerts to a standard logger-style writer.
type LogSink struct {
	printf func(format string, args ...any)
}

// NewLogSink creates a LogSink using the provided printf-style function.
func NewLogSink(printf func(format string, args ...any)) *LogSink {
	return &LogSink{printf: printf}
}

// Send emits the alert via the configured printf function.
func (l *LogSink) Send(a Alert) error {
	l.printf("%s\n", a.String())
	return nil
}
