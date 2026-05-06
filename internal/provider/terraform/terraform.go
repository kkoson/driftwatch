// Package terraform provides a driftwatch provider that reads infrastructure
// snapshots from Terraform state files (local or remote via HTTP).
package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/example/driftwatch/internal/provider"
	"github.com/example/driftwatch/internal/snapshot"
)

const providerType = "terraform"

// StateClient abstracts reading a Terraform state payload.
type StateClient interface {
	FetchState(ctx context.Context) ([]byte, error)
}

// Provider collects snapshots by parsing a Terraform state source.
type Provider struct {
	client StateClient
}

// New creates a Provider backed by the given StateClient.
func New(c StateClient) *Provider {
	return &Provider{client: c}
}

// Collect parses the Terraform state and returns one snapshot per resource.
func (p *Provider) Collect(ctx context.Context) ([]*snapshot.Snapshot, error) {
	data, err := p.client.FetchState(ctx)
	if err != nil {
		return nil, fmt.Errorf("terraform: fetch state: %w", err)
	}

	var state tfState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("terraform: parse state: %w", err)
	}

	var snaps []*snapshot.Snapshot
	for _, res := range state.Resources {
		for _, inst := range res.Instances {
			attrs := make(map[string]string, len(inst.Attributes)+2)
			for k, v := range inst.Attributes {
				attrs[k] = fmt.Sprintf("%v", v)
			}
			attrs["tf_type"] = res.Type
			attrs["tf_name"] = res.Name
			id := fmt.Sprintf("%s.%s", res.Type, res.Name)
			snaps = append(snaps, snapshot.New(id, providerType, attrs))
		}
	}
	return snaps, nil
}

// tfState mirrors the subset of Terraform state JSON we need.
type tfState struct {
	Resources []tfResource `json:"resources"`
}

type tfResource struct {
	Type      string       `json:"type"`
	Name      string       `json:"name"`
	Instances []tfInstance `json:"instances"`
}

type tfInstance struct {
	Attributes map[string]interface{} `json:"attributes"`
}

// fileClient reads state from a local file path.
type fileClient struct{ path string }

func (f *fileClient) FetchState(_ context.Context) ([]byte, error) {
	return os.ReadFile(f.path)
}

// httpClient reads state from a remote HTTP endpoint.
type httpClient struct {
	url    string
	client *http.Client
}

func (h *httpClient) FetchState(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func init() {
	provider.Register(providerType, func(cfg map[string]string) (provider.Provider, error) {
		if path, ok := cfg["state_file"]; ok {
			return New(&fileClient{path: path}), nil
		}
		if url, ok := cfg["state_url"]; ok {
			return New(&httpClient{
				url:    url,
				client: &http.Client{Timeout: 15 * time.Second},
			}), nil
		}
		return nil, fmt.Errorf("terraform provider requires 'state_file' or 'state_url'")
	})
}
