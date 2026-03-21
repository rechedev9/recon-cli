package output

import (
	"testing"

	"github.com/rechedev9/CLIClaudeCode/internal/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatJSON(t *testing.T) {
	report := &scanner.Report{
		Path: "/tmp/test",
		Tree: &scanner.TreeReport{TotalFiles: 5, TotalDirs: 2},
	}

	data, err := FormatJSON(report)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"total_files": 5`)
	assert.Contains(t, string(data), `"path": "/tmp/test"`)
}

func TestFormatJSONOmitsNil(t *testing.T) {
	report := &scanner.Report{Path: "/tmp/test"}

	data, err := FormatJSON(report)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"tree"`)
	assert.NotContains(t, string(data), `"git"`)
}

func TestFormatMarkdown(t *testing.T) {
	report := &scanner.Report{
		Path: "/tmp/test",
		Languages: &scanner.LangReport{
			TotalFiles: 10,
			TotalLOC:   500,
			Languages: []scanner.LangEntry{
				{Name: "Go", Files: 8, LOC: 450},
				{Name: "Markdown", Files: 2, LOC: 50},
			},
		},
		Git: &scanner.GitReport{
			Branch:     "main",
			DirtyFiles: 2,
			LastCommits: []scanner.CommitInfo{
				{Hash: "abc1234", Subject: "feat: init", Date: "2026-03-21"},
			},
		},
	}

	md := FormatMarkdown(report)
	assert.Contains(t, md, "# /tmp/test")
	assert.Contains(t, md, "Go: 8 files, 450 LOC")
	assert.Contains(t, md, "## Git (main, 2 dirty)")
	assert.Contains(t, md, "abc1234 feat: init (2026-03-21)")
}
