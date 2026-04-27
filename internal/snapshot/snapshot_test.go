package snapshot

import (
	"testing"
)

func baseAttrs() map[string]any {
	return map[string]any{
		"region":        "us-east-1",
		"instance_type": "t3.micro",
		"tags": map[string]string{
			"env": "prod",
		},
	}
}

func TestNew_SetsChecksum(t *testing.T) {
	snap, err := New("i-123", "aws", "ec2_instance", baseAttrs(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestDiffersFrom_SameAttributes(t *testing.T) {
	a, _ := New("i-123", "aws", "ec2_instance", baseAttrs(), nil)
	b, _ := New("i-123", "aws", "ec2_instance", baseAttrs(), nil)
	if a.DiffersFrom(b) {
		t.Error("expected snapshots with identical attributes to not differ")
	}
}

func TestDiffersFrom_ChangedAttribute(t *testing.T) {
	a, _ := New("i-123", "aws", "ec2_instance", baseAttrs(), nil)
	modified := baseAttrs()
	modified["instance_type"] = "t3.large"
	b, _ := New("i-123", "aws", "ec2_instance", modified, nil)
	if !a.DiffersFrom(b) {
		t.Error("expected snapshots with different attributes to differ")
	}
}

func TestChangedKeys(t *testing.T) {
	a, _ := New("i-123", "aws", "ec2_instance", baseAttrs(), nil)
	modified := baseAttrs()
	modified["instance_type"] = "t3.large"
	modified["new_key"] = "value"
	b, _ := New("i-123", "aws", "ec2_instance", modified, nil)

	keys := a.ChangedKeys(b)
	found := map[string]bool{}
	for _, k := range keys {
		found[k] = true
	}
	if !found["instance_type"] {
		t.Error("expected instance_type to be in changed keys")
	}
	if !found["new_key"] {
		t.Error("expected new_key to be in changed keys")
	}
}
