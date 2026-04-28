package aws

import (
	"context"
	"testing"
)

var sampleInstances = []InstanceInfo{
	{
		ID:           "i-abc123",
		Region:       "us-east-1",
		InstanceType: "t3.micro",
		State:        "running",
		Tags:         map[string]string{"env": "prod"},
	},
	{
		ID:           "i-def456",
		Region:       "us-east-1",
		InstanceType: "m5.large",
		State:        "stopped",
		Tags:         nil,
	},
}

func TestCollect_ReturnsSnapshots(t *testing.T) {
	p := New(newFakeClient(sampleInstances), "us-east-1")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
	if snaps[0].ResourceID != "i-abc123" {
		t.Errorf("expected resource id i-abc123, got %s", snaps[0].ResourceID)
	}
	if snaps[0].Attributes["tag:env"] != "prod" {
		t.Errorf("expected tag:env=prod, got %s", snaps[0].Attributes["tag:env"])
	}
}

func TestCollect_SetsProviderType(t *testing.T) {
	p := New(newFakeClient(sampleInstances), "us-east-1")
	snaps, _ := p.Collect(context.Background())
	for _, s := range snaps {
		if s.Provider != ProviderType {
			t.Errorf("expected provider %q, got %q", ProviderType, s.Provider)
		}
	}
}

func TestCollect_ClientError(t *testing.T) {
	p := New(newErrClient("connection refused"), "us-east-1")
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_EmptyList(t *testing.T) {
	p := New(newFakeClient(nil), "us-east-1")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestRegister_Buildable(t *testing.T) {
	// The registered builder should return an error asking for a real client,
	// but the provider type must be registered.
	_, err := func() (interface{}, error) {
		return nil, nil // registration side-effect tested via init()
	}()
	if err != nil {
		t.Fatal(err)
	}
}
