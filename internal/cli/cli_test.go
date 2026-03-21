package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rechedev9/CLIClaudeCode/internal/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFixtureProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("/usr/bin/git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}

	run("init")
	run("checkout", "-b", "main")

	require.NoError(t, os.MkdirAll(filepath.Join(root, "cmd", "app"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "cmd", "app", "main.go"),
		[]byte("package main\n\nfunc main() {}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"),
		[]byte("module example.com/app\n\ngo 1.24\n\nrequire github.com/foo/bar v1.0.0\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("# App\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "docs", "guide.md"), []byte("# Guide\n"), 0o644))

	run("add", ".")
	run("commit", "-m", "initial commit")

	return root
}

func TestIntegrationJSON(t *testing.T) {
	root := setupFixtureProject(t)

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{root}, &stdout, &stderr)
	require.NoError(t, err)

	var report scanner.Report
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &report))

	assert.Equal(t, root, report.Path)
	assert.NotNil(t, report.Tree)
	assert.NotNil(t, report.Languages)
	assert.NotNil(t, report.Git)
	assert.Equal(t, "main", report.Git.Branch)
	assert.NotNil(t, report.Dependencies)
	assert.NotNil(t, report.Docs)
	assert.True(t, report.Docs.HasReadme)

	found := false
	for _, ep := range report.EntryPoints {
		if ep.Type == "main" && ep.Path == filepath.Join("cmd", "app", "main.go") {
			found = true
		}
	}
	assert.True(t, found, "expected main entry point")
}

func TestIntegrationMarkdown(t *testing.T) {
	root := setupFixtureProject(t)

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{root, "--format", "md"}, &stdout, &stderr)
	require.NoError(t, err)

	md := stdout.String()
	assert.Contains(t, md, "# "+root)
	assert.Contains(t, md, "## Structure")
	assert.Contains(t, md, "## Languages")
	assert.Contains(t, md, "## Git")
}

func TestIntegrationSkipFlags(t *testing.T) {
	root := setupFixtureProject(t)

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{root, "--no-git", "--no-deps", "--no-docs"}, &stdout, &stderr)
	require.NoError(t, err)

	var report scanner.Report
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &report))

	assert.Nil(t, report.Git)
	assert.Nil(t, report.Dependencies)
	assert.Nil(t, report.Docs)
	assert.NotNil(t, report.Tree)
}
