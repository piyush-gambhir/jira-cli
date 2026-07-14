package client

// ListStatuses returns all issue statuses defined on the instance
// (GET /rest/api/3/status -> array of Status).
func (c *Client) ListStatuses() ([]Status, error) {
	var out []Status
	if err := c.GetJSON(c.api("status"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetStatus returns a single status by id or name
// (GET /rest/api/3/status/{idOrName}).
func (c *Client) GetStatus(idOrName string) (*Status, error) {
	var out Status
	if err := c.GetJSON(c.api("status/%s", idOrName), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
