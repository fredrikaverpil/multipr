package job

import (
	"errors"
	"fmt"
	"sync"

	"github.com/fredrikaverpil/multipr/internal/command"
	"github.com/fredrikaverpil/multipr/internal/git"
)

// processRepository handles the changes for a single repository.
func (m *Manager) processRepository(repo *git.Repo) error {
	// Check out default branch
	if err := repo.CheckoutDefaultBranch(); err != nil {
		return fmt.Errorf("failed to checkout default branch for %s: %w", repo.LocalPath(), err)
	}

	// Create new branch for changes
	branchName := m.config.PR.GitHub.Branch
	if err := repo.CheckoutNewBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch for %s: %w", repo.LocalPath(), err)
	}

	// Apply each change
	if err := m.applyChanges(repo); err != nil {
		return err
	}

	// Check if there are any changes
	hasChanges, err := repo.HasChanges()
	if err != nil {
		return fmt.Errorf("failed to check for changes in %s: %w", repo.LocalPath(), err)
	}

	if !hasChanges {
		return fmt.Errorf("no changes detected (likely a problem?) in %s", repo.LocalPath())
	}

	// Handle stages, diffs and commits
	if err = m.handleStagingAndCommits(repo); err != nil {
		return err
	}

	return nil
}

// applyChanges applies all configured changes to a repository.
func (m *Manager) applyChanges(repo *git.Repo) error {
	for _, change := range m.config.Changes {
		_, err := m.exec.ExecuteWithShell(change.Cmd, command.WithDir(repo.LocalPath()))
		if err != nil {
			return fmt.Errorf("failed to apply change '%s' to %s: %w", change.Name, repo.LocalPath(), err)
		}
	}
	return nil
}

// handleStagingAndCommits manages the staging, diffing, and committing of changes.
func (m *Manager) handleStagingAndCommits(repo *git.Repo) error {
	// Stage changes first
	if !m.options.ManualCommit {
		_, err := m.exec.Execute("git", []string{"add", "-A"}, command.WithDir(repo.LocalPath()))
		if err != nil {
			return fmt.Errorf("failed to stage changes for %s: %w", repo.LocalPath(), err)
		}
	}

	// Now show the staged diff
	if m.options.ShowDiffs {
		err := repo.ShowDiff()
		if err != nil {
			return fmt.Errorf("failed to show diff for %s: %w", repo.LocalPath(), err)
		}
	}

	// Create commit if not manual
	if !m.options.ManualCommit {
		// Now commit the staged changes
		_, err := m.exec.Execute(
			"git",
			[]string{"commit", "-m", m.config.PR.GitHub.Title},
			command.WithDir(repo.LocalPath()),
		)
		if err != nil {
			return fmt.Errorf("failed to create commit for %s: %w", repo.LocalPath(), err)
		}
	}

	return nil
}

func (m *Manager) processRepositories(repos []*git.Repo) ([]*git.Repo, error) {
	m.log.Info("Processing repositories and preparing changes...")

	var mu sync.Mutex
	var processedRepos []*git.Repo
	var errs []error

	for _, repo := range repos {
		m.pool.Submit(func() {
			err := m.processRepository(repo)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			// Add to processed repos list
			mu.Lock()
			processedRepos = append(processedRepos, repo)
			mu.Unlock()
		})
	}

	m.pool.Wait()

	if len(errs) > 0 {
		return processedRepos, errors.Join(errs...)
	}

	return processedRepos, nil
}
