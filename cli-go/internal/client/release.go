package client

import (
	"fmt"
	"strconv"
)

// Version is a Jira project version (the UI calls these "Releases").
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

// VersionIssueCounts is the response of GET /version/{id}/relatedIssueCounts.
type VersionIssueCounts struct {
	Self                                     string `json:"self,omitempty"`
	IssuesFixedCount                         int    `json:"issuesFixedCount"`
	IssuesAffectedCount                      int    `json:"issuesAffectedCount"`
	IssueCountWithCustomFieldsShowingVersion int    `json:"issueCountWithCustomFieldsShowingVersion"`
}

// VersionUnresolvedCount is the response of GET /version/{id}/unresolvedIssueCount.
type VersionUnresolvedCount struct {
	Self                  string `json:"self,omitempty"`
	IssuesUnresolvedCount int    `json:"issuesUnresolvedCount"`
	IssuesCount           int    `json:"issuesCount"`
}

// ListVersions returns the versions of a project (array endpoint).
func (c *Client) ListVersions(projectIDOrKey string) ([]Version, error) {
	var out []Version
	if err := c.GetJSON(c.api("project/%s/versions", projectIDOrKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetVersion returns a single version by id.
func (c *Client) GetVersion(id string) (*Version, error) {
	var out Version
	if err := c.GetJSON(c.api("version/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateVersion creates a project version. body must carry at least "name" and
// "projectId".
func (c *Client) CreateVersion(body map[string]any) (*Version, error) {
	var out Version
	if err := c.PostJSON(c.api("version"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateVersion updates a version's fields and returns the updated resource.
func (c *Client) UpdateVersion(id string, body map[string]any) (*Version, error) {
	var out Version
	if err := c.PutJSON(c.api("version/%s", id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MergeVersion merges version id into target version targetID (moving fix/affects
// references), then deletes id. This endpoint returns 204 No Content.
func (c *Client) MergeVersion(id, targetID string) error {
	return c.PutJSON(c.api("version/%s/mergeto/%s", id, targetID), nil, nil, nil)
}

// DeleteVersionAndSwap deletes a version, optionally re-pointing the issues that
// reference it. moveFixTo/moveAffectedTo are version ids; pass "" to skip. This
// endpoint returns 204 No Content.
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

// VersionRelatedCounts returns the related issue counts for a version.
func (c *Client) VersionRelatedCounts(id string) (*VersionIssueCounts, error) {
	var out VersionIssueCounts
	if err := c.GetJSON(c.api("version/%s/relatedIssueCounts", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VersionUnresolvedIssues returns the count of unresolved issues for a version.
func (c *Client) VersionUnresolvedIssues(id string) (*VersionUnresolvedCount, error) {
	var out VersionUnresolvedCount
	if err := c.GetJSON(c.api("version/%s/unresolvedIssueCount", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResolveProjectID turns a project id-or-key into the numeric projectId required
// by POST /version. A numeric input is returned as-is; a key is resolved via
// GetProject.
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
