package client

import "net/url"

func (c *Client) GetComment(key, id string) (*Comment, error) {
	var out Comment
	if err := c.GetJSON(c.api("issue/%s/comment/%s", key, id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateComment(key, id string, body any) (*Comment, error) {
	payload := map[string]any{"body": body}
	var out Comment
	if err := c.PutJSON(c.api("issue/%s/comment/%s", key, id), nil, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteComment(key, id string) error {
	return c.Delete(c.api("issue/%s/comment/%s", key, id), nil)
}

func (c *Client) GetWorklog(key, id string) (*Worklog, error) {
	var out Worklog
	if err := c.GetJSON(c.api("issue/%s/worklog/%s", key, id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateWorklog(key, id string, body map[string]any) (*Worklog, error) {
	var out Worklog
	if err := c.PutJSON(c.api("issue/%s/worklog/%s", key, id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteWorklog(key, id string, query url.Values) error {
	return c.Delete(c.api("issue/%s/worklog/%s", key, id), query)
}
