package client

import "net/url"

// ListLinkTypes returns the available issue link types and their direction labels.
func (c *Client) ListLinkTypes() ([]LinkType, error) {
	var out struct {
		IssueLinkTypes []LinkType `json:"issueLinkTypes"`
	}
	if err := c.GetJSON(c.api("issueLinkType"), nil, &out); err != nil {
		return nil, err
	}
	return out.IssueLinkTypes, nil
}

// LinkIssues creates a link between two issues. linkType is the type name (e.g.
// "Blocks"); comment is optional (ADF). Returns no body on success (201).
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

// DeleteLink removes an issue link by id.
func (c *Client) DeleteLink(linkID string) error {
	return c.Delete(c.api("issueLink/%s", linkID), nil)
}

// GetWatchers returns the watchers of an issue.
func (c *Client) GetWatchers(key string) (*Watchers, error) {
	var out Watchers
	if err := c.GetJSON(c.api("issue/%s/watchers", key), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddWatcher adds a watcher. An empty accountID adds the calling user. The add
// endpoint takes a bare JSON string body (the accountId).
func (c *Client) AddWatcher(key, accountID string) error {
	if accountID == "" {
		return c.PostJSON(c.api("issue/%s/watchers", key), nil, nil, nil)
	}
	return c.PostJSON(c.api("issue/%s/watchers", key), nil, accountID, nil)
}

// RemoveWatcher removes a watcher (identified by accountId query param).
func (c *Client) RemoveWatcher(key, accountID string) error {
	return c.Delete(c.api("issue/%s/watchers", key), url.Values{"accountId": {accountID}})
}

// GetVotes returns vote info for an issue.
func (c *Client) GetVotes(key string) (*Votes, error) {
	var out Votes
	if err := c.GetJSON(c.api("issue/%s/votes", key), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddVote casts the calling user's vote (no body).
func (c *Client) AddVote(key string) error {
	return c.PostJSON(c.api("issue/%s/votes", key), nil, nil, nil)
}

// RemoveVote retracts the calling user's vote.
func (c *Client) RemoveVote(key string) error {
	return c.Delete(c.api("issue/%s/votes", key), nil)
}
