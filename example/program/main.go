package main

import (
	"log"

	"github.com/go-zoox/watcher/example/program/config"
	zd "github.com/go-zoox/zoox/default"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	r := zd.Default()

	r.Run(":10080")
}
