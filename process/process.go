package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/go-zoox/logger"
)

type Process interface {
	Start() error
	Stop() error
	Restart() error
}

type process struct {
	command     string
	context     string
	environment map[string]string
	//
	start   chan int
	stop    chan int
	restart chan int
	//
	isStarted bool
	cmd       *exec.Cmd
	mu        sync.Mutex
}

type Options struct {
	Context     string
	Environment map[string]string
}

func New(command string, options ...*Options) Process {
	context := ""
	environment := make(map[string]string)
	if len(options) > 0 && options[0] != nil {
		context = options[0].Context
		environment = options[0].Environment
	}

	return &process{
		command:     command,
		context:     context,
		environment: environment,
		//
		start:   make(chan int),
		stop:    make(chan int),
		restart: make(chan int),
	}
}

func (p *process) Start() (err error) {
	// logger.Info("[process] start: %s", p.command)

	go p.routine()

	p.start <- 1

	// logger.Info("[process] start: %s done", p.command)
	return nil
}

func (p *process) Stop() error {
	p.stop <- 1
	return nil
}

func (p *process) Restart() (err error) {
	if !p.isStarted {
		err = p.Start()
		if err != nil {
			return err
		}
	}

	p.restart <- 1
	return nil
}

func (p *process) routine() {
	if p.isStarted {
		return
	}
	p.isStarted = true

	for {
		select {
		case <-p.stop:
			go func() {
				p.mu.Lock()
				defer p.mu.Unlock()
				if err := p.kill(); err != nil {
					logger.Error("[process] kill error: %s", err)
				}
			}()
			return

		case <-p.restart:
			go func() {
				p.mu.Lock()
				defer p.mu.Unlock()
				if err := p.kill(); err != nil {
					logger.Error("[process] kill error: %s", err)
				}

				if err := p.run(); err != nil {
					logger.Error("[process] failed to restart process: %s", err)
				}
			}()

		case <-p.start:
			go func() {
				p.mu.Lock()
				defer p.mu.Unlock()
				if err := p.run(); err != nil {
					logger.Error("[process] failed to start process (1): %s", err)
				}
			}()
		}
	}
}

func (p *process) run() error {
	environment := os.Environ()
	for k, v := range p.environment {
		environment = append(environment, fmt.Sprintf("%s=%s", k, v))
	}

	cmd := exec.Command("sh", "-c", p.command)
	// cmd := exec.Command("/bin/sh", "-c", p.command)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	//
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	//
	cmd.Dir = p.context
	cmd.Env = environment
	// Attach stdio directly so cmd.Wait() in kill() is not blocked waiting for
	// pipe-draining goroutines (which can deadlock under test output capture).
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	p.cmd = cmd

	return cmd.Start()
}

func (p *process) kill() error {
	cmd := p.cmd
	if cmd == nil {
		return nil
	}

	if cmd.Process == nil {
		p.cmd = nil
		return nil
	}

	pid := cmd.Process.Pid

	// Kill the whole process group (shell + children, e.g. the server holding the port).
	// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	// Wait until the process exits and is reaped; otherwise the next start can race
	// the old listener and hit "address already in use".
	if err := cmd.Wait(); err != nil {
		// Non-zero exit or signal after SIGKILL is expected.
		logger.Debugf("[process] wait after kill (pid %d): %s", pid, err)
	}

	p.cmd = nil
	return nil
}
