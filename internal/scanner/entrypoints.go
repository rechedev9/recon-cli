package scanner

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type entryPointScanner struct{}

func (s *entryPointScanner) Scan(ctx context.Context, root string) ([]EntryPoint, error) {
	var points []EntryPoint

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() && shouldSkip(d.Name()) && path != root {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		name := d.Name()

		switch {
		case name == "main.go" && isGoMain(path):
			points = append(points, EntryPoint{Type: "main", Path: rel})
		case name == "main.rs" && filepath.Dir(rel) == "src":
			points = append(points, EntryPoint{Type: "main", Path: rel})
		case name == "__main__.py":
			points = append(points, EntryPoint{Type: "main", Path: rel})
		case name == "Makefile":
			points = append(points, EntryPoint{Type: "makefile", Path: rel})
		case name == "Dockerfile" || name == "docker-compose.yml" || name == "docker-compose.yaml":
			points = append(points, EntryPoint{Type: "dockerfile", Path: rel})
		case name == "package.json" && hasScriptsOrMain(path):
			points = append(points, EntryPoint{Type: "package.json", Path: rel})
		case name == "setup.py":
			points = append(points, EntryPoint{Type: "pyproject", Path: rel})
		case name == "pyproject.toml":
			points = append(points, EntryPoint{Type: "pyproject", Path: rel})
		case filepath.Dir(rel) == "scripts" && isExecutable(path):
			points = append(points, EntryPoint{Type: "script", Path: rel})
		}

		return nil
	})

	return points, err
}

func isGoMain(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	hasPackageMain := false
	hasFuncMain := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "package main" {
			hasPackageMain = true
		}
		if strings.HasPrefix(line, "func main()") {
			hasFuncMain = true
		}
	}
	return hasPackageMain && hasFuncMain
}

func hasScriptsOrMain(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var pkg map[string]json.RawMessage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	_, hasScripts := pkg["scripts"]
	_, hasMain := pkg["main"]
	_, hasBin := pkg["bin"]
	return hasScripts || hasMain || hasBin
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0o111 != 0
}
