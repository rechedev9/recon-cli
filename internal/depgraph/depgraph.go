package depgraph

import (
	"fmt"
	"os"
	"path/filepath"
)

func BuildGraph(root string) (*GraphReport, error) {
	builders := []struct {
		check   string
		builder GraphBuilder
	}{
		{"go.mod", &GoGraphBuilder{}},
		{"package-lock.json", &NodeGraphBuilder{}},
		{"requirements.txt", &PythonGraphBuilder{}},
	}

	for _, b := range builders {
		if _, err := os.Stat(filepath.Join(root, b.check)); err == nil {
			report, err := b.builder.Build(root)
			if err != nil {
				return nil, fmt.Errorf("build %s graph: %w", b.check, err)
			}
			if report != nil {
				return report, nil
			}
		}
	}

	return nil, fmt.Errorf("no supported dependency manifest found in %s", root)
}
