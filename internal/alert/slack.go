package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultSlackTimeout = 10 * time.Second

// slackPayload is the JSON body sent to a Slack incoming webhook.
type slackPayload struct {
	Text string `json:"text"`
}

// SlackSink delivers alerts to a Slack channel via an incoming webhook URL.
type SlackSink struct {
	webhookURL string
	client     *http.Client
}

// NewSlackSink creates a SlackSink that posts to the given Slack webhook URL.
// Returns an error if webhookURL is empty.
func NewSlackSink(webhookURL string, opts ...func(*SlackSink)) (*SlackSink, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack: webhook URL must not be empty")
	}
	s := &SlackSink{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: defaultSlackTimeout},
	}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// WithSlackHTTPClient overrides the HTTP client used by SlackSink.
func WithSlackHTTPClient(c *http.Client) func(*SlackSink) {
	return func(s *SlackSink) { s.client = c }
}

// Send formats the alert and posts it to the Slack webhook.
func (s *SlackSink) Send(a Alert) error {
	payload := slackPayload{
		Text: fmt.Sprintf("*[%s] Drift detected — %s/%s*\nChanged keys: %v\nDetected at: %s",
			a.Severity, a.ProviderType, a.ResourceID,
			a.ChangedKeys, a.DetectedAt.Format(time.RFC3339)),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}
	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
