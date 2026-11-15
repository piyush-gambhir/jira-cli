package client

import (
	"net/url"
	"strings"
)

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

func (c *Client) SearchDashboards(query string, limit int) ([]Dashboard, error) {
	if c.IsServer() {
		return c.searchDashboardsServer(query, limit)
	}
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

func (c *Client) searchDashboardsServer(query string, limit int) ([]Dashboard, error) {
	needle := strings.ToLower(strings.TrimSpace(query))
	out := []Dashboard{}
	q := url.Values{"maxResults": {"100"}}
	for startAt := 0; ; {
		q.Set("startAt", itoa(startAt))
		var page struct {
			StartAt    int         `json:"startAt"`
			MaxResults int         `json:"maxResults"`
			Total      int         `json:"total"`
			Dashboards []Dashboard `json:"dashboards"`
		}
		if err := c.GetJSON(c.api("dashboard"), q, &page); err != nil {
			return nil, err
		}
		for _, d := range page.Dashboards {
			if needle != "" && !strings.Contains(strings.ToLower(d.Name), needle) {
				continue
			}
			out = append(out, d)
			if limit > 0 && len(out) == limit {
				return out, nil
			}
		}
		if len(page.Dashboards) == 0 || page.MaxResults <= 0 || page.StartAt+page.MaxResults >= page.Total {
			return out, nil
		}
		startAt = page.StartAt + page.MaxResults
		q.Set("maxResults", itoa(page.MaxResults))
	}
}

func (c *Client) GetDashboard(id string) (*Dashboard, error) {
	var out Dashboard
	if err := c.GetJSON(c.api("dashboard/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
