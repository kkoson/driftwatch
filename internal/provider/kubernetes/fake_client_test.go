package kubernetes

import (
	"context"
	"errors"
)

type fakeClient struct {
	resources []Resource
}

func newFakeClient(resources []Resource) *fakeClient {
	return &fakeClient{resources: resources}
}

func (f *fakeClient) ListDeployments(_ context.Context, _ string) ([]Resource, error) {
	return f.resources, nil
}

type errClient struct{ msg string }

func newErrClient(msg string) *errClient { return &errClient{msg: msg} }

func (e *errClient) ListDeployments(_ context.Context, _ string) ([]Resource, error) {
	return nil, errors.New(e.msg)
}
