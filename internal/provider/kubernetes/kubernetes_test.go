package kubernetes

import (
	"context"
	"testing"
)

func TestCollect_ReturnsSnapshots(t *testing.T) {
	resources := []Resource{
		{Name: "api", Namespace: "prod", Replicas: 3, Image: "api:v1", Labels: map[string]string{"app": "api"}},
		{Name: "worker", Namespace: "prod", Replicas: 1, Image: "worker:v2", Labels: nil},
	}
	p := New(newFakeClient(resources), "prod")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestCollect_SetsProviderType(t *testing.T) {
	resources := []Resource{
		{Name: "svc", Namespace: "default", Replicas: 1, Image: "svc:latest"},
	}
	p := New(newFakeClient(resources), "default")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps[0].ProviderType != "kubernetes" {
		t.Errorf("expected provider type 'kubernetes', got %q", snaps[0].ProviderType)
	}
}

func TestCollect_ClientError(t *testing.T) {
	p := New(newErrClient("connection refused"), "default")
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_EmptyList(t *testing.T) {
	p := New(newFakeClient(nil), "default")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestCollect_SnapshotID(t *testing.T) {
	resources := []Resource{
		{Name: "frontend", Namespace: "staging", Replicas: 2, Image: "fe:v3"},
	}
	p := New(newFakeClient(resources), "staging")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "k8s/staging/frontend"
	if snaps[0].ID != want {
		t.Errorf("expected snapshot ID %q, got %q", want, snaps[0].ID)
	}
}
