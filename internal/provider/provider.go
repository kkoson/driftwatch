// Package provider defines the interface and registry for cloud infrastructure
// providers. Each provider is responsible for collecting snapshots of resources
// within a specific cloud environment (e.g., AWS, GCP, Azure).
package provider

import (
	"context"
	"fmt"

	"github.com/yourorg/driftwatch/internal/snapshot"
)

// Provider is the interface that every cloud provider adapter must implement.
// Collect returns a slice of snapshots representing the current observed state
// of all resources the provider is responsible for monitoring.
type Provider interface {
	// Name returns a human-readable identifier for the provider (e.g., "aws", "gcp").
	Name() string

	// Collect gathers the current state of all monitored resources and returns
	// them as snapshots. Implementations should respect context cancellation.
	Collect(ctx context.Context) ([]*snapshot.Snapshot, error)
}

// FactoryFunc is a constructor function that creates a new Provider instance
// from an arbitrary configuration map. The map values come directly from the
// parsed YAML/JSON configuration file.
type FactoryFunc func(cfg map[string]any) (Provider, error)

// registry holds all registered provider factories, keyed by provider type
// name (e.g., "aws", "gcp", "azure").
var registry = map[string]FactoryFunc{}

// Register associates a provider type name with its factory function.
// It is intended to be called from provider-specific init() functions so that
// providers are automatically available when their package is imported.
//
// Register panics if the same name is registered more than once to surface
// misconfiguration early during startup.
func Register(name string, fn FactoryFunc) {
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("provider: duplicate registration for %q", name))
	}
	registry[name] = fn
}

// Build looks up the factory for the given provider type and constructs a new
// Provider using the supplied configuration map. It returns an error if the
// provider type is unknown or if the factory itself returns an error.
func Build(providerType string, cfg map[string]any) (Provider, error)	{
	fn, ok := registry[providerType]
	if !ok {
		return nil, fmt.Errorf("provider: unknown provider type %q (is the package imported?)", providerType)
	}
	return fn(cfg)
}

// Registered returns the names of all currently registered provider types.
// The order of the returned slice is non-deterministic.
func Registered() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
