package scanner

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
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

// IgnoreChecker determines which files/dirs should be excluded from scans.
type IgnoreChecker interface {
	ShouldIgnore(relPath string) bool
	ShouldIgnoreDir(name string) bool
}

// gitIgnoreChecker uses `git ls-files` output to determine ignored files.
type gitIgnoreChecker struct {
	knownFiles map[string]bool
	knownDirs  map[string]bool
}

func (c *gitIgnoreChecker) ShouldIgnore(relPath string) bool {
	return !c.knownFiles[relPath]
}

func (c *gitIgnoreChecker) ShouldIgnoreDir(name string) bool {
	return shouldSkipDir(name)
}

// fallbackIgnoreChecker is used for non-git repos; relies on dir-level skipping only.
type fallbackIgnoreChecker struct{}

func (c *fallbackIgnoreChecker) ShouldIgnore(_ string) bool {
	return false
}

func (c *fallbackIgnoreChecker) ShouldIgnoreDir(name string) bool {
	return shouldSkipDir(name)
}

// NewIgnoreChecker creates the appropriate checker for the given root directory.
// If root is a git repo, it uses git ls-files; otherwise falls back to dir-skip only.
func NewIgnoreChecker(ctx context.Context, root string) IgnoreChecker {
	// Check if this is a git repo.
	revParse := exec.CommandContext(ctx, "/usr/bin/git", "rev-parse", "--git-dir")
	revParse.Dir = root
	if err := revParse.Run(); err != nil {
		return &fallbackIgnoreChecker{}
	}

	// List all non-ignored files (tracked + untracked that are not excluded).
	lsFiles := exec.CommandContext(ctx, "/usr/bin/git", "ls-files", "-co", "--exclude-standard")
	lsFiles.Dir = root
	out, err := lsFiles.Output()
	if err != nil {
		return &fallbackIgnoreChecker{}
	}

	knownFiles := make(map[string]bool)
	knownDirs := make(map[string]bool)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// Normalize to OS path separators.
		rel := filepath.FromSlash(line)
		knownFiles[rel] = true
		// Record all parent directories.
		dir := filepath.Dir(rel)
		for dir != "." && dir != "" {
			if knownDirs[dir] {
				break
			}
			knownDirs[dir] = true
			dir = filepath.Dir(dir)
		}
	}

	return &gitIgnoreChecker{
		knownFiles: knownFiles,
		knownDirs:  knownDirs,
	}
}

type treeScanner struct {
	ignore IgnoreChecker
}

func (s *treeScanner) Scan(ctx context.Context, root string, depth int) (*TreeReport, error) {
	ignore := s.ignore
	if ignore == nil {
		ignore = &fallbackIgnoreChecker{}
	}
	report := &TreeReport{}
	entries, err := s.walk(ctx, root, root, 0, depth, report, ignore)
	if err != nil {
		return nil, err
	}
	report.Entries = entries
	return report, nil
}

func (s *treeScanner) walk(ctx context.Context, base, dir string, currentDepth, maxDepth int, report *TreeReport, ignore IgnoreChecker) ([]TreeNode, error) {
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

		if entry.IsDir() && ignore.ShouldIgnoreDir(name) {
			continue
		}

		rel, _ := filepath.Rel(base, filepath.Join(dir, name))

		if entry.IsDir() {
			report.TotalDirs++
			node := TreeNode{Path: rel, IsDir: true}

			subDir := filepath.Join(dir, name)
			children, err := s.walk(ctx, base, subDir, currentDepth+1, maxDepth, report, ignore)
			if err != nil {
				return nil, err
			}
			node.Children = children
			node.Files = countFilesInNodes(children)
			nodes = append(nodes, node)
		} else {
			if ignore.ShouldIgnore(rel) {
				continue
			}
			report.TotalFiles++
			nodes = append(nodes, TreeNode{Path: rel, IsDir: false})
		}
	}

	return nodes, nil
}

func countFilesInNodes(nodes []TreeNode) int {
	count := 0
	for _, n := range nodes {
		if !n.IsDir {
			count++
		}
	}
	return count
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
