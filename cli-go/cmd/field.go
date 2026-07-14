package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newFieldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "field",
		Aliases: []string{"fields"},
		Short:   "List system and custom fields",
	}
	cmd.AddCommand(newFieldListCmd())
	return cmd
}

func newFieldListCmd() *cobra.Command {
	var custom bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all fields (with their JQL clause names)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := jiraClient.ListFields()
			if err != nil {
				return err
			}
			if custom {
				filtered := fields[:0]
				for _, f := range fields {
					if f.Custom {
						filtered = append(filtered, f)
					}
				}
				fields = filtered
			}
			def := &output.TableDef{
				Headers: []string{"ID", "NAME", "CUSTOM", "JQL CLAUSES"},
				RowFunc: func(item interface{}) []string {
					f := item.(client.Field)
					return []string{f.ID, f.Name, fmt.Sprintf("%v", f.Custom), truncate(strings.Join(f.ClauseNames, ", "), 40)}
				},
			}
			return render(fields, def)
		},
	}
	cmd.Flags().BoolVar(&custom, "custom", false, "Only show custom fields")
	return cmd
}
