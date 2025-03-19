package job

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fredrikaverpil/multipr/internal/command"
	"github.com/fredrikaverpil/multipr/internal/config"
	"github.com/fredrikaverpil/multipr/internal/log"
	"github.com/fredrikaverpil/multipr/internal/worker"
)

const (
	DefaultFilePerms = 0o755
)

// CLIOptions holds runtime options for the job runner.
type CLIOptions struct {
	Clean        bool
	Debug        bool
	Draft        bool
	ManualCommit bool
	Publish      bool
	ReviewSteps  bool
	Shell        string
	ShowDiffs    bool
	SkipSearch   bool
	Workers      int
}

// Manager manages the execution of a job.
type Manager struct {
	config   *config.JobConfig
	options  *CLIOptions
	workDir  string
	reposDir string
	log      *log.Logger
	exec     *command.Executor
	pool     *worker.Pool
}

// NewManager creates a new Runner.
func NewManager(config *config.JobConfig, opts *CLIOptions, jobFilePath string) (*Manager, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	jobName := config.Name
	if jobName == "" {
		jobFileName := filepath.Base(jobFilePath)
		jobName = strings.TrimSuffix(jobFileName, filepath.Ext(jobFileName))
		config.Name = jobName
	}
	workDir := filepath.Join(currentDir, "jobs", jobName)
	reposDir := filepath.Join(workDir, "repos")

	logger, err := log.NewLogger(log.Options{
		LevelDebug: opts.Debug,
		LogFile:    filepath.Join(workDir, "joblog.json"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	logger.Debug("Config loaded: %v", config)
	exec := command.NewExecutor(opts.Debug, opts.Shell, logger)
	pool := worker.NewWorkerPool(opts.Workers)

	return &Manager{
		config:   config,
		options:  opts,
		workDir:  workDir,
		reposDir: reposDir,
		log:      logger,
		exec:     exec,
		pool:     pool,
	}, nil
}
