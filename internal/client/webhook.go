package client

import (
	"net/http"
	"net/url"
)

// Webhook is a dynamic (OAuth-app) webhook registration.
type Webhook struct {
	ID             int      `json:"id"`
	JqlFilter      string   `json:"jqlFilter,omitempty"`
	Events         []string `json:"events,omitempty"`
	ExpirationDate string   `json:"expirationDate,omitempty"`
	FieldIDsFilter []string `json:"fieldIdsFilter,omitempty"`
}

// WebhookRegistration is one webhook entry in a register request.
type WebhookRegistration struct {
	JqlFilter      string   `json:"jqlFilter"`
	Events         []string `json:"events"`
	FieldIDsFilter []string `json:"fieldIdsFilter,omitempty"`
}

// WebhookRegistrationResult is one entry in the register response: either a
// created id or per-webhook validation errors.
type WebhookRegistrationResult struct {
	CreatedWebhookID int      `json:"createdWebhookId,omitempty"`
	Errors           []string `json:"errors,omitempty"`
}

// ListWebhooks returns the dynamic webhooks registered for the OAuth app.
func (c *Client) ListWebhooks(limit int) ([]Webhook, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []Webhook `json:"values"`
	}
	if err := c.GetJSON(c.api("webhook"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// RegisterWebhooks registers one or more webhooks for the given callback URL and
// returns the per-webhook results (created id or errors), in request order.
func (c *Client) RegisterWebhooks(callbackURL string, webhooks []WebhookRegistration) ([]WebhookRegistrationResult, error) {
	body := map[string]any{"url": callbackURL, "webhooks": webhooks}
	var out struct {
		WebhookRegistrationResult []WebhookRegistrationResult `json:"webhookRegistrationResult"`
	}
	if err := c.PostJSON(c.api("webhook"), nil, body, &out); err != nil {
		return nil, err
	}
	return out.WebhookRegistrationResult, nil
}

// DeleteWebhooks removes the webhooks with the given integer ids. The DELETE
// carries a JSON body, so it goes through doJSON directly (c.Delete has no body).
func (c *Client) DeleteWebhooks(ids []int) error {
	body := map[string]any{"webhookIds": ids}
	return c.doJSON(http.MethodDelete, c.api("webhook"), nil, body, nil, true)
}

// RefreshWebhooks extends the expiry of the given webhook ids and returns the
// new expiration date.
func (c *Client) RefreshWebhooks(ids []int) (string, error) {
	body := map[string]any{"webhookIds": ids}
	var out struct {
		ExpirationDate string `json:"expirationDate"`
	}
	if err := c.PutJSON(c.api("webhook/refresh"), nil, body, &out); err != nil {
		return "", err
	}
	return out.ExpirationDate, nil
}
