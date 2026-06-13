package client

import "net/url"

// Priority is an issue priority definition.
type Priority struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

// Resolution is an issue resolution definition.
type Resolution struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListIssueTypes returns all issue types visible to the user.
func (c *Client) ListIssueTypes() ([]IssueType, error) {
	var out []IssueType
	if err := c.GetJSON(c.api("issuetype"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListIssueTypesForProject returns the issue types associated with a project id.
func (c *Client) ListIssueTypesForProject(projectID string) ([]IssueType, error) {
	q := url.Values{"projectId": {projectID}}
	var out []IssueType
	if err := c.GetJSON(c.api("issuetype/project"), q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetIssueType returns a single issue type by id.
func (c *Client) GetIssueType(id string) (*IssueType, error) {
	var out IssueType
	if err := c.GetJSON(c.api("issuetype/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPriorities returns all issue priorities.
func (c *Client) ListPriorities() ([]Priority, error) {
	var out []Priority
	if err := c.GetJSON(c.api("priority"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPriority returns a single priority by id.
func (c *Client) GetPriority(id string) (*Priority, error) {
	var out Priority
	if err := c.GetJSON(c.api("priority/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListResolutions returns all issue resolutions.
func (c *Client) ListResolutions() ([]Resolution, error) {
	var out []Resolution
	if err := c.GetJSON(c.api("resolution"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetResolution returns a single resolution by id.
func (c *Client) GetResolution(id string) (*Resolution, error) {
	var out Resolution
	if err := c.GetJSON(c.api("resolution/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListLabels returns available issue labels (the PageBean values array). If
// limit is positive at most that many labels are returned.
func (c *Client) ListLabels(limit int) ([]string, error) {
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
