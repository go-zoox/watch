package watch

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-zoox/fs"
)

type Watch interface {
	Command() string
	Context() string
	Environment() map[string]string
	PID() int
	//
	Run() error
	Kill() error
	Restart() error
	// Reload() error
}

type Config struct {
	Context     string
	Environment map[string]string
}

type cmd struct {
	core *exec.Cmd
	//
	command     string
	context     string
	environment map[string]string
}

func New(command string, cfg ...*Config) Watch {
	context := fs.CurrentDir()
	environment := make(map[string]string)
	if len(cfg) > 0 && cfg[0] != nil {
		context = cfg[0].Context
		environment = cfg[0].Environment
	}

	return &cmd{
		command:     command,
		context:     context,
		environment: environment,
	}
}

func (c *cmd) Command() string {
	return c.command
}

func (c *cmd) Context() string {
	return c.context
}

func (c *cmd) Environment() map[string]string {
	return c.environment
}

func (c *cmd) PID() int {
	if c.core == nil || c.core.Process == nil {
		return 0
	}
	return c.core.Process.Pid
}

func (c *cmd) Run() error {
	// c.core = exec.Command("/bin/sh", "-c", c.command)

	nameAndArgs := strings.Split(c.command, " ")
	c.core = exec.Command(nameAndArgs[0], nameAndArgs[1:]...)

	// cmd := exec.Command(nameAndArgs[0])
	// cmd.Args = nameAndArgs

	cmd := c.core
	environment := os.Environ()
	for k, v := range c.environment {
		environment = append(environment, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Env = environment
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = c.context

	return cmd.Run()
}

func (c *cmd) Kill() error {
	if c.core == nil || c.core.Process == nil {
		return nil
	}

	if c.core.ProcessState != nil && c.core.ProcessState.Exited() {
		fmt.Printf("pid %d stat %v\n", c.PID(), c.core.ProcessState.Exited())
		return nil
	}

	return c.core.Process.Kill()
}

func (c *cmd) Restart() error {
	if err := c.Kill(); err != nil {
		return fmt.Errorf("failed to kill: %s", err)
	}

	if err := c.Run(); err != nil {
		return fmt.Errorf("failed to run: %s", err)
	}

	return nil
}
