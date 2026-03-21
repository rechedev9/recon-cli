package teststatcli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/rechedev9/recon-cli/internal/cli"
	"github.com/rechedev9/recon-cli/internal/output"
	"github.com/rechedev9/recon-cli/internal/teststat"
	"github.com/spf13/cobra"
)

var version = "dev"

type options struct {
	format string
	dryRun bool
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	opts := &options{}

	root := &cobra.Command{
		Use:           "teststat [path]",
		Short:         "Test runner and result summarizer",
		Version:       version,
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			runner := teststat.DetectRunner(absPath)
			if runner == nil {
				return &cli.UsageError{Msg: fmt.Sprintf("no supported test framework in %s", absPath)}
			}

			if opts.dryRun {
				fmt.Fprintln(stdout, runner.Command(absPath))
				return nil
			}

			result, err := runner.Run(cmd.Context(), absPath)
			if err != nil {
				return fmt.Errorf("run tests: %w", err)
			}

			switch opts.format {
			case "json":
				data, err := output.FormatJSON(result)
				if err != nil {
					return fmt.Errorf("format json: %w", err)
				}
				fmt.Fprintln(stdout, string(data))
			case "md":
				fmt.Fprint(stdout, formatResultMarkdown(result))
			default:
				return &cli.UsageError{Msg: fmt.Sprintf("unknown format %q (use json or md)", opts.format)}
			}

			return nil
		},
	}

	root.Flags().StringVar(&opts.format, "format", "json", "Output format: json or md")
	root.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Show command without executing")

	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	return root.ExecuteContext(ctx)
}

func formatResultMarkdown(r *teststat.TestResult) string {
	var b strings.Builder
	status := "PASS"
	if r.Failed > 0 {
		status = "FAIL"
	}
	fmt.Fprintf(&b, "# Test Results (%s)\n\n", status)
	fmt.Fprintf(&b, "**Framework:** %s\n", r.Framework)
	fmt.Fprintf(&b, "**Command:** `%s`\n", r.Command)
	if r.Duration != "" {
		fmt.Fprintf(&b, "**Duration:** %s\n", r.Duration)
	}
	fmt.Fprintf(&b, "\n| Metric | Count |\n|--------|-------|\n")
	fmt.Fprintf(&b, "| Total | %d |\n", r.TotalTests)
	fmt.Fprintf(&b, "| Passed | %d |\n", r.Passed)
	fmt.Fprintf(&b, "| Failed | %d |\n", r.Failed)
	fmt.Fprintf(&b, "| Skipped | %d |\n", r.Skipped)

	if len(r.Failures) > 0 {
		fmt.Fprintf(&b, "\n## Failures\n\n")
		for _, f := range r.Failures {
			fmt.Fprintf(&b, "### %s\n", f.Name)
			if f.Package != "" {
				fmt.Fprintf(&b, "Package: %s\n", f.Package)
			}
			fmt.Fprintf(&b, "```\n%s\n```\n\n", f.Message)
		}
	}

	return b.String()
}
