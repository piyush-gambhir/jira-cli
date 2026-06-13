package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

// issueMetaFieldRow flattens a FieldMeta map entry (keyed by field id) for table
// rendering. Used by both createmeta (with --type) and editmeta.
type issueMetaFieldRow struct {
	FieldID    string   `json:"fieldId"`
	Name       string   `json:"name"`
	Required   bool     `json:"required"`
	Type       string   `json:"type,omitempty"`
	Operations []string `json:"operations,omitempty"`
	Allowed    []string `json:"allowedValues,omitempty"`
}

// issueMetaFieldRows turns a {fieldId: FieldMeta} map into id-sorted rows.
func issueMetaFieldRows(fields map[string]client.FieldMeta) []issueMetaFieldRow {
	rows := make([]issueMetaFieldRow, 0, len(fields))
	for id, fm := range fields {
		rows = append(rows, issueMetaFieldRow{
			FieldID:    id,
			Name:       fm.Name,
			Required:   fm.Required,
			Type:       fm.Schema.Type,
			Operations: fm.Operations,
			Allowed:    issueMetaAllowedNames(fm.AllowedValues),
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].FieldID < rows[j].FieldID })
	return rows
}

// issueMetaAllowedNames extracts human-readable labels from a field's allowed
// values (each is an arbitrary object; prefer name/value/key/id).
func issueMetaAllowedNames(vals []any) []string {
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		m, ok := v.(map[string]any)
		if !ok {
			out = append(out, fmt.Sprintf("%v", v))
			continue
		}
		for _, k := range []string{"name", "value", "key", "id"} {
			if s, ok := m[k].(string); ok && s != "" {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

func issueMetaTypeTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "SUBTASK"},
		RowFunc: func(item interface{}) []string {
			t := item.(client.IssueType)
			return []string{t.ID, t.Name, strconv.FormatBool(t.Subtask)}
		},
	}
}

func issueMetaFieldTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"FIELD", "NAME", "REQUIRED", "TYPE", "ALLOWED"},
		RowFunc: func(item interface{}) []string {
			r := item.(issueMetaFieldRow)
			return []string{
				r.FieldID,
				dash(r.Name),
				strconv.FormatBool(r.Required),
				dash(r.Type),
				dash(truncate(strings.Join(r.Allowed, ", "), 40)),
			}
		},
	}
}

func issueMetaEditTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"FIELD", "NAME", "REQUIRED", "OPERATIONS"},
		RowFunc: func(item interface{}) []string {
			r := item.(issueMetaFieldRow)
			return []string{
				r.FieldID,
				dash(r.Name),
				strconv.FormatBool(r.Required),
				dash(strings.Join(r.Operations, ",")),
			}
		},
	}
}

// newIssueCreateMetaCmd: issue createmeta <projectKey> [--type <id|name>]
func newIssueCreateMetaCmd() *cobra.Command {
	var issueType string
	cmd := &cobra.Command{
		Use:   "createmeta <projectKey>",
		Short: "Show creatable issue types, or the create-screen fields for a type",
		Long: `List the issue types a project accepts. With --type (an issue type id or
name) it instead lists that type's create-screen fields (id, required, type,
allowed values).

Examples:
  jira issue createmeta ABC
  jira issue createmeta ABC --type Bug
  jira issue createmeta ABC --type 10004 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectKey := args[0]
			types, err := jiraClient.CreateMetaIssueTypes(projectKey)
			if err != nil {
				return err
			}
			if issueType == "" {
				return render(types, issueMetaTypeTable())
			}
			typeID, err := issueMetaResolveType(types, issueType)
			if err != nil {
				return err
			}
			fields, err := jiraClient.CreateMetaFields(projectKey, typeID)
			if err != nil {
				return err
			}
			return render(issueMetaFieldRows(fields), issueMetaFieldTable())
		},
	}
	cmd.Flags().StringVar(&issueType, "type", "", "Issue type id or name; show its create-screen fields")
	return cmd
}

// issueMetaResolveType resolves an issue-type reference (exact id, or
// case-insensitive name) to its id within the project's creatable types.
func issueMetaResolveType(types []client.IssueType, ref string) (string, error) {
	for _, t := range types {
		if t.ID == ref {
			return t.ID, nil
		}
	}
	for _, t := range types {
		if strings.EqualFold(t.Name, ref) {
			return t.ID, nil
		}
	}
	return "", fmt.Errorf("no creatable issue type matching %q for this project", ref)
}

// newIssueEditMetaCmd: issue editmeta <key>
func newIssueEditMetaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "editmeta <key>",
		Short: "Show the fields editable on an issue's edit screen",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			meta, err := jiraClient.GetEditMeta(args[0])
			if err != nil {
				return err
			}
			return render(issueMetaFieldRows(meta.Fields), issueMetaEditTable())
		},
	}
}

// issueMetaChangeRow is a flattened single field change for table rendering.
type issueMetaChangeRow struct {
	Created string `json:"created"`
	Author  string `json:"author"`
	Field   string `json:"field"`
	From    string `json:"from"`
	To      string `json:"to"`
}

// newIssueChangelogCmd: issue changelog <key> [--limit -n]
func newIssueChangelogCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "changelog <key>",
		Short: "Show an issue's change history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := jiraClient.GetChangelog(args[0], 0, limit)
			if err != nil {
				return err
			}
			var rows []issueMetaChangeRow
			for _, e := range entries {
				author := userName(e.Author)
				for _, it := range e.Items {
					rows = append(rows, issueMetaChangeRow{
						Created: e.Created,
						Author:  author,
						Field:   it.Field,
						From:    it.FromString,
						To:      it.ToString,
					})
				}
			}
			def := &output.TableDef{
				Headers: []string{"CREATED", "AUTHOR", "FIELD", "FROM -> TO"},
				RowFunc: func(item interface{}) []string {
					r := item.(issueMetaChangeRow)
					return []string{
						dash(r.Created),
						dash(r.Author),
						dash(r.Field),
						fmt.Sprintf("%s -> %s", dash(truncate(r.From, 24)), dash(truncate(r.To, 24))),
					}
				},
			}
			return render(rows, def)
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum changelog entries to fetch")
	return cmd
}
