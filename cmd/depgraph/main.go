package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rechedev9/recon-cli/internal/cli"
	"github.com/rechedev9/recon-cli/internal/depgraphcli"
)

func main() {
	ctx := context.Background()
	if err := depgraphcli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCode(err))
	}
}
