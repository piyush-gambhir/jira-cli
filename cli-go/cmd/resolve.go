package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// resolution describes what arguments would run without executing the command.
type resolution struct {
	Command     string            `json:"command"`
	Effect      string            `json:"effect"`
	Conditional bool              `json:"conditional"`
	Interactive bool              `json:"interactive"`
	Secret      bool              `json:"secret_output"`
	Flags       map[string]string `json:"flags"`
	Args        []string          `json:"args"`
}

func newCommandsCmd() *cobra.Command {
	parent := &cobra.Command{Use: "commands", Short: "Offline command discovery for agents"}
	parent.AddCommand(&cobra.Command{
		Use:                "resolve -- <arguments>",
		Short:              "Show the command, effect and flags that arguments would run, without running them",
		DisableFlagParsing: true,
		Example:            "  jira commands resolve -- issue get PROJ-123\n  jira commands resolve -- -o json issue delete PROJ-123 --yes",
		// Override the root hooks so resolution never loads config or checks for updates.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			outputFormat, OutputFormat, outFormat = "json", "json", output.FormatJSON
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error { return nil },
		RunE: func(cmd *cobra.Command, args []string) error {
			for i, arg := range args {
				if arg == "--" {
					args = args[i+1:]
					break
				}
			}
			probe := newRootCmd()
			probe.SetIn(strings.NewReader(""))
			probe.SetOut(io.Discard)
			probe.SetErr(io.Discard)
			target, rest, err := probe.Find(args)
			if err != nil || target == probe || !target.Runnable() || target.Annotations[groupDispatcherKey] == "true" {
				return fmt.Errorf("arguments do not name a runnable jira command")
			}
			if err := target.ParseFlags(rest); err != nil {
				return err
			}
			res := describeResolution(target)
			return output.Print(cmd.OutOrStdout(), output.FormatJSON, res, nil)
		},
	})
	return parent
}

func describeResolution(target *cobra.Command) resolution {
	res := resolution{Command: target.CommandPath(), Effect: "read", Flags: map[string]string{}, Args: target.Flags().Args()}
	if res.Args == nil {
		res.Args = []string{}
	}
	ann := target.Annotations
	switch {
	case ann["mutates"] == "true" || ann["mutates"] == "conditional":
		res.Effect, res.Conditional = "remote_write", ann["mutates"] == "conditional"
	case ann["mutates-when-flag"] != "":
		if target.Flags().Changed(ann["mutates-when-flag"]) {
			res.Effect = "remote_write"
		}
	case ann["writes-local"] == "true" || ann["writes-local"] == "conditional" || ann["self-update"] != "" ||
		(ann["writes-local-when-flag"] != "" && target.Flags().Changed(ann["writes-local-when-flag"])):
		res.Effect, res.Conditional = "local_write", ann["writes-local"] == "conditional"
	}
	res.Interactive = ann["interactive"] == "true"
	res.Secret = ann["secret-output"] == "true"
	target.Flags().Visit(func(f *pflag.Flag) { res.Flags[f.Name] = f.Value.String() })
	return res
}
