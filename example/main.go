package main

import (
	"os"
	"path"

	"github.com/go-zoox/logger"
	watch "github.com/go-zoox/watch"
)

func main() {
	pwd, _ := os.Getwd()
	// watcher := watcher.New(&watcher.Config{
	// 	Context: path.Join(pwd, "example/go"),
	// 	Ignores: []string{},
	// 	Commands: []string{
	// 		"go run .",
	// 		// "go build -o /tmp/example && /tmp/example",
	// 	},
	// })

	watcher := watch.New(&watch.Config{
		Context: path.Join(pwd, "example/node"),
		Ignores: []string{},
		Commands: []string{
			"node app.js",
		},
	})

	if err := watcher.Watch(); err != nil {
		logger.Error("failed to watch: %s", err)
	}
}
