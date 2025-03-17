package command

import "fmt"

type ExecError struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

func (e *ExecError) Error() string {
	return fmt.Sprintf("command '%s' failed with exit code %d: %s",
		e.Command, e.ExitCode, e.Stderr)
}

func (e *ExecError) Unwrap() error {
	return e.Err
}
