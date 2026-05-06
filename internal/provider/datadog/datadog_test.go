package datadog

import (
	"context"
	"testing"

	"github.com/youorg/driftwatch/internal/provider"
)

func TestCollect_ReturnsSnapshots(t *testing.T) {
	monitors := []Monitor{
		{ID: 1, Name: "CPU High", Type: "metric alert", Query: "avg:system.cpu.user{*} > 90", Message: "CPU is high"},
		{ID: 2, Name: "Disk Low", Type: "metric alert", Query: "avg:system.disk.in_use{*} > 0.9", Message: "Disk is low"},
	}
	p := New(newFakeClient(monitors))
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestCollect_SetsProviderType(t *testing.T) {
	monitors := []Monitor{{ID: 42, Name: "Latency", Type: "metric alert", Query: "q", Message: "m"}}
	p := New(newFakeClient(monitors))
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps[0].ProviderType != providerType {
		t.Errorf("expected provider type %q, got %q", providerType, snaps[0].ProviderType)
	}
}

func TestCollect_ClientError(t *testing.T) {
	p := New(newErrClient("api unavailable"))
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_EmptyList(t *testing.T) {
	p := New(newFakeClient(nil))
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestRegister_Buildable(t *testing.T) {
	registered := provider.Registered()
	for _, name := range registered {
		if name == providerType {
			return
		}
	}
	t.Errorf("provider %q not found in registered list: %v", providerType, registered)
}
