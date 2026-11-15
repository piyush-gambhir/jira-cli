package client

import (
	"net/http"
	"net/url"
)

type Webhook struct {
	ID             int      `json:"id"`
	JqlFilter      string   `json:"jqlFilter,omitempty"`
	Events         []string `json:"events,omitempty"`
	ExpirationDate string   `json:"expirationDate,omitempty"`
	FieldIDsFilter []string `json:"fieldIdsFilter,omitempty"`
}

type WebhookRegistration struct {
	JqlFilter      string   `json:"jqlFilter"`
	Events         []string `json:"events"`
	FieldIDsFilter []string `json:"fieldIdsFilter,omitempty"`
}

type WebhookRegistrationResult struct {
	CreatedWebhookID int      `json:"createdWebhookId,omitempty"`
	Errors           []string `json:"errors,omitempty"`
}

func errWebhooksCloudOnly() error {
	return errCloudOnly("webhook management", "on Server/Data Center a Jira administrator manages webhooks under Administration > System > WebHooks")
}

func (c *Client) ListWebhooks(limit int) ([]Webhook, error) {
	if c.IsServer() {
		return nil, errWebhooksCloudOnly()
	}
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

func (c *Client) RegisterWebhooks(callbackURL string, webhooks []WebhookRegistration) ([]WebhookRegistrationResult, error) {
	if c.IsServer() {
		return nil, errWebhooksCloudOnly()
	}
	body := map[string]any{"url": callbackURL, "webhooks": webhooks}
	var out struct {
		WebhookRegistrationResult []WebhookRegistrationResult `json:"webhookRegistrationResult"`
	}
	if err := c.PostJSON(c.api("webhook"), nil, body, &out); err != nil {
		return nil, err
	}
	return out.WebhookRegistrationResult, nil
}

func (c *Client) DeleteWebhooks(ids []int) error {
	if c.IsServer() {
		return errWebhooksCloudOnly()
	}
	body := map[string]any{"webhookIds": ids}
	return c.doJSON(http.MethodDelete, c.api("webhook"), nil, body, nil, true)
}

func (c *Client) RefreshWebhooks(ids []int) (string, error) {
	if c.IsServer() {
		return "", errWebhooksCloudOnly()
	}
	body := map[string]any{"webhookIds": ids}
	var out struct {
		ExpirationDate string `json:"expirationDate"`
	}
	if err := c.PutJSON(c.api("webhook/refresh"), nil, body, &out); err != nil {
		return "", err
	}
	return out.ExpirationDate, nil
}
