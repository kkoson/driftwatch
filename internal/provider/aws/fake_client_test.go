package aws

import (
	"context"
	"errors"
)

// fakeClient implements Client for unit tests.
type fakeClient struct {
	instances []InstanceInfo
	err       error
}

func (f *fakeClient) ListInstances(_ context.Context) ([]InstanceInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.instances, nil
}

func newFakeClient(instances []InstanceInfo) *fakeClient {
	return &fakeClient{instances: instances}
}

func newErrClient(msg string) *fakeClient {
	return &fakeClient{err: errors.New(msg)}
}
