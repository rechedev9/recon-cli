package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

type docsScanner struct{}

func (s *docsScanner) Scan(_ context.Context, root string) (*DocsReport, error) {
	report := &DocsReport{}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)

		if lower == "readme.md" {
			report.HasReadme = true
			report.Files = append(report.Files, name)
		} else if name == "CLAUDE.md" {
			report.HasClaudeMD = true
			report.Files = append(report.Files, name)
		} else if name == "AGENTS.md" {
			report.Files = append(report.Files, name)
		} else if strings.HasSuffix(lower, ".md") {
			report.Files = append(report.Files, name)
		}
	}

	docsDir := filepath.Join(root, "docs")
	if info, err := os.Stat(docsDir); err == nil && info.IsDir() {
		report.DocsDir = true
		docsEntries, err := os.ReadDir(docsDir)
		if err == nil {
			for _, e := range docsEntries {
				if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
					report.Files = append(report.Files, filepath.Join("docs", e.Name()))
				}
			}
		}
	}

	return report, nil
}
