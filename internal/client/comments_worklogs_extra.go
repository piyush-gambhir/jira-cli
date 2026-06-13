package client

import "net/url"

// GetComment returns a single comment on an issue.
func (c *Client) GetComment(key, id string) (*Comment, error) {
	var out Comment
	if err := c.GetJSON(c.api("issue/%s/comment/%s", key, id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateComment edits a comment's body. body is the new comment body (an ADF
// document on Cloud v3, or a plain string on v2).
func (c *Client) UpdateComment(key, id string, body any) (*Comment, error) {
	payload := map[string]any{"body": body}
	var out Comment
	if err := c.PutJSON(c.api("issue/%s/comment/%s", key, id), nil, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteComment deletes a comment from an issue.
func (c *Client) DeleteComment(key, id string) error {
	return c.Delete(c.api("issue/%s/comment/%s", key, id), nil)
}

// GetWorklog returns a single worklog entry on an issue.
func (c *Client) GetWorklog(key, id string) (*Worklog, error) {
	var out Worklog
	if err := c.GetJSON(c.api("issue/%s/worklog/%s", key, id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateWorklog edits a worklog entry. body carries any of timeSpent/
// timeSpentSeconds, started, and comment (ADF).
func (c *Client) UpdateWorklog(key, id string, body map[string]any) (*Worklog, error) {
	var out Worklog
	if err := c.PutJSON(c.api("issue/%s/worklog/%s", key, id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteWorklog deletes a worklog entry. query may carry adjustEstimate,
// newEstimate, and increaseBy to control how the remaining estimate is updated.
func (c *Client) DeleteWorklog(key, id string, query url.Values) error {
	return c.Delete(c.api("issue/%s/worklog/%s", key, id), query)
}
