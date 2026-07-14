package client

import "net/url"

// ListAllUsers returns all users on the site (GET /user/search with an empty
// query returns every user). Use limit to cap the page size.
func (c *Client) ListAllUsers(limit int) ([]User, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out []User
	if err := c.GetJSON(c.api("users/search"), q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BulkUsers returns the users for the given accountIds in a single request
// (GET /user/bulk?accountId=..&accountId=..). Returns a PageBean of User.
func (c *Client) BulkUsers(ids []string) ([]User, error) {
	q := url.Values{"accountId": ids}
	q.Set("maxResults", itoa(10))
	var out struct {
		Values []User `json:"values"`
	}
	if err := c.GetJSON(c.api("user/bulk"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// AssignableUsers returns users assignable to issues, optionally filtered by a
// query (name/email). Scope it with a project key or an issue key (issueKey
// wins when both are set). This is the correct picker for issue assign/create.
func (c *Client) AssignableUsers(query, project, issueKey string, limit int) ([]User, error) {
	q := url.Values{}
	if query != "" {
		q.Set("query", query)
	}
	if issueKey != "" {
		q.Set("issueKey", issueKey)
	} else if project != "" {
		q.Set("project", project)
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out []User
	if err := c.GetJSON(c.api("user/assignable/search"), q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UserGroups returns the groups a user (accountId) belongs to
// (GET /user/groups?accountId= -> array of {name,groupId}).
func (c *Client) UserGroups(accountID string) ([]Group, error) {
	var out []Group
	if err := c.GetJSON(c.api("user/groups"), url.Values{"accountId": {accountID}}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
