package mock_test

import (
	"context"
	"testing"

	"github.com/yourorg/driftwatch/internal/provider"
	"github.com/yourorg/driftwatch/internal/provider/mock"
)

func TestNew_EmptySeed(t *testing.T) {
	mp, err := mock.New(map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snaps, err := mp.Collect(context.Background())
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestNew_InvalidSeed(t *testing.T) {
	_, err := mock.New(map[string]any{"resources": "bad"})
	if err == nil {
		t.Fatal("expected error for invalid resources seed")
	}
}

func TestCollect_ReturnsSnapshots(t *testing.T) {
	mp, _ := mock.New(map[string]any{})
	mp.SetResource("res-1", map[string]any{"region": "us-east-1", "size": "t3.micro"})
	mp.SetResource("res-2", map[string]any{"region": "eu-west-1", "size": "t3.small"})

	snaps, err := mp.Collect(context.Background())
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(snaps) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestCollect_AfterDelete(t *testing.T) {
	mp, _ := mock.New(map[string]any{})
	mp.SetResource("res-1", map[string]any{"size": "t3.micro"})
	mp.DeleteResource("res-1")

	snaps, err := mp.Collect(context.Background())
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots after delete, got %d", len(snaps))
	}
}

func TestRegister_Buildable(t *testing.T) {
	// Ensure the init() registration wired the provider correctly.
	p, err := provider.Build(mock.ProviderType, map[string]any{})
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}
