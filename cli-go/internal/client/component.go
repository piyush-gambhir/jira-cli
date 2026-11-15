package client

import "net/url"

type Component struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	Lead         *User  `json:"lead,omitempty"`
	AssigneeType string `json:"assigneeType,omitempty"`
	Project      string `json:"project,omitempty"`
	Self         string `json:"self,omitempty"`
}

func (c *Client) ListComponents(projectIDOrKey string) ([]Component, error) {
	var out []Component
	if err := c.GetJSON(c.api("project/%s/components", projectIDOrKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetComponent(id string) (*Component, error) {
	var out Component
	if err := c.GetJSON(c.api("component/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateComponent(body map[string]any) (*Component, error) {
	var out Component
	if err := c.PostJSON(c.api("component"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateComponent(id string, body map[string]any) (*Component, error) {
	var out Component
	if err := c.PutJSON(c.api("component/%s", id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteComponent(id, moveIssuesTo string) error {
	q := url.Values{}
	if moveIssuesTo != "" {
		q.Set("moveIssuesTo", moveIssuesTo)
	}
	return c.Delete(c.api("component/%s", id), q)
}

func (c *Client) ComponentIssueCount(id string) (int, error) {
	var out struct {
		IssueCount int `json:"issueCount"`
	}
	if err := c.GetJSON(c.api("component/%s/relatedIssueCounts", id), nil, &out); err != nil {
		return 0, err
	}
	return out.IssueCount, nil
}
