package datadog

import (
	"context"
	"errors"
)

type fakeClient struct {
	monitors []Monitor
}

func newFakeClient(monitors []Monitor) Client {
	return &fakeClient{monitors: monitors}
}

func (f *fakeClient) ListMonitors(_ context.Context) ([]Monitor, error) {
	return f.monitors, nil
}

type errClient struct{ err error }

func newErrClient(msg string) Client {
	return &errClient{err: errors.New(msg)}
}

func (e *errClient) ListMonitors(_ context.Context) ([]Monitor, error) {
	return nil, e.err
}

// newHTTPClient is a stub so the package compiles without a real HTTP client.
func newHTTPClient(_, _ string) Client {
	return newFakeClient(nil)
}
