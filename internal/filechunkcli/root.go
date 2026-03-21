package filechunkcli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/rechedev9/recon-cli/internal/chunker"
	"github.com/rechedev9/recon-cli/internal/cli"
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
		Use:           "filechunk <file>",
		Short:         "Extract function signatures and type definitions from source files",
		Version:       version,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			chunk, err := chunker.Chunk(path)
			if err != nil {
				return fmt.Errorf("chunk %s: %w", path, err)
			}

			switch opts.format {
			case "json":
				data, err := output.FormatJSON(chunk)
				if err != nil {
					return fmt.Errorf("format json: %w", err)
				}
				fmt.Fprintln(stdout, string(data))
			case "md":
				fmt.Fprint(stdout, formatChunkMarkdown(chunk))
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

func formatChunkMarkdown(chunk *chunker.FileChunk) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s (%s)\n\n", chunk.Path, chunk.Language)
	for _, sym := range chunk.Symbols {
		fmt.Fprintf(&b, "- [%s] **%s** (line %d): `%s`\n", sym.Kind, sym.Name, sym.Line, sym.Signature)
	}
	return b.String()
}
