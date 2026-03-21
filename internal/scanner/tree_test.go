package scanner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeScan(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "foo", "bar"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "baz"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "top.go"), []byte("package main"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "foo", "a.go"), []byte("package foo"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "foo", "bar", "deep.go"), []byte("package bar"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "baz", "b.go"), []byte("package baz"), 0o644))

	ts := &treeScanner{}
	report, err := ts.Scan(context.Background(), root, 4)
	require.NoError(t, err)

	assert.Equal(t, 4, report.TotalFiles)
	assert.Equal(t, 3, report.TotalDirs)
}

func TestTreeScanSkipsDirs(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "node_modules", "pkg"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".git", "objects"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "src"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "node_modules", "pkg", "index.js"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".git", "objects", "abc"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main"), 0o644))

	ts := &treeScanner{}
	report, err := ts.Scan(context.Background(), root, 4)
	require.NoError(t, err)

	assert.Equal(t, 1, report.TotalFiles)
	assert.Equal(t, 1, report.TotalDirs)
}

func TestTreeScanDepthLimit(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "a", "b", "c", "d"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "a", "b", "c", "d", "e.go"), []byte(""), 0o644))

	ts := &treeScanner{}
	report, err := ts.Scan(context.Background(), root, 2)
	require.NoError(t, err)

	// At depth 2, we see a/ and a/b/ but not deeper
	assert.Equal(t, 2, report.TotalDirs)
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("/usr/bin/git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
}

func TestGitIgnoreChecker(t *testing.T) {
	root := t.TempDir()

	gitRun(t, root, "init")

	// Create .gitignore
	require.NoError(t, os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored/\n*.log\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "ignored"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "ignored", "secret.go"), []byte("package secret"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "debug.log"), []byte("log"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0o644))

	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-m", "init")

	checker := NewIgnoreChecker(context.Background(), root)

	// main.go should NOT be ignored
	assert.False(t, checker.ShouldIgnore("main.go"))
	// .gitignore itself should NOT be ignored
	assert.False(t, checker.ShouldIgnore(".gitignore"))
	// ignored/secret.go should be ignored
	assert.True(t, checker.ShouldIgnore(filepath.Join("ignored", "secret.go")))
	// debug.log should be ignored
	assert.True(t, checker.ShouldIgnore("debug.log"))
}

func TestFallbackIgnoreChecker(t *testing.T) {
	root := t.TempDir() // not a git repo

	checker := NewIgnoreChecker(context.Background(), root)

	// Fallback never ignores files
	assert.False(t, checker.ShouldIgnore("anything.go"))
	// But still skips known dirs
	assert.True(t, checker.ShouldIgnoreDir("node_modules"))
	assert.True(t, checker.ShouldIgnoreDir(".git"))
	assert.False(t, checker.ShouldIgnoreDir("src"))
}

func TestTreeScanRespectsGitignore(t *testing.T) {
	root := t.TempDir()

	gitRun(t, root, "init")
	require.NoError(t, os.WriteFile(filepath.Join(root, ".gitignore"), []byte("build/\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "build"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "build", "output.js"), []byte(""), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "src"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main"), 0o644))
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-m", "init")

	checker := NewIgnoreChecker(context.Background(), root)
	ts := &treeScanner{ignore: checker}
	report, err := ts.Scan(context.Background(), root, 4)
	require.NoError(t, err)

	// Should only see src/ and its files, not build/
	assert.Equal(t, 1, report.TotalDirs) // just src/
	assert.Equal(t, 2, report.TotalFiles) // .gitignore + src/main.go
}
