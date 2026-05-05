package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookSink sends alerts as JSON POST requests to a configured URL.
type WebhookSink struct {
	url    string
	client *http.Client
}

// webhookPayload is the JSON body sent to the webhook endpoint.
type webhookPayload struct {
	ResourceID   string            `json:"resource_id"`
	ProviderType string            `json:"provider_type"`
	ChangedKeys  []string          `json:"changed_keys"`
	Attributes   map[string]string `json:"attributes"`
	DetectedAt   time.Time         `json:"detected_at"`
}

// NewWebhookSink creates a WebhookSink that posts alerts to url.
// A zero timeout defaults to 10 seconds.
func NewWebhookSink(url string, timeout time.Duration) *WebhookSink {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookSink{
		url: url,
		client: &http.Client{Timeout: timeout},
	}
}

// Send marshals the alert and posts it to the configured webhook URL.
func (w *WebhookSink) Send(ctx context.Context, a Alert) error {
	payload := webhookPayload{
		ResourceID:   a.ResourceID,
		ProviderType: a.ProviderType,
		ChangedKeys:  a.ChangedKeys,
		Attributes:   a.Current,
		DetectedAt:   a.DetectedAt,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
