package depgraph

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GoGraphBuilder struct{}

func (b *GoGraphBuilder) Build(root string) (*GraphReport, error) {
	modPath := filepath.Join(root, "go.mod")
	if _, err := os.Stat(modPath); err != nil {
		return nil, nil
	}

	modName, err := readModuleName(modPath)
	if err != nil {
		return nil, fmt.Errorf("read module name: %w", err)
	}

	goBin := findGoBinary()
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, goBin, "mod", "graph")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go mod graph: %w", err)
	}

	nodes := parseGoModGraphOutput(string(out), modName)

	return &GraphReport{
		Manager: "go.mod",
		Root:    modName,
		Nodes:   nodes,
	}, nil
}

func readModuleName(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}
	return "", fmt.Errorf("no module directive found")
}

func findGoBinary() string {
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}
	// Common fallback locations
	for _, p := range []string{"/usr/local/go/bin/go", "/usr/lib/go/bin/go"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "go" // let exec fail with a clear error
}

type modDep struct {
	name    string
	version string
}

func parseGoModGraphOutput(output string, rootModule string) []DepNode {
	children := make(map[string][]modDep)

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		parent := parts[0]
		child := parts[1]

		parentName := parent
		if idx := strings.LastIndex(parent, "@"); idx > 0 {
			parentName = parent[:idx]
		}

		childName, childVersion := splitAtVersion(child)

		children[parentName] = append(children[parentName], modDep{name: childName, version: childVersion})
	}

	return buildTree(children, rootModule, make(map[string]bool))
}

func splitAtVersion(s string) (string, string) {
	if idx := strings.LastIndex(s, "@"); idx > 0 {
		return s[:idx], s[idx+1:]
	}
	return s, ""
}

func buildTree(children map[string][]modDep, name string, visited map[string]bool) []DepNode {
	deps, ok := children[name]
	if !ok {
		return nil
	}

	var nodes []DepNode
	for _, d := range deps {
		node := DepNode{Name: d.name, Version: d.version}
		if !visited[d.name] {
			visited[d.name] = true
			node.Children = buildTree(children, d.name, visited)
		}
		nodes = append(nodes, node)
	}
	return nodes
}
