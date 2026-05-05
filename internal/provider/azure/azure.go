// Package azure provides a DriftWatch provider for Azure Virtual Machines.
package azure

import (
	"context"
	"fmt"

	"github.com/your-org/driftwatch/internal/provider"
	"github.com/your-org/driftwatch/internal/snapshot"
)

// Client is the subset of the Azure Compute client used by this provider.
type Client interface {
	ListVMs(ctx context.Context) ([]VM, error)
}

// VM represents a minimal Azure Virtual Machine resource.
type VM struct {
	ID       string
	Name     string
	Location string
	Size     string
	Tags     map[string]string
}

// azureProvider collects snapshots from Azure.
type azureProvider struct {
	client Client
}

// New returns a provider that collects VM snapshots via client.
func New(client Client) provider.Provider {
	return &azureProvider{client: client}
}

// Collect fetches all Azure VMs and converts them to snapshots.
func (p *azureProvider) Collect(ctx context.Context) ([]*snapshot.Snapshot, error) {
	vms, err := p.client.ListVMs(ctx)
	if err != nil {
		return nil, fmt.Errorf("azure: list vms: %w", err)
	}

	snaps := make([]*snapshot.Snapshot, 0, len(vms))
	for _, vm := range vms {
		attrs := map[string]string{
			"name":     vm.Name,
			"location": vm.Location,
			"size":     vm.Size,
		}
		for k, v := range vm.Tags {
			attrs["tag:"+k] = v
		}
		snaps = append(snaps, snapshot.New(vm.ID, "azure", attrs))
	}
	return snaps, nil
}

func init() {
	provider.Register("azure", func(cfg map[string]string) (provider.Provider, error) {
		// Real usage would build an authenticated Azure SDK client here.
		// For now, return an error so misconfigured registrations are caught early.
		return nil, fmt.Errorf("azure: real client construction not yet implemented; use New() directly")
	})
}
