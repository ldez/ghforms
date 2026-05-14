package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ldez/ghforms/internal"
	"github.com/ldez/ghforms/internal/form"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
	"gitlab.com/greyxor/slogor"
)

//go:generate npm run build:css

var version = "dev"

func main() {
	opts := []slogor.OptionFn{
		slogor.SetLevel(slog.LevelInfo),
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		opts = append(opts, slogor.DisableColor())
	}

	slog.SetDefault(slog.New(slogor.NewHandler(os.Stdout, opts...)))

	app := createRootCommand()
	app.Version = version

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("%s version %s %s/%s\n", app.Name, cmd.Version, runtime.GOOS, runtime.GOARCH)
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		slog.Error("Fatal error", slog.Any("error", err))

		os.Exit(1)
	}
}

func createRootCommand() *cli.Command {
	return &cli.Command{
		Name:                  "ghforms",
		Usage:                 "GitHub Forms Live Preview and Validation (Issues and Discussions)",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Usage: "HTTP listen address",
				Value: ":8080",
				Local: true,
			},
			&cli.StringFlag{
				Name:  "dir",
				Usage: "Path to the forms directory",
				Value: ".github/ISSUE_TEMPLATE",
				Local: true,
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.NArg() > 0 && cmd.Command(cmd.Args().First()) == nil {
				return ctx, errors.New("unknown command")
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return internal.Run(cmd.String("addr"), cmd.String("dir"))
		},
		Commands: []*cli.Command{
			createVerifyCommand(),
		},
	}
}

func createVerifyCommand() *cli.Command {
	return &cli.Command{
		Name:  "verify",
		Usage: "Verify the forms",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "Path to the issue templates directory",
				Value: ".github",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			loader, err := form.New()
			if err != nil {
				return err
			}

			return filepath.WalkDir(cmd.String("dir"), func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !entry.IsDir() {
					return nil
				}

				if !strings.EqualFold(entry.Name(), "ISSUE_TEMPLATE") && !strings.EqualFold(entry.Name(), "DISCUSSION_TEMPLATE") {
					return nil
				}

				slog.Info("Verifying forms", slog.String("dir", path))

				_, err = loader.Load(path)
				if err != nil {
					return err
				}

				return nil
			})
		},
	}
}
