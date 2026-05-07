package consul

import (
	"context"
	"testing"

	"github.com/your-org/driftwatch/internal/provider"
)

func TestCollect_ReturnsSnapshots(t *testing.T) {
	client := newFakeClient(map[string]map[string]string{
		"web": {"version": "1.2.3", "region": "us-east-1"},
		"db":  {"version": "5.7"},
	})
	p := New(client, "agent-1")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestCollect_SetsProviderType(t *testing.T) {
	client := newFakeClient(map[string]map[string]string{
		"api": {"port": "8080"},
	})
	p := New(client, "agent-2")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps[0].ProviderType != "consul" {
		t.Errorf("expected provider type %q, got %q", "consul", snaps[0].ProviderType)
	}
}

func TestCollect_ClientError(t *testing.T) {
	p := New(newErrClient("connection refused"), "agent-3")
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_EmptyList(t *testing.T) {
	p := New(newFakeClient(map[string]map[string]string{}), "agent-4")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestCollect_AgentInAttributes(t *testing.T) {
	client := newFakeClient(map[string]map[string]string{
		"cache": {},
	})
	p := New(client, "my-agent")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps[0].Attributes["agent"] != "my-agent" {
		t.Errorf("expected agent attribute %q, got %q", "my-agent", snaps[0].Attributes["agent"])
	}
}

func TestRegister_Buildable(t *testing.T) {
	names := provider.Registered()
	for _, n := range names {
		if n == "consul" {
			return
		}
	}
	t.Error("consul provider not registered")
}
