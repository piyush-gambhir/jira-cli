package client

import "net/url"

// Dashboard is a Jira dashboard (subset of fields).
type Dashboard struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Owner       *User  `json:"owner,omitempty"`
	View        string `json:"view,omitempty"`
	IsFavourite bool   `json:"isFavourite,omitempty"`
	Popularity  int    `json:"popularity,omitempty"`
	Self        string `json:"self,omitempty"`
}

// ListDashboards returns dashboards visible to the user, optionally filtered by
// "my", "favourite", or "public" (GET /dashboard returns {dashboards:[...]}).
func (c *Client) ListDashboards(filter string, limit int) ([]Dashboard, error) {
	q := url.Values{}
	if filter != "" {
		q.Set("filter", filter)
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Dashboards []Dashboard `json:"dashboards"`
	}
	if err := c.GetJSON(c.api("dashboard"), q, &out); err != nil {
		return nil, err
	}
	return out.Dashboards, nil
}

// SearchDashboards searches dashboards by name (GET /dashboard/search, a
// PageBean returning {values:[...]}).
func (c *Client) SearchDashboards(query string, limit int) ([]Dashboard, error) {
	q := url.Values{}
	if query != "" {
		q.Set("dashboardName", query)
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []Dashboard `json:"values"`
	}
	if err := c.GetJSON(c.api("dashboard/search"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// GetDashboard returns a single dashboard by id.
func (c *Client) GetDashboard(id string) (*Dashboard, error) {
	var out Dashboard
	if err := c.GetJSON(c.api("dashboard/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
