// Package provider defines the Provider interface and a registry for
// building named provider instances from configuration maps.
package provider

import (
	"context"
	"fmt"

	"github.com/your-org/driftwatch/internal/snapshot"
)

// Provider collects infrastructure snapshots from a cloud environment.
type Provider interface {
	Collect(ctx context.Context) ([]*snapshot.Snapshot, error)
}

// BuildFunc constructs a Provider from a string→string config map.
type BuildFunc func(cfg map[string]string) (Provider, error)

var registry = map[string]BuildFunc{}

// Register associates name with a BuildFunc. It is typically called from
// provider package init() functions.
func Register(name string, fn BuildFunc) {
	registry[name] = fn
}

// Build creates a Provider by name using the supplied config.
func Build(name string, cfg map[string]string) (Provider, error) {
	fn, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("provider %q is not registered", name)
	}
	return fn(cfg)
}

// Registered returns the names of all registered providers.
func Registered() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	return names
}
