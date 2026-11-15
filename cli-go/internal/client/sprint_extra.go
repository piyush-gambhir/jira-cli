package client

func (c *Client) SprintPartialUpdate(id int, body map[string]any) (*Sprint, error) {
	var out Sprint
	if err := c.PostJSON(c.agile("sprint/%d", id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSprint(id int) error {
	return c.Delete(c.agile("sprint/%d", id), nil)
}
