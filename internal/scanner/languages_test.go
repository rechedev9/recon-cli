package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLangScan(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "util.go"), []byte("package main\n\nfunc util() {}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("# Hello\n"), 0o644))

	ls := &langScanner{}
	report, err := ls.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.Equal(t, 3, report.TotalFiles)
	assert.Equal(t, 5, report.TotalLOC) // 2 + 2 + 1 (blank lines excluded)

	require.Len(t, report.Languages, 2)
	assert.Equal(t, "Go", report.Languages[0].Name)
	assert.Equal(t, 2, report.Languages[0].Files)
	assert.Equal(t, 4, report.Languages[0].LOC)
	assert.Equal(t, "Markdown", report.Languages[1].Name)
}

func TestLangScanSkipsDirs(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "node_modules"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "node_modules", "pkg.js"), []byte("var x = 1;\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "app.js"), []byte("var y = 2;\n"), 0o644))

	ls := &langScanner{}
	report, err := ls.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.Equal(t, 1, report.TotalFiles)
}

func TestLangScanEmptyLines(t *testing.T) {
	root := t.TempDir()

	content := "package main\n\n\n   \nfunc main() {}\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "main.go"), []byte(content), 0o644))

	ls := &langScanner{}
	report, err := ls.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.Equal(t, 2, report.TotalLOC)
}
