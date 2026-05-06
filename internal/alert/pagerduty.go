package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultPagerDutyURL     = "https://events.pagerduty.com/v2/enqueue"
	defaultPagerDutyTimeout = 10 * time.Second
)

// pagerDutyPayload is the Events API v2 envelope.
type pagerDutyPayload struct {
	RoutingKey  string         `json:"routing_key"`
	EventAction string         `json:"event_action"`
	Payload     pdInnerPayload `json:"payload"`
}

type pdInnerPayload struct {
	Summary  string            `json:"summary"`
	Source   string            `json:"source"`
	Severity string            `json:"severity"`
	Custom   map[string]string `json:"custom_details,omitempty"`
}

// PagerDutySink sends alerts to PagerDuty via the Events API v2.
type PagerDutySink struct {
	routingKey string
	endpoint   string
	client     *http.Client
}

// NewPagerDutySink constructs a PagerDutySink. routingKey must be non-empty.
func NewPagerDutySink(routingKey, endpoint string) (*PagerDutySink, error) {
	if routingKey == "" {
		return nil, fmt.Errorf("pagerduty: routing key must not be empty")
	}
	if endpoint == "" {
		endpoint = defaultPagerDutyURL
	}
	return &PagerDutySink{
		routingKey: routingKey,
		endpoint:   endpoint,
		client:     &http.Client{Timeout: defaultPagerDutyTimeout},
	}, nil
}

// Send delivers a drift alert to PagerDuty.
func (p *PagerDutySink) Send(ctx context.Context, a Alert) error {
	details := make(map[string]string, len(a.ChangedKeys))
	for _, k := range a.ChangedKeys {
		details[k] = "changed"
	}

	body := pagerDutyPayload{
		RoutingKey:  p.routingKey,
		EventAction: "trigger",
		Payload: pdInnerPayload{
			Summary:  a.String(),
			Source:   a.ResourceID,
			Severity: "warning",
			Custom:   details,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("pagerduty: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
