package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fredrikaverpil/multipr/internal/command"
	"github.com/fredrikaverpil/multipr/internal/log"
)

const (
	DefaultFilePerms = 0o755
)

type Repo struct {
	Host     string
	FullName string
	ReposDir string
	executor *command.Executor
	log      *log.Logger
}

func NewRepo(host, fullName, reposDir string, executor *command.Executor, logger *log.Logger) *Repo {
	return &Repo{
		Host:     host,     // e.g. "github.com"
		FullName: fullName, // e.g. "fredrikaverpil/multipr"
		ReposDir: reposDir,
		executor: executor,
		log:      logger,
	}
}

func (r *Repo) String() string {
	return fmt.Sprintf("%s/%s", r.Host, r.FullName)
}

func (r *Repo) LocalPath() string {
	parts := strings.Split(r.FullName, "/")
	username := parts[0]
	repoName := parts[1]
	return filepath.Join(r.ReposDir, r.Host, username, repoName)
}

func (r *Repo) Clone() error {
	if _, err := os.Stat(r.LocalPath()); err == nil {
		r.log.Info(fmt.Sprintf("Repository already exists at %s", r.LocalPath()))
		return nil
	}

	// TODO: switch/case on host (or use interface?)
	return r.executor.GHClone(r.FullName, r.LocalPath())
}

// CheckoutDefaultBranch checks out the default branch and resets it.
func (r *Repo) CheckoutDefaultBranch() error {
	// Use git symbolic-ref to get the default branch reference
	result, err := r.executor.Execute(
		"git",
		[]string{"symbolic-ref", "refs/remotes/origin/HEAD"},
		command.WithDir(r.LocalPath()),
	)
	if err != nil {
		return fmt.Errorf("failed to get default branch: %w", err)
	}

	// Extract just the branch name from the full reference path
	// Input example: "refs/remotes/origin/main"
	refPath := strings.TrimSpace(result.Stdout)
	defaultBranch := strings.TrimPrefix(refPath, "refs/remotes/origin/")

	if err = r.executor.GitFetchAll(r.LocalPath()); err != nil {
		return err
	}
	if err = r.executor.GitCheckout(r.LocalPath(), defaultBranch); err != nil {
		return err
	}
	if err = r.executor.GitResetHard(r.LocalPath(), defaultBranch); err != nil {
		return err
	}

	return nil
}

func (r *Repo) ShowDiff() error {
	return r.executor.GitDiff(r.LocalPath())
}

// CheckoutNewBranch creates and checks out a new branch.
func (r *Repo) CheckoutNewBranch(branchName string) error {
	return r.executor.GitCheckout(r.LocalPath(), branchName)
}

// CheckPRExists checks if a PR already exists for the given branch.
func (r *Repo) CheckPRExists(branchName string) (bool, string, error) {
	result, err := r.executor.ExecuteWithShell(
		fmt.Sprintf("gh pr list --head %s --json number --jq '.[0].number'", branchName),
		"",
		command.WithDir(r.LocalPath()))
	if err != nil {
		return false, "", fmt.Errorf("failed to check for existing PR: %w", err)
	}

	// If output is empty, no PR exists
	if result.Stdout == "" {
		return false, "", nil
	}

	return true, result.Stdout, nil
}

// PushBranch pushes the current branch to the remote.
func (r *Repo) PushBranch(branchName string) error {
	return r.executor.GitPushForce(r.LocalPath(), branchName)
}

// CreateCommit creates a commit with the given message.
func (r *Repo) CreateCommit(message string) error {
	if err := r.executor.GitAddAll(r.LocalPath()); err != nil {
		return err
	}
	if err := r.executor.GitCommit(r.LocalPath(), message); err != nil {
		return err
	}

	return nil
}

func (r *Repo) HasChanges() (bool, error) {
	result, err := r.executor.Execute("git", []string{"status", "--porcelain"}, command.WithDir(r.LocalPath()))
	if err != nil {
		return false, fmt.Errorf("failed to check status: %w", err)
	}

	return result.Stdout != "", nil
}

// TODO: implement
// func (r *Repo) Publish() error {
//  // TODO: switch/case on host (or use interface?)
// 	return r.executor.GHPublish(r.LocalPath())
// }
