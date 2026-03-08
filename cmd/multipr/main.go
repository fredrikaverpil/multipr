package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fredrikaverpil/multipr/internal/config"
	"github.com/fredrikaverpil/multipr/internal/job"
)

const cpuMultiplier = 2

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	clean := flag.Bool("clean", false, "Remove cloned repositories before run")
	debug := flag.Bool("debug", false, "Print all commands and their output")
	draft := flag.Bool("draft", false, "Make PRs into drafts")
	help := flag.Bool("help", false, "Show help")
	jobFile := flag.String("job", "", "Path to the YAML job file (required)")
	manualCommit := flag.Bool("manual-commit", false, "User manages git commits in shell commands")
	publish := flag.Bool("publish", false, "Publish PRs")
	reviewSteps := flag.Bool("review", false, "Manual review of each major step")
	showDiffs := flag.Bool("show-diffs", true, "Show each git diff")
	skipSearch := flag.Bool("skip-search", false, "Skip search for repositories")
	workers := flag.Int("workers", 0, "Number of workers to use for concurrency (default: 2x CPU cores)")
	shell := flag.String("shell", "bash", "Shell to use for executing commands")

	flag.Parse()

	if *help {
		flag.Usage()
		return nil
	}
	if *jobFile == "" {
		flag.Usage()
		return errors.New("job file is required")
	}

	// Load configuration from YAML file
	cfg, err := config.LoadFromFile(*jobFile)
	if err != nil {
		return fmt.Errorf("loading job file: %w", err)
	}

	// Set default workers if auto-detection requested
	if *workers <= 0 {
		cpus := runtime.NumCPU()
		*workers = cpus * cpuMultiplier // Reasonable for I/O-bound work
	}

	// Create run options
	opts := &job.CLIOptions{
		Clean:        *clean,
		Debug:        *debug,
		Draft:        *draft,
		ManualCommit: *manualCommit,
		Publish:      *publish,
		ReviewSteps:  *reviewSteps,
		Shell:        *shell,
		ShowDiffs:    *showDiffs,
		SkipSearch:   *skipSearch,
		Workers:      *workers,
	}

	// Create context that cancels on interrupt signals
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create and run the workflow
	r, err := job.NewManager(cfg, opts, *jobFile)
	if err != nil {
		return err
	}
	return r.RunWorkflow(ctx)
}
