package client

// SprintPartialUpdate applies a partial update to a sprint via
// POST /rest/agile/1.0/sprint/{id}. Only the keys present in body are changed
// (e.g. {"state":"active"}, {"name":...,"goal":...}). The updated sprint is
// returned.
func (c *Client) SprintPartialUpdate(id int, body map[string]any) (*Sprint, error) {
	var out Sprint
	if err := c.PostJSON(c.agile("sprint/%d", id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSprint deletes a sprint via DELETE /rest/agile/1.0/sprint/{id}.
func (c *Client) DeleteSprint(id int) error {
	return c.Delete(c.agile("sprint/%d", id), nil)
}
