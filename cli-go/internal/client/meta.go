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

// SearchUsers finds users matching a query (display name or email; Server/DC
// also matches the username).
func (c *Client) SearchUsers(query string, limit int) ([]User, error) {
	q := url.Values{c.userSearchParam(): {query}}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out []User
	if err := c.GetJSON(c.api("user/search"), q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetUser returns a single user by accountId (Cloud) or username (Server/DC).
func (c *Client) GetUser(id string) (*User, error) {
	var out User
	if err := c.GetJSON(c.api("user"), c.userQuery(id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UserID returns the identifier the deployment uses for u: the accountId on
// Cloud, the username on Server/DC.
func (c *Client) UserID(u User) string {
	if c.IsServer() {
		return u.Name
	}
	return u.AccountID
}

// UserRef is the JSON object that identifies a user in a request body:
// {"accountId": id} on Cloud, {"name": id} on Server/DC. id may also be "-1"
// (default assignee) or nil (unassigned) where the endpoint allows it.
func (c *Client) UserRef(id any) map[string]any {
	if c.IsServer() {
		return map[string]any{"name": id}
	}
	return map[string]any{"accountId": id}
}

// userQuery identifies a user in a query string: accountId on Cloud, username
// on Server/DC.
func (c *Client) userQuery(id string) url.Values {
	if c.IsServer() {
		return url.Values{"username": {id}}
	}
	return url.Values{"accountId": {id}}
}

// userSearchParam is the user-search query parameter: Cloud's "query", or
// Server/DC's "username" (which matches username, name, or email).
func (c *Client) userSearchParam() string {
	if c.IsServer() {
		return "username"
	}
	return "query"
}

// errCloudOnly reports an endpoint that Jira Server/Data Center does not have.
// It is returned before any request is sent. hint, if set, names what to use
// on Server/Data Center instead.
func errCloudOnly(what, hint string) error {
	if hint != "" {
		return fmt.Errorf("%s is only available on Jira Cloud (REST API v3); %s", what, hint)
	}
	return fmt.Errorf("%s is only available on Jira Cloud (REST API v3)", what)
}

// ResolveUser turns a user reference into the user identifier the deployment
// uses (accountId on Cloud, username on Server/DC). It accepts "@me"/"me" (the
// current user), an explicit "id:<accountId or username>", or a name/email
// which is resolved via user search (first match wins).
func (c *Client) ResolveUser(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	switch ref {
	case "@me", "me", "":
		me, err := c.Myself()
		if err != nil {
			return "", err
		}
		return c.UserID(*me), nil
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
	return c.UserID(users[0]), nil
}

// ListProjects returns projects, optionally filtered by a query string.
func (c *Client) ListProjects(query string, limit int) ([]Project, error) {
	if c.IsServer() {
		return c.listProjectsServer(query, limit)
	}
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

// listProjectsServer lists projects on Server/DC, which has no project/search:
// GET project returns every visible project as an array, so the key/name filter
// (case-insensitive, like Cloud's query) and the limit are applied here.
func (c *Client) listProjectsServer(query string, limit int) ([]Project, error) {
	var all []Project
	if err := c.GetJSON(c.api("project"), nil, &all); err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	out := []Project{}
	for _, p := range all {
		if needle != "" && !strings.Contains(strings.ToLower(p.Key), needle) && !strings.Contains(strings.ToLower(p.Name), needle) {
			continue
		}
		out = append(out, p)
		if limit > 0 && len(out) == limit {
			break
		}
	}
	return out, nil
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
