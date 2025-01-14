package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"

	art "actshad.dev/go-atomicredteam"
)

func main() {
	app := &cli.App{
		Name:    "goart",
		Usage:   "Standalone Atomic Red Team Executor (written in Go)",
		Version: art.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "emulation-path",
				Aliases: []string{"E"},
				Usage:   "path to file with list of atomics",
			},
			&cli.StringFlag{
				Name:    "technique",
				Aliases: []string{"t"},
				Usage:   "technique ID",
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "test name",
			},
			&cli.StringFlag{
				Name:    "guid",
				Aliases: []string{"G"},
				Usage:   "test guid",
				Value:   "",
			},

			&cli.IntFlag{
				Name:    "index",
				Aliases: []string{"i"},
				Usage:   "test index",
				Value:   -1,
			},
			&cli.StringSliceFlag{
				Name:  "input",
				Usage: "input key=value pairs",
			},
			&cli.StringSliceFlag{
				Name:    "env",
				Aliases: []string{"e"},
				Usage:   "env variable key=value pairs",
			},
			&cli.StringFlag{
				Name:    "local-atomics-path",
				Aliases: []string{"l"},
				Usage:   "directory containing additional/custom atomic test definitions",
			},
			&cli.StringFlag{
				Name:    "dump-technique",
				Aliases: []string{"d"},
				Usage:   "directory to dump the given technique test config to",
			},
			&cli.StringFlag{
				Name:    "results-file",
				Aliases: []string{"o"},
				Usage:   "file to write test results to (auto-generated by default)",
			},
			&cli.StringFlag{
				Name:    "results-format",
				Aliases: []string{"f"},
				Usage:   "format to use when writing results to file (json, yaml)",
				Value:   "yaml",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "disable printing info to terminal when executing a test",
			},
		},
		Action: func(ctx *cli.Context) error {
			art.Configure(ctx)
			defer art.Teardown(ctx)
			if emulationPath := ctx.String("emulation-path"); emulationPath != "" {
				return art.InvokeEmulation(ctx)
			} else {
				return art.InvokeAtomic(ctx)
			}
		},
	}

	if art.REPO == "" {
		app.Flags = append(
			app.Flags,
			&cli.StringFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Value:   "redcanaryco/master",
				Usage:   "Atomic Red Team repo/branch name",
			},
		)
	}

	if runtime.GOOS != "windows" {
		app.Flags = append(
			app.Flags,
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "disable printing colors to terminal",
			},
		)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	art.Println()
}
