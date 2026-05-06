package alert

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newSlackAlert() Alert {
	return Alert{
		ResourceID:   "i-0abc123",
		ProviderType: "aws",
		ChangedKeys:  []string{"instance_type", "tags"},
		Severity:     "high",
		DetectedAt:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
	}
}

func TestNewSlackSink_EmptyURL(t *testing.T) {
	_, err := NewSlackSink("")
	if err == nil {
		t.Fatal("expected error for empty webhook URL, got nil")
	}
}

func TestNewSlackSink_DefaultTimeout(t *testing.T) {
	s, err := NewSlackSink("https://hooks.slack.com/services/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != defaultSlackTimeout {
		t.Errorf("expected timeout %v, got %v", defaultSlackTimeout, s.client.Timeout)
	}
}

func TestSlackSink_Send_Success(t *testing.T) {
	var gotContentType string
	var gotBody bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotBody = r.ContentLength > 0
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s, err := NewSlackSink(server.URL, WithSlackHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Send(newSlackAlert()); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", gotContentType)
	}
	if !gotBody {
		t.Error("expected non-empty request body")
	}
}

func TestSlackSink_Send_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s, err := NewSlackSink(server.URL, WithSlackHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Send(newSlackAlert()); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestSlackSink_Send_InvalidURL(t *testing.T) {
	s, err := NewSlackSink("http://127.0.0.1:0/no-server")
	if err != nil {
		t.Fatalf("unexpected error creating sink: %v", err)
	}
	if err := s.Send(newSlackAlert()); err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}
