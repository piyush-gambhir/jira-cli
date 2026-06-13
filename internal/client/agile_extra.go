package client

import "net/url"

// BoardEpic is an epic as returned by the board epic listing.
type BoardEpic struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Self  string `json:"self,omitempty"`
	Done  bool   `json:"done"`
	Color struct {
		Key string `json:"key,omitempty"`
	} `json:"color,omitempty"`
}

// BoardCreate creates an Agile board. body is the raw create payload
// (name, type, filterId, optional location).
func (c *Client) BoardCreate(body map[string]any) (*Board, error) {
	var out Board
	if err := c.PostJSON(c.agile("board"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBoard deletes a board by id.
func (c *Client) DeleteBoard(boardID int) error {
	return c.Delete(c.agile("board/%d", boardID), nil)
}

// BoardConfig returns a board's configuration (columns, estimation, ranking field).
func (c *Client) BoardConfig(boardID int) (map[string]any, error) {
	var out map[string]any
	if err := c.GetJSON(c.agile("board/%d/configuration", boardID), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BoardEpics returns the epics associated with a board. When done is true only
// completed epics are returned.
func (c *Client) BoardEpics(boardID int, done bool, limit int) ([]BoardEpic, error) {
	q := url.Values{}
	if done {
		q.Set("done", "true")
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	var out struct {
		Values []BoardEpic `json:"values"`
	}
	if err := c.GetJSON(c.agile("board/%d/epic", boardID), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// EpicUpdate partially updates an epic (name, summary, color, done). body is the
// raw partial-update payload.
func (c *Client) EpicUpdate(epicKey string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.PostJSON(c.agile("epic/%s", epicKey), nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EpicAddIssues moves issues into an epic (max 50).
func (c *Client) EpicAddIssues(epicKey string, issueKeys []string) error {
	return c.PostJSON(c.agile("epic/%s/issue", epicKey), nil, map[string]any{"issues": issueKeys}, nil)
}

// EpicRemoveIssues removes issues from their epic (POST to the reserved "none"
// epic; max 50).
func (c *Client) EpicRemoveIssues(issueKeys []string) error {
	return c.PostJSON(c.agile("epic/none/issue"), nil, map[string]any{"issues": issueKeys}, nil)
}

// MoveIssuesToBoardBacklog moves issues to the backlog of a specific board (max
// 50). Use MoveIssuesToBacklog for the board-agnostic endpoint.
func (c *Client) MoveIssuesToBoardBacklog(boardID int, issueKeys []string) error {
	return c.PostJSON(c.agile("backlog/%d/issue", boardID), nil, map[string]any{"issues": issueKeys}, nil)
}
