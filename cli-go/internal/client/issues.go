package client

import (
	"net/url"
	"strconv"
)

// IssueUpdate is the create/edit/transition request body (fields + update verbs).
type IssueUpdate struct {
	Fields     map[string]any   `json:"fields,omitempty"`
	Update     map[string]any   `json:"update,omitempty"`
	Transition map[string]any   `json:"transition,omitempty"`
	Properties []map[string]any `json:"properties,omitempty"`
}

// CreateIssue creates an issue. fields is the raw fields map (project, issuetype,
// summary, description-as-ADF, etc.). Returns {id,key,self}.
func (c *Client) CreateIssue(update IssueUpdate) (*Issue, error) {
	var out Issue
	if err := c.PostJSON(c.api("issue"), nil, update, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIssue fetches an issue into the typed struct. fields/expand are optional CSV.
func (c *Client) GetIssue(key, fields, expand string) (*Issue, error) {
	var out Issue
	if err := c.GetJSON(c.api("issue/%s", key), issueQuery(fields, expand), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIssueRaw fetches an issue as generic JSON for full-fidelity output.
func (c *Client) GetIssueRaw(key, fields, expand string) (map[string]any, error) {
	var out map[string]any
	if err := c.GetJSON(c.api("issue/%s", key), issueQuery(fields, expand), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func issueQuery(fields, expand string) url.Values {
	q := url.Values{}
	if fields != "" {
		q.Set("fields", fields)
	}
	if expand != "" {
		q.Set("expand", expand)
	}
	return q
}

// EditIssue applies field/update changes. notifyUsers controls watcher emails.
func (c *Client) EditIssue(key string, update IssueUpdate, notifyUsers bool) error {
	q := url.Values{}
	if !notifyUsers {
		q.Set("notifyUsers", "false")
	}
	return c.PutJSON(c.api("issue/%s", key), q, update, nil)
}

// DeleteIssue deletes an issue (optionally its subtasks).
func (c *Client) DeleteIssue(key string, deleteSubtasks bool) error {
	q := url.Values{}
	if deleteSubtasks {
		q.Set("deleteSubtasks", "true")
	}
	return c.Delete(c.api("issue/%s", key), q)
}

// AssignIssue assigns an issue. accountID may be a real id, "-1" (default
// assignee), or nil (unassign).
func (c *Client) AssignIssue(key string, accountID any) error {
	return c.PutJSON(c.api("issue/%s/assignee", key), nil, map[string]any{"accountId": accountID}, nil)
}

// GetTransitions lists the transitions available on an issue (with screen fields).
func (c *Client) GetTransitions(key string) ([]Transition, error) {
	var out struct {
		Transitions []Transition `json:"transitions"`
	}
	q := url.Values{"expand": {"transitions.fields"}}
	if err := c.GetJSON(c.api("issue/%s/transitions", key), q, &out); err != nil {
		return nil, err
	}
	return out.Transitions, nil
}

// DoTransition performs a transition, optionally setting fields/update (e.g. a
// resolution or a comment) on the transition screen.
func (c *Client) DoTransition(key, transitionID string, fields, update map[string]any) error {
	body := IssueUpdate{Transition: map[string]any{"id": transitionID}}
	if len(fields) > 0 {
		body.Fields = fields
	}
	if len(update) > 0 {
		body.Update = update
	}
	return c.PostJSON(c.api("issue/%s/transitions", key), nil, body, nil)
}

// FieldMeta describes a field on the create/edit screen.
type FieldMeta struct {
	Required      bool     `json:"required"`
	Name          string   `json:"name"`
	Key           string   `json:"key"`
	HasDefault    bool     `json:"hasDefaultValue"`
	Operations    []string `json:"operations"`
	AllowedValues []any    `json:"allowedValues,omitempty"`
	Schema        struct {
		Type   string `json:"type"`
		Items  string `json:"items"`
		System string `json:"system"`
		Custom string `json:"custom"`
	} `json:"schema"`
}

// CreateMetaIssueTypes lists creatable issue types for a project.
func (c *Client) CreateMetaIssueTypes(projectKey string) ([]IssueType, error) {
	var out struct {
		IssueTypes []IssueType `json:"issueTypes"`
	}
	q := url.Values{"maxResults": {"200"}}
	if err := c.GetJSON(c.api("issue/createmeta/%s/issuetypes", projectKey), q, &out); err != nil {
		return nil, err
	}
	return out.IssueTypes, nil
}

// CreateMetaFields lists the create-screen fields for a project + issue type.
func (c *Client) CreateMetaFields(projectKey, issueTypeID string) (map[string]FieldMeta, error) {
	var out struct {
		Fields []struct {
			FieldMeta
			FieldID string `json:"fieldId"`
		} `json:"fields"`
	}
	q := url.Values{"maxResults": {"200"}}
	if err := c.GetJSON(c.api("issue/createmeta/%s/issuetypes/%s", projectKey, issueTypeID), q, &out); err != nil {
		return nil, err
	}
	m := make(map[string]FieldMeta, len(out.Fields))
	for _, f := range out.Fields {
		id := f.FieldID
		if id == "" {
			id = f.Key
		}
		m[id] = f.FieldMeta
	}
	return m, nil
}

// itoa is a tiny helper for building query values.
func itoa(n int) string { return strconv.Itoa(n) }
