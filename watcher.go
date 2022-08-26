package watcher

// references:
//	https://studygolang.com/articles/34169
//  https://github.com/bep/debounce/blob/master/debounce.go
//  https://nathanleclaire.com/blog/2014/08/03/write-a-function-similar-to-underscore-dot-jss-debounce-in-golang/
//	https://drailing.net/2018/01/debounce-function-for-golang/

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/watcher/process"
)

type Watcher interface {
	Watch() error
	Stop() error
}

type watcher struct {
	cfg       *Config
	processes []process.Process
}

type Config struct {
	Context  string
	Paths    []string
	Ignores  []string
	Commands []string
	Env      map[string]string
	//

}

func New(cfg *Config) Watcher {
	processes := []process.Process{}
	for _, command := range cfg.Commands {
		processes = append(processes, process.New(command, &process.Options{
			Context:     cfg.Context,
			Environment: cfg.Env,
		}))
	}

	return &watcher{
		cfg:       cfg,
		processes: processes,
	}
}

func (w *watcher) Watch() error {
	paths := append(w.cfg.Paths, w.cfg.Context)

	go func() {
		fmt.Printf(`#################
#  Watcher Start #
#################
# Command: %s
# Context: %s
#################

`, w.cfg.Commands, w.cfg.Context)
		if err := w.run(); err != nil {
			logger.Error("failed to run processes: %s", err)
		}
	}()

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

		logger.Info("[watch] file changes, restart processes: %s", strings.Join(w.cfg.Commands, ", "))
		if err := w.restart(); err != nil {
			logger.Error("[watch] failed to restart processes: %s", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to create watcher: %s", err)
	}

	logger.Info("[watch] stopping ...")
	return nil
}

func (w *watcher) run() error {
	for _, p := range w.processes {
		if err := p.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (w *watcher) restart() error {
	for _, p := range w.processes {
		if err := p.Restart(); err != nil {
			return err
		}
	}

	return nil
}

func (w *watcher) Stop() error {
	for _, p := range w.processes {
		if err := p.Stop(); err != nil {
			return err
		}
	}

	return nil
}
