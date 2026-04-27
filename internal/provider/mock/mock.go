// Package mock provides a mock cloud provider for testing and local development.
package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/yourorg/driftwatch/internal/provider"
	"github.com/yourorg/driftwatch/internal/snapshot"
)

const ProviderType = "mock"

func init() {
	provider.Register(ProviderType, func(cfg map[string]any) (provider.Provider, error) {
		return New(cfg)
	})
}

// MockProvider is an in-memory provider whose resources can be mutated at
// runtime, making it useful for unit tests and local demos.
type MockProvider struct {
	mu        sync.RWMutex
	resources map[string]map[string]any
}

// New creates a MockProvider. The optional "resources" key in cfg may contain
// a map[string]map[string]any seed value.
func New(cfg map[string]any) (*MockProvider, error) {
	mp := &MockProvider{
		resources: make(map[string]map[string]any),
	}
	if seed, ok := cfg["resources"]; ok {
		seeded, ok := seed.(map[string]map[string]any)
		if !ok {
			return nil, fmt.Errorf("mock: 'resources' must be map[string]map[string]any")
		}
		for id, attrs := range seeded {
			mp.resources[id] = attrs
		}
	}
	return mp, nil
}

// SetResource inserts or replaces a resource, safe for concurrent use.
func (m *MockProvider) SetResource(id string, attrs map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resources[id] = attrs
}

// DeleteResource removes a resource, safe for concurrent use.
func (m *MockProvider) DeleteResource(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.resources, id)
}

// Collect implements provider.Provider and returns a snapshot per resource.
func (m *MockProvider) Collect(_ context.Context) ([]*snapshot.Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snaps := make([]*snapshot.Snapshot, 0, len(m.resources))
	for id, attrs := range m.resources {
		s, err := snapshot.New(ProviderType, id, attrs)
		if err != nil {
			return nil, fmt.Errorf("mock: building snapshot for %s: %w", id, err)
		}
		snaps = append(snaps, s)
	}
	return snaps, nil
}
