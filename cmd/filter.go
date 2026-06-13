package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

func newFilterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "filter",
		Aliases: []string{"filters"},
		Short:   "List, inspect, and manage saved JQL filters",
		Long: `List, search, run, create, update, and delete saved JQL filters, and
manage which filters are marked as favourites.

Examples:
  jira filter list
  jira filter favourites
  jira filter search --query "release" -n 20
  jira filter create --name "My open bugs" --jql "assignee = currentUser() AND type = Bug" --favourite
  jira filter run 10001 --all`,
	}
	cmd.AddCommand(
		newFilterListCmd(),
		newFilterFavouritesCmd(),
		newFilterSearchCmd(),
		newFilterGetCmd(),
		newFilterCreateCmd(),
		newFilterUpdateCmd(),
		newFilterDeleteCmd(),
		newFilterFavouriteCmd(),
		newFilterUnfavouriteCmd(),
		newFilterRunCmd(),
	)
	return cmd
}

func filterTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "FAV", "OWNER", "JQL"},
		RowFunc: func(item interface{}) []string {
			f := item.(client.Filter)
			return []string{
				dash(f.ID),
				dash(f.Name),
				filterBool(f.Favourite),
				userName(f.Owner),
				truncate(f.JQL, 60),
			}
		},
	}
}

func filterBool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func newFilterListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List filters owned by the current user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filters, err := jiraClient.ListMyFilters()
			if err != nil {
				return err
			}
			return render(filters, filterTable())
		},
	}
}

func newFilterFavouritesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "favourites",
		Aliases: []string{"favorites"},
		Short:   "List the current user's favourite filters",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filters, err := jiraClient.ListFavouriteFilters()
			if err != nil {
				return err
			}
			return render(filters, filterTable())
		},
	}
}

func newFilterSearchCmd() *cobra.Command {
	var query string
	var limit int
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search filters by name",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filters, err := jiraClient.SearchFilters(query, limit)
			if err != nil {
				return err
			}
			return render(filters, filterTable())
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "Filter by (partial) name")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum filters to return")
	return cmd
}

func newFilterGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a filter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := jiraClient.GetFilter(args[0])
			if err != nil {
				return err
			}
			return render(*f, filterTable())
		},
	}
}

func newFilterCreateCmd() *cobra.Command {
	var name, jql, description string
	var favourite bool
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create a saved filter",
		Annotations: mutates,
		Args:        cobra.NoArgs,
		Long: `Create a saved JQL filter. --name and --jql are required.

Examples:
  jira filter create --name "My open bugs" \
    --jql "assignee = currentUser() AND type = Bug AND statusCategory != Done" \
    --description "Bugs I still need to fix" --favourite`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" || jql == "" {
				return fmt.Errorf("--name and --jql are required")
			}
			body := map[string]any{"name": name, "jql": jql, "favourite": favourite}
			if description != "" {
				body["description"] = description
			}
			f, err := jiraClient.CreateFilter(body)
			if err != nil {
				return err
			}
			info("Created filter %s", f.ID)
			return render(*f, filterTable())
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Filter name (required)")
	cmd.Flags().StringVar(&jql, "jql", "", "JQL query (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Filter description")
	cmd.Flags().BoolVar(&favourite, "favourite", false, "Mark the new filter as a favourite")
	return cmd
}

func newFilterUpdateCmd() *cobra.Command {
	var name, jql, description string
	var favourite bool
	cmd := &cobra.Command{
		Use:         "update <id>",
		Short:       "Update a filter's fields",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if jql != "" {
				body["jql"] = jql
			}
			if description != "" {
				body["description"] = description
			}
			if cmd.Flags().Changed("favourite") {
				body["favourite"] = favourite
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update; pass --name/--jql/--description/--favourite")
			}
			f, err := jiraClient.UpdateFilter(args[0], body)
			if err != nil {
				return err
			}
			info("Updated filter %s", args[0])
			return render(*f, filterTable())
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New filter name")
	cmd.Flags().StringVar(&jql, "jql", "", "New JQL query")
	cmd.Flags().StringVarP(&description, "description", "d", "", "New description")
	cmd.Flags().BoolVar(&favourite, "favourite", false, "Set favourite state (true/false)")
	return cmd
}

func newFilterDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "delete <id>",
		Short:       "Delete a filter",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete filter %s without --yes", args[0])
				}
				ans, _ := prompt(fmt.Sprintf("Delete filter %s? [y/N]: ", args[0]))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteFilter(args[0]); err != nil {
				return err
			}
			info("Deleted filter %s", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newFilterFavouriteCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "favourite <id>",
		Aliases:     []string{"favorite"},
		Short:       "Mark a filter as a favourite",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := jiraClient.FavouriteFilter(args[0])
			if err != nil {
				return err
			}
			info("Favourited filter %s", args[0])
			return render(*f, filterTable())
		},
	}
}

func newFilterUnfavouriteCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "unfavourite <id>",
		Aliases:     []string{"unfavorite"},
		Short:       "Remove a filter from your favourites",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := jiraClient.UnfavouriteFilter(args[0])
			if err != nil {
				return err
			}
			info("Unfavourited filter %s", args[0])
			return render(*f, filterTable())
		},
	}
}

func newFilterRunCmd() *cobra.Command {
	var limit int
	var all bool
	cmd := &cobra.Command{
		Use:   "run <id>",
		Short: "Run a filter's JQL and list the matching issues",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := jiraClient.GetFilter(args[0])
			if err != nil {
				return err
			}
			if strings.TrimSpace(f.JQL) == "" {
				return fmt.Errorf("filter %s has no JQL to run", args[0])
			}
			info("JQL: %s", f.JQL)
			issues, err := jiraClient.SearchIssues(f.JQL, nil, limit, all)
			if err != nil {
				return err
			}
			return render(issues, issueListTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum issues to return (ignored with --all)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all matching issues (paginate)")
	return cmd
}
