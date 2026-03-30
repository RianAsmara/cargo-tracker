package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/RianAsmara/cargo-tracker/internal/domain"
)

type WebhookDispatcher interface {
	Dispatch(ctx context.Context, tenant *domain.Tenant, event *domain.TrackingEvent) error
}

type httpWebhookDispatcher struct {
	client *http.Client
}

func NewWebhookDispatcher() WebhookDispatcher {
	return &httpWebhookDispatcher{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type WebhookPayload struct {
	Event   string                `json:"event"`
	Tenant  string                `json:"tenant"`
	Details *domain.TrackingEvent `json:"details"`
}

func (d *httpWebhookDispatcher) Dispatch(ctx context.Context, tenant *domain.Tenant, event *domain.TrackingEvent) error {
	if tenant.WebhookURL == "" {
		return nil // Nothing to dispatch
	}

	payload := WebhookPayload{
		Event:   "SHIPMENT_STATUS_UPDATE",
		Tenant:  tenant.Name,
		Details: event,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tenant.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Just a mock fire for n8n or generic webhook logic, ignore payload response body.
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
