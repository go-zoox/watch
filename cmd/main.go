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
		Version: "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
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
		},
	})

	app.Command(func(ctx *cli.Context) error {
		watcher := watch.New(&watch.Config{
			Context: ctx.String("context"),
			Commands: []string{
				ctx.String("command"),
			},
		})

		return watcher.Watch()
	})

	app.Run()
}
