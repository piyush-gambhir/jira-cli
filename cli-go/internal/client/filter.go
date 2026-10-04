package client

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
)

// Filter is a saved JQL filter.
type Filter struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Owner       *User  `json:"owner,omitempty"`
	JQL         string `json:"jql,omitempty"`
	Favourite   bool   `json:"favourite,omitempty"`
	ViewURL     string `json:"viewUrl,omitempty"`
	Self        string `json:"self,omitempty"`
}

// ListMyFilters returns the filters owned by the caller (favourite state
// expanded). Server/DC has no filter/my.
func (c *Client) ListMyFilters() ([]Filter, error) {
	if c.IsServer() {
		return nil, errCloudOnly("listing your own filters", "on Server/Data Center use `jira filter favourites`")
	}
	var out []Filter
	if err := c.GetJSON(c.api("filter/my"), url.Values{"expand": {"favourite"}}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListFavouriteFilters returns the caller's favourite filters.
func (c *Client) ListFavouriteFilters() ([]Filter, error) {
	var out []Filter
	if err := c.GetJSON(c.api("filter/favourite"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchFilters finds filters by (partial) name. Returns at most `limit`
// results. Server/DC has no filter/search.
func (c *Client) SearchFilters(name string, limit int) ([]Filter, error) {
	if c.IsServer() {
		return nil, errCloudOnly("filter search", "on Server/Data Center use `jira filter favourites` or `jira filter get <id>`")
	}
	q := url.Values{}
	if name != "" {
		q.Set("name", name)
	}
	if limit > 0 {
		q.Set("maxResults", itoa(limit))
	}
	q.Set("expand", "jql,owner,favourite,viewUrl,description")
	var out struct {
		Values []Filter `json:"values"`
	}
	if err := c.GetJSON(c.api("filter/search"), q, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// GetFilter returns a single filter by id.
func (c *Client) GetFilter(id string) (*Filter, error) {
	var out Filter
	if err := c.GetJSON(c.api("filter/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateFilter creates a saved filter. body holds name, jql, description, favourite.
func (c *Client) CreateFilter(body map[string]any) (*Filter, error) {
	var out Filter
	if err := c.PostJSON(c.api("filter"), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateFilter updates a filter's fields and returns the updated filter.
// Server/DC's PUT filter ignores "favourite", so there it is set separately.
func (c *Client) UpdateFilter(id string, body map[string]any) (*Filter, error) {
	if fav, ok := body["favourite"].(bool); ok && c.IsServer() {
		rest := maps.Clone(body)
		delete(rest, "favourite")
		if len(rest) > 0 {
			if _, err := c.UpdateFilter(id, rest); err != nil {
				return nil, err
			}
		}
		if fav {
			return c.FavouriteFilter(id)
		}
		return c.UnfavouriteFilter(id)
	}
	var out Filter
	if err := c.PutJSON(c.api("filter/%s", id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteFilter deletes a filter by id.
func (c *Client) DeleteFilter(id string) error {
	return c.Delete(c.api("filter/%s", id), nil)
}

// FavouriteFilter marks a filter as a favourite and returns the updated filter.
func (c *Client) FavouriteFilter(id string) (*Filter, error) {
	if c.IsServer() {
		if err := c.setFilterFavouriteServer(http.MethodPut, id); err != nil {
			return nil, err
		}
		return c.GetFilter(id)
	}
	var out Filter
	if err := c.PutJSON(c.api("filter/%s/favourite", id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UnfavouriteFilter removes a filter from the caller's favourites.
func (c *Client) UnfavouriteFilter(id string) (*Filter, error) {
	if c.IsServer() {
		if err := c.setFilterFavouriteServer(http.MethodDelete, id); err != nil {
			return nil, err
		}
		return c.GetFilter(id)
	}
	if err := c.Delete(c.api("filter/%s/favourite", id), nil); err != nil {
		return nil, err
	}
	return c.GetFilter(id)
}

// setFilterFavouriteServer stars (PUT) or unstars (DELETE) a filter on
// Server/DC. REST v2 has no favourite endpoint; the Data Center reference points
// to rest/api/1.0/filters/{id}/favourite instead. The bodyless write carries
// X-Atlassian-Token so Jira's XSRF check accepts it.
func (c *Client) setFilterFavouriteServer(method, id string) error {
	path := fmt.Sprintf("/rest/api/1.0/filters/%s/favourite", id)
	return c.doJSONHeader(method, path, nil, nil, nil, true, http.Header{"X-Atlassian-Token": {"no-check"}})
}
