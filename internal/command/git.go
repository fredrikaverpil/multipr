package command

import "fmt"

func (e *Executor) GitClone(repo, path string) error {
	_, err := e.Execute("git", []string{"clone", repo, path})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	return nil
}

func (e *Executor) GitFetchAll(dir string) error {
	_, err := e.Execute("git", []string{"fetch", "--all"}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to fetch all: %w", err)
	}
	return nil
}

func (e *Executor) GitCheckout(dir, branch string) error {
	_, err := e.Execute("git", []string{"checkout", "-B", branch}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}
	return nil
}

func (e *Executor) GitResetHard(dir, branch string) error {
	_, err := e.Execute("git", []string{"reset", "--hard", "origin/" + branch}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}
	return nil
}

func (e *Executor) GitDiff(dir string) error {
	e.outputMutex.Lock()
	defer e.outputMutex.Unlock()

	_, err := e.Execute("pwd", []string{}, WithDir(dir), WithTee())
	if err != nil {
		return fmt.Errorf("failed to show pwd: %w", err)
	}
	_, err = e.Execute("git", []string{"diff", "--color=always", "--cached"}, WithDir(dir), WithTee())
	if err != nil {
		return fmt.Errorf("failed to show diff: %w", err)
	}
	return nil
}

func (e *Executor) GitPushForce(dir, branch string) error {
	_, err := e.Execute("git", []string{"push", "-u", "origin", branch, "--force-with-lease"}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}
	return nil
}

func (e *Executor) GitAddAll(dir string) error {
	_, err := e.Execute("git", []string{"add", "-A"}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}
	return nil
}

func (e *Executor) GitCommit(dir, message string) error {
	_, err := e.Execute("git", []string{"commit", "-m", message}, WithDir(dir))
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}
