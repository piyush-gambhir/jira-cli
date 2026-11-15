package client

import "net/url"

type Priority struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

type Resolution struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *Client) ListIssueTypes() ([]IssueType, error) {
	var out []IssueType
	if err := c.GetJSON(c.api("issuetype"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListIssueTypesForProject(projectID string) ([]IssueType, error) {
	if c.IsServer() {
		var p struct {
			IssueTypes []IssueType `json:"issueTypes"`
		}
		if err := c.GetJSON(c.api("project/%s", projectID), nil, &p); err != nil {
			return nil, err
		}
		return p.IssueTypes, nil
	}
	q := url.Values{"projectId": {projectID}}
	var out []IssueType
	if err := c.GetJSON(c.api("issuetype/project"), q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetIssueType(id string) (*IssueType, error) {
	var out IssueType
	if err := c.GetJSON(c.api("issuetype/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListPriorities() ([]Priority, error) {
	var out []Priority
	if err := c.GetJSON(c.api("priority"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetPriority(id string) (*Priority, error) {
	var out Priority
	if err := c.GetJSON(c.api("priority/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListResolutions() ([]Resolution, error) {
	var out []Resolution
	if err := c.GetJSON(c.api("resolution"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetResolution(id string) (*Resolution, error) {
	var out Resolution
	if err := c.GetJSON(c.api("resolution/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListLabels(limit int) ([]string, error) {
	if c.IsServer() {
		return nil, errCloudOnly("label list", "")
	}
	q := url.Values{}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []string `json:"values"`
	}
	if err := c.GetJSON(c.api("label"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}
