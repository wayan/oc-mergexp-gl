package cmd

import (
	"context"
	"os"

	"log/slog"

	"github.com/urfave/cli/v3"
)

func Run(cli *cli.Command, err error) {
	if err != nil {
		slog.Error("building cli failed: %", err.Error(), "exitCode", 1)
		os.Exit(1)
	}
	if err := cli.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error(), "exitCode", 1)
		os.Exit(1)
	}
}
