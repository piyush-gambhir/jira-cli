package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code,omitempty"`
}

func WriteError(w io.Writer, format Format, err error, statusCode int) {
	if format == FormatJSON {
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error(), StatusCode: statusCode})
	} else {
		fmt.Fprintf(w, "Error: %v\n", err)
	}
}
