package client

import (
	"net/url"
	"strings"
)

type Permission struct {
	Key            string `json:"key,omitempty"`
	ID             string `json:"id,omitempty"`
	Type           string `json:"type,omitempty"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	HavePermission bool   `json:"havePermission"`
}

func (c *Client) MyPermissions(keys []string, projectKey, issueKey string) ([]Permission, error) {
	q := url.Values{"permissions": {strings.Join(keys, ",")}}
	if projectKey != "" {
		q.Set("projectKey", projectKey)
	}
	if issueKey != "" {
		q.Set("issueKey", issueKey)
	}
	var resp struct {
		Permissions map[string]Permission `json:"permissions"`
	}
	if err := c.GetJSON(c.api("mypermissions"), q, &resp); err != nil {
		return nil, err
	}
	out := make([]Permission, 0, len(resp.Permissions))
	for k, p := range resp.Permissions {
		if p.Key == "" {
			p.Key = k
		}
		out = append(out, p)
	}
	return out, nil
}

func (c *Client) ListPermissions() ([]Permission, error) {
	var resp struct {
		Permissions map[string]Permission `json:"permissions"`
	}
	if err := c.GetJSON(c.api("permissions"), nil, &resp); err != nil {
		return nil, err
	}
	out := make([]Permission, 0, len(resp.Permissions))
	for k, p := range resp.Permissions {
		if p.Key == "" {
			p.Key = k
		}
		out = append(out, p)
	}
	return out, nil
}

func (c *Client) PermittedProjects(keys []string) ([]string, error) {
	if c.IsServer() {
		return nil, errCloudOnly("permission permitted-projects", "on Server/Data Center check one project with `jira permission mine --project <key>`")
	}
	body := map[string]any{"permissions": keys}
	var resp struct {
		Projects []struct {
			ID flexString `json:"id"`
		} `json:"projects"`
	}
	if err := c.PostJSONRetryable(c.api("permissions/project"), nil, body, &resp); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(resp.Projects))
	for _, p := range resp.Projects {
		out = append(out, string(p.ID))
	}
	return out, nil
}
