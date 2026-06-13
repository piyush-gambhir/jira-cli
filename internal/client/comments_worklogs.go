package client

import "net/url"

// ListComments returns an issue's comments (most recent first, up to limit).
func (c *Client) ListComments(key string, limit int) ([]Comment, error) {
	var out struct {
		Comments []Comment `json:"comments"`
	}
	q := url.Values{"orderBy": {"-created"}}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	if err := c.GetJSON(c.api("issue/%s/comment", key), q, &out); err != nil {
		return nil, err
	}
	return out.Comments, nil
}

// AddComment adds a comment. body is the comment body (an ADF document on Cloud
// v3, or a plain string on v2); visibility is optional.
func (c *Client) AddComment(key string, body any, visibility map[string]any) (*Comment, error) {
	payload := map[string]any{"body": body}
	if len(visibility) > 0 {
		payload["visibility"] = visibility
	}
	var out Comment
	if err := c.PostJSON(c.api("issue/%s/comment", key), nil, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListWorklogs returns an issue's worklog entries (up to limit).
func (c *Client) ListWorklogs(key string, limit int) ([]Worklog, error) {
	var out struct {
		Worklogs []Worklog `json:"worklogs"`
	}
	q := url.Values{}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	if err := c.GetJSON(c.api("issue/%s/worklog", key), q, &out); err != nil {
		return nil, err
	}
	return out.Worklogs, nil
}

// AddWorklog logs work on an issue. body carries timeSpent/timeSpentSeconds,
// started, and optional comment (ADF). adjustEstimate/newEstimate are optional.
func (c *Client) AddWorklog(key string, body map[string]any, adjustEstimate, newEstimate string) (*Worklog, error) {
	q := url.Values{}
	if adjustEstimate != "" {
		q.Set("adjustEstimate", adjustEstimate)
	}
	if newEstimate != "" {
		q.Set("newEstimate", newEstimate)
	}
	var out Worklog
	if err := c.PostJSON(c.api("issue/%s/worklog", key), q, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
