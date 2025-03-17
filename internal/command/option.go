package command

type Option func(*execOptions)

type execOptions struct {
	dir string
	tee bool
}

// WithDir sets the working directory for the command.
func WithDir(dir string) Option {
	return func(o *execOptions) {
		o.dir = dir
	}
}

// WithTee duplicates the command's output to both stdout and stderr.
func WithTee() Option {
	return func(o *execOptions) {
		o.tee = true
	}
}
