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
			eligible, checkoutErr := m.isRepoEligible(repo)
			if checkoutErr != nil {
				mu.Lock()
				errs = append(errs, checkoutErr)
				mu.Unlock()
				return
			}

			if eligible {
				mu.Lock()
				eligibleRepos = append(eligibleRepos, repo)
				mu.Unlock()
			}
		})
	}

	m.pool.Wait()

	if len(errs) > 0 {
		return eligibleRepos, errors.Join(errs...)
	}

	m.logEligibleRepos(eligibleRepos)
	return eligibleRepos, nil
}

// isRepoEligible checks if a repository is eligible based on identification commands.
func (m *Manager) isRepoEligible(repo *git.Repo) (bool, error) {
	checkoutErr := repo.CheckoutDefaultBranch()
	if checkoutErr != nil {
		return false, fmt.Errorf("failed to checkout default branch for %s: %w", repo.LocalPath(), checkoutErr)
	}

	// Try each identification command
	for _, identify := range m.config.Identify {
		if m.options.Debug {
			m.log.Debug(fmt.Sprintf("Running identification command '%s' on %s\n", identify.Name, repo.LocalPath()))
		}

		// Run identification command
		result, cmdErr := m.exec.ExecuteWithShell(identify.Cmd, identify.Shell, command.WithDir(repo.LocalPath()))
		if cmdErr != nil {
			if result == nil {
				return false, fmt.Errorf("identification command '%s' failed for %s: %w", identify.Name, repo.LocalPath(), cmdErr)
			}
			continue
		}

		// Check if eligible (exit code 0)
		if result.ExitCode == 0 {
			return true, nil // Repository is eligible
		}
	}

	return false, nil
}

// logEligibleRepos logs information about eligible repositories.
func (m *Manager) logEligibleRepos(eligibleRepos []*git.Repo) {
	m.log.Info(fmt.Sprintf("Found %d eligible repositories", len(eligibleRepos)))
	for _, repo := range eligibleRepos {
		m.log.Info(fmt.Sprintf("  - %s", repo.String()))
	}
}

// FIXME: move into git/repo.go?
func rebuildRepos(reposDir string, executor *command.Executor, log *log.Logger) ([]*git.Repo, error) {
	var repos []*git.Repo

	hosts, readErr := os.ReadDir(reposDir)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", reposDir, readErr)
	}

	for _, host := range hosts {
		if !host.IsDir() {
			continue
		}

		hostRepos, hostErr := processHostDir(reposDir, host.Name(), executor, log)
		if hostErr != nil {
			return nil, hostErr
		}

		repos = append(repos, hostRepos...)
	}

	return repos, nil
}

// processHostDir processes a host directory and returns repositories.
func processHostDir(reposDir, hostName string, executor *command.Executor, log *log.Logger) ([]*git.Repo, error) {
	var repos []*git.Repo
	var fullNames []string

	hostPath := filepath.Join(reposDir, hostName)
	usernames, readErr := os.ReadDir(hostPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", hostPath, readErr)
	}

	for _, username := range usernames {
		if !username.IsDir() {
			continue
		}

		userFullNames, userErr := getUserProjects(reposDir, hostName, username.Name())
		if userErr != nil {
			return nil, userErr
		}

		fullNames = append(fullNames, userFullNames...)
	}

	for _, fullName := range fullNames {
		repo := git.NewRepo(hostName, fullName, reposDir, executor, log)
		repos = append(repos, repo)
	}

	return repos, nil
}

// getUserProjects gets a list of project full names for a user.
func getUserProjects(reposDir, hostName, userName string) ([]string, error) {
	var fullNames []string

	userPath := filepath.Join(reposDir, hostName, userName)
	projects, readErr := os.ReadDir(userPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", userPath, readErr)
	}

	for _, project := range projects {
		if !project.IsDir() {
			continue
		}
		fullNames = append(fullNames, fmt.Sprintf("%s/%s", userName, project.Name()))
	}

	return fullNames, nil
}
