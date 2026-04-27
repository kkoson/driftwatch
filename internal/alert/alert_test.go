package alert_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/driftwatch/internal/alert"
)

func makeAlert() alert.Alert {
	return alert.Alert{
		ResourceID:  "i-0abc123",
		Provider:    "aws",
		ChangedKeys: []string{"instance_type", "tags"},
		Severity:    alert.SeverityWarning,
		DetectedAt:  time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Message:     "instance type changed",
	}
}

func TestAlert_String(t *testing.T) {
	a := makeAlert()
	s := a.String()
	if !strings.Contains(s, "i-0abc123") {
		t.Errorf("expected resource id in string, got: %s", s)
	}
	if !strings.Contains(s, "warning") {
		t.Errorf("expected severity in string, got: %s", s)
	}
	if !strings.Contains(s, "instance_type") {
		t.Errorf("expected changed key in string, got: %s", s)
	}
}

func TestLogSink_Send(t *testing.T) {
	var captured string
	sink := alert.NewLogSink(func(format string, args ...any) {
		captured = fmt.Sprintf(format, args...)
	})
	if err := sink.Send(makeAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(captured, "aws") {
		t.Errorf("expected provider in log output, got: %s", captured)
	}
}

func TestFanout_SendAll(t *testing.T) {
	var count int
	countSink := &captureSink{onSend: func(alert.Alert) error { count++; return nil }}
	f := alert.NewFanout(countSink, countSink)
	if err := f.Send(makeAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 sends, got %d", count)
	}
}

func TestFanout_CollectsErrors(t *testing.T) {
	errSink := &captureSink{onSend: func(alert.Alert) error { return errors.New("sink down") }}
	f := alert.NewFanout(errSink, errSink)
	err := f.Send(makeAlert())
	if err == nil {
		t.Fatal("expected error from fanout, got nil")
	}
	if !strings.Contains(err.Error(), "fanout") {
		t.Errorf("expected fanout prefix in error, got: %v", err)
	}
}

func TestFanout_Add(t *testing.T) {
	f := alert.NewFanout()
	if f.Len() != 0 {
		t.Fatalf("expected 0 sinks, got %d", f.Len())
	}
	f.Add(&captureSink{})
	if f.Len() != 1 {
		t.Errorf("expected 1 sink after Add, got %d", f.Len())
	}
}

type captureSink struct {
	onSend func(alert.Alert) error
}

func (c *captureSink) Send(a alert.Alert) error {
	if c.onSend != nil {
		return c.onSend(a)
	}
	return nil
}
