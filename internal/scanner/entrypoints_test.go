package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryPointScan(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "cmd", "app"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "cmd", "app", "main.go"),
		[]byte("package main\n\nfunc main() {}\n"), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(root, "Makefile"), []byte("build:\n\tgo build"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "Dockerfile"), []byte("FROM golang"), 0o644))

	require.NoError(t, os.MkdirAll(filepath.Join(root, "scripts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "scripts", "run.sh"), []byte("#!/bin/bash"), 0o755))

	eps := &entryPointScanner{}
	points, err := eps.Scan(context.Background(), root)
	require.NoError(t, err)

	types := make(map[string]bool)
	for _, ep := range points {
		types[ep.Type] = true
	}
	assert.True(t, types["main"])
	assert.True(t, types["makefile"])
	assert.True(t, types["dockerfile"])
	assert.True(t, types["script"])
}

func TestEntryPointPackageJSON(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "package.json"),
		[]byte(`{"name":"app","scripts":{"start":"node index.js"}}`), 0o644))

	eps := &entryPointScanner{}
	points, err := eps.Scan(context.Background(), root)
	require.NoError(t, err)

	require.Len(t, points, 1)
	assert.Equal(t, "package.json", points[0].Type)
}
