package shell

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
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
}

func (s Shell) Run(ctx context.Context) error {
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
	cmd.Stdin = s.Stdin
	cmd.Stdout = s.Stdout
	cmd.Stderr = s.Stderr
	cmd.Dir = s.Dir
	for k, v := range s.Env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Cancel = func() error {
		slog.Debug("cancel recieved")
		return cmd.Process.Signal(os.Interrupt)
	}

	slog.Info("running shell", "command", s.Command, "args", s.Args)
	return cmd.Run()
}
