package client

import "net/url"

type Group struct {
	Name    string `json:"name,omitempty"`
	GroupID string `json:"groupId,omitempty"`
	HTML    string `json:"html,omitempty"`
}

func (c *Client) ListGroups(limit int) ([]Group, int, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	if c.IsServer() {
		var out struct {
			Total  int     `json:"total"`
			Groups []Group `json:"groups"`
		}
		if err := c.GetJSON(c.api("groups/picker"), q, &out); err != nil {
			return nil, 0, err
		}
		return out.Groups, out.Total, nil
	}
	var out struct {
		Total  int     `json:"total"`
		Values []Group `json:"values"`
	}
	if err := c.GetJSON(c.api("group/bulk"), q, &out); err != nil {
		return nil, 0, err
	}
	return out.Values, out.Total, nil
}

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

func (c *Client) GroupMembers(name, groupID string, limit int, includeInactive bool) ([]User, error) {
	if groupID != "" && c.IsServer() {
		return nil, errCloudOnly("--group-id", "Server/Data Center groups have no ids, so pass the group name")
	}
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

func (c *Client) AddGroupUser(name, accountID string) error {
	q := url.Values{"groupname": {name}}
	return c.PostJSON(c.api("group/user"), q, c.UserRef(accountID), nil)
}

func (c *Client) RemoveGroupUser(name, accountID string) error {
	q := c.userQuery(accountID)
	q.Set("groupname", name)
	return c.Delete(c.api("group/user"), q)
}

func (c *Client) CreateGroup(name string) (*Group, error) {
	var out Group
	if err := c.PostJSON(c.api("group"), nil, map[string]any{"name": name}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteGroup(name, swapGroup string) error {
	q := url.Values{"groupname": {name}}
	if swapGroup != "" {
		q.Set("swapGroup", swapGroup)
	}
	return c.Delete(c.api("group"), q)
}
