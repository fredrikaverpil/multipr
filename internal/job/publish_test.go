package job

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fredrikaverpil/multipr/internal/command"
	"github.com/fredrikaverpil/multipr/internal/log"
)

func newManagerForTest(t *testing.T, jobContent string) *Manager {
	t.Helper()

	dir := t.TempDir()
	jobPath := filepath.Join(dir, "job.yaml")
	if err := os.WriteFile(jobPath, []byte(jobContent), 0o644); err != nil {
		t.Fatal(err)
	}

	logger, err := log.NewLogger(log.Options{LevelDebug: false})
	if err != nil {
		t.Fatal(err)
	}

	exec := command.NewExecutor(false, "sh", logger)

	// Minimal manager with only the fields needed for processing
	m := &Manager{
		exec:        exec,
		log:         logger,
		jobFilePath: jobPath,
	}
	return m
}

func TestProcessBodyTemplate_NoPlaceholder(t *testing.T) {
	m := newManagerForTest(t, "a: 1\n")
	body := "no placeholders here"
	got := m.processBodyTemplate(body)
	if got != body {
		t.Fatalf("expected unchanged body, got:\n%s", got)
	}
}

func TestProcessBodyTemplate_NoIndent(t *testing.T) {
	m := newManagerForTest(t, "a: 1\nb: 2\n")
	body := "Header\n{yaml}\nFooter"
	got := m.processBodyTemplate(body)

	want := "Header\n```yaml\na: 1\nb: 2\n```\nFooter"
	if got != want {
		t.Fatalf("mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestProcessBodyTemplate_SpacesIndent(t *testing.T) {
	m := newManagerForTest(t, "a: 1\nb: 2\n")
	body := "Header\n    {yaml}\nFooter"
	got := m.processBodyTemplate(body)

	want := strings.Join([]string{
		"Header",
		"    ```yaml",
		"    a: 1",
		"    b: 2",
		"    ```",
		"Footer",
	}, "\n")

	if got != want {
		t.Fatalf("mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestProcessBodyTemplate_TabsIndent(t *testing.T) {
	m := newManagerForTest(t, "a: 1\n\nb: 2\n")
	body := "Header\n\t{yaml}\nFooter"
	got := m.processBodyTemplate(body)

	want := strings.Join([]string{
		"Header",
		"\t```yaml",
		"\ta: 1",
		"\t",
		"\tb: 2",
		"\t```",
		"Footer",
	}, "\n")

	if got != want {
		t.Fatalf("mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestProcessBodyTemplate_SuffixAfterPlaceholder(t *testing.T) {
	m := newManagerForTest(t, "x: y\n")
	body := "Intro\n  {yaml} trailing\nOutro"
	got := m.processBodyTemplate(body)

	want := strings.Join([]string{
		"Intro",
		"  ```yaml",
		"  x: y",
		"  ``` trailing",
		"Outro",
	}, "\n")

	if got != want {
		t.Fatalf("mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestProcessBodyTemplate_MultiplePlaceholders(t *testing.T) {
	m := newManagerForTest(t, "a: 1\n")
	body := "{yaml}\n  {yaml}"
	got := m.processBodyTemplate(body)

	want := strings.Join([]string{
		"```yaml",
		"a: 1",
		"```",
		"  ```yaml",
		"  a: 1",
		"  ```",
	}, "\n")

	if got != want {
		t.Fatalf("mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestProcessBodyTemplate_NoTrailingNewlineInYAML(t *testing.T) {
	m := newManagerForTest(t, "k: v") // no trailing newline
	body := "{yaml}"
	got := m.processBodyTemplate(body)

	want := "```yaml\nk: v\n```"
	if got != want {
		t.Fatalf("mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
