package client

import "net/url"

// Group is a Jira user group (name plus the immutable groupId).
type Group struct {
	Name    string `json:"name,omitempty"`
	GroupID string `json:"groupId,omitempty"`
	HTML    string `json:"html,omitempty"` // present in groups picker results
}

// ListGroups returns groups via the bulk endpoint (PageBean of {name,groupId}).
func (c *Client) ListGroups(limit int) ([]Group, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []Group `json:"values"`
	}
	if err := c.GetJSON(c.api("group/bulk"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// FindGroups searches for groups whose name matches query (groups picker).
func (c *Client) FindGroups(query string) ([]Group, error) {
	q := url.Values{}
	if query != "" {
		q.Set("query", query)
	}
	var out struct {
		Groups []Group `json:"groups"`
	}
	if err := c.GetJSON(c.api("groups/picker"), q, &out); err != nil {
		return nil, err
	}
	return out.Groups, nil
}

// GroupMembers returns the members of a group. Identify the group by name or, if
// groupID is set, by id. includeInactive also returns deactivated users.
func (c *Client) GroupMembers(name, groupID string, limit int, includeInactive bool) ([]User, error) {
	q := url.Values{}
	if groupID != "" {
		q.Set("groupId", groupID)
	} else {
		q.Set("groupname", name)
	}
	if includeInactive {
		q.Set("includeInactiveUsers", "true")
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []User `json:"values"`
	}
	if err := c.GetJSON(c.api("group/member"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// AddGroupUser adds the user (accountId) to the named group.
func (c *Client) AddGroupUser(name, accountID string) error {
	q := url.Values{"groupname": {name}}
	body := map[string]any{"accountId": accountID}
	return c.PostJSON(c.api("group/user"), q, body, nil)
}

// RemoveGroupUser removes the user (accountId) from the named group.
func (c *Client) RemoveGroupUser(name, accountID string) error {
	q := url.Values{"groupname": {name}, "accountId": {accountID}}
	return c.Delete(c.api("group/user"), q)
}

// CreateGroup creates a group with the given name and returns it.
func (c *Client) CreateGroup(name string) (*Group, error) {
	var out Group
	if err := c.PostJSON(c.api("group"), nil, map[string]any{"name": name}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteGroup deletes the named group. swapGroup, if set, receives any
// restrictions/permissions that referenced the deleted group.
func (c *Client) DeleteGroup(name, swapGroup string) error {
	q := url.Values{"groupname": {name}}
	if swapGroup != "" {
		q.Set("swapGroup", swapGroup)
	}
	return c.Delete(c.api("group"), q)
}
