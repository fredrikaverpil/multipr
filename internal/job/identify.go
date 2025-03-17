package job

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fredrikaverpil/multipr/internal/command"
	"github.com/fredrikaverpil/multipr/internal/git"
	"github.com/fredrikaverpil/multipr/internal/log"
)

func (m *Manager) identifyEligibleRepos(reposDir string) ([]*git.Repo, error) {
	m.log.Info("Identifying eligible repositories...")

	var mu sync.Mutex
	var eligibleRepos []*git.Repo
	var errs []error

	repos, err := rebuildRepos(reposDir, m.exec, m.log)
	if err != nil {
		return nil, fmt.Errorf("failed to rebuild repositories: %w", err)
	}

	// If there are no identification commands, consider all repos eligible
	if len(m.config.Identify) == 0 {
		eligibleRepos = append(eligibleRepos, repos...)
		return eligibleRepos, nil
	}

	for _, repo := range repos {
		m.pool.Submit(func() {
			if err := repo.CheckoutDefaultBranch(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to checkout default branch for %s: %w", repo.LocalPath(), err))
				mu.Unlock()
				return // Skip this repository
			}

			// Try each identification command
			for _, identify := range m.config.Identify {
				if m.options.Debug {
					m.log.Debug(fmt.Sprintf("Running identification command '%s' on %s\n", identify.Name, repo.LocalPath()))
				}

				// Run identification command
				result, err := m.exec.ExecuteWithShell(identify.Cmd, command.WithDir(repo.LocalPath()))
				if err != nil {
					if result == nil {
						mu.Lock()
						errs = append(
							errs,
							fmt.Errorf("identification command '%s' failed for %s: %w", identify.Name, repo.LocalPath(), err),
						)
						mu.Unlock()
						continue
					}
				}

				// Check if eligible (exit code 0)
				if result.ExitCode == 0 {
					mu.Lock()
					eligibleRepos = append(eligibleRepos, repo)
					mu.Unlock()
					return // Repository is eligible, no need to check other commands
				}
			}
		})
	}

	m.pool.Wait()

	if len(errs) > 0 {
		return eligibleRepos, errors.Join(errs...)
	}

	m.log.Info(fmt.Sprintf("Found %d eligible repositories", len(eligibleRepos)))
	for _, repo := range eligibleRepos {
		m.log.Info(fmt.Sprintf("  - %s", repo.String()))
	}

	return eligibleRepos, nil
}

// FIXME: move into git/repo.go?
func rebuildRepos(reposDir string, executor *command.Executor, log *log.Logger) ([]*git.Repo, error) {
	var repos []*git.Repo
	hosts, err := os.ReadDir(reposDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", reposDir, err)
	}
	for _, host := range hosts {
		if !host.IsDir() {
			continue
		}
		var fullNames []string
		usernames, err := os.ReadDir(filepath.Join(reposDir, host.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", reposDir, err)
		}
		for _, username := range usernames {
			if !username.IsDir() {
				continue
			}
			projects, err := os.ReadDir(filepath.Join(reposDir, host.Name(), username.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read directory %s: %w", username.Name(), err)
			}
			for _, project := range projects {
				if !project.IsDir() {
					continue
				}
				fullNames = append(fullNames, fmt.Sprintf("%s/%s", username.Name(), project.Name()))
			}
		}

		for _, fullName := range fullNames {
			repo := git.NewRepo(host.Name(), fullName, reposDir, executor, log)
			repos = append(repos, repo)
		}
	}

	return repos, nil
}
