package consul

import (
	"context"
	"errors"
)

type fakeClient struct {
	services map[string]map[string]string
}

func newFakeClient(services map[string]map[string]string) *fakeClient {
	return &fakeClient{services: services}
}

func (f *fakeClient) Services(_ context.Context) (map[string]map[string]string, error) {
	return f.services, nil
}

type errClient struct{ msg string }

func newErrClient(msg string) *errClient { return &errClient{msg: msg} }

func (e *errClient) Services(_ context.Context) (map[string]map[string]string, error) {
	return nil, errors.New(e.msg)
}
