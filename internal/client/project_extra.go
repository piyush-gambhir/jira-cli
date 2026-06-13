package client

// ProjectStatusType is one entry of GET /project/{key}/statuses: an issue type
// together with the statuses available to it in the project's workflow.
type ProjectStatusType struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Subtask  bool     `json:"subtask,omitempty"`
	Statuses []Status `json:"statuses,omitempty"`
}

// ProjectRoleActor is a member of a project role (a user or a group).
type ProjectRoleActor struct {
	ID          int    `json:"id,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	ActorUser   *struct {
		AccountID string `json:"accountId,omitempty"`
	} `json:"actorUser,omitempty"`
	ActorGroup *struct {
		Name        string `json:"name,omitempty"`
		DisplayName string `json:"displayName,omitempty"`
		GroupID     string `json:"groupId,omitempty"`
	} `json:"actorGroup,omitempty"`
}

// ProjectRole is the detail of a single project role (GET /project/{key}/role/{id}).
type ProjectRole struct {
	ID          int                `json:"id,omitempty"`
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	Self        string             `json:"self,omitempty"`
	Actors      []ProjectRoleActor `json:"actors,omitempty"`
}

// ProjectCategory is a project category.
type ProjectCategory struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
}

// ProjectStatuses returns, for each issue type in the project, the statuses
// available to it (GET /rest/api/3/project/{key}/statuses).
func (c *Client) ProjectStatuses(idOrKey string) ([]ProjectStatusType, error) {
	var out []ProjectStatusType
	if err := c.GetJSON(c.api("project/%s/statuses", idOrKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ArchiveProject archives a project (POST /project/{key}/archive).
func (c *Client) ArchiveProject(idOrKey string) error {
	return c.PostJSON(c.api("project/%s/archive", idOrKey), nil, nil, nil)
}

// RestoreProject restores an archived (or trashed) project (POST /project/{key}/restore).
func (c *Client) RestoreProject(idOrKey string) error {
	return c.PostJSON(c.api("project/%s/restore", idOrKey), nil, nil, nil)
}

// ProjectRoles returns the project's roles as a name -> role-detail-URL map
// (GET /project/{key}/role).
func (c *Client) ProjectRoles(idOrKey string) (map[string]string, error) {
	var out map[string]string
	if err := c.GetJSON(c.api("project/%s/role", idOrKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectRole returns a single project role with its actors
// (GET /project/{key}/role/{id}).
func (c *Client) GetProjectRole(idOrKey, roleID string) (*ProjectRole, error) {
	var out ProjectRole
	if err := c.GetJSON(c.api("project/%s/role/%s", idOrKey, roleID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ProjectCategories returns all project categories (GET /rest/api/3/projectCategory).
func (c *Client) ProjectCategories() ([]ProjectCategory, error) {
	var out []ProjectCategory
	if err := c.GetJSON(c.api("projectCategory"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProjectCategory creates a project category (POST /projectCategory).
func (c *Client) CreateProjectCategory(body map[string]any) (*ProjectCategory, error) {
	var out ProjectCategory
	if err := c.PostJSON(c.api("projectCategory"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteProjectCategory deletes a project category (DELETE /projectCategory/{id}).
func (c *Client) DeleteProjectCategory(id string) error {
	return c.Delete(c.api("projectCategory/%s", id), nil)
}
