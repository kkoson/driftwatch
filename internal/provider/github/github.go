// Package github provides a driftwatch provider that collects configuration
// snapshots from GitHub repository and organization settings.
package github

import (
	"context"
	"fmt"

	"github.com/youorg/driftwatch/internal/provider"
	"github.com/youorg/driftwatch/internal/snapshot"
)

// Client is the interface for fetching GitHub repository metadata.
type Client interface {
	ListRepos(ctx context.Context, org string) ([]Repo, error)
}

// Repo represents a minimal GitHub repository configuration.
type Repo struct {
	Name          string
	Private       bool
	DefaultBranch string
	Archived      bool
	HasWiki       bool
	HasIssues     bool
}

// githubProvider collects snapshots from a GitHub organisation.
type githubProvider struct {
	client Client
	org    string
}

// New creates a new GitHub provider for the given organisation.
func New(client Client, org string) (provider.Provider, error) {
	if org == "" {
		return nil, fmt.Errorf("github: org must not be empty")
	}
	if client == nil {
		return nil, fmt.Errorf("github: client must not be nil")
	}
	return &githubProvider{client: client, org: org}, nil
}

// Collect fetches all repositories for the configured org and returns snapshots.
func (p *githubProvider) Collect(ctx context.Context) ([]*snapshot.Snapshot, error) {
	repos, err := p.client.ListRepos(ctx, p.org)
	if err != nil {
		return nil, fmt.Errorf("github: list repos: %w", err)
	}

	snaps := make([]*snapshot.Snapshot, 0, len(repos))
	for _, r := range repos {
		attrs := map[string]string{
			"private":        fmt.Sprintf("%t", r.Private),
			"default_branch": r.DefaultBranch,
			"archived":       fmt.Sprintf("%t", r.Archived),
			"has_wiki":       fmt.Sprintf("%t", r.HasWiki),
			"has_issues":     fmt.Sprintf("%t", r.HasIssues),
		}
		id := fmt.Sprintf("github/%s/%s", p.org, r.Name)
		snaps = append(snaps, snapshot.New(id, "github", attrs))
	}
	return snaps, nil
}

func init() {
	provider.Register("github", func(cfg map[string]string) (provider.Provider, error) {
		org, ok := cfg["org"]
		if !ok || org == "" {
			return nil, fmt.Errorf("github: missing required config key 'org'")
		}
		// Real HTTP client construction would happen here.
		return nil, fmt.Errorf("github: real HTTP client not wired in init; use New() directly")
	})
}
