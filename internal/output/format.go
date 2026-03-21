package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rechedev9/CLIClaudeCode/internal/scanner"
)

func FormatJSON(report *scanner.Report) ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return data, nil
}

func FormatMarkdown(report *scanner.Report) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n", report.Path)

	if report.Tree != nil {
		fmt.Fprintf(&b, "\n## Structure (%d files, %d dirs)\n", report.Tree.TotalFiles, report.Tree.TotalDirs)
		for _, e := range report.Tree.Entries {
			if e.IsDir {
				fmt.Fprintf(&b, "%s/ (%d files)\n", e.Path, e.Files)
			}
		}
	}

	if report.Languages != nil {
		fmt.Fprintf(&b, "\n## Languages (%d LOC)\n", report.Languages.TotalLOC)
		for _, l := range report.Languages.Languages {
			fmt.Fprintf(&b, "%s: %d files, %d LOC\n", l.Name, l.Files, l.LOC)
		}
	}

	if len(report.EntryPoints) > 0 {
		fmt.Fprintf(&b, "\n## Entry Points\n")
		for _, ep := range report.EntryPoints {
			fmt.Fprintf(&b, "%s: %s\n", ep.Type, ep.Path)
		}
	}

	if report.Git != nil {
		fmt.Fprintf(&b, "\n## Git (%s, %d dirty)\n", report.Git.Branch, report.Git.DirtyFiles)
		for _, c := range report.Git.LastCommits {
			fmt.Fprintf(&b, "%s %s (%s)\n", c.Hash, c.Subject, c.Date)
		}
	}

	if report.Dependencies != nil && len(report.Dependencies.Manifests) > 0 {
		fmt.Fprintf(&b, "\n## Dependencies\n")
		for _, m := range report.Dependencies.Manifests {
			for _, d := range m.Deps {
				fmt.Fprintf(&b, "%s: %s %s\n", m.Manager, d.Name, d.Version)
			}
		}
	}

	if report.Docs != nil {
		fmt.Fprintf(&b, "\n## Docs\n")
		fmt.Fprintf(&b, "%s\n", strings.Join(report.Docs.Files, ", "))
	}

	return b.String()
}
