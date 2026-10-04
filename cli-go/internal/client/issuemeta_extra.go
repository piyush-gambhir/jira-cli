package client

import "net/url"

// EditMeta is the response of GET /issue/{key}/editmeta: the fields editable on
// the issue's edit screen, keyed by field id.
type EditMeta struct {
	Fields map[string]FieldMeta `json:"fields"`
}

// GetEditMeta returns the edit-screen field metadata for an issue.
func (c *Client) GetEditMeta(key string) (*EditMeta, error) {
	var out EditMeta
	if err := c.GetJSON(c.api("issue/%s/editmeta", key), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ChangelogItem is a single field change within a changelog entry.
type ChangelogItem struct {
	Field      string `json:"field"`
	FieldType  string `json:"fieldtype,omitempty"`
	From       string `json:"from,omitempty"`
	FromString string `json:"fromString,omitempty"`
	To         string `json:"to,omitempty"`
	ToString   string `json:"toString,omitempty"`
}

// ChangelogEntry is one history record (an author's set of field changes).
type ChangelogEntry struct {
	ID      string          `json:"id,omitempty"`
	Author  *User           `json:"author,omitempty"`
	Created string          `json:"created,omitempty"`
	Items   []ChangelogItem `json:"items,omitempty"`
}

// GetChangelog returns an issue's change history (offset-paginated). Server/DC
// has no issue changelog resource, so it reads the issue with
// expand=changelog (every history, oldest first) and pages locally.
func (c *Client) GetChangelog(key string, startAt, max int) ([]ChangelogEntry, error) {
	if c.IsServer() {
		var issue struct {
			Changelog struct {
				Histories []ChangelogEntry `json:"histories"`
			} `json:"changelog"`
		}
		if err := c.GetJSON(c.api("issue/%s", key), issueQuery("summary", "changelog"), &issue); err != nil {
			return nil, err
		}
		h := issue.Changelog.Histories
		if startAt > 0 {
			h = h[min(startAt, len(h)):]
		}
		if max > 0 {
			h = h[:min(max, len(h))]
		}
		return h, nil
	}
	q := url.Values{}
	if startAt > 0 {
		q.Set("startAt", itoa(startAt))
	}
	if max > 0 {
		q.Set("maxResults", itoa(max))
	}
	var out struct {
		Values []ChangelogEntry `json:"values"`
	}
	if err := c.GetJSON(c.api("issue/%s/changelog", key), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}
