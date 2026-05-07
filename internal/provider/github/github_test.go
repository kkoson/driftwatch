package github

import (
	"context"
	"errors"
	"testing"
)

// fakeClient implements Client for testing.
type fakeClient struct {
	repos []Repo
	err   error
}

func (f *fakeClient) ListRepos(_ context.Context, _ string) ([]Repo, error) {
	return f.repos, f.err
}

func newFakeClient(repos []Repo) Client { return &fakeClient{repos: repos} }
func newErrClient() Client              { return &fakeClient{err: errors.New("api error")} }

func TestCollect_ReturnsSnapshots(t *testing.T) {
	client := newFakeClient([]Repo{
		{Name: "alpha", Private: true, DefaultBranch: "main", Archived: false, HasWiki: true, HasIssues: true},
		{Name: "beta", Private: false, DefaultBranch: "master", Archived: true, HasWiki: false, HasIssues: false},
	})
	p, err := New(client, "myorg")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestCollect_SetsProviderType(t *testing.T) {
	client := newFakeClient([]Repo{{Name: "repo1", DefaultBranch: "main"}})
	p, _ := New(client, "myorg")
	snaps, _ := p.Collect(context.Background())
	if snaps[0].ProviderType != "github" {
		t.Errorf("expected provider type 'github', got %q", snaps[0].ProviderType)
	}
}

func TestCollect_ClientError(t *testing.T) {
	p, _ := New(newErrClient(), "myorg")
	_, err := p.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_EmptyList(t *testing.T) {
	p, _ := New(newFakeClient(nil), "myorg")
	snaps, err := p.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestNew_EmptyOrg(t *testing.T) {
	_, err := New(newFakeClient(nil), "")
	if err == nil {
		t.Fatal("expected error for empty org")
	}
}

func TestNew_NilClient(t *testing.T) {
	_, err := New(nil, "myorg")
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}
