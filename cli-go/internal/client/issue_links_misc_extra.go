package client

import "encoding/json"

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

func (c *Client) RemoteLinks(key string) ([]RemoteLink, error) {
	var out []RemoteLink
	if err := c.GetJSON(c.api("issue/%s/remotelink", key), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AddRemoteLink(key string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.PostJSON(c.api("issue/%s/remotelink", key), nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DeleteRemoteLink(key, id string) error {
	return c.Delete(c.api("issue/%s/remotelink/%s", key, id), nil)
}

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

func (c *Client) SetIssueProperty(key, propKey string, value json.RawMessage) error {
	return c.PutJSON(c.api("issue/%s/properties/%s", key, propKey), nil, value, nil)
}

func (c *Client) DeleteIssueProperty(key, propKey string) error {
	return c.Delete(c.api("issue/%s/properties/%s", key, propKey), nil)
}

func (c *Client) Notify(key string, body map[string]any) error {
	return c.PostJSON(c.api("issue/%s/notify", key), nil, body, nil)
}
