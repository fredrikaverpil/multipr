package job

import (
	"errors"
	"fmt"
	"sync"

	"github.com/fredrikaverpil/multipr/internal/command"
	"github.com/fredrikaverpil/multipr/internal/git"
)

func (m *Manager) processRepositories(repos []*git.Repo) ([]*git.Repo, error) {
	m.log.Info("Processing repositories and preparing changes...")

	var mu sync.Mutex
	var processedRepos []*git.Repo
	var errs []error

	for _, repo := range repos {
		m.pool.Submit(func() {
			// Check out default branch
			if err := repo.CheckoutDefaultBranch(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to checkout default branch for %s: %w", repo.LocalPath(), err))
				mu.Unlock()
				return
			}

			// Create new branch for changes
			branchName := m.config.PR.GitHub.Branch
			if err := repo.CheckoutNewBranch(branchName); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to create branch for %s: %w", repo.LocalPath(), err))
				mu.Unlock()
				return
			}

			// Apply each change
			for _, change := range m.config.Changes {
				_, err := m.exec.ExecuteWithShell(change.Cmd, command.WithDir(repo.LocalPath()))
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to apply change '%s' to %s: %w", change.Name, repo.LocalPath(), err))
					mu.Unlock()
					return
				}
			}

			// Check if there are any changes
			hasChanges, err := repo.HasChanges()
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to check for changes in %s: %w", repo.LocalPath(), err))
				mu.Unlock()
				return
			}

			if !hasChanges {
				mu.Lock()
				errs = append(errs, fmt.Errorf("no changes detected (likely a problem?) in %s", repo.LocalPath()))
				mu.Unlock()
				return
			}

			// Stage changes first
			if !m.options.ManualCommit {
				_, err = m.exec.Execute("git", []string{"add", "-A"}, command.WithDir(repo.LocalPath()))
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to stage changes for %s: %w", repo.LocalPath(), err))
					mu.Unlock()
					return
				}
			}

			// Now show the staged diff
			if m.options.ShowDiffs {
				if err := repo.ShowDiff(); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to show diff for %s: %w", repo.LocalPath(), err))
					mu.Unlock()
					return
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
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to create commit for %s: %w", repo.LocalPath(), err))
					mu.Unlock()
					return
				}
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
