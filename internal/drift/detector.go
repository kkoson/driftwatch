package drift

import (
	"fmt"
	"time"

	"github.com/yourorg/driftwatch/internal/snapshot"
)

// Alert represents a detected drift event.
type Alert struct {
	ResourceID  string
	ChangedKeys []string
	OldChecksum string
	NewChecksum string
	DetectedAt  time.Time
}

func (a Alert) String() string {
	return fmt.Sprintf("drift detected on %s at %s: changed keys=%v",
		a.ResourceID, a.DetectedAt.Format(time.RFC3339), a.ChangedKeys)
}

// Detector compares incoming snapshots against a store and emits alerts.
type Detector struct {
	store  *snapshot.Store
	alerts chan Alert
}

// NewDetector creates a Detector backed by the given store.
func NewDetector(store *snapshot.Store) *Detector {
	return &Detector{
		store:  store,
		alerts: make(chan Alert, 64),
	}
}

// Alerts returns a read-only channel of drift alerts.
func (d *Detector) Alerts() <-chan Alert {
	return d.alerts
}

// Observe processes a new snapshot. If a previous snapshot exists and differs,
// an Alert is sent on the Alerts channel. The store is updated regardless.
func (d *Detector) Observe(snap *snapshot.Snapshot) error {
	prev, err := d.store.Get(snap.ResourceID)
	if err == nil && prev.DiffersFrom(snap) {
		d.alerts <- Alert{
			ResourceID:  snap.ResourceID,
			ChangedKeys: prev.ChangedKeys(snap),
			OldChecksum: prev.Checksum,
			NewChecksum: snap.Checksum,
			DetectedAt:  time.Now().UTC(),
		}
	}
	return d.store.Put(snap)
}
