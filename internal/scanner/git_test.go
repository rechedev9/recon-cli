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

func initTestRepo(t *testing.T) string {
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
	require.NoError(t, os.WriteFile(filepath.Join(root, "file.go"), []byte("package main"), 0o644))
	run("add", "file.go")
	run("commit", "-m", "initial commit")

	return root
}

func TestGitScan(t *testing.T) {
	root := initTestRepo(t)

	gs := &gitScanner{}
	report, err := gs.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.Equal(t, "main", report.Branch)
	assert.Equal(t, 0, report.DirtyFiles)
	require.Len(t, report.LastCommits, 1)
	assert.Equal(t, "initial commit", report.LastCommits[0].Subject)
}

func TestGitScanDirtyFiles(t *testing.T) {
	root := initTestRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(root, "new.go"), []byte("package main"), 0o644))

	gs := &gitScanner{}
	report, err := gs.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.Equal(t, 1, report.DirtyFiles)
}

func TestGitScanNotARepo(t *testing.T) {
	root := t.TempDir()

	gs := &gitScanner{}
	report, err := gs.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.Nil(t, report)
}
