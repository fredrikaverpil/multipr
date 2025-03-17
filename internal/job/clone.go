package job

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/fredrikaverpil/multipr/internal/git"
)

func (m *Manager) cloneRepositories(repos []*git.Repo) ([]*git.Repo, error) {
	var mu sync.Mutex
	var clonedRepos []*git.Repo
	var errs []error

	m.log.Info("Cloning repositories...")

	if err := os.MkdirAll(m.reposDir, DefaultFilePerms); err != nil {
		return nil, fmt.Errorf("failed to create repositories directory: %w", err)
	}

	reposToClone := []*git.Repo{}
	for _, repo := range repos {
		if _, err := os.Stat(repo.LocalPath()); err == nil {
			m.log.Info(fmt.Sprintf("Repository %s already exists at %s", repo.FullName, repo.LocalPath()))
			continue
		}
		reposToClone = append(reposToClone, repo)
	}

	for _, repo := range reposToClone {
		m.pool.Submit(func() {
			m.log.Info(fmt.Sprintf("Cloning repository %s...", repo.String()))
			if err := repo.Clone(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to clone %s: %w", repo.FullName, err))
				mu.Unlock()

				return
			}

			if err := repo.CheckoutDefaultBranch(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to checkout default branch for %s: %w", repo.FullName, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			clonedRepos = append(clonedRepos, repo)
			mu.Unlock()
		})
	}

	m.pool.Wait()

	if len(errs) > 0 {
		return clonedRepos, errors.Join(errs...)
	}

	if len(clonedRepos) > 0 {
		m.log.Info(fmt.Sprintf("Successfully cloned %d repositories", len(clonedRepos)))
		for _, repo := range clonedRepos {
			m.log.Info(fmt.Sprintf("  - %s -> %s", repo.String(), repo.LocalPath()))
		}
	}

	return clonedRepos, nil
}
