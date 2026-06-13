package client

import (
	"fmt"
	"net/url"
	"strings"
)

// ServerInfo returns instance metadata (works unauthenticated on Cloud).
func (c *Client) ServerInfo() (*ServerInfo, error) {
	var out ServerInfo
	if err := c.GetJSON(c.api("serverInfo"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Myself returns the authenticated user (a 200 confirms credentials work).
func (c *Client) Myself() (*User, error) {
	var out User
	if err := c.GetJSON(c.api("myself"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SearchUsers finds users matching a query (display name or email).
func (c *Client) SearchUsers(query string, limit int) ([]User, error) {
	q := url.Values{"query": {query}}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out []User
	if err := c.GetJSON(c.api("user/search"), q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetUser returns a single user by accountId.
func (c *Client) GetUser(accountID string) (*User, error) {
	var out User
	if err := c.GetJSON(c.api("user"), url.Values{"accountId": {accountID}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResolveUser turns a user reference into an accountId. It accepts "@me"/"me"
// (the current user), an explicit "id:<accountId>", or a name/email which is
// resolved via user search (first match wins).
func (c *Client) ResolveUser(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	switch ref {
	case "@me", "me", "":
		me, err := c.Myself()
		if err != nil {
			return "", err
		}
		return me.AccountID, nil
	}
	if strings.HasPrefix(ref, "id:") {
		return strings.TrimPrefix(ref, "id:"), nil
	}
	users, err := c.SearchUsers(ref, 2)
	if err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", fmt.Errorf("no user found matching %q", ref)
	}
	return users[0].AccountID, nil
}

// ListProjects returns projects, optionally filtered by a query string.
func (c *Client) ListProjects(query string, limit int) ([]Project, error) {
	q := url.Values{}
	if query != "" {
		q.Set("query", query)
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []Project `json:"values"`
	}
	if err := c.GetJSON(c.api("project/search"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// GetProject returns a single project by id or key.
func (c *Client) GetProject(idOrKey string) (*Project, error) {
	var out Project
	if err := c.GetJSON(c.api("project/%s", idOrKey), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateProject creates a project (requires the manage:jira-project scope /
// project-admin permission). body is the raw CreateProjectDetails. Returns the
// {self,id,key} identifiers.
func (c *Client) CreateProject(body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.PostJSON(c.api("project"), nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProject updates a project's fields.
func (c *Client) UpdateProject(idOrKey string, body map[string]any) (*Project, error) {
	var out Project
	if err := c.PutJSON(c.api("project/%s", idOrKey), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteProject deletes a project. enableUndo moves it to the recycle bin instead.
func (c *Client) DeleteProject(idOrKey string, enableUndo bool) error {
	q := url.Values{}
	if enableUndo {
		q.Set("enableUndo", "true")
	}
	return c.Delete(c.api("project/%s", idOrKey), q)
}

// ListFields returns all system and custom fields.
func (c *Client) ListFields() ([]Field, error) {
	var out []Field
	if err := c.GetJSON(c.api("field"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
