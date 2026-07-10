package client

// DefaultSearchFields is the field set requested when the caller doesn't specify
// one (the new /search/jql defaults to id only, so we must be explicit).
var DefaultSearchFields = []string{"summary", "status", "assignee", "priority", "issuetype", "updated"}

// SearchIssues runs a JQL query. On Cloud (v3) it uses the enhanced
// /search/jql cursor endpoint; on Server/DC (v2) it uses the classic /search.
// If fetchAll is false, at most `limit` issues are returned.
func (c *Client) SearchIssues(jql string, fields []string, limit int, fetchAll bool) ([]Issue, error) {
	if len(fields) == 0 {
		fields = DefaultSearchFields
	}
	if c.apiVersion == "2" {
		return c.searchClassic(jql, fields, limit, fetchAll)
	}
	return c.searchJQL(jql, fields, limit, fetchAll)
}

// searchJQL paginates the Cloud enhanced search via nextPageToken.
func (c *Client) searchJQL(jql string, fields []string, limit int, fetchAll bool) ([]Issue, error) {
	pageSize := pageSizeFor(limit)
	all := []Issue{}
	token := ""
	for {
		body := map[string]any{"jql": jql, "maxResults": pageSize, "fields": fields}
		if token != "" {
			body["nextPageToken"] = token
		}
		var resp struct {
			Issues        []Issue `json:"issues"`
			NextPageToken string  `json:"nextPageToken"`
			IsLast        bool    `json:"isLast"`
		}
		if err := c.PostJSONRetryable(c.api("search/jql"), nil, body, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Issues...)
		if !fetchAll && limit > 0 && len(all) >= limit {
			return all[:limit], nil
		}
		if resp.IsLast || resp.NextPageToken == "" || len(resp.Issues) == 0 {
			break
		}
		token = resp.NextPageToken
	}
	return all, nil
}

// searchClassic paginates the deprecated/Server search via startAt/total.
func (c *Client) searchClassic(jql string, fields []string, limit int, fetchAll bool) ([]Issue, error) {
	all := []Issue{}
	startAt := 0
	pageSize := pageSizeFor(limit)
	for {
		body := map[string]any{"jql": jql, "startAt": startAt, "maxResults": pageSize, "fields": fields}
		var resp struct {
			Issues     []Issue `json:"issues"`
			Total      int     `json:"total"`
			StartAt    int     `json:"startAt"`
			MaxResults int     `json:"maxResults"`
		}
		if err := c.PostJSONRetryable(c.api("search"), nil, body, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Issues...)
		if !fetchAll && limit > 0 && len(all) >= limit {
			return all[:limit], nil
		}
		if len(resp.Issues) == 0 || startAt+len(resp.Issues) >= resp.Total {
			break
		}
		startAt += len(resp.Issues)
	}
	return all, nil
}

// ApproximateCount returns the (approximate) number of issues matching a JQL query.
func (c *Client) ApproximateCount(jql string) (int, error) {
	if c.apiVersion == "2" {
		// DC has no approximate-count; ask classic search for the total only.
		var resp struct {
			Total int `json:"total"`
		}
		body := map[string]any{"jql": jql, "startAt": 0, "maxResults": 0}
		if err := c.PostJSONRetryable(c.api("search"), nil, body, &resp); err != nil {
			return 0, err
		}
		return resp.Total, nil
	}
	var resp struct {
		Count int `json:"count"`
	}
	if err := c.PostJSONRetryable(c.api("search/approximate-count"), nil, map[string]any{"jql": jql}, &resp); err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func pageSizeFor(limit int) int {
	if limit > 0 && limit < 100 {
		return limit
	}
	return 100
}
