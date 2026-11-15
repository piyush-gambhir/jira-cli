package client

func (c *Client) ListStatuses() ([]Status, error) {
	var out []Status
	if err := c.GetJSON(c.api("status"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetStatus(idOrName string) (*Status, error) {
	var out Status
	if err := c.GetJSON(c.api("status/%s", idOrName), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
