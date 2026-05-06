// Package datadog provides a driftwatch provider that collects monitor
// and dashboard configuration snapshots from the Datadog API.
package datadog

import (
	"context"
	"fmt"
	"strconv"

	"github.com/youorg/driftwatch/internal/provider"
	"github.com/youorg/driftwatch/internal/snapshot"
)

const providerType = "datadog"

// Client is the subset of the Datadog API used by this provider.
type Client interface {
	ListMonitors(ctx context.Context) ([]Monitor, error)
}

// Monitor represents a Datadog monitor returned by the API.
type Monitor struct {
	ID      int64
	Name    string
	Type    string
	Query   string
	Message string
}

// datadogProvider collects snapshots from Datadog.
type datadogProvider struct {
	client Client
}

// New returns a new Datadog provider backed by the given client.
func New(client Client) provider.Provider {
	return &datadogProvider{client: client}
}

// Collect fetches all Datadog monitors and returns them as snapshots.
func (p *datadogProvider) Collect(ctx context.Context) ([]*snapshot.Snapshot, error) {
	monitors, err := p.client.ListMonitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("datadog: list monitors: %w", err)
	}

	snaps := make([]*snapshot.Snapshot, 0, len(monitors))
	for _, m := range monitors {
		attrs := map[string]string{
			"name":    m.Name,
			"type":    m.Type,
			"query":   m.Query,
			"message": m.Message,
		}
		id := fmt.Sprintf("%s/monitor/%s", providerType, strconv.FormatInt(m.ID, 10))
		snaps = append(snaps, snapshot.New(id, providerType, attrs))
	}
	return snaps, nil
}

func init() {
	provider.Register(providerType, func(cfg map[string]string) (provider.Provider, error) {
		apiKey := cfg["api_key"]
		appKey := cfg["app_key"]
		if apiKey == "" || appKey == "" {
			return nil, fmt.Errorf("datadog: api_key and app_key are required")
		}
		client := newHTTPClient(apiKey, appKey)
		return New(client), nil
	})
}
