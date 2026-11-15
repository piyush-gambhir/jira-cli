package client

type ProjectStatusType struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Subtask  bool     `json:"subtask,omitempty"`
	Statuses []Status `json:"statuses,omitempty"`
}

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

type ProjectRole struct {
	ID          int                `json:"id,omitempty"`
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	Self        string             `json:"self,omitempty"`
	Actors      []ProjectRoleActor `json:"actors,omitempty"`
}

type ProjectCategory struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
}

func (c *Client) ProjectStatuses(idOrKey string) ([]ProjectStatusType, error) {
	var out []ProjectStatusType
	if err := c.GetJSON(c.api("project/%s/statuses", idOrKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ArchiveProject(idOrKey string) error {
	if c.IsServer() {
		return c.PutJSON(c.api("project/%s/archive", idOrKey), nil, nil, nil)
	}
	return c.PostJSON(c.api("project/%s/archive", idOrKey), nil, nil, nil)
}

func (c *Client) RestoreProject(idOrKey string) error {
	if c.IsServer() {
		return c.PutJSON(c.api("project/%s/restore", idOrKey), nil, nil, nil)
	}
	return c.PostJSON(c.api("project/%s/restore", idOrKey), nil, nil, nil)
}

func (c *Client) ProjectRoles(idOrKey string) (map[string]string, error) {
	var out map[string]string
	if err := c.GetJSON(c.api("project/%s/role", idOrKey), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetProjectRole(idOrKey, roleID string) (*ProjectRole, error) {
	var out ProjectRole
	if err := c.GetJSON(c.api("project/%s/role/%s", idOrKey, roleID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ProjectCategories() ([]ProjectCategory, error) {
	var out []ProjectCategory
	if err := c.GetJSON(c.api("projectCategory"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateProjectCategory(body map[string]any) (*ProjectCategory, error) {
	var out ProjectCategory
	if err := c.PostJSON(c.api("projectCategory"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteProjectCategory(id string) error {
	return c.Delete(c.api("projectCategory/%s", id), nil)
}
