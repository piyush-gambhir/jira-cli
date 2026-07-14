package main

import (
	"errors"
	"os"

	"github.com/piyush-gambhir/jira-cli/cli-go/cmd"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func main() {
	if err := cmd.RootCmd().Execute(); err != nil {
		statusCode := 0
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			statusCode = apiErr.StatusCode
		}
		// Use the output format set during PersistentPreRunE via Cobra flag parsing
		outFmt, _ := output.ParseFormat(cmd.OutputFormat)
		output.WriteError(os.Stderr, outFmt, err, statusCode)
		os.Exit(1)
	}
}
