package gcp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/youorg/driftwatch/internal/provider"
	"github.com/youorg/driftwatch/internal/provider/gcp"
)

// fakeClient implements gcp.Client for testing.
type fakeClient struct {
	instances []gcp.Instance
	err       error
}

func (f *fakeClient) ListInstances(_ context.Context, _ string) ([]gcp.Instance, error) {
	return f.instances, f.err
}

func TestCollect_ReturnsSnapshots(t *testing.T) {
	client := &fakeClient{
		instances: []gcp.Instance{
			{ID: "i-1", Name: "web-1", Zone: "us-central1-a", Status: "RUNNING", Labels: map[string]string{"env": "prod"}},
			{ID: "i-2", Name: "web-2", Zone: "us-central1-b", Status: "TERMINATED", Labels: nil},
		},
	}
	p := gcp.New(client, "my-project")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
	if snaps[0].Attributes["name"] != "web-1" {
		t.Errorf("expected name=web-1, got %s", snaps[0].Attributes["name"])
	}
	if snaps[0].Attributes["label:env"] != "prod" {
		t.Errorf("expected label:env=prod, got %s", snaps[0].Attributes["label:env"])
	}
}

func TestCollect_SetsProviderType(t *testing.T) {
	client := &fakeClient{
		instances: []gcp.Instance{
			{ID: "i-3", Name: "db-1", Zone: "europe-west1-b", Status: "RUNNING"},
		},
	}
	p := gcp.New(client, "my-project")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps[0].ProviderType != "gcp" {
		t.Errorf("expected provider type gcp, got %s", snaps[0].ProviderType)
	}
}

func TestCollect_ClientError(t *testing.T) {
	client := &fakeClient{err: errors.New("permission denied")}
	p := gcp.New(client, "my-project")
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_EmptyList(t *testing.T) {
	client := &fakeClient{instances: []gcp.Instance{}}
	p := gcp.New(client, "my-project")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestRegister_Buildable(t *testing.T) {
	_ = gcp.New // ensure package is loaded and init() ran
	registered := provider.Registered()
	for _, name := range registered {
		if name == "gcp" {
			return
		}
	}
	t.Error("gcp provider not found in registry")
}
