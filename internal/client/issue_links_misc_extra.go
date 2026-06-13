package client

import "encoding/json"

// RemoteLink is a remote (web) link on an issue. The visible title/url/summary
// live under the "object" sub-document.
type RemoteLink struct {
	ID           int    `json:"id"`
	GlobalID     string `json:"globalId,omitempty"`
	Relationship string `json:"relationship,omitempty"`
	Object       struct {
		URL     string `json:"url,omitempty"`
		Title   string `json:"title,omitempty"`
		Summary string `json:"summary,omitempty"`
	} `json:"object"`
}

// RemoteLinks returns the remote (web) links attached to an issue.
func (c *Client) RemoteLinks(key string) ([]RemoteLink, error) {
	var out []RemoteLink
	if err := c.GetJSON(c.api("issue/%s/remotelink", key), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AddRemoteLink creates a remote (web) link on an issue. body is the raw request
// (typically {"object":{"url","title","summary"}}). Returns {id,self}.
func (c *Client) AddRemoteLink(key string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.PostJSON(c.api("issue/%s/remotelink", key), nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteRemoteLink removes a remote link from an issue by its id.
func (c *Client) DeleteRemoteLink(key, id string) error {
	return c.Delete(c.api("issue/%s/remotelink/%s", key, id), nil)
}

// ListIssueProperties returns the keys of all properties stored on an issue.
func (c *Client) ListIssueProperties(key string) ([]string, error) {
	var out struct {
		Keys []struct {
			Key  string `json:"key"`
			Self string `json:"self"`
		} `json:"keys"`
	}
	if err := c.GetJSON(c.api("issue/%s/properties", key), nil, &out); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(out.Keys))
	for _, k := range out.Keys {
		keys = append(keys, k.Key)
	}
	return keys, nil
}

// GetIssueProperty returns a single issue property value (the raw "value").
func (c *Client) GetIssueProperty(key, propKey string) (json.RawMessage, error) {
	var out struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}
	if err := c.GetJSON(c.api("issue/%s/properties/%s", key, propKey), nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// SetIssueProperty stores a property value on an issue. value is the raw JSON
// document to persist. Returns no body on success.
func (c *Client) SetIssueProperty(key, propKey string, value json.RawMessage) error {
	return c.PutJSON(c.api("issue/%s/properties/%s", key, propKey), nil, value, nil)
}

// DeleteIssueProperty removes a property from an issue.
func (c *Client) DeleteIssueProperty(key, propKey string) error {
	return c.Delete(c.api("issue/%s/properties/%s", key, propKey), nil)
}

// Notify sends an email notification about an issue to a chosen audience. body
// carries {subject, textBody/htmlBody, to:{reporter,assignee,watchers,voters,
// users:[{accountId}]}}. Returns no body on success (204).
func (c *Client) Notify(key string, body map[string]any) error {
	return c.PostJSON(c.api("issue/%s/notify", key), nil, body, nil)
}
