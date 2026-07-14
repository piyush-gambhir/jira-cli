package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIError represents an error response from the Jira REST API. Jira returns a
// consistent envelope: {"errorMessages":[...], "errors":{"field":"msg"}}.
type APIError struct {
	StatusCode  int
	Status      string
	Messages    []string
	FieldErrors map[string]string
	URL         string
	Raw         string
}

func (e *APIError) Error() string {
	var parts []string
	parts = append(parts, e.Messages...)
	// Stable order isn't guaranteed for a map, but error display doesn't need it.
	for f, m := range e.FieldErrors {
		parts = append(parts, fmt.Sprintf("%s: %s", f, m))
	}
	msg := strings.Join(parts, "; ")
	if msg == "" {
		msg = e.Raw
	}
	if msg == "" {
		return fmt.Sprintf("jira API error: %s (status %d) url=%s", e.Status, e.StatusCode, e.URL)
	}
	return fmt.Sprintf("jira API error (status %d): %s", e.StatusCode, msg)
}

// parseAPIError builds an APIError from a non-2xx response body.
func parseAPIError(statusCode int, status, url string, body []byte) *APIError {
	e := &APIError{StatusCode: statusCode, Status: status, URL: url, Raw: strings.TrimSpace(string(body))}
	var env struct {
		ErrorMessages []string          `json:"errorMessages"`
		Errors        map[string]string `json:"errors"`
	}
	if json.Unmarshal(body, &env) == nil {
		e.Messages = env.ErrorMessages
		e.FieldErrors = env.Errors
	}
	// If the body wasn't the JSON envelope (e.g. an HTML gateway/proxy page),
	// keep a truncated raw snippet so the message is still useful.
	if len(e.Messages) == 0 && len(e.FieldErrors) == 0 && len(e.Raw) > 500 {
		e.Raw = e.Raw[:500] + "..."
	}
	return e
}
