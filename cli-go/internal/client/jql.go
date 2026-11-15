package client

import (
	"encoding/json"
	"net/url"
)

type JQLFieldRef struct {
	Value       string   `json:"value"`
	DisplayName string   `json:"displayName"`
	Operators   []string `json:"operators,omitempty"`
	Types       []string `json:"types,omitempty"`
}

type JQLFunctionRef struct {
	Value       string   `json:"value"`
	DisplayName string   `json:"displayName"`
	IsList      string   `json:"isList,omitempty"`
	Types       []string `json:"types,omitempty"`
}

type JQLAutocompleteData struct {
	VisibleFieldNames    []JQLFieldRef    `json:"visibleFieldNames"`
	VisibleFunctionNames []JQLFunctionRef `json:"visibleFunctionNames"`
	JQLReservedWords     []string         `json:"jqlReservedWords"`
}

type JQLSuggestion struct {
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
}

type JQLParsedError struct {
	Query     string          `json:"query"`
	Structure json.RawMessage `json:"structure,omitempty"`
	Errors    []string        `json:"errors,omitempty"`
}

func (c *Client) JQLAutocompleteData() (*JQLAutocompleteData, error) {
	var out JQLAutocompleteData
	if err := c.GetJSON(c.api("jql/autocompletedata"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

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

func (c *Client) JQLParse(queries []string, validation string) ([]JQLParsedError, error) {
	if c.IsServer() {
		return nil, errCloudOnly("jql parse", "on Server/Data Center run the query with `jira issue search` to validate it")
	}
	q := url.Values{}
	if validation != "" {
		q.Set("validation", validation)
	}
	body := map[string]any{"queries": queries}
	var out struct {
		Queries []JQLParsedError `json:"queries"`
	}
	if err := c.PostJSONRetryable(c.api("jql/parse"), q, body, &out); err != nil {
		return nil, err
	}
	return out.Queries, nil
}
