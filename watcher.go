package watcher

// references:
//	https://studygolang.com/articles/34169
//  https://github.com/bep/debounce/blob/master/debounce.go
//  https://nathanleclaire.com/blog/2014/08/03/write-a-function-similar-to-underscore-dot-jss-debounce-in-golang/
//	https://drailing.net/2018/01/debounce-function-for-golang/

import (
	"fmt"
	iofs "io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
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

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %s", err)
	}
	runner := createRunner(w.cfg.Commands, w.cfg.Context)

	for _, path := range paths {
		logger.Info("[watch] watching path: %s ...", path)

		err := fs.WalkDir(path, func(path string, d iofs.DirEntry, err error) error {
			if d.IsDir() {
				if err := watcher.Add(path); err != nil {
					return fmt.Errorf("failed to watch directory: %s (err: %s)", path, err)
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk dir: %v", err)
		}
	}

	if err := runner(); err != nil {
		logger.Error("[watch] failed to run runner: %s", err)
	}

	for {
		select {
		case e := <-watcher.Events:
			if e.Op == fsnotify.Chmod {
				continue
			}

			ignored := false
			for _, ignore := range w.cfg.Ignores {
				if ok, err := regexp.MatchString(ignore, e.Name); err != nil {
					return fmt.Errorf("failed to match ignore: %s (err: %s)", ignore, err)
				} else if ok {
					ignored = true
					break
				}
			}
			if ignored {
				continue
			}

			// logger.Info("[watch] file change: %s (%s)", e.Name, e.Op.String())
			if err := runner(); err != nil {
				logger.Error("[watch] failed to run runner: %s", err)
			}

		case err := <-watcher.Errors:
			logger.Error("[watch] error: %s", err)

		case <-w.stop:
			logger.Info("[watch] stopping ...")
			if err := watcher.Close(); err != nil {
				logger.Error("[watch] failed to close watcher: %s", err)
			}

			return nil
		}
	}
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
