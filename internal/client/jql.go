package client

import (
	"encoding/json"
	"net/url"
)

// JQLFieldRef is a field that can be used in JQL, as returned by the
// autocomplete data endpoint (value is the JQL clause name).
type JQLFieldRef struct {
	Value       string   `json:"value"`
	DisplayName string   `json:"displayName"`
	Operators   []string `json:"operators,omitempty"`
	Types       []string `json:"types,omitempty"`
}

// JQLFunctionRef is a JQL function reference from the autocomplete data.
type JQLFunctionRef struct {
	Value       string   `json:"value"`
	DisplayName string   `json:"displayName"`
	IsList      string   `json:"isList,omitempty"`
	Types       []string `json:"types,omitempty"`
}

// JQLAutocompleteData is the response of GET /jql/autocompletedata: the fields,
// functions and reserved words usable when composing JQL.
type JQLAutocompleteData struct {
	VisibleFieldNames    []JQLFieldRef    `json:"visibleFieldNames"`
	VisibleFunctionNames []JQLFunctionRef `json:"visibleFunctionNames"`
	JQLReservedWords     []string         `json:"jqlReservedWords"`
}

// JQLSuggestion is a single suggested value for a field (autocomplete).
type JQLSuggestion struct {
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
}

// JQLParsedError describes a problem with a parsed query (empty when valid).
type JQLParsedError struct {
	Query     string          `json:"query"`
	Structure json.RawMessage `json:"structure,omitempty"`
	Errors    []string        `json:"errors,omitempty"`
}

// JQLAutocompleteData returns the fields, functions and reserved words that can
// be used to build JQL queries on this site.
func (c *Client) JQLAutocompleteData() (*JQLAutocompleteData, error) {
	var out JQLAutocompleteData
	if err := c.GetJSON(c.api("jql/autocompletedata"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// JQLSuggestions returns autocomplete suggestions for a field. fieldValue is an
// optional value prefix; predicateName is an optional predicate (e.g. "by").
func (c *Client) JQLSuggestions(fieldName, fieldValue, predicateName string) ([]JQLSuggestion, error) {
	q := url.Values{}
	if fieldName != "" {
		q.Set("fieldName", fieldName)
	}
	if fieldValue != "" {
		q.Set("fieldValue", fieldValue)
	}
	if predicateName != "" {
		q.Set("predicateName", predicateName)
	}
	var out struct {
		Results []JQLSuggestion `json:"results"`
	}
	if err := c.GetJSON(c.api("jql/autocompletedata/suggestions"), q, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// JQLParse parses (and optionally validates) one or more JQL queries. validation
// is one of "strict", "warn" or "none" (empty uses the server default). Each
// returned entry carries the parsed structure and any per-query errors.
func (c *Client) JQLParse(queries []string, validation string) ([]JQLParsedError, error) {
	q := url.Values{}
	if validation != "" {
		q.Set("validation", validation)
	}
	body := map[string]any{"queries": queries}
	var out struct {
		Queries []JQLParsedError `json:"queries"`
	}
	if err := c.PostJSON(c.api("jql/parse"), q, body, &out); err != nil {
		return nil, err
	}
	return out.Queries, nil
}
