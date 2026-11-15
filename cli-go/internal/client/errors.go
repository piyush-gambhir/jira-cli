package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

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
	if len(e.Messages) == 0 && len(e.FieldErrors) == 0 && len(e.Raw) > 500 {
		e.Raw = e.Raw[:500] + "..."
	}
	return e
}
