package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
	".next":        true,
	"dist":         true,
	"build":        true,
}

type treeScanner struct{}

func (s *treeScanner) Scan(ctx context.Context, root string, depth int) (*TreeReport, error) {
	report := &TreeReport{}
	entries, err := s.walk(ctx, root, root, 0, depth, report)
	if err != nil {
		return nil, err
	}
	report.Entries = entries
	return report, nil
}

func (s *treeScanner) walk(ctx context.Context, base, dir string, currentDepth, maxDepth int, report *TreeReport) ([]TreeNode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if currentDepth >= maxDepth {
		return nil, nil
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var nodes []TreeNode
	for _, entry := range dirEntries {
		name := entry.Name()

		if entry.IsDir() && shouldSkipDir(name) {
			continue
		}

		rel, _ := filepath.Rel(base, filepath.Join(dir, name))

		if entry.IsDir() {
			report.TotalDirs++
			node := TreeNode{Path: rel, IsDir: true}

			children, err := s.walk(ctx, base, filepath.Join(dir, name), currentDepth+1, maxDepth, report)
			if err != nil {
				return nil, err
			}
			node.Children = children
			node.Files = countDirectFiles(filepath.Join(dir, name))
			nodes = append(nodes, node)
		} else {
			report.TotalFiles++
			nodes = append(nodes, TreeNode{Path: rel, IsDir: false})
		}
	}

	return nodes, nil
}

func countDirectFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && !shouldSkip(e.Name()) {
			count++
		}
	}
	return count
}

func shouldSkip(name string) bool {
	return shouldSkipDir(name)
}

func shouldSkipDir(name string) bool {
	if skipDirs[name] {
		return true
	}
	if strings.HasPrefix(name, ".") && name != "." && name != ".." {
		return true
	}
	return false
}

func isHiddenFile(name string) bool {
	return strings.HasPrefix(name, ".") && name != "." && name != ".."
}
