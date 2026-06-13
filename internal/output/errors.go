package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// ErrorResponse represents a structured error for JSON output.
type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code,omitempty"`
}

// WriteError writes an error in the appropriate format. When format is JSON,
// it writes a structured JSON error object. Otherwise it writes a plain text
// error message.
func WriteError(w io.Writer, format Format, err error, statusCode int) {
	if format == FormatJSON {
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error(), StatusCode: statusCode})
	} else {
		fmt.Fprintf(w, "Error: %v\n", err)
	}
}
