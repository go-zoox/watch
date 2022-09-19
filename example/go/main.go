package main

import (
	"log"

	"github.com/go-zoox/watch/example/go/config"
	"github.com/go-zoox/zoox"
	zd "github.com/go-zoox/zoox/default"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	r := zd.Default()

	r.Get("/hi", func(ctx *zoox.Context) {
		ctx.String(200, "hello world 888")
	})

	r.Run(":10080")
}
