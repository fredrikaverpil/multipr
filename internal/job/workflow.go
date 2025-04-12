package job

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fredrikaverpil/multipr/internal/git"
)

// RunWorkflow executes the job workflow.
func (m *Manager) RunWorkflow() error {
	m.logJobStart()

	if err := m.handleCleanup(); err != nil {
		return err
	}

	repos, err := m.handleRepositorySearch()
	if err != nil {
		return err
	}

	if len(repos) > 0 {
		if err = m.handleRepositoryCloning(repos); err != nil {
			return err
		}
	}

	eligibleRepos, err := m.handleEligibleRepoIdentification()
	if err != nil {
		return err
	}

	processedRepos, err := m.handleRepositoryProcessing(eligibleRepos)
	if err != nil {
		return err
	}

	if len(processedRepos) > 0 {
		if err = m.handlePublishing(processedRepos); err != nil {
			return err
		}
	}

	m.logJobCompletion()
	return nil
}

func (m *Manager) logJobStart() {
	if m.options.Publish {
		m.log.Info(fmt.Sprintf("🚀 Executing job: %s", m.config.Name))
	} else {
		m.log.Info(fmt.Sprintf("☂️ Executing job (dry-run): %s", m.config.Name))
	}
	m.log.Debug(fmt.Sprintf("Using concurrency pool with %d workers", m.options.Workers))
}

func (m *Manager) handleCleanup() error {
	if !m.options.Clean {
		return nil
	}

	if m.options.ReviewSteps && !m.confirmStep("Remove all previously cloned repos?") {
		return nil
	}

	m.log.Info("Removing all previously cloned repos...")
	if err := os.RemoveAll(m.reposDir); err != nil {
		return fmt.Errorf("failed to remove repos directory: %w", err)
	}
	return nil
}

func (m *Manager) handleRepositorySearch() ([]*git.Repo, error) {
	if m.options.SkipSearch {
		return nil, nil
	}

	if m.options.ReviewSteps && !m.confirmStep("Search for repositories?") {
		return nil, nil
	}

	repos, err := m.searchRepositories()
	if err != nil {
		return nil, fmt.Errorf("error searching repositories: %w", err)
	}
	return repos, nil
}

func (m *Manager) handleRepositoryCloning(repos []*git.Repo) error {
	if m.options.ReviewSteps && !m.confirmStep(fmt.Sprintf("Clone down repositories to local disk? [%s]", m.reposDir)) {
		return nil
	}

	_, err := m.cloneRepositories(repos)
	if err != nil {
		return fmt.Errorf("error cloning repositories: %w", err)
	}
	return nil
}

func (m *Manager) handleEligibleRepoIdentification() ([]*git.Repo, error) {
	if m.options.ReviewSteps && !m.confirmStep("Identify eligible repositories?") {
		return nil, nil
	}

	eligibleRepos, err := m.identifyEligibleRepos(m.reposDir)
	if err != nil {
		return nil, fmt.Errorf("error identifying eligible repositories: %w", err)
	}
	return eligibleRepos, nil
}

func (m *Manager) handleRepositoryProcessing(eligibleRepos []*git.Repo) ([]*git.Repo, error) {
	if m.options.ReviewSteps && !m.confirmStep("Process desired changes in repositories?") {
		return nil, nil
	}

	processedRepos, err := m.processRepositories(eligibleRepos)
	if err != nil {
		return nil, fmt.Errorf("error processing repositories: %w", err)
	}
	return processedRepos, nil
}

func (m *Manager) handlePublishing(processedRepos []*git.Repo) error {
	if m.options.ReviewSteps && !m.confirmStep("Publish changes as PRs?") {
		return nil
	}

	err := m.publishRepositories(processedRepos)
	if err != nil {
		return fmt.Errorf("error publishing PRs: %w", err)
	}
	return nil
}

func (m *Manager) logJobCompletion() {
	if !m.options.Publish {
		m.log.Info("☂️ To publish PRs, use the -publish flag")
	} else {
		m.log.Info("🚀 PRs published successfully!")
	}
}

// confirmStep asks the user to confirm a step.
func (m *Manager) confirmStep(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	// Replace forbidden fmt.Printf with logger
	m.log.Info(message + " (y/n): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	return strings.ToLower(text) == "y" || strings.ToLower(text) == "yes"
}
