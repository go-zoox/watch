package watch

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

	"github.com/go-zoox/core-utils/array"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/watch/process"
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
	Commands []string
	Context  string
	Paths    []string
	Ignores  []string
	Exts     []string
	Env      map[string]string
	// go | golang
	Mode string
}

func New(cfg *Config) Watcher {
	processes := []process.Process{}
	for _, command := range cfg.Commands {
		processes = append(processes, process.New(command, &process.Options{
			Context:     cfg.Context,
			Environment: cfg.Env,
		}))
	}

	// if cfg.Mode == "" {
	// 	cfg.Mode = "gpm"
	// }

	cfg.Ignores = append(cfg.Ignores, fs.JoinPath(cfg.Context, ".git"))

	return &watcher{
		cfg:       cfg,
		processes: processes,
	}
}

func (w *watcher) Watch() error {
	if w.cfg.Mode == "gpm" {
		err := w.watchGpm()
		if err != nil {
			logger.Error("watch gpm error: %s", err)
		}

		return err
	}

	paths := append(w.cfg.Paths, w.cfg.Context)

	go func() {
		logger.Infof("[watch] start watching ...")

		logger.Debugf(`#################
#  Watcher Start #
#################
# Command: %s
# Context: %s
#################

`, w.cfg.Commands, w.cfg.Context)
		if err := w.run(); err != nil {
			logger.Errorf("failed to run processes: %s", err)
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
				logger.Errorf("failed to match ignore: %s (err: %s)", ignore, err)
				return
			} else if ok {
				ignored = true
				break
			}
		}
		if ignored {
			return
		}

		// check ext, if not match anyone, ignore
		if len(w.cfg.Exts) > 0 {
			vExt := fs.ExtName(filepath)
			ok := array.Some(w.cfg.Exts, func(element string, index int) bool {
				return element == vExt
			})
			if !ok {
				return
			}
		}

		logger.Infof("[watch] file changes, restart processes (event: %s) ...", event)
		if err := w.restart(); err != nil {
			logger.Errorf("[watch] failed to restart processes: %s", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to create watcher: %s", err)
	}

	logger.Infof("[watch] stopping ...")
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

func (w *watcher) watchGpm() error {
	if err := exec.Command("which", "gpm").Run(); err != nil {
		return fmt.Errorf("gpm is not installed, you can install it by `npm i -g @cliz/gpm`")
	}

	cmd := exec.Command("gpm", "watch", "--exec", w.cfg.Commands[0])
	cmd.Dir = w.cfg.Context
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
