// Package gcp provides a DriftWatch provider for Google Cloud Platform resources.
package gcp

import (
	"context"
	"fmt"

	"github.com/youorg/driftwatch/internal/provider"
	"github.com/youorg/driftwatch/internal/snapshot"
)

const providerType = "gcp"

// Client is the interface for fetching GCP instance data.
type Client interface {
	ListInstances(ctx context.Context, project string) ([]Instance, error)
}

// Instance represents a minimal GCP Compute Engine instance record.
type Instance struct {
	ID     string
	Name   string
	Zone   string
	Status string
	Labels map[string]string
}

// gcpProvider collects snapshots from GCP Compute Engine.
type gcpProvider struct {
	client  Client
	project string
}

// New creates a new GCP provider using the supplied client and project ID.
func New(client Client, project string) *gcpProvider {
	return &gcpProvider{client: client, project: project}
}

// Collect fetches current GCP instances and returns them as snapshots.
func (p *gcpProvider) Collect(ctx context.Context) ([]*snapshot.Snapshot, error) {
	instances, err := p.client.ListInstances(ctx, p.project)
	if err != nil {
		return nil, fmt.Errorf("gcp: list instances: %w", err)
	}

	snaps := make([]*snapshot.Snapshot, 0, len(instances))
	for _, inst := range instances {
		attrs := map[string]string{
			"name":   inst.Name,
			"zone":   inst.Zone,
			"status": inst.Status,
		}
		for k, v := range inst.Labels {
			attrs["label:"+k] = v
		}
		snaps = append(snaps, snapshot.New(inst.ID, providerType, attrs))
	}
	return snaps, nil
}

// init registers the GCP provider with the global provider registry.
func init() {
	provider.Register(providerType, func(cfg map[string]string) (provider.Provider, error) {
		project, ok := cfg["project"]
		if !ok || project == "" {
			return nil, fmt.Errorf("gcp: missing required config key \"project\"")
		}
		// Real usage would construct an authenticated GCP client here.
		return nil, fmt.Errorf("gcp: real client construction not implemented; use New() directly")
	})
}
