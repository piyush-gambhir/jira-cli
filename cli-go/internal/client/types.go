package client

import "strings"

// flexString unmarshals from a JSON string OR number. Some Jira endpoints return
// ids both ways (e.g. attachment id is a string in an issue's attachment list but
// a number from GET /attachment/{id}).
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		*f = ""
		return nil
	}
	*f = flexString(strings.Trim(s, `"`))
	return nil
}

// User is a Jira user (identified by accountId on Cloud).
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	AccountType  string `json:"accountType,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Active       bool   `json:"active"`
	TimeZone     string `json:"timeZone,omitempty"`
	Self         string `json:"self,omitempty"`
}

// Named is the common {id,name} shape (priority, issuetype, resolution, …).
type Named struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// StatusCategory is the workflow-independent bucket of a status.
type StatusCategory struct {
	Key       string `json:"key,omitempty"`
	Name      string `json:"name,omitempty"`
	ColorName string `json:"colorName,omitempty"`
}

// Status is an issue status.
type Status struct {
	ID             string          `json:"id,omitempty"`
	Name           string          `json:"name,omitempty"`
	StatusCategory *StatusCategory `json:"statusCategory,omitempty"`
}

// Project is a Jira project (subset of fields).
type Project struct {
	ID             string `json:"id,omitempty"`
	Key            string `json:"key,omitempty"`
	Name           string `json:"name,omitempty"`
	ProjectTypeKey string `json:"projectTypeKey,omitempty"`
	Style          string `json:"style,omitempty"`
	Simplified     bool   `json:"simplified,omitempty"`
	Lead           *User  `json:"lead,omitempty"`
	Self           string `json:"self,omitempty"`
}

// IssueFields holds the common navigable fields of an issue.
type IssueFields struct {
	Summary     string   `json:"summary,omitempty"`
	Status      *Status  `json:"status,omitempty"`
	Priority    *Named   `json:"priority,omitempty"`
	IssueType   *Named   `json:"issuetype,omitempty"`
	Assignee    *User    `json:"assignee,omitempty"`
	Reporter    *User    `json:"reporter,omitempty"`
	Project     *Project `json:"project,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
	DueDate     string   `json:"duedate,omitempty"`
	Resolution  *Named   `json:"resolution,omitempty"`
	Parent      *Issue   `json:"parent,omitempty"`
	Description any      `json:"description,omitempty"` // ADF (Cloud v3) or string (v2)
}

// Issue is a Jira issue (typed subset; use GetIssueRaw for full fidelity).
type Issue struct {
	ID     string      `json:"id,omitempty"`
	Key    string      `json:"key,omitempty"`
	Self   string      `json:"self,omitempty"`
	Fields IssueFields `json:"fields,omitempty"`
}

// Transition is an available workflow transition for an issue.
type Transition struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	To        *Status `json:"to,omitempty"`
	HasScreen bool    `json:"hasScreen,omitempty"`
}

// Comment is an issue comment (body is ADF on Cloud v3).
type Comment struct {
	ID      string `json:"id,omitempty"`
	Self    string `json:"self,omitempty"`
	Author  *User  `json:"author,omitempty"`
	Body    any    `json:"body,omitempty"`
	Created string `json:"created,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// Worklog is a worklog entry.
type Worklog struct {
	ID               string `json:"id,omitempty"`
	Author           *User  `json:"author,omitempty"`
	Comment          any    `json:"comment,omitempty"`
	Started          string `json:"started,omitempty"`
	TimeSpent        string `json:"timeSpent,omitempty"`
	TimeSpentSeconds int    `json:"timeSpentSeconds,omitempty"`
	Created          string `json:"created,omitempty"`
}

// Attachment is issue attachment metadata.
type Attachment struct {
	ID        flexString `json:"id,omitempty"`
	Filename  string     `json:"filename,omitempty"`
	Author    *User      `json:"author,omitempty"`
	Created   string     `json:"created,omitempty"`
	Size      int64      `json:"size,omitempty"`
	MimeType  string     `json:"mimeType,omitempty"`
	Content   string     `json:"content,omitempty"`
	Thumbnail string     `json:"thumbnail,omitempty"`
}

// IssueType is an issue type.
type IssueType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
}

// Field is a system or custom field definition.
type Field struct {
	ID          string   `json:"id"`
	Key         string   `json:"key,omitempty"`
	Name        string   `json:"name"`
	Custom      bool     `json:"custom"`
	Navigable   bool     `json:"navigable"`
	Searchable  bool     `json:"searchable"`
	ClauseNames []string `json:"clauseNames,omitempty"`
}

// LinkType is an issue link type with its directional labels.
type LinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

// Board is an Agile board.
type Board struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Self string `json:"self,omitempty"`
}

// Sprint is an Agile sprint.
type Sprint struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	State         string `json:"state"`
	StartDate     string `json:"startDate,omitempty"`
	EndDate       string `json:"endDate,omitempty"`
	CompleteDate  string `json:"completeDate,omitempty"`
	Goal          string `json:"goal,omitempty"`
	OriginBoardID int    `json:"originBoardId,omitempty"`
}

// ServerInfo is the response of GET /serverInfo.
type ServerInfo struct {
	BaseURL        string `json:"baseUrl"`
	Version        string `json:"version"`
	DeploymentType string `json:"deploymentType"`
	BuildNumber    int    `json:"buildNumber"`
	ServerTitle    string `json:"serverTitle"`
	ServerTime     string `json:"serverTime"`
}

// Watchers is the response of GET issue watchers.
type Watchers struct {
	WatchCount int    `json:"watchCount"`
	IsWatching bool   `json:"isWatching"`
	Watchers   []User `json:"watchers"`
}

// Votes is the response of GET issue votes.
type Votes struct {
	Votes    int    `json:"votes"`
	HasVoted bool   `json:"hasVoted"`
	Voters   []User `json:"voters"`
}
