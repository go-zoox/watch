package watcher

// references:
//	https://studygolang.com/articles/34169
//  https://github.com/bep/debounce/blob/master/debounce.go
//  https://nathanleclaire.com/blog/2014/08/03/write-a-function-similar-to-underscore-dot-jss-debounce-in-golang/
//	https://drailing.net/2018/01/debounce-function-for-golang/

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/go-zoox/debounce"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
)

type Watcher interface {
	Watch() error
	Stop() error
}

type watcher struct {
	cfg  *Config
	stop chan int
}

type Config struct {
	Context  string
	Paths    []string
	Ignores  []string
	Commands []string
}

func New(cfg *Config) Watcher {
	return &watcher{
		cfg:  cfg,
		stop: make(chan int),
	}
}

func (w *watcher) Watch() error {
	paths := append(w.cfg.Paths, w.cfg.Context)
	runner := createRunner(w.cfg.Commands, w.cfg.Context)

	if err := runner(); err != nil {
		logger.Error("[watch] failed to run runner: %s", err)
	}

	err := fs.WatchDir(context.Background(), paths, func(err error, event, filepath string) {
		if err != nil {
			logger.Error("[watch] error: %s", err)
			return
		}

		ignored := false
		for _, ignore := range w.cfg.Ignores {
			if ok, err := regexp.MatchString(ignore, filepath); err != nil {
				logger.Error("failed to match ignore: %s (err: %s)", ignore, err)
				return
			} else if ok {
				ignored = true
				break
			}
		}
		if ignored {
			return
		}

		// logger.Info("[watch] file change: %s (%s)", e.Name, e.Op.String())
		if err := runner(); err != nil {
			logger.Error("[watch] failed to run runner: %s", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to create watcher: %s", err)
	}

	logger.Info("[watch] stopping ...")
	return nil
}

func (w *watcher) Stop() error {
	w.stop <- 1
	return nil
}

func createRunner(commands []string, context string) func() error {
	cancels := make([]func() error, 0)

	return debounce.New(func() {
		logger.Info("[watch] file changing, exec `%s`", strings.Join(commands, ", "))

		for _, cancel := range cancels {
			if err := cancel(); err != nil {
				logger.Error("[watch] failed to cancel command: %s (err: %s)", cancel, err)
			}
		}

		// reset cancels
		cancels = make([]func() error, 0)
		for _, cmd := range commands {
			cancels = append(cancels, execCommand(cmd, context))
		}
	}, 300*time.Millisecond)
}

func execCommand(command string, context string) func() error {
	logger.Info("[watch] running command: %s ...", command)
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = context

	go func() {
		if err := cmd.Run(); err != nil {
			logger.Error("[watch] failed to run command: %s (err: %s)", command, err)
		}
	}()

	return func() error {
		if cmd != nil && cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			return nil
		}

		return cmd.Process.Signal(syscall.SIGTERM)
	}
}
