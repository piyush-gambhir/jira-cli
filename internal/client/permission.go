package client

import (
	"net/url"
	"strings"
)

// Permission is a single permission entry as returned by GET /mypermissions
// (key is the map key, not a body field, so callers fill Key in themselves).
type Permission struct {
	Key            string `json:"key,omitempty"`
	ID             string `json:"id,omitempty"`
	Type           string `json:"type,omitempty"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	HavePermission bool   `json:"havePermission"`
}

// MyPermissions returns the caller's permissions, optionally scoped to a project
// and/or issue. The permissions query param is required by Jira Cloud, so keys
// must be non-empty. The response shape is {permissions:{KEY:{...}}}; the map key
// is copied into each Permission's Key field for stable rendering.
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

// ListPermissions returns all permissions the instance knows about
// (GET /permissions -> {permissions:{KEY:{key,name,type,description}}}). The map
// key is copied into Key when the body omits it.
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

// PermittedProjects returns the ids of projects in which the caller holds all of
// the given permissions (POST /permissions/project, body {permissions:[...]} ->
// {projects:[{id}]}).
func (c *Client) PermittedProjects(keys []string) ([]string, error) {
	body := map[string]any{"permissions": keys}
	var resp struct {
		Projects []struct {
			ID flexString `json:"id"`
		} `json:"projects"`
	}
	if err := c.PostJSON(c.api("permissions/project"), nil, body, &resp); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(resp.Projects))
	for _, p := range resp.Projects {
		out = append(out, string(p.ID))
	}
	return out, nil
}
