package main

import (
	zd "github.com/go-zoox/zoox/default"
)

func main() {
	r := zd.Default()

	r.Run(":10080")
}
