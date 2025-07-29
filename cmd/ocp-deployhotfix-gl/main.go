package main

import (
	"context"
	"os"

	"log/slog"

	"github.com/wayan/oc-mergexp-gl/cmd"
)

func run() error {
	cli, err := cmd.CliDeployHotfix(cmd.OCP)
	if err != nil {
		return err
	}
	return cli.Run(context.Background(), os.Args)
}

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error(), "exitCode", 1)
		os.Exit(1)
	}
}
