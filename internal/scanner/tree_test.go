package scanner

import (
	"context"
	"os"
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
