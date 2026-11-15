package client

func (c *Client) ListLinkTypes() ([]LinkType, error) {
	var out struct {
		IssueLinkTypes []LinkType `json:"issueLinkTypes"`
	}
	if err := c.GetJSON(c.api("issueLinkType"), nil, &out); err != nil {
		return nil, err
	}
	return out.IssueLinkTypes, nil
}

func (c *Client) LinkIssues(linkType, inwardKey, outwardKey string, comment any) error {
	body := map[string]any{
		"type":         map[string]string{"name": linkType},
		"inwardIssue":  map[string]string{"key": inwardKey},
		"outwardIssue": map[string]string{"key": outwardKey},
	}
	if comment != nil {
		body["comment"] = map[string]any{"body": comment}
	}
	return c.PostJSON(c.api("issueLink"), nil, body, nil)
}

func (c *Client) DeleteLink(linkID string) error {
	return c.Delete(c.api("issueLink/%s", linkID), nil)
}

func (c *Client) GetWatchers(key string) (*Watchers, error) {
	var out Watchers
	if err := c.GetJSON(c.api("issue/%s/watchers", key), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AddWatcher(key, accountID string) error {
	if accountID == "" {
		return c.PostJSON(c.api("issue/%s/watchers", key), nil, nil, nil)
	}
	return c.PostJSON(c.api("issue/%s/watchers", key), nil, accountID, nil)
}

func (c *Client) RemoveWatcher(key, accountID string) error {
	return c.Delete(c.api("issue/%s/watchers", key), c.userQuery(accountID))
}

func (c *Client) GetVotes(key string) (*Votes, error) {
	var out Votes
	if err := c.GetJSON(c.api("issue/%s/votes", key), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AddVote(key string) error {
	return c.PostJSON(c.api("issue/%s/votes", key), nil, nil, nil)
}

func (c *Client) RemoveVote(key string) error {
	return c.Delete(c.api("issue/%s/votes", key), nil)
}
