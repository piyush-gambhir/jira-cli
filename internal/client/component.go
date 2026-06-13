package client

import "net/url"

// Component is a project component.
type Component struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	Lead         *User  `json:"lead,omitempty"`
	AssigneeType string `json:"assigneeType,omitempty"`
	Project      string `json:"project,omitempty"` // project key
	Self         string `json:"self,omitempty"`
}

// ListComponents returns the components of a project (by id or key).
func (c *Client) ListComponents(projectIDOrKey string) ([]Component, error) {
	var out []Component
	if err := c.GetJSON(c.api("project/%s/components", projectIDOrKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetComponent returns a single component by id.
func (c *Client) GetComponent(id string) (*Component, error) {
	var out Component
	if err := c.GetJSON(c.api("component/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateComponent creates a component. body is the raw component details
// (name, project, description, leadAccountId, assigneeType).
func (c *Client) CreateComponent(body map[string]any) (*Component, error) {
	var out Component
	if err := c.PostJSON(c.api("component"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateComponent updates a component's fields.
func (c *Client) UpdateComponent(id string, body map[string]any) (*Component, error) {
	var out Component
	if err := c.PutJSON(c.api("component/%s", id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteComponent deletes a component. If moveIssuesTo is non-empty, issues that
// reference the deleted component are reassigned to that component id.
func (c *Client) DeleteComponent(id, moveIssuesTo string) error {
	q := url.Values{}
	if moveIssuesTo != "" {
		q.Set("moveIssuesTo", moveIssuesTo)
	}
	return c.Delete(c.api("component/%s", id), q)
}

// ComponentIssueCount returns the number of issues using a component.
func (c *Client) ComponentIssueCount(id string) (int, error) {
	var out struct {
		IssueCount int `json:"issueCount"`
	}
	if err := c.GetJSON(c.api("component/%s/relatedIssueCounts", id), nil, &out); err != nil {
		return 0, err
	}
	return out.IssueCount, nil
}
