package main

import (
	"github.com/go-zoox/cli"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/watch"
)

func main() {
	app := cli.NewSingleProgram(&cli.SingleProgramConfig{
		Name:    "watch",
		Usage:   "The command watcher",
		Version: watch.Version,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "command",
				Usage:    "the command",
				Aliases:  []string{"c"},
				Required: true,
			},
			&cli.StringFlag{
				Name:  "context",
				Usage: "the command context",
				Value: fs.CurrentDir(),
			},
			&cli.StringSliceFlag{
				Name:  "ignore",
				Usage: "the ignored files",
			},
		},
	})

	app.Command(func(ctx *cli.Context) error {
		watcher := watch.New(&watch.Config{
			Context:  ctx.String("context"),
			Commands: ctx.StringSlice("command"),
			Ignores:  ctx.StringSlice("ignore"),
		})

		return watcher.Watch()
	})

	app.Run()
}
