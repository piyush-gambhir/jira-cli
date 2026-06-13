package client

import "net/url"

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

// ListMyFilters returns the filters owned by the caller (favourite state expanded).
func (c *Client) ListMyFilters() ([]Filter, error) {
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

// SearchFilters finds filters by (partial) name. Returns at most `limit` results.
func (c *Client) SearchFilters(name string, limit int) ([]Filter, error) {
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
func (c *Client) UpdateFilter(id string, body map[string]any) (*Filter, error) {
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
	var out Filter
	if err := c.PutJSON(c.api("filter/%s/favourite", id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UnfavouriteFilter removes a filter from the caller's favourites.
func (c *Client) UnfavouriteFilter(id string) (*Filter, error) {
	if err := c.Delete(c.api("filter/%s/favourite", id), nil); err != nil {
		return nil, err
	}
	return c.GetFilter(id)
}
