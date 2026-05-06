// Package kubernetes provides a driftwatch provider for Kubernetes resources.
package kubernetes

import (
	"context"
	"fmt"

	"github.com/your-org/driftwatch/internal/provider"
	"github.com/your-org/driftwatch/internal/snapshot"
)

// Client is the interface for listing Kubernetes resources.
type Client interface {
	ListDeployments(ctx context.Context, namespace string) ([]Resource, error)
}

// Resource represents a simplified Kubernetes resource.
type Resource struct {
	Name      string
	Namespace string
	Replicas  int32
	Image     string
	Labels    map[string]string
}

// k8sProvider collects snapshots from a Kubernetes cluster.
type k8sProvider struct {
	client    Client
	namespace string
}

// New creates a new Kubernetes provider.
func New(client Client, namespace string) provider.Provider {
	return &k8sProvider{client: client, namespace: namespace}
}

// Collect fetches Kubernetes deployments and returns snapshots.
func (p *k8sProvider) Collect(ctx context.Context) ([]*snapshot.Snapshot, error) {
	resources, err := p.client.ListDeployments(ctx, p.namespace)
	if err != nil {
		return nil, fmt.Errorf("kubernetes: list deployments: %w", err)
	}

	snaps := make([]*snapshot.Snapshot, 0, len(resources))
	for _, r := range resources {
		attrs := map[string]string{
			"namespace": r.Namespace,
			"replicas":  fmt.Sprintf("%d", r.Replicas),
			"image":     r.Image,
		}
		for k, v := range r.Labels {
			attrs["label:"+k] = v
		}
		snaps = append(snaps, snapshot.New(
			fmt.Sprintf("k8s/%s/%s", r.Namespace, r.Name),
			"kubernetes",
			attrs,
		))
	}
	return snaps, nil
}

func init() {
	provider.Register("kubernetes", func(cfg map[string]string) (provider.Provider, error) {
		namespace := cfg["namespace"]
		if namespace == "" {
			namespace = "default"
		}
		// Real usage would build a proper client from kubeconfig/in-cluster config.
		return nil, fmt.Errorf("kubernetes: use New() directly with a configured client")
		_ = namespace
	})
}
