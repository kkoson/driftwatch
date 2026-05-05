package alert_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/driftwatch/internal/alert"
)

func newWebhookAlert() alert.Alert {
	return makeAlert("aws", "i-abc123", []string{"instance_type"})
}

func TestWebhookSink_Send_Success(t *testing.T) {
	var received map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sink := alert.NewWebhookSink(srv.URL, 5*time.Second)
	a := newWebhookAlert()

	if err := sink.Send(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["resource_id"] != "i-abc123" {
		t.Errorf("resource_id mismatch: %v", received["resource_id"])
	}
	if received["provider_type"] != "aws" {
		t.Errorf("provider_type mismatch: %v", received["provider_type"])
	}
	keys, ok := received["changed_keys"].([]interface{})
	if !ok || len(keys) == 0 || keys[0] != "instance_type" {
		t.Errorf("changed_keys mismatch: %v", received["changed_keys"])
	}
}

func TestWebhookSink_Send_NonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	sink := alert.NewWebhookSink(srv.URL, 5*time.Second)
	err := sink.Send(context.Background(), newWebhookAlert())
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestWebhookSink_Send_InvalidURL(t *testing.T) {
	sink := alert.NewWebhookSink("http://127.0.0.1:0", 500*time.Millisecond)
	err := sink.Send(context.Background(), newWebhookAlert())
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

func TestWebhookSink_DefaultTimeout(t *testing.T) {
	// Passing zero timeout should not panic and should use the default.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	sink := alert.NewWebhookSink(srv.URL, 0)
	if err := sink.Send(context.Background(), newWebhookAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
