package command

import (
	"context"
	"fmt"
)

func (e *Executor) GitClone(ctx context.Context, repo, path string) error {
	_, err := e.Execute(ctx, "git", []string{"clone", repo, path})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	return nil
}

func (e *Executor) GitFetchAll(ctx context.Context, dir string) error {
	_, err := e.Execute(ctx, "git", []string{"fetch", "--all"}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to fetch all: %w", err)
	}
	return nil
}

func (e *Executor) GitCheckout(ctx context.Context, dir, branch string) error {
	_, err := e.Execute(ctx, "git", []string{"checkout", "-B", branch}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}
	return nil
}

func (e *Executor) GitResetHard(ctx context.Context, dir, branch string) error {
	_, err := e.Execute(ctx, "git", []string{"reset", "--hard", "origin/" + branch}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}
	return nil
}

func (e *Executor) GitDiff(ctx context.Context, dir string) error {
	e.outputMutex.Lock()
	defer e.outputMutex.Unlock()

	_, err := e.Execute(ctx, "pwd", []string{}, WithDir(dir), WithTee())
	if err != nil {
		return fmt.Errorf("failed to show pwd: %w", err)
	}
	_, err = e.Execute(ctx, "git", []string{"diff", "--color=always", "--cached"}, WithDir(dir), WithTee())
	if err != nil {
		return fmt.Errorf("failed to show diff: %w", err)
	}
	return nil
}

func (e *Executor) GitPushForce(ctx context.Context, dir, branch string) error {
	_, err := e.Execute(ctx, "git", []string{"push", "-u", "origin", branch, "--force-with-lease"}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}
	return nil
}

func (e *Executor) GitAddAll(ctx context.Context, dir string) error {
	_, err := e.Execute(ctx, "git", []string{"add", "-A"}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}
	return nil
}

func (e *Executor) GitCommit(ctx context.Context, dir, message string) error {
	_, err := e.Execute(ctx, "git", []string{"commit", "-m", message}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}
