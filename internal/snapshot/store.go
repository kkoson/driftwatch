package snapshot

import (
	"fmt"
	"sync"
)

// Store holds the most recent snapshot for each resource.
type Store struct {
	mu    sync.RWMutex
	items map[string]*Snapshot
}

// NewStore initialises an empty snapshot store.
func NewStore() *Store {
	return &Store{
		items: make(map[string]*Snapshot),
	}
}

// Put saves a snapshot, overwriting any previous entry for the same resource.
func (s *Store) Put(snap *Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[snap.ResourceID] = snap
}

// Get retrieves the snapshot for a resource. Returns an error if not found.
func (s *Store) Get(resourceID string) (*Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.items[resourceID]
	if !ok {
		return nil, fmt.Errorf("snapshot not found for resource %q", resourceID)
	}
	return snap, nil
}

// Delete removes the snapshot for a resource.
func (s *Store) Delete(resourceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, resourceID)
}

// All returns a copy of all stored snapshots.
func (s *Store) All() []*Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Snapshot, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	return out
}

// Len returns the number of snapshots in the store.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}
