package job

import (
	"fmt"
	"os"

	"github.com/fredrikaverpil/multipr/internal/git"
)

func (m *Manager) searchRepositories() ([]*git.Repo, error) {
	if err := os.MkdirAll(m.reposDir, DefaultFilePerms); err != nil {
		return nil, fmt.Errorf("failed to create repositories directory: %w", err)
	}

	// Check if GitHub search is configured
	if m.config.Search.GitHub.Method != "" {
		method := m.config.Search.GitHub.Method
		query := m.config.Search.GitHub.Query

		m.log.Info(fmt.Sprintf("Searching GitHub using method '%s' with query: %s", method, query))

		var fullNames []string
		var err error

		switch method {
		case "code":
			fullNames, err = m.exec.GHSearchCode(query, 0)
		case "repos":
			fullNames, err = m.exec.GHSearchRepos(query, 0)
		default:
			return nil, fmt.Errorf("unsupported GitHub search method: %s", method)
		}

		if err != nil {
			return nil, fmt.Errorf("failed GitHub search: %w", err)
		}

		m.log.Info(fmt.Sprintf("Found %d repositories", len(fullNames)))
		for _, fullName := range fullNames {
			m.log.Info(fmt.Sprintf("  - github.com/%s", fullName))
		}

		// Convert to repos
		var repos []*git.Repo
		for _, fullName := range fullNames {
			newRepo := git.NewRepo("github.com", fullName, m.reposDir, m.exec, m.log)
			repos = append(repos, newRepo)
		}

		return repos, nil
	}

	return []*git.Repo{}, nil
}
