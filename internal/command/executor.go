package command

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fredrikaverpil/multipr/internal/log"
)

type Result struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}

type Executor struct {
	log         *log.Logger
	debug       bool
	outputMutex sync.Mutex
}

func NewExecutor(logger *log.Logger, debug bool) *Executor {
	return &Executor{log: logger, debug: debug}
}

func (e *Executor) ExecuteWithShell(command string, opts ...Option) (*Result, error) {
	cmd := exec.Command("bash", "-c", command)
	return e.execute(cmd, opts...)
}

func (e *Executor) Execute(command string, args []string, opts ...Option) (*Result, error) {
	cmd := exec.Command(command, args...)
	return e.execute(cmd, opts...)
}

func (e *Executor) execute(cmd *exec.Cmd, opts ...Option) (*Result, error) {
	// Set up options
	options := &execOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.dir != "" {
		cmd.Dir = options.dir
	}
	multiWriter := options.tee || e.debug

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if multiWriter {
		e.log.Debug(fmt.Sprintf("Executing command: %s", cmd.String()))
		// Use MultiWriter to tee output to both terminal and buffer
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	}

	err := cmd.Run()

	result := &Result{
		Command:  cmd.String(),
		ExitCode: 0,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()

			// NOTE: the identification command can fail with exit code 1, and it can be expected.
			// In that case it just means that the repo was not elibible for a code change.
			//
			// TODO: should we instead consider any failing command a problem, log and exit?
			//
			// e.log.Error(
			// 	fmt.Sprintf("Command failed with exit code %d", result.ExitCode),
			// 	slog.String("stderr", result.Stderr),
			// 	slog.String("stdout", result.Stdout),
			// )

			// Create and return custom error with structured information
			return result, &ExecError{
				Command:  result.Command,
				ExitCode: result.ExitCode,
				Stdout:   result.Stdout,
				Stderr:   result.Stderr,
				Err:      err,
			}
		}
		// For non-exit errors (like command not found)
		return result, fmt.Errorf("failed to execute command '%s': %w", result.Command, err)
	}

	e.log.Debug(fmt.Sprintf("Exit code: %d", result.ExitCode))

	return result, nil
}
