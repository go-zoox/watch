package main

import (
	"os"
	"path"

	"github.com/go-zoox/logger"
	"github.com/go-zoox/watcher"
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

	watcher := watcher.New(&watcher.Config{
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
