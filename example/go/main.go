package main

import (
	"log"
	"net/http"

	"github.com/go-zoox/watch/example/go/config"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	http.HandleFunc("/hi", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world 888"))
	})

	log.Fatal(http.ListenAndServe(":10080", nil))
}
