package client

import (
	"fmt"
	"strconv"
)

type Version struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Released    bool   `json:"released"`
	Archived    bool   `json:"archived"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	ProjectID   int    `json:"projectId,omitempty"`
	Self        string `json:"self,omitempty"`
}

type VersionIssueCounts struct {
	Self                                     string `json:"self,omitempty"`
	IssuesFixedCount                         int    `json:"issuesFixedCount"`
	IssuesAffectedCount                      int    `json:"issuesAffectedCount"`
	IssueCountWithCustomFieldsShowingVersion int    `json:"issueCountWithCustomFieldsShowingVersion"`
}

type VersionUnresolvedCount struct {
	Self                  string `json:"self,omitempty"`
	IssuesUnresolvedCount int    `json:"issuesUnresolvedCount"`
	IssuesCount           int    `json:"issuesCount"`
}

func (c *Client) ListVersions(projectIDOrKey string) ([]Version, error) {
	var out []Version
	if err := c.GetJSON(c.api("project/%s/versions", projectIDOrKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetVersion(id string) (*Version, error) {
	var out Version
	if err := c.GetJSON(c.api("version/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateVersion(body map[string]any) (*Version, error) {
	var out Version
	if err := c.PostJSON(c.api("version"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateVersion(id string, body map[string]any) (*Version, error) {
	var out Version
	if err := c.PutJSON(c.api("version/%s", id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) MergeVersion(id, targetID string) error {
	return c.PutJSON(c.api("version/%s/mergeto/%s", id, targetID), nil, nil, nil)
}

func (c *Client) DeleteVersionAndSwap(id, moveFixTo, moveAffectedTo string) error {
	body := map[string]any{}
	if moveFixTo != "" {
		body["moveFixIssuesTo"] = moveFixTo
	}
	if moveAffectedTo != "" {
		body["moveAffectedIssuesTo"] = moveAffectedTo
	}
	return c.PostJSON(c.api("version/%s/removeAndSwap", id), nil, body, nil)
}

func (c *Client) VersionRelatedCounts(id string) (*VersionIssueCounts, error) {
	var out VersionIssueCounts
	if err := c.GetJSON(c.api("version/%s/relatedIssueCounts", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) VersionUnresolvedIssues(id string) (*VersionUnresolvedCount, error) {
	var out VersionUnresolvedCount
	if err := c.GetJSON(c.api("version/%s/unresolvedIssueCount", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ResolveProjectID(projectIDOrKey string) (int, error) {
	if n, err := strconv.Atoi(projectIDOrKey); err == nil {
		return n, nil
	}
	p, err := c.GetProject(projectIDOrKey)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(p.ID)
	if err != nil {
		return 0, fmt.Errorf("could not parse project id %q: %w", p.ID, err)
	}
	return n, nil
}
