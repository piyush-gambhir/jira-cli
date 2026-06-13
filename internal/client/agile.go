package client

import "net/url"

// ListBoards returns Agile boards, optionally filtered by project and type.
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

// GetBoard returns a single board.
func (c *Client) GetBoard(boardID int) (*Board, error) {
	var out Board
	if err := c.GetJSON(c.agile("board/%d", boardID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSprints returns sprints for a board, optionally filtered by state
// (comma-separated: future,active,closed).
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

// GetSprint returns a single sprint.
func (c *Client) GetSprint(sprintID int) (*Sprint, error) {
	var out Sprint
	if err := c.GetJSON(c.agile("sprint/%d", sprintID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateSprint creates a sprint on a board.
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

// SprintIssues returns issues in a sprint.
func (c *Client) SprintIssues(sprintID, limit int) ([]Issue, error) {
	return c.agileIssues(c.agile("sprint/%d/issue", sprintID), limit)
}

// BoardIssues returns all issues on a board.
func (c *Client) BoardIssues(boardID, limit int) ([]Issue, error) {
	return c.agileIssues(c.agile("board/%d/issue", boardID), limit)
}

// BacklogIssues returns a board's backlog issues.
func (c *Client) BacklogIssues(boardID, limit int) ([]Issue, error) {
	return c.agileIssues(c.agile("board/%d/backlog", boardID), limit)
}

// EpicIssues returns issues belonging to an epic.
func (c *Client) EpicIssues(epicKey string, limit int) ([]Issue, error) {
	return c.agileIssues(c.agile("epic/%s/issue", epicKey), limit)
}

// GetEpic returns an epic.
func (c *Client) GetEpic(epicKey string) (map[string]any, error) {
	var out map[string]any
	if err := c.GetJSON(c.agile("epic/%s", epicKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MoveIssuesToSprint moves issues into a sprint (max 50).
func (c *Client) MoveIssuesToSprint(sprintID int, issueKeys []string) error {
	return c.PostJSON(c.agile("sprint/%d/issue", sprintID), nil, map[string]any{"issues": issueKeys}, nil)
}

// MoveIssuesToBacklog moves issues to the backlog (max 50).
func (c *Client) MoveIssuesToBacklog(issueKeys []string) error {
	return c.PostJSON(c.agile("backlog/issue"), nil, map[string]any{"issues": issueKeys}, nil)
}

// agileIssues is the shared offset-paginated issue fetch for agile endpoints.
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
