package client

import (
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) ServerInfo() (*ServerInfo, error) {
	var out ServerInfo
	if err := c.GetJSON(c.api("serverInfo"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Myself() (*User, error) {
	var out User
	if err := c.GetJSON(c.api("myself"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

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

func (c *Client) GetUser(id string) (*User, error) {
	var out User
	if err := c.GetJSON(c.api("user"), c.userQuery(id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UserID(u User) string {
	if c.IsServer() {
		return u.Name
	}
	return u.AccountID
}

func (c *Client) UserRef(id any) map[string]any {
	if c.IsServer() {
		return map[string]any{"name": id}
	}
	return map[string]any{"accountId": id}
}

func (c *Client) userQuery(id string) url.Values {
	if c.IsServer() {
		return url.Values{"username": {id}}
	}
	return url.Values{"accountId": {id}}
}

func (c *Client) userSearchParam() string {
	if c.IsServer() {
		return "username"
	}
	return "query"
}

func errCloudOnly(what, hint string) error {
	if hint != "" {
		return fmt.Errorf("%s is only available on Jira Cloud (REST API v3); %s", what, hint)
	}
	return fmt.Errorf("%s is only available on Jira Cloud (REST API v3)", what)
}

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

func (c *Client) GetProject(idOrKey string) (*Project, error) {
	var out Project
	if err := c.GetJSON(c.api("project/%s", idOrKey), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateProject(body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.PostJSON(c.api("project"), nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateProject(idOrKey string, body map[string]any) (*Project, error) {
	var out Project
	if err := c.PutJSON(c.api("project/%s", idOrKey), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteProject(idOrKey string, enableUndo bool) error {
	q := url.Values{}
	if enableUndo {
		q.Set("enableUndo", "true")
	}
	return c.Delete(c.api("project/%s", idOrKey), q)
}

func (c *Client) ListFields() ([]Field, error) {
	var out []Field
	if err := c.GetJSON(c.api("field"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
