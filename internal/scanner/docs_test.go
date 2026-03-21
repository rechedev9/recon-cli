package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsScan(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("# App"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("# Rules"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "docs", "guide.md"), []byte("# Guide"), 0o644))

	ds := &docsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.True(t, report.HasReadme)
	assert.True(t, report.HasClaudeMD)
	assert.True(t, report.DocsDir)
	assert.Contains(t, report.Files, "README.md")
	assert.Contains(t, report.Files, "CLAUDE.md")
	assert.Contains(t, report.Files, "docs/guide.md")
}

func TestDocsScanCaseInsensitiveReadme(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "readme.md"), []byte("# App"), 0o644))

	ds := &docsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.True(t, report.HasReadme)
}

func TestDocsScanEmpty(t *testing.T) {
	root := t.TempDir()

	ds := &docsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.False(t, report.HasReadme)
	assert.False(t, report.HasClaudeMD)
	assert.False(t, report.DocsDir)
	assert.Empty(t, report.Files)
}
