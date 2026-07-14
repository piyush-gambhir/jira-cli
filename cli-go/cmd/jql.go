package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newJQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jql",
		Short: "JQL helper tooling (autocomplete, suggest, parse)",
		Long: `Tools for composing and validating JQL queries — handy for agents.

Discover the fields, functions and reserved words available on the site, fetch
value suggestions for a field, and parse/validate queries before running them.

Examples:
  jira jql autocomplete
  jira jql autocomplete -o json
  jira jql suggest --field assignee --value jo
  jira jql parse "project = ABC AND statusCategory != Done"
  jira jql parse "project = ABC" "assignee = bogus(" --validation strict`,
	}
	cmd.AddCommand(
		newJQLAutocompleteCmd(),
		newJQLSuggestCmd(),
		newJQLParseCmd(),
	)
	return cmd
}

func jqlFieldTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"VALUE", "DISPLAY NAME", "OPERATORS", "TYPES"},
		RowFunc: func(item interface{}) []string {
			f := item.(client.JQLFieldRef)
			return []string{
				f.Value,
				dash(f.DisplayName),
				truncate(strings.Join(f.Operators, " "), 40),
				truncate(strings.Join(f.Types, ", "), 30),
			}
		},
	}
}

func jqlSuggestionTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"VALUE", "DISPLAY NAME"},
		RowFunc: func(item interface{}) []string {
			s := item.(client.JQLSuggestion)
			return []string{s.Value, dash(s.DisplayName)}
		},
	}
}

func newJQLAutocompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "autocomplete",
		Aliases: []string{"autocompletedata"},
		Short:   "Show the fields, functions and reserved words usable in JQL",
		Long: `Fetch JQL autocomplete data: the visible field names (with their valid
operators and value types), function names, and reserved words.

In table mode this lists the field clause names and their display names. Use
-o json to get the full object (fields, functions and reserved words).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := jiraClient.JQLAutocompleteData()
			if err != nil {
				return err
			}
			// Table mode shows the fields; json/yaml emit the whole object.
			if outFormat == output.FormatTable {
				return render(data.VisibleFieldNames, jqlFieldTable())
			}
			return render(data, nil)
		},
	}
}

func newJQLSuggestCmd() *cobra.Command {
	var field, value, predicate string
	cmd := &cobra.Command{
		Use:   "suggest",
		Short: "Suggest autocomplete values for a JQL field",
		Long: `Return autocomplete value suggestions for a field. --field is required;
pass --value to filter by a prefix, and --predicate to scope to a predicate.

Examples:
  jira jql suggest --field assignee --value jo
  jira jql suggest --field status
  jira jql suggest --field labels --value reg -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if field == "" {
				return fmt.Errorf("--field is required")
			}
			suggestions, err := jiraClient.JQLSuggestions(field, value, predicate)
			if err != nil {
				return err
			}
			return render(suggestions, jqlSuggestionTable())
		},
	}
	cmd.Flags().StringVar(&field, "field", "", "Field name to suggest values for (required)")
	cmd.Flags().StringVar(&value, "value", "", "Value prefix to filter suggestions")
	cmd.Flags().StringVar(&predicate, "predicate", "", "Predicate name to scope suggestions (e.g. by)")
	return cmd
}

func newJQLParseCmd() *cobra.Command {
	var validation string
	cmd := &cobra.Command{
		Use:   "parse <jql>...",
		Short: "Parse and validate one or more JQL queries",
		Long: `Parse (and optionally validate) JQL queries without running them. Pass one
or more queries as arguments. --validation controls strictness: strict, warn,
or none.

Per-query errors are printed to stderr; the command exits nonzero if any query
has errors. Use -o json to get the parsed structures and errors as data.

Examples:
  jira jql parse "project = ABC AND statusCategory != Done"
  jira jql parse "project = ABC" "assignee = " --validation strict
  jira jql parse "project = ABC" -o json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parsed, err := jiraClient.JQLParse(args, validation)
			if err != nil {
				return err
			}
			def := &output.TableDef{
				Headers: []string{"QUERY", "VALID", "ERRORS"},
				RowFunc: func(item interface{}) []string {
					p := item.(client.JQLParsedError)
					valid := "yes"
					if len(p.Errors) > 0 {
						valid = "no"
					}
					return []string{
						truncate(p.Query, 50),
						valid,
						truncate(strings.Join(p.Errors, "; "), 60),
					}
				},
			}
			if err := render(parsed, def); err != nil {
				return err
			}
			// Report problems and exit nonzero if any query failed to parse.
			var bad int
			for _, p := range parsed {
				for _, e := range p.Errors {
					info("error in %q: %s", p.Query, e)
					bad++
				}
			}
			if bad > 0 {
				return fmt.Errorf("%d JQL query error(s) found", bad)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&validation, "validation", "", "Validation level: strict, warn, none")
	return cmd
}
