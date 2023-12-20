package process

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

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
				if err := p.kill(); err != nil {
					logger.Error("[process] kill error: %s", err)
				}
			}()
			return

		case <-p.restart:
			go func() {
				if err := p.kill(); err != nil {
					logger.Error("[process] kill error: %s", err)
				}

				if err := p.run(); err != nil {
					logger.Error("[process] failed to restart process: %s", err)
				}
			}()

		case <-p.start:
			go func() {
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

	p.cmd = cmd

	// go func() {
	// 	time.Sleep(time.Second)
	// 	fmt.Println("[process] run:", p.command, cmd.Process.Pid)
	// }()

	// return cmd.Run()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	// // wait release any resources associated with the Cmd
	// go func() {
	// 	time.Sleep(100 * time.Millisecond)

	// 	if err := cmd.Wait(); err != nil {
	// 		logger.Error("wait child process exit error: %s (pid: %d)", err, cmd.Process.Pid)
	// 	}
	// }()

	return cmd.Start()
}

func (p *process) kill() error {
	cmd := p.cmd

	// fmt.Println("[process] kill:", cmd.Process.Pid)

	if cmd != nil {
		// wait release any resources associated with the Cmd
		// go func() {
		// 	if err := cmd.Wait(); err != nil {
		// 		logger.Error("wait child process exit error: %s (pid: %d)", err, cmd.Process.Pid)
		// 	}
		// }()

		time.Sleep(100 * time.Millisecond)

		// // https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
		// if err := cmd.Process.Kill(); err != nil {
		// 	return fmt.Errorf("failed to kill process: %s", err)
		// }

		// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to kill process: %s", err)
		}

		// p, err := ps.NewProcess(int32(cmd.Process.Pid))
		// if err != nil {
		// 	return fmt.Errorf("failed to find process when kill: %s", err)
		// }
		// if err := p.Kill(); err != nil {
		// 	return fmt.Errorf("failed to kill process: %s", err)
		// }
	}

	// fmt.Println("[process] kill done:", p.command, p.cmd.ProcessState)

	time.Sleep(100 * time.Millisecond)

	return nil
}
