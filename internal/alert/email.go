package alert

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/example/driftwatch/internal/drift"
)

// EmailSink sends drift alerts via SMTP email.
type EmailSink struct {
	addr     string
	auth     smtp.Auth
	from     string
	to       []string
	smtpSend func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// EmailConfig holds configuration for the email sink.
type EmailConfig struct {
	SMTPAddr string
	Username string
	Password string
	From     string
	To       []string
}

// NewEmailSink creates an EmailSink from the provided config.
func NewEmailSink(cfg EmailConfig) (*EmailSink, error) {
	if cfg.SMTPAddr == "" {
		return nil, fmt.Errorf("alert/email: SMTPAddr must not be empty")
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("alert/email: From must not be empty")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("alert/email: To must contain at least one recipient")
	}

	host := cfg.SMTPAddr
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, host)
	}

	return &EmailSink{
		addr:     cfg.SMTPAddr,
		auth:     auth,
		from:     cfg.From,
		to:       cfg.To,
		smtpSend: smtp.SendMail,
	}, nil
}

// Send delivers the drift alert as an email message.
func (e *EmailSink) Send(a drift.Alert) error {
	subject := fmt.Sprintf("[driftwatch] Drift detected: %s (%s)", a.ResourceID, a.ProviderType)
	body := fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(e.to, ", "),
		e.from,
		subject,
		a.String(),
	)

	if err := e.smtpSend(e.addr, e.auth, e.from, e.to, []byte(body)); err != nil {
		return fmt.Errorf("alert/email: send failed: %w", err)
	}
	return nil
}
