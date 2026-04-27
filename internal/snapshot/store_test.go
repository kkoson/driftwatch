package snapshot

import (
	"testing"
)

func makeSnap(id string) *Snapshot {
	s, _ := New(id, "aws", "ec2_instance", baseAttrs(), nil)
	return s
}

func TestStore_PutAndGet(t *testing.T) {
	store := NewStore()
	snap := makeSnap("i-abc")
	store.Put(snap)

	got, err := store.Get("i-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ResourceID != "i-abc" {
		t.Errorf("expected resource id i-abc, got %s", got.ResourceID)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	store := NewStore()
	_, err := store.Get("missing")
	if err == nil {
		t.Error("expected error for missing resource")
	}
}

func TestStore_Delete(t *testing.T) {
	store := NewStore()
	store.Put(makeSnap("i-del"))
	store.Delete("i-del")
	_, err := store.Get("i-del")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestStore_All(t *testing.T) {
	store := NewStore()
	store.Put(makeSnap("i-1"))
	store.Put(makeSnap("i-2"))
	store.Put(makeSnap("i-3"))

	all := store.All()
	if len(all) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(all))
	}
}

func TestStore_Len(t *testing.T) {
	store := NewStore()
	if store.Len() != 0 {
		t.Error("expected empty store")
	}
	store.Put(makeSnap("i-x"))
	if store.Len() != 1 {
		t.Errorf("expected len 1, got %d", store.Len())
	}
}
