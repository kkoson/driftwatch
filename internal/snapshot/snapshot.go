package snapshot

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// Snapshot represents a point-in-time capture of a resource's configuration.
type Snapshot struct {
	ResourceID   string            `json:"resource_id"`
	Provider     string            `json:"provider"`
	ResourceType string            `json:"resource_type"`
	Attributes   map[string]any    `json:"attributes"`
	Checksum     string            `json:"checksum"`
	CapturedAt   time.Time         `json:"captured_at"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// New creates a new Snapshot for the given resource and computes its checksum.
func New(resourceID, provider, resourceType string, attributes map[string]any, labels map[string]string) (*Snapshot, error) {
	checksum, err := computeChecksum(attributes)
	if err != nil {
		return nil, fmt.Errorf("computing checksum: %w", err)
	}

	return &Snapshot{
		ResourceID:   resourceID,
		Provider:     provider,
		ResourceType: resourceType,
		Attributes:   attributes,
		Checksum:     checksum,
		CapturedAt:   time.Now().UTC(),
		Labels:       labels,
	}, nil
}

// DiffersFrom returns true if the snapshot's checksum differs from another snapshot.
func (s *Snapshot) DiffersFrom(other *Snapshot) bool {
	return s.Checksum != other.Checksum
}

// ChangedKeys returns the list of attribute keys that differ between two snapshots.
func (s *Snapshot) ChangedKeys(other *Snapshot) []string {
	var changed []string
	for k, v := range s.Attributes {
		ov, ok := other.Attributes[k]
		if !ok {
			changed = append(changed, k)
			continue
		}
		if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", ov) {
			changed = append(changed, k)
		}
	}
	for k := range other.Attributes {
		if _, ok := s.Attributes[k]; !ok {
			changed = append(changed, k)
		}
	}
	return changed
}

func computeChecksum(attributes map[string]any) (string, error) {
	data, err := json.Marshal(attributes)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}
