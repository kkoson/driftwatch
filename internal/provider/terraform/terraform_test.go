package terraform

import (
	"context"
	"errors"
	"testing"

	"github.com/example/driftwatch/internal/provider"
)

// stubClient implements StateClient for tests.
type stubClient struct {
	data []byte
	err  error
}

func (s *stubClient) FetchState(_ context.Context) ([]byte, error) {
	return s.data, s.err
}

const sampleState = `{
  "resources": [
    {
      "type": "aws_instance",
      "name": "web",
      "instances": [
        {"attributes": {"id": "i-abc123", "instance_type": "t3.micro"}}
      ]
    },
    {
      "type": "aws_s3_bucket",
      "name": "assets",
      "instances": [
        {"attributes": {"id": "my-assets", "region": "us-east-1"}}
      ]
    }
  ]
}`

func TestCollect_ReturnsSnapshots(t *testing.T) {
	p := New(&stubClient{data: []byte(sampleState)})
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestCollect_SetsProviderType(t *testing.T) {
	p := New(&stubClient{data: []byte(sampleState)})
	snaps, _ := p.Collect(context.Background())
	for _, s := range snaps {
		if s.ProviderType != providerType {
			t.Errorf("expected provider %q, got %q", providerType, s.ProviderType)
		}
	}
}

func TestCollect_SetsResourceAttributes(t *testing.T) {
	p := New(&stubClient{data: []byte(sampleState)})
	snaps, _ := p.Collect(context.Background())
	attrs := snaps[0].Attributes
	if attrs["tf_type"] != "aws_instance" {
		t.Errorf("expected tf_type=aws_instance, got %q", attrs["tf_type"])
	}
	if attrs["tf_name"] != "web" {
		t.Errorf("expected tf_name=web, got %q", attrs["tf_name"])
	}
}

func TestCollect_ClientError(t *testing.T) {
	p := New(&stubClient{err: errors.New("network failure")})
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_EmptyState(t *testing.T) {
	p := New(&stubClient{data: []byte(`{"resources":[]}`)}) 
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestCollect_InvalidJSON(t *testing.T) {
	p := New(&stubClient{data: []byte(`not-json`)})
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestRegister_Buildable(t *testing.T) {
	_, err := provider.Build(providerType, map[string]string{"state_file": "/tmp/terraform.tfstate"})
	if err != nil {
		t.Fatalf("expected provider to build, got: %v", err)
	}
}

func TestRegister_MissingConfig(t *testing.T) {
	_, err := provider.Build(providerType, map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}
