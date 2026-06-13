package cmd

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/output"
)

type profileSummary struct {
	Name     string `json:"name"`
	Current  bool   `json:"current"`
	AuthType string `json:"auth_type"`
	Site     string `json:"site"`
	Account  string `json:"account,omitempty"`
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured profiles",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var summaries []profileSummary
			for name, p := range cfg.Profiles {
				acct := p.Email
				if acct == "" {
					acct = p.Username
				}
				summaries = append(summaries, profileSummary{
					Name:     name,
					Current:  name == cfg.CurrentProfile,
					AuthType: p.EffectiveAuthType(),
					Site:     p.Site,
					Account:  acct,
				})
			}
			if len(summaries) == 0 {
				info("No profiles configured. Run 'jira auth login'.")
			}
			def := &output.TableDef{
				Headers: []string{"CURRENT", "NAME", "AUTH TYPE", "SITE", "ACCOUNT"},
				RowFunc: func(item interface{}) []string {
					s := item.(profileSummary)
					cur := ""
					if s.Current {
						cur = "*"
					}
					return []string{cur, s.Name, s.AuthType, dash(s.Site), dash(s.Account)}
				},
			}
			return render(summaries, def)
		},
	}
}
