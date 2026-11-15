package client

import "net/url"

func (c *Client) ListBoards(projectKeyOrID, boardType string, limit int) ([]Board, error) {
	q := url.Values{}
	if projectKeyOrID != "" {
		q.Set("projectKeyOrId", projectKeyOrID)
	}
	if boardType != "" {
		q.Set("type", boardType)
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []Board `json:"values"`
	}
	if err := c.GetJSON(c.agile("board"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

func (c *Client) GetBoard(boardID int) (*Board, error) {
	var out Board
	if err := c.GetJSON(c.agile("board/%d", boardID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListSprints(boardID int, state string, limit int) ([]Sprint, error) {
	q := url.Values{}
	if state != "" {
		q.Set("state", state)
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []Sprint `json:"values"`
	}
	if err := c.GetJSON(c.agile("board/%d/sprint", boardID), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

func (c *Client) GetSprint(sprintID int) (*Sprint, error) {
	var out Sprint
	if err := c.GetJSON(c.agile("sprint/%d", sprintID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateSprint(name string, originBoardID int, startDate, endDate, goal string) (*Sprint, error) {
	body := map[string]any{"name": name, "originBoardId": originBoardID}
	if startDate != "" {
		body["startDate"] = startDate
	}
	if endDate != "" {
		body["endDate"] = endDate
	}
	if goal != "" {
		body["goal"] = goal
	}
	var out Sprint
	if err := c.PostJSON(c.agile("sprint"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SprintIssues(sprintID, limit int) ([]Issue, error) {
	return c.agileIssues(c.agile("sprint/%d/issue", sprintID), limit)
}

func (c *Client) BoardIssues(boardID, limit int) ([]Issue, error) {
	return c.agileIssues(c.agile("board/%d/issue", boardID), limit)
}

func (c *Client) BacklogIssues(boardID, limit int) ([]Issue, error) {
	return c.agileIssues(c.agile("board/%d/backlog", boardID), limit)
}

func (c *Client) EpicIssues(epicKey string, limit int) ([]Issue, error) {
	return c.agileIssues(c.agile("epic/%s/issue", epicKey), limit)
}

func (c *Client) GetEpic(epicKey string) (map[string]any, error) {
	var out map[string]any
	if err := c.GetJSON(c.agile("epic/%s", epicKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) MoveIssuesToSprint(sprintID int, issueKeys []string) error {
	return c.PostJSON(c.agile("sprint/%d/issue", sprintID), nil, map[string]any{"issues": issueKeys}, nil)
}

func (c *Client) MoveIssuesToBacklog(issueKeys []string) error {
	return c.PostJSON(c.agile("backlog/issue"), nil, map[string]any{"issues": issueKeys}, nil)
}

func (c *Client) agileIssues(path string, limit int) ([]Issue, error) {
	q := url.Values{"fields": {"summary,status,assignee,priority,issuetype,updated"}}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Issues []Issue `json:"issues"`
	}
	if err := c.GetJSON(path, q, &out); err != nil {
		return nil, err
	}
	return out.Issues, nil
}
