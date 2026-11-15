package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) ListAllUsers(limit int) ([]User, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	if c.IsServer() {
		var page struct {
			Values []User `json:"values"`
		}
		err := c.GetJSON(c.api("user/list"), q, &page)
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("this Jira version has no user list endpoint (GET /rest/api/2/user/list needs Data Center 11.0 or a recent 10.3 LTS release); use `jira user search <query>` instead: %w", err)
		}
		if err != nil {
			return nil, err
		}
		return page.Values, nil
	}
	var out []User
	if err := c.GetJSON(c.api("users/search"), q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) BulkUsers(ids []string) ([]User, error) {
	if c.IsServer() {
		return nil, errCloudOnly("bulk user lookup", "")
	}
	all := []User{}
	for startAt := 0; startAt < len(ids); {
		q := url.Values{"accountId": ids}
		q.Set("startAt", itoa(startAt))
		q.Set("maxResults", itoa(len(ids)))
		var page struct {
			Values []User `json:"values"`
			IsLast bool   `json:"isLast"`
		}
		if err := c.GetJSON(c.api("user/bulk"), q, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Values...)
		if page.IsLast || len(page.Values) == 0 {
			break
		}
		startAt += len(page.Values)
	}
	return all, nil
}

func (c *Client) AssignableUsers(query, project, issueKey string, limit int) ([]User, error) {
	q := url.Values{}
	if query != "" {
		q.Set(c.userSearchParam(), query)
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

func (c *Client) UserGroups(accountID string) ([]Group, error) {
	if c.IsServer() {
		q := c.userQuery(accountID)
		q.Set("expand", "groups")
		var u struct {
			Groups struct {
				Items []Group `json:"items"`
			} `json:"groups"`
		}
		if err := c.GetJSON(c.api("user"), q, &u); err != nil {
			return nil, err
		}
		return u.Groups.Items, nil
	}
	var out []Group
	if err := c.GetJSON(c.api("user/groups"), url.Values{"accountId": {accountID}}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
