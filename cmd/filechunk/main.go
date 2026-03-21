package main

import (
	"context"
	"os"

	"github.com/rechedev9/recon-cli/internal/cli"
	"github.com/rechedev9/recon-cli/internal/filechunkcli"
)

func main() {
	ctx := context.Background()
	if err := filechunkcli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		os.Exit(cli.ExitCode(err))
	}
}
