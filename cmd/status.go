package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

type statusResult struct {
	Site          string             `json:"site"`
	Authenticated bool               `json:"authenticated"`
	User          string             `json:"user,omitempty"`
	Auth          string             `json:"auth"`
	ServerInfo    *client.ServerInfo `json:"serverInfo,omitempty"`
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show site connectivity and authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			res := statusResult{Site: jiraClient.BaseURL(), Auth: jiraClient.AuthDescription()}

			si, siErr := jiraClient.ServerInfo()
			if siErr == nil {
				res.ServerInfo = si
			}
			if me, err := jiraClient.Myself(); err == nil {
				res.Authenticated = true
				res.User = me.DisplayName
			}

			def := &output.TableDef{
				Headers: []string{"SITE", "VERSION", "DEPLOYMENT", "AUTHENTICATED", "USER"},
				RowFunc: func(item interface{}) []string {
					r := item.(statusResult)
					ver, dep := "-", "-"
					if r.ServerInfo != nil {
						ver, dep = dash(r.ServerInfo.Version), dash(r.ServerInfo.DeploymentType)
					}
					return []string{r.Site, ver, dep, fmt.Sprintf("%v", r.Authenticated), dash(r.User)}
				},
			}
			if err := render(res, def); err != nil {
				return err
			}
			if !res.Authenticated {
				return fmt.Errorf("not authenticated for %s (run 'jira auth login')", res.Site)
			}
			return nil
		},
	}
}
