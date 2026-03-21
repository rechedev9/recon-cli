package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/rechedev9/CLIClaudeCode/internal/output"
	"github.com/rechedev9/CLIClaudeCode/internal/scanner"
	"github.com/spf13/cobra"
)

type options struct {
	format  string
	depth   int
	noGit   bool
	noDeps  bool
	noDocs  bool
	timeout int
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	opts := &options{}

	root := &cobra.Command{
		Use:           "codestat [path]",
		Short:         "Project structure summarizer",
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

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), time.Duration(opts.timeout)*time.Second)
			defer cancel()

			s := scanner.NewDefault()
			if opts.noGit {
				s.Git = nil
			}
			if opts.noDeps {
				s.Deps = nil
			}
			if opts.noDocs {
				s.Docs = nil
			}

			report, err := s.Run(timeoutCtx, absPath, opts.depth)
			if err != nil {
				return fmt.Errorf("scan %s: %w", absPath, err)
			}

			switch opts.format {
			case "json":
				data, err := output.FormatJSON(report)
				if err != nil {
					return fmt.Errorf("format json: %w", err)
				}
				fmt.Fprintln(stdout, string(data))
			case "md":
				fmt.Fprint(stdout, output.FormatMarkdown(report))
			default:
				return &UsageError{Msg: fmt.Sprintf("unknown format %q (use json or md)", opts.format)}
			}

			return nil
		},
	}

	root.Flags().StringVar(&opts.format, "format", "json", "Output format: json or md")
	root.Flags().IntVar(&opts.depth, "depth", 4, "Directory tree depth")
	root.Flags().BoolVar(&opts.noGit, "no-git", false, "Skip git summary")
	root.Flags().BoolVar(&opts.noDeps, "no-deps", false, "Skip dependency analysis")
	root.Flags().BoolVar(&opts.noDocs, "no-docs", false, "Skip doc index")
	root.Flags().IntVar(&opts.timeout, "timeout", 30, "Timeout in seconds for external commands")

	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	return root.ExecuteContext(ctx)
}
