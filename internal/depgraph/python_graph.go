package depgraph

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PythonGraphBuilder struct{}

var pyDepRe = regexp.MustCompile(`^([a-zA-Z0-9_-]+)\s*([><=!~]+.+)?$`)

func (b *PythonGraphBuilder) Build(root string) (*GraphReport, error) {
	reqPath := filepath.Join(root, "requirements.txt")
	if _, err := os.Stat(reqPath); err != nil {
		return nil, nil
	}

	f, err := os.Open(reqPath)
	if err != nil {
		return nil, fmt.Errorf("open requirements.txt: %w", err)
	}
	defer f.Close()

	var nodes []DepNode
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		matches := pyDepRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		nodes = append(nodes, DepNode{
			Name:    matches[1],
			Version: strings.TrimSpace(matches[2]),
		})
	}

	return &GraphReport{
		Manager: "requirements.txt",
		Root:    filepath.Base(root),
		Nodes:   nodes,
	}, sc.Err()
}
