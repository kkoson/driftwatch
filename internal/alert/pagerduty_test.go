package alert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newPDAlert() Alert {
	return Alert{
		ResourceID:   "i-0abc123",
		ProviderType: "aws",
		ChangedKeys:  []string{"instance_type", "tags"},
		DetectedAt:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
	}
}

func TestNewPagerDutySink_EmptyRoutingKey(t *testing.T) {
	_, err := NewPagerDutySink("", "")
	if err == nil {
		t.Fatal("expected error for empty routing key")
	}
}

func TestNewPagerDutySink_DefaultEndpoint(t *testing.T) {
	sink, err := NewPagerDutySink("test-key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sink.endpoint != defaultPagerDutyURL {
		t.Errorf("want %q, got %q", defaultPagerDutyURL, sink.endpoint)
	}
}

func TestPagerDutySink_Send_Success(t *testing.T) {
	var received pagerDutyPayload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("want POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("want application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	sink, err := NewPagerDutySink("rk-abc", srv.URL)
	if err != nil {
		t.Fatalf("NewPagerDutySink: %v", err)
	}

	a := newPDAlert()
	if err := sink.Send(context.Background(), a); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if received.RoutingKey != "rk-abc" {
		t.Errorf("routing key: want %q, got %q", "rk-abc", received.RoutingKey)
	}
	if received.EventAction != "trigger" {
		t.Errorf("event_action: want trigger, got %q", received.EventAction)
	}
	if received.Payload.Source != a.ResourceID {
		t.Errorf("source: want %q, got %q", a.ResourceID, received.Payload.Source)
	}
	if received.Payload.Severity != "warning" {
		t.Errorf("severity: want warning, got %q", received.Payload.Severity)
	}
	if _, ok := received.Payload.Custom["instance_type"]; !ok {
		t.Error("expected instance_type in custom_details")
	}
}

func TestPagerDutySink_Send_NonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	sink, _ := NewPagerDutySink("rk-abc", srv.URL)
	err := sink.Send(context.Background(), newPDAlert())
	if err == nil {
		t.Fatal("expected error on non-2xx status")
	}
}

func TestPagerDutySink_Send_InvalidURL(t *testing.T) {
	sink, _ := NewPagerDutySink("rk-abc", "http://127.0.0.1:0")
	err := sink.Send(context.Background(), newPDAlert())
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
