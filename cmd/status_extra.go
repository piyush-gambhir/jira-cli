package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/config"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

// statusCategoryName renders a status' workflow category for table cells.
func statusCategoryName(s client.Status) string {
	if s.StatusCategory == nil {
		return "-"
	}
	if s.StatusCategory.Name != "" {
		return s.StatusCategory.Name
	}
	return dash(s.StatusCategory.Key)
}

func statusTableDef() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "CATEGORY"},
		RowFunc: func(item interface{}) []string {
			s := item.(client.Status)
			return []string{dash(s.ID), dash(s.Name), statusCategoryName(s)}
		},
	}
}

func newStatusListCmd() *cobra.Command {
	var statusProject string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List issue statuses",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			statuses, err := jiraClient.ListStatuses()
			if err != nil {
				return err
			}
			// The /status endpoint returns instance-wide statuses (no project
			// field), so --project is accepted but only emits a note when set.
			if statusProject != "" {
				info("Note: --project does not filter the instance status list; showing all statuses.")
			}
			return render(statuses, statusTableDef())
		},
	}
	cmd.Flags().StringVar(&statusProject, "project", "", "Project key (informational; status list is instance-wide)")
	return cmd
}

func newStatusGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <idOrName>",
		Short: "Show a single issue status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := jiraClient.GetStatus(args[0])
			if err != nil {
				return err
			}
			return render(*s, statusTableDef())
		},
	}
}

// serverInfoResult is the rendered view of GET /serverInfo.
type serverInfoResult struct {
	BaseURL        string `json:"baseUrl"`
	Version        string `json:"version"`
	DeploymentType string `json:"deploymentType"`
	BuildNumber    int    `json:"buildNumber"`
	ServerTitle    string `json:"serverTitle,omitempty"`
	ServerTime     string `json:"serverTime,omitempty"`
}

func newServerInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serverinfo",
		Short: "Show Jira instance metadata",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			si, err := jiraClient.ServerInfo()
			if err != nil {
				return err
			}
			res := serverInfoResult{
				BaseURL:        si.BaseURL,
				Version:        si.Version,
				DeploymentType: si.DeploymentType,
				BuildNumber:    si.BuildNumber,
				ServerTitle:    si.ServerTitle,
				ServerTime:     si.ServerTime,
			}
			def := &output.TableDef{
				Headers: []string{"BASE URL", "VERSION", "DEPLOYMENT", "BUILD", "SERVER TIME"},
				RowFunc: func(item interface{}) []string {
					r := item.(serverInfoResult)
					return []string{
						dash(r.BaseURL),
						dash(r.Version),
						dash(r.DeploymentType),
						strconv.Itoa(r.BuildNumber),
						dash(r.ServerTime),
					}
				},
			}
			return render(res, def)
		},
	}
}

// authCurrentResult is the rendered view of the active profile.
type authCurrentResult struct {
	Name     string `json:"name"`
	Site     string `json:"site"`
	AuthType string `json:"auth_type"`
	Account  string `json:"account,omitempty"`
}

func newAuthCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the active authentication profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Load()
			if err != nil {
				return err
			}
			name := c.CurrentProfile
			if name == "" {
				return fmt.Errorf("no current profile set (run 'jira auth login' or 'jira auth use <profile>')")
			}
			p, ok := c.Profiles[name]
			if !ok {
				return fmt.Errorf("current profile %q not found in profiles", name)
			}
			acct := p.Email
			if acct == "" {
				acct = p.Username
			}
			res := authCurrentResult{
				Name:     name,
				Site:     p.Site,
				AuthType: p.EffectiveAuthType(),
				Account:  acct,
			}
			def := &output.TableDef{
				Headers: []string{"NAME", "SITE", "AUTH TYPE", "ACCOUNT"},
				RowFunc: func(item interface{}) []string {
					r := item.(authCurrentResult)
					return []string{dash(r.Name), dash(r.Site), dash(r.AuthType), dash(r.Account)}
				},
			}
			return render(res, def)
		},
	}
}

func newAuthRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <oldName> <newName>",
		Short: "Rename an authentication profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName, newName := args[0], args[1]
			if err := config.Update(func(c *config.Config) error {
				p, ok := c.Profiles[oldName]
				if !ok {
					return fmt.Errorf("profile %q not found", oldName)
				}
				if _, exists := c.Profiles[newName]; exists {
					return fmt.Errorf("profile %q already exists", newName)
				}
				delete(c.Profiles, oldName)
				c.Profiles[newName] = p
				if c.CurrentProfile == oldName {
					c.CurrentProfile = newName
				}
				return nil
			}); err != nil {
				return err
			}
			info("Renamed profile %q to %q", oldName, newName)
			return nil
		},
	}
}
