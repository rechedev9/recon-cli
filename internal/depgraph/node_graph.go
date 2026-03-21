package depgraph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type NodeGraphBuilder struct{}

type packageLock struct {
	Name     string                    `json:"name"`
	Packages map[string]packageLockPkg `json:"packages"`
}

type packageLockPkg struct {
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

func (b *NodeGraphBuilder) Build(root string) (*GraphReport, error) {
	lockPath := filepath.Join(root, "package-lock.json")
	if _, err := os.Stat(lockPath); err != nil {
		return nil, nil
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("read package-lock.json: %w", err)
	}

	var lock packageLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parse package-lock.json: %w", err)
	}

	rootPkg, ok := lock.Packages[""]
	if !ok {
		return &GraphReport{Manager: "package-lock.json", Root: lock.Name}, nil
	}

	nodes := buildNodeTree(lock.Packages, rootPkg.Dependencies)

	return &GraphReport{
		Manager: "package-lock.json",
		Root:    lock.Name,
		Nodes:   nodes,
	}, nil
}

func buildNodeTree(packages map[string]packageLockPkg, deps map[string]string) []DepNode {
	var nodes []DepNode
	names := make([]string, 0, len(deps))
	for name := range deps {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		key := "node_modules/" + name
		pkg, ok := packages[key]
		if !ok {
			continue
		}
		node := DepNode{
			Name:    name,
			Version: pkg.Version,
		}
		if len(pkg.Dependencies) > 0 {
			node.Children = buildNodeTree(packages, pkg.Dependencies)
		}
		nodes = append(nodes, node)
	}
	return nodes
}
