// Package consul provides a driftwatch provider for HashiCorp Consul
// service and key-value configuration.
package consul

import (
	"context"
	"fmt"

	"github.com/your-org/driftwatch/internal/provider"
	"github.com/your-org/driftwatch/internal/snapshot"
)

// Client is the interface satisfied by the Consul agent used in production
// and by the fake used in tests.
type Client interface {
	Services(ctx context.Context) (map[string]map[string]string, error)
}

// Provider collects snapshots from a Consul agent.
type Provider struct {
	client Client
	agent  string
}

// New returns a Provider that queries the given Consul client.
func New(client Client, agent string) *Provider {
	return &Provider{client: client, agent: agent}
}

// Collect fetches all registered services from Consul and returns one
// snapshot per service.
func (p *Provider) Collect(ctx context.Context) ([]*snapshot.Snapshot, error) {
	services, err := p.client.Services(ctx)
	if err != nil {
		return nil, fmt.Errorf("consul: list services: %w", err)
	}

	snaps := make([]*snapshot.Snapshot, 0, len(services))
	for name, tags := range services {
		attrs := make(map[string]string, len(tags)+1)
		attrs["agent"] = p.agent
		for k, v := range tags {
			attrs[k] = v
		}
		snaps = append(snaps, snapshot.New(
			fmt.Sprintf("consul/%s/%s", p.agent, name),
			"consul",
			attrs,
		))
	}
	return snaps, nil
}

func init() {
	provider.Register("consul", func(cfg map[string]string) (provider.Provider, error) {
		agent, ok := cfg["agent"]
		if !ok || agent == "" {
			return nil, fmt.Errorf("consul: missing required config key \"agent\"")
		}
		// In production a real Consul HTTP client would be constructed here.
		return nil, fmt.Errorf("consul: real client construction not implemented; use New() directly")
	})
}
