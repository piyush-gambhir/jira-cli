package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

func newDashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dashboard",
		Aliases: []string{"dashboards"},
		Short:   "List, search, and inspect dashboards",
		Long: `List, search, and inspect Jira dashboards.

Dashboards are visual gadget collections; these commands are read-only.

Examples:
  jira dashboard list --filter favourite
  jira dashboard search --query "Team" -o json
  jira dashboard get 10000`,
	}
	cmd.AddCommand(
		newDashboardListCmd(),
		newDashboardSearchCmd(),
		newDashboardGetCmd(),
	)
	return cmd
}

func dashboardTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "OWNER", "FAV", "POP"},
		RowFunc: func(item interface{}) []string {
			d := item.(client.Dashboard)
			fav := "-"
			if d.IsFavourite {
				fav = "yes"
			}
			return []string{dash(d.ID), dash(d.Name), userName(d.Owner), fav, strconv.Itoa(d.Popularity)}
		},
	}
}

func newDashboardListCmd() *cobra.Command {
	var filter string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List dashboards",
		Long: `List dashboards visible to you, optionally filtered.

--filter accepts: my, favourite, public.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dashboards, err := jiraClient.ListDashboards(filter, limit)
			if err != nil {
				return err
			}
			return render(dashboards, dashboardTable())
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Filter dashboards: my, favourite, public")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum dashboards to return")
	return cmd
}

func newDashboardSearchCmd() *cobra.Command {
	var query string
	var limit int
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search dashboards by name",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dashboards, err := jiraClient.SearchDashboards(query, limit)
			if err != nil {
				return err
			}
			return render(dashboards, dashboardTable())
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "Filter by dashboard name")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum dashboards to return")
	return cmd
}

func newDashboardGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a dashboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := jiraClient.GetDashboard(args[0])
			if err != nil {
				return err
			}
			return render(*d, dashboardTable())
		},
	}
}
