package drift_test

import (
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/snapshot"
)

func makeSnap(id string, attrs map[string]string) *snapshot.Snapshot {
	s, _ := snapshot.New(id, "aws", attrs, time.Now())
	return s
}

func newDetector() *drift.Detector {
	store := snapshot.NewStore()
	return drift.NewDetector(store)
}

func TestObserve_NoPreviousSnapshot_NoAlert(t *testing.T) {
	d := newDetector()
	snap := makeSnap("res-1", map[string]string{"region": "us-east-1"})
	if err := d.Observe(snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case a := <-d.Alerts():
		t.Fatalf("expected no alert, got: %v", a)
	default:
	}
}

func TestObserve_UnchangedSnapshot_NoAlert(t *testing.T) {
	d := newDetector()
	attrs := map[string]string{"region": "us-east-1"}
	d.Observe(makeSnap("res-1", attrs)) //nolint:errcheck
	if err := d.Observe(makeSnap("res-1", attrs)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case a := <-d.Alerts():
		t.Fatalf("expected no alert, got: %v", a)
	default:
	}
}

func TestObserve_ChangedSnapshot_EmitsAlert(t *testing.T) {
	d := newDetector()
	d.Observe(makeSnap("res-2", map[string]string{"region": "us-east-1"})) //nolint:errcheck
	if err := d.Observe(makeSnap("res-2", map[string]string{"region": "eu-west-1"})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case a := <-d.Alerts():
		if a.ResourceID != "res-2" {
			t.Errorf("expected resource res-2, got %s", a.ResourceID)
		}
		if len(a.ChangedKeys) == 0 {
			t.Error("expected changed keys, got none")
		}
		if a.OldChecksum == a.NewChecksum {
			t.Error("expected different checksums")
		}
	default:
		t.Fatal("expected an alert but channel was empty")
	}
}

func TestAlert_String(t *testing.T) {
	a := drift.Alert{
		ResourceID:  "res-3",
		ChangedKeys: []string{"size"},
		DetectedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	s := a.String()
	if s == "" {
		t.Error("expected non-empty string from Alert.String()")
	}
}
