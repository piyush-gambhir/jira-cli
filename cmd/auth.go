package cmd

import (
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication profiles",
		Long: `Authenticate to Jira and manage connection profiles.

Supported auth methods (see docs/CREDENTIALS.md for the full reference):
  api_token     Cloud — Basic auth with an Atlassian API token (default)
  scoped_token  Cloud — Basic auth with a scoped API token (api.atlassian.com gateway)
  oauth2        Cloud — OAuth 2.0 (3LO) browser flow with refresh tokens
  pat           Server/Data Center — Bearer personal access token
  basic         Server/Data Center — Basic auth with username + password

Examples:
  jira auth login                          # interactive Cloud API-token login
  jira auth login --type pat --site https://jira.company.com
  jira auth login --type oauth2 --client-id ... --client-secret ...
  jira auth list
  jira auth use staging
  jira auth logout --profile staging`,
	}
	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthListCmd())
	cmd.AddCommand(newAuthUseCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	return cmd
}
