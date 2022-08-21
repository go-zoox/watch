# Debounce

> Function debouncer.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-zoox/debounce)](https://pkg.go.dev/github.com/go-zoox/debounce)
[![Build Status](https://github.com/go-zoox/debounce/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/go-zoox/debounce/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-zoox/debounce)](https://goreportcard.com/report/github.com/go-zoox/debounce)
[![Coverage Status](https://coveralls.io/repos/github/go-zoox/debounce/badge.svg?branch=master)](https://coveralls.io/github/go-zoox/debounce?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/go-zoox/debounce.svg)](https://github.com/go-zoox/debounce/issues)
[![Release](https://img.shields.io/github/tag/go-zoox/debounce.svg?label=Release)](https://github.com/go-zoox/debounce/tags)

## Installation

To install the package, run:

```bash
go get github.com/go-zoox/debounce
```

## Getting Started

```go
import (
  "testing"
  "github.com/go-zoox/debounce"
)

func main(t *testing.T) {
	count := 0

	fn := func() {
		count++
	}
	debouncedFn := New(fn, 10*time.Millisecond)

	for i := 0; i < 100; i++ {
		debouncedFn()
	}

	time.Sleep(20 * time.Millisecond)

	fmt.Println(count)
}
```

## Inspired By

- [bep/debounce](https://github.com/bep/debounce) - A debouncer written in Go.

## License

GoZoox is released under the [MIT License](./LICENSE).
