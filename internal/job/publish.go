package job

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
			// Push branch to remote
			if err := repo.PushBranch(m.config.PR.GitHub.Branch); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to push branch for %s: %w", repo.LocalPath(), err))
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

			var processErr error
			if exists {
				processErr = m.updateExistingPR(repo, prNumber)
			} else {
				processErr = m.createNewPR(repo)
			}

			if processErr != nil {
				mu.Lock()
				errs = append(errs, processErr)
				mu.Unlock()
			}
		})
	}

	m.pool.Wait()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (m *Manager) updateExistingPR(repo *git.Repo, prNumber string) error {
	repoName := filepath.Base(repo.LocalPath())
	m.log.Info(fmt.Sprintf("Editing existing PR #%s for %s", prNumber, repoName))

	processedBody := m.processBodyTemplate(m.config.PR.GitHub.Body)

	_, err := m.exec.Execute(
		"gh",
		[]string{"pr", "edit", prNumber, "--title", m.config.PR.GitHub.Title, "--body", processedBody},
		command.WithDir(repo.LocalPath()),
	)
	if err != nil {
		return fmt.Errorf("failed to edit PR for %s: %w", repo.LocalPath(), err)
	}

	var draftCmd string
	if m.options.Draft {
		m.log.Info(fmt.Sprintf("Marking existing PR #%s as draft", prNumber))
		draftCmd = fmt.Sprintf("gh pr ready --undo %s", prNumber)
	} else {
		m.log.Info(fmt.Sprintf("Marking existing PR #%s as ready for review", prNumber))
		draftCmd = fmt.Sprintf("gh pr ready %s", prNumber)
	}

	_, err = m.exec.ExecuteWithShell(draftCmd, "", command.WithDir(repo.LocalPath()))
	if err != nil {
		return fmt.Errorf("failed to update PR draft status for %s: %w", repo.LocalPath(), err)
	}

	return nil
}

func (m *Manager) createNewPR(repo *git.Repo) error {
	repoName := filepath.Base(repo.LocalPath())
	m.log.Info(fmt.Sprintf("Creating PR for %s", repoName))

	processedBody := m.processBodyTemplate(m.config.PR.GitHub.Body)

	args := []string{
		"pr",
		"create",
		"--assignee",
		"@me",
		"--title",
		m.config.PR.GitHub.Title,
		"--body",
		processedBody,
	}

	if m.options.Draft {
		args = append(args, "--draft")
	}

	_, err := m.exec.Execute("gh", args, command.WithDir(repo.LocalPath()))
	if err != nil {
		return fmt.Errorf("failed to create PR for %s: %w", repo.LocalPath(), err)
	}

	return nil
}

// processBodyTemplate replaces {yaml} placeholder with the job YAML content.
// It preserves the indentation level of the line containing {yaml}.
func (m *Manager) processBodyTemplate(body string) string {
	if !strings.Contains(body, "{yaml}") {
		return body
	}

	// Read the original job file
	yamlContent, err := os.ReadFile(m.jobFilePath)
	if err != nil {
		m.log.Debug("Failed to read job file for template replacement: %v", err)
		return body
	}

	// Prepare content lines split with normalized newlines
	// Keep trailing newline behavior consistent with source
	content := string(yamlContent)
	// Ensure content ends with a newline, so closing fence is on its own line
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// Replace all occurrences while respecting per-line indentation
	var b strings.Builder
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		idx := strings.Index(line, "{yaml}")
		if idx == -1 {
			// No placeholder on this line; write as-is
			b.WriteString(line)
		} else {
			// Capture indentation (leading spaces/tabs) up to the placeholder start
			indent := leadingWhitespace(line)

			// Construct indented fenced block
			b.WriteString(indent)
			b.WriteString("```yaml\n")

			// Indent each YAML content line
			contentLines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
			for j, yLine := range contentLines {
				b.WriteString(indent)
				b.WriteString(yLine)
				if j < len(contentLines)-1 {
					b.WriteByte('\n')
				}
			}
			b.WriteByte('\n')

			// Close fence
			b.WriteString(indent)
			b.WriteString("```")

			// If there is any suffix after {yaml} on the same line, keep it
			suffix := line[idx+len("{yaml}"):]
			if len(suffix) > 0 {
				b.WriteString(suffix)
			}
		}

		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// leadingWhitespace returns the run of spaces/tabs from line start.
func leadingWhitespace(s string) string {
	for i, r := range s {
		if r != ' ' && r != '\t' {
			return s[:i]
		}
	}
	return s
}
