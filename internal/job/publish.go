package job

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/fredrikaverpil/multipr/internal/command"
	"github.com/fredrikaverpil/multipr/internal/git"
)

func (m *Manager) publishRepositories(repos []*git.Repo) error {
	var mu sync.Mutex
	var errs []error

	if !m.options.Publish {
		return nil
	}

	m.log.Info("Publishing PRs for repositories...")

	for _, repo := range repos {
		m.pool.Submit(func() {
			draftFlag := ""
			if m.options.Draft {
				draftFlag = "--draft"
			}

			// Push branch to remote
			if err := repo.PushBranch(m.config.PR.GitHub.Branch); err != nil {
				mu.Lock()
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to push branch for %s: %w", repo.LocalPath(), err))
				mu.Unlock()
				mu.Unlock()
				return
			}

			// Check if PR already exists
			exists, prNumber, err := repo.CheckPRExists(m.config.PR.GitHub.Branch)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to check if PR exists for %s: %w", repo.LocalPath(), err))
				mu.Unlock()
				return
			}

			if exists {
				m.log.Info(fmt.Sprintf("Editing existing PR #%s for %s", prNumber, filepath.Base(repo.LocalPath())))
				_, err = m.exec.Execute(
					"gh",
					[]string{"pr", "edit", prNumber, "--title", m.config.PR.GitHub.Title, "--body", m.config.PR.GitHub.Body},
					command.WithDir(repo.LocalPath()),
				)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to edit PR for %s: %w", repo.LocalPath(), err))
					mu.Unlock()
					return
				}

				var draftCmd string
				if m.options.Draft {
					m.log.Info(fmt.Sprintf("Marking existing PR #%s as draft", prNumber))
					draftCmd = fmt.Sprintf("gh pr ready --undo %s", prNumber)
				} else {
					m.log.Info(fmt.Sprintf("Marking existing PR #%s as ready for review", prNumber))
					draftCmd = fmt.Sprintf("gh pr ready %s", prNumber)
				}

				_, err = m.exec.ExecuteWithShell(draftCmd, command.WithDir(repo.LocalPath()))
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to update PR draft status for %s: %w", repo.LocalPath(), err))
					mu.Unlock()
					return
				}
			} else {
				m.log.Info(fmt.Sprintf("Creating PR for %s", filepath.Base(repo.LocalPath())))
				_, err = m.exec.Execute("gh", []string{
					"pr",
					"create",
					"--assignee",
					"@me",
					"--title",
					m.config.PR.GitHub.Title,
					"--body",
					m.config.PR.GitHub.Body,
					draftFlag,
				}, command.WithDir(repo.LocalPath()))
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to create PR for %s: %w", repo.LocalPath(), err))
					mu.Unlock()
					return
				}
			}
		})
	}

	m.pool.Wait()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
