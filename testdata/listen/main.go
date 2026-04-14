// Command listen binds to 127.0.0.1:<port> and blocks until killed.
// Used by process tests to assert TCP ports are released across restarts.
package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: listen <port>")
		os.Exit(2)
	}
	port := os.Args[1]
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer ln.Close()

	// Block until SIGKILL (or similar) tears the process down.
	// Avoid select{} or recv on an empty chan in main alone — the runtime may
	// report a fatal deadlock. Accept blocks without tripping that check.
	for {
		if _, err := ln.Accept(); err != nil {
			return
		}
	}
}
