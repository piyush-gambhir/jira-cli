package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/adf"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func attachmentTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "FILENAME", "SIZE", "MIME", "AUTHOR"},
		RowFunc: func(item interface{}) []string {
			a := item.(client.Attachment)
			return []string{string(a.ID), a.Filename, fmt.Sprintf("%d", a.Size), dash(a.MimeType), userName(a.Author)}
		},
	}
}

func newIssueAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "attach <key> <file>...",
		Short:       "Attach one or more files to an issue",
		Annotations: mutates,
		Args:        cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, files := args[0], args[1:]
			attached, err := jiraClient.AddAttachment(key, files)
			if err != nil {
				return err
			}
			info("Attached %d file(s) to %s", len(attached), key)
			return render(attached, attachmentTable())
		},
	}
	return cmd
}

func newIssueAttachmentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "attachments <key>",
		Short: "List an issue's attachments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			attachments, err := jiraClient.ListAttachments(args[0])
			if err != nil {
				return err
			}
			return render(attachments, attachmentTable())
		},
	}
}

func newIssueDownloadCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "download <attachment-id>",
		Short: "Download an attachment by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, filename, err := jiraClient.DownloadAttachment(args[0])
			if err != nil {
				return err
			}
			if out == "" {
				out = filename
			}
			if out == "" {
				out = "attachment-" + args[0]
			}
			if err := os.WriteFile(out, data, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", out, err)
			}
			info("Downloaded %d bytes to %s", len(data), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "Output file path (defaults to the attachment's filename)")
	return cmd
}

func newIssueLinkCmd() *cobra.Command {
	var linkType, comment string
	var markdown bool
	cmd := &cobra.Command{
		Use:         "link <fromKey> <toKey>",
		Short:       "Link two issues",
		Annotations: mutates,
		Args:        cobra.ExactArgs(2),
		Long: `Create a link "<fromKey> <type-outward> <toKey>" (e.g. with --type Blocks,
"fromKey Blocks toKey"). Run 'jira issue link-types' to see available types.

Examples:
  jira issue link ABC-1 ABC-2 --type Blocks
  jira issue link ABC-1 ABC-2 --type "Relates" --comment "see also"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if linkType == "" {
				return fmt.Errorf("--type is required (see 'jira issue link-types')")
			}
			var c any
			if comment != "" {
				if markdown {
					c = adf.FromMarkdown(comment)
				} else {
					c = adf.FromPlainText(comment)
				}
			}
			// from = outwardIssue, to = inwardIssue.
			if err := jiraClient.LinkIssues(linkType, args[1], args[0], c); err != nil {
				return err
			}
			info("Linked %s %s %s", args[0], linkType, args[1])
			return nil
		},
	}
	cmd.Flags().StringVar(&linkType, "type", "", "Link type name, e.g. Blocks/Relates/Duplicate (required)")
	cmd.Flags().StringVar(&comment, "comment", "", "Optional comment to add with the link")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Interpret --comment as lightweight markdown")
	return cmd
}

func newIssueLinkTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "link-types",
		Short: "List available issue link types",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			types, err := jiraClient.ListLinkTypes()
			if err != nil {
				return err
			}
			def := &output.TableDef{
				Headers: []string{"ID", "NAME", "OUTWARD", "INWARD"},
				RowFunc: func(item interface{}) []string {
					t := item.(client.LinkType)
					return []string{t.ID, t.Name, t.Outward, t.Inward}
				},
			}
			return render(types, def)
		},
	}
}
