package azure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/driftwatch/internal/provider"
	"github.com/your-org/driftwatch/internal/provider/azure"
)

// fakeClient implements azure.Client for testing.
type fakeClient struct {
	vms []azure.VM
	err error
}

func (f *fakeClient) ListVMs(_ context.Context) ([]azure.VM, error) {
	return f.vms, f.err
}

func newFakeClient(vms []azure.VM) *fakeClient { return &fakeClient{vms: vms} }
func newErrClient(err error) *fakeClient       { return &fakeClient{err: err} }

func TestCollect_ReturnsSnapshots(t *testing.T) {
	vms := []azure.VM{
		{ID: "vm-1", Name: "web", Location: "eastus", Size: "Standard_B1s", Tags: map[string]string{"env": "prod"}},
		{ID: "vm-2", Name: "db", Location: "westus", Size: "Standard_D2s_v3"},
	}
	p := azure.New(newFakeClient(vms))
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("want 2 snapshots, got %d", len(snaps))
	}
	if snaps[0].ResourceID != "vm-1" {
		t.Errorf("want resource id vm-1, got %s", snaps[0].ResourceID)
	}
}

func TestCollect_SetsProviderType(t *testing.T) {
	vms := []azure.VM{{ID: "vm-3", Name: "svc", Location: "northeurope", Size: "Standard_A1"}}
	p := azure.New(newFakeClient(vms))
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps[0].ProviderType != "azure" {
		t.Errorf("want provider type azure, got %s", snaps[0].ProviderType)
	}
}

func TestCollect_ClientError(t *testing.T) {
	p := azure.New(newErrClient(errors.New("auth failure")))
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_EmptyList(t *testing.T) {
	p := azure.New(newFakeClient(nil))
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("want 0 snapshots, got %d", len(snaps))
	}
}

func TestRegister_Buildable(t *testing.T) {
	names := provider.Registered()
	for _, n := range names {
		if n == "azure" {
			return
		}
	}
	t.Error("azure provider not found in registry")
}
