// Package aws provides an AWS infrastructure snapshot provider for driftwatch.
package aws

import (
	"context"
	"fmt"

	"github.com/your-org/driftwatch/internal/provider"
	"github.com/your-org/driftwatch/internal/snapshot"
)

const ProviderType = "aws"

// Client is the interface for fetching AWS resource attributes.
// It is satisfied by the real AWS SDK client and by fakes in tests.
type Client interface {
	ListInstances(ctx context.Context) ([]InstanceInfo, error)
}

// InstanceInfo holds the raw attributes returned by the AWS client.
type InstanceInfo struct {
	ID         string
	Region     string
	InstanceType string
	State      string
	Tags       map[string]string
}

type awsProvider struct {
	client Client
	region string
}

// New creates an AWS provider using the supplied client and region.
func New(client Client, region string) provider.Provider {
	return &awsProvider{client: client, region: region}
}

// Collect fetches EC2 instance state and returns snapshots.
func (p *awsProvider) Collect(ctx context.Context) ([]*snapshot.Snapshot, error) {
	instances, err := p.client.ListInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws: list instances: %w", err)
	}

	snaps := make([]*snapshot.Snapshot, 0, len(instances))
	for _, inst := range instances {
		attrs := map[string]string{
			"region":        inst.Region,
			"instance_type": inst.InstanceType,
			"state":         inst.State,
		}
		for k, v := range inst.Tags {
			attrs["tag:"+k] = v
		}
		snaps = append(snaps, snapshot.New(ProviderType, inst.ID, attrs))
	}
	return snaps, nil
}

func init() {
	provider.Register(ProviderType, func(cfg map[string]string) (provider.Provider, error) {
		region, ok := cfg["region"]
		if !ok || region == "" {
			return nil, fmt.Errorf("aws provider: missing required config key 'region'")
		}
		// Real usage would construct an SDK client here.
		return nil, fmt.Errorf("aws provider: use New() directly with a real SDK client")
	})
}
