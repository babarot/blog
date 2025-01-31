package shell

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

	"github.com/babarot/blog/internal/env"
)

func New(command string, args ...string) Shell {
	return Shell{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Env:     map[string]string{},
		Command: command,
		Args:    args,
	}
}

type Shell struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Env     map[string]string
	Command string
	Args    []string
	Dir     string

	StartingMsg string
}

func (s Shell) Run(ctx context.Context) error {
	if msg := s.StartingMsg; msg != "" {
		slog.Info(s.StartingMsg)
	}

	command := s.Command
	if _, err := exec.LookPath(command); err != nil {
		return err
	}
	for _, arg := range s.Args {
		command += " " + arg
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	file, err := os.OpenFile(env.BLOG_LOG_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		cmd.Stderr = file
		cmd.Stdout = file
	}
	cmd.Stdin = s.Stdin
	cmd.Dir = s.Dir
	for k, v := range s.Env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Cancel = func() error {
		slog.Debug("cancel recieved")
		return cmd.Process.Signal(os.Interrupt)
	}

	return cmd.Run()
}
