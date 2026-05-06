package alert

import (
	"errors"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

func newEmailAlert() drift.Alert {
	return drift.Alert{
		ResourceID:   "i-0abc123",
		ProviderType: "aws",
		ChangedKeys:  []string{"instance_type", "ami"},
		DetectedAt:   time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestNewEmailSink_EmptySMTPAddr(t *testing.T) {
	_, err := NewEmailSink(EmailConfig{From: "a@b.com", To: []string{"c@d.com"}})
	if err == nil || !strings.Contains(err.Error(), "SMTPAddr") {
		t.Fatalf("expected SMTPAddr error, got %v", err)
	}
}

func TestNewEmailSink_EmptyFrom(t *testing.T) {
	_, err := NewEmailSink(EmailConfig{SMTPAddr: "localhost:25", To: []string{"c@d.com"}})
	if err == nil || !strings.Contains(err.Error(), "From") {
		t.Fatalf("expected From error, got %v", err)
	}
}

func TestNewEmailSink_EmptyTo(t *testing.T) {
	_, err := NewEmailSink(EmailConfig{SMTPAddr: "localhost:25", From: "a@b.com"})
	if err == nil || !strings.Contains(err.Error(), "To") {
		t.Fatalf("expected To error, got %v", err)
	}
}

func TestEmailSink_Send_Success(t *testing.T) {
	sink, err := NewEmailSink(EmailConfig{
		SMTPAddr: "localhost:25",
		From:     "drift@example.com",
		To:       []string{"ops@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var capturedMsg []byte
	sink.smtpSend = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		capturedMsg = msg
		return nil
	}

	if err := sink.Send(newEmailAlert()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	body := string(capturedMsg)
	if !strings.Contains(body, "i-0abc123") {
		t.Errorf("expected resource ID in message body, got: %s", body)
	}
	if !strings.Contains(body, "[driftwatch]") {
		t.Errorf("expected subject prefix in message body, got: %s", body)
	}
}

func TestEmailSink_Send_SMTPError(t *testing.T) {
	sink, err := NewEmailSink(EmailConfig{
		SMTPAddr: "localhost:25",
		From:     "drift@example.com",
		To:       []string{"ops@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sink.smtpSend = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("connection refused")
	}

	err = sink.Send(newEmailAlert())
	if err == nil || !strings.Contains(err.Error(), "send failed") {
		t.Fatalf("expected send failed error, got %v", err)
	}
}
