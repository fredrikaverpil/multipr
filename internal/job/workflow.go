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
	if m.options.Publish {
		m.log.Info(fmt.Sprintf("🚀 Executing job: %s", m.config.Name))
	} else {
		m.log.Info(fmt.Sprintf("☂️ Executing job (dry-run): %s", m.config.Name))
	}
	m.log.Debug(fmt.Sprintf("Using concurrency pool with %d workers", m.options.Workers))

	// Clean up
	if m.options.Clean {
		if m.options.ReviewSteps {
			if !m.confirmStep("Remove all previously cloned repos?") {
				return nil
			}
		}
		m.log.Info("Removing all previously cloned repos...")
		if err := os.RemoveAll(m.reposDir); err != nil {
			return fmt.Errorf("failed to remove repos directory: %w", err)
		}
	}

	// Search for repositories
	var repos []*git.Repo
	if !m.options.SkipSearch {
		if m.options.ReviewSteps {
			if !m.confirmStep("Search for repositories?") {
				return nil
			}
		}
		var err error
		repos, err = m.searchRepositories()
		if err != nil {
			return fmt.Errorf("error searching repositories: %w", err)
		}
	}

	// Clone repositories
	if len(repos) > 0 {
		if m.options.ReviewSteps {
			if !m.confirmStep(fmt.Sprintf("Clone down repositories to local disk? [%s]", m.reposDir)) {
				return nil
			}
		}
		_, err := m.cloneRepositories(repos)
		if err != nil {
			return fmt.Errorf("error cloning repositories: %w", err)
		}
	}

	// Identify eligible repos
	if m.options.ReviewSteps {
		if !m.confirmStep("Identify eligible repositories?") {
			return nil
		}
	}
	eligibleRepos, err := m.identifyEligibleRepos(m.reposDir)
	if err != nil {
		return fmt.Errorf("error identifying eligible repositories: %w", err)
	}

	// Process changes in repos
	if m.options.ReviewSteps {
		if !m.confirmStep("Process desired changes in repositories?") {
			return nil
		}
	}
	processedRepos, err := m.processRepositories(eligibleRepos)
	if err != nil {
		return fmt.Errorf("error processing repositories: %w", err)
	}

	// Publish PRs
	if len(processedRepos) > 0 {
		if m.options.ReviewSteps {
			if !m.confirmStep("Publish changes as PRs?") {
				return nil
			}
		}
		err = m.publishRepositories(processedRepos)
		if err != nil {
			return fmt.Errorf("error publishing PRs: %w", err)
		}
	}

	if !m.options.Publish {
		m.log.Info("☂️ To publish PRs, use the --publish flag")
	} else {
		m.log.Info("🚀 PRs published successfully!")
	}
	return nil
}

// confirmStep asks the user to confirm a step.
func (m *Manager) confirmStep(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s (y/n): ", message)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	return strings.ToLower(text) == "y" || strings.ToLower(text) == "yes"
}
