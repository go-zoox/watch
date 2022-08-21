# Watcher

> Make it create watcher easier.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-zoox/watcher)](https://pkg.go.dev/github.com/go-zoox/watcher)
[![Build Status](https://github.com/go-zoox/watcher/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/go-zoox/watcher/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-zoox/watcher)](https://goreportcard.com/report/github.com/go-zoox/watcher)
[![Coverage Status](https://coveralls.io/repos/github/go-zoox/watcher/badge.svg?branch=master)](https://coveralls.io/github/go-zoox/watcher?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/go-zoox/watcher.svg)](https://github.com/go-zoox/watcher/issues)
[![Release](https://img.shields.io/github/tag/go-zoox/watcher.svg?label=Release)](https://github.com/go-zoox/watcher/tags)

## Installation

To install the package, run:

```bash
go get github.com/go-zoox/watcher
```

## Getting Started

```go
import (
  "testing"
  "github.com/go-zoox/watcher"
)

func main(t *testing.T) {
	pwd, _ := os.Getwd()
	watcher := watcher.New(&watcher.Config{
		Context: path.Join(pwd, "example/program"),
		Ignores: []string{},
		Commands: []string{
			"go run .",
		},
	})

	if err := watcher.Watch(); err != nil {
		logger.Error("failed to watch: %s", err)
	}
}
```

## Inspired By

- [silenceper/gowatch](https://github.com/silenceper/gowatch) - 🚀 gowatch is a
  command line tool that builds and (re)starts your go project everytime you
  save a Go or template file.
- [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) - Cross-platform
  file system notifications for Go.

## License

GoZoox is released under the [MIT License](./LICENSE).
