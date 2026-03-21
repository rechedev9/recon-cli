package depgraphcli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/rechedev9/recon-cli/internal/cli"
	"github.com/rechedev9/recon-cli/internal/depgraph"
	"github.com/rechedev9/recon-cli/internal/output"
	"github.com/spf13/cobra"
)

var version = "dev"

type options struct {
	format string
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	opts := &options{}

	root := &cobra.Command{
		Use:           "depgraph [path]",
		Short:         "Dependency tree analyzer",
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

			report, err := depgraph.BuildGraph(absPath)
			if err != nil {
				return fmt.Errorf("build graph: %w", err)
			}

			switch opts.format {
			case "json":
				data, err := output.FormatJSON(report)
				if err != nil {
					return fmt.Errorf("format json: %w", err)
				}
				fmt.Fprintln(stdout, string(data))
			case "md":
				fmt.Fprint(stdout, formatGraphMarkdown(report))
			default:
				return &cli.UsageError{Msg: fmt.Sprintf("unknown format %q (use json or md)", opts.format)}
			}

			return nil
		},
	}

	root.Flags().StringVar(&opts.format, "format", "json", "Output format: json or md")

	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	return root.ExecuteContext(ctx)
}

func formatGraphMarkdown(report *depgraph.GraphReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s (%s)\n\n", report.Root, report.Manager)
	printNodes(&b, report.Nodes, 0)
	return b.String()
}

func printNodes(b *strings.Builder, nodes []depgraph.DepNode, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, n := range nodes {
		ver := ""
		if n.Version != "" {
			ver = " " + n.Version
		}
		fmt.Fprintf(b, "%s- %s%s\n", prefix, n.Name, ver)
		if len(n.Children) > 0 {
			printNodes(b, n.Children, indent+1)
		}
	}
}
