package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Shell struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Env     map[string]string
	Command string
	Dir     string
}

func (s Shell) exec(ctx context.Context) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", s.Command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", s.Command)
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

	slog.Info("running shell", "command", s.Command)
	return cmd
}

func (s Shell) Run(ctx context.Context) error {
	if s.Command == "" {
		return errors.New("command not found")
	}
	return s.exec(ctx).Run()
}

func Command(command ...string) *exec.Cmd {
	return Shell{
		Command: strings.Join(command, " "),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  io.Discard,
	}.exec(context.Background())
}

func ExpandHome(input string) (string, error) {
	result := input

	// 1. expand tilda
	if strings.HasPrefix(result, "~/") {
		home := os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("HOME environment variable is not set")
		}
		result = strings.Replace(result, "~/", home+"/", 1)
	} else if result == "~" {
		home := os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("HOME environment variable is not set")
		}
		result = home
	}

	// 2. expand env, e.g. $HOME、${HOME}
	for {
		start := strings.Index(result, "$")
		if start == -1 {
			break
		}

		var end int
		var varName string

		if strings.HasPrefix(result[start:], "${") {
			// case of ${VAR} format
			end = strings.Index(result[start:], "}")
			if end == -1 {
				return "", fmt.Errorf("unclosed variable brace in input: %s", input)
			}
			end += start
			varName = result[start+2 : end]
			end++ // go to next of "}"
		} else {
			for i := start + 1; i < len(result); i++ {
				if !isShellVarChar(result[i]) {
					end = i
					break
				}
			}
			if end == 0 {
				end = len(result)
			}
			varName = result[start+1 : end]
		}

		value := os.Getenv(varName)
		if value == "" {
			value = ""
		}

		result = result[:start] + value + result[end:]
	}

	return result, nil
}

func isShellVarChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
