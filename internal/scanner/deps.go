package scanner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

var knownManifests = map[string]bool{
	"go.mod":           true,
	"package.json":     true,
	"requirements.txt": true,
	"Cargo.toml":       true,
	"pyproject.toml":   true,
}

type depsScanner struct{}

func (s *depsScanner) Scan(ctx context.Context, root string) (*DepsReport, error) {
	report := &DepsReport{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() && shouldSkipDir(d.Name()) && path != root {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if !knownManifests[name] {
			return nil
		}

		rel, _ := filepath.Rel(root, path)

		var deps []DepEntry
		var parseErr error

		switch name {
		case "go.mod":
			deps, parseErr = parseGoMod(path)
		case "package.json":
			deps, parseErr = parsePackageJSON(path)
		case "requirements.txt":
			deps, parseErr = parseRequirementsTxt(path)
		case "Cargo.toml":
			deps, parseErr = parseCargoToml(path)
		case "pyproject.toml":
			deps, parseErr = parsePyprojectToml(path)
		}

		if parseErr != nil {
			return nil
		}

		report.Manifests = append(report.Manifests, ManifestReport{
			Manager: name,
			Path:    rel,
			Deps:    deps,
		})
		return nil
	})

	return report, err
}

func parseGoMod(path string) ([]DepEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open go.mod: %w", err)
	}
	defer f.Close()

	var deps []DepEntry
	inRequire := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}

		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				deps = append(deps, DepEntry{Name: parts[1], Version: parts[2]})
			}
			continue
		}

		if inRequire {
			if strings.Contains(line, "// indirect") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deps = append(deps, DepEntry{Name: parts[0], Version: parts[1]})
			}
		}
	}

	return deps, sc.Err()
}

func parsePackageJSON(path string) ([]DepEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read package.json: %w", err)
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	var deps []DepEntry
	for name, ver := range pkg.Dependencies {
		deps = append(deps, DepEntry{Name: name, Version: ver})
	}
	for name, ver := range pkg.DevDependencies {
		deps = append(deps, DepEntry{Name: name, Version: ver})
	}
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})
	return deps, nil
}

func parseRequirementsTxt(path string) ([]DepEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open requirements.txt: %w", err)
	}
	defer f.Close()

	var deps []DepEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		if parts := strings.SplitN(line, "==", 2); len(parts) == 2 {
			deps = append(deps, DepEntry{Name: parts[0], Version: parts[1]})
		}
	}
	return deps, sc.Err()
}

func parseCargoToml(path string) ([]DepEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Cargo.toml: %w", err)
	}

	var cargo struct {
		Dependencies    map[string]any `toml:"dependencies"`
		DevDependencies map[string]any `toml:"dev-dependencies"`
	}
	if err := toml.Unmarshal(data, &cargo); err != nil {
		return nil, fmt.Errorf("parse Cargo.toml: %w", err)
	}

	var deps []DepEntry
	for name, val := range cargo.Dependencies {
		deps = append(deps, DepEntry{Name: name, Version: cargoVersion(val)})
	}
	for name, val := range cargo.DevDependencies {
		deps = append(deps, DepEntry{Name: name, Version: cargoVersion(val)})
	}
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})
	return deps, nil
}

func cargoVersion(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case map[string]any:
		if ver, ok := v["version"]; ok {
			if s, ok := ver.(string); ok {
				return s
			}
		}
	}
	return ""
}

func parsePyprojectToml(path string) ([]DepEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read pyproject.toml: %w", err)
	}

	var pyproj struct {
		Project struct {
			Dependencies []string `toml:"dependencies"`
		} `toml:"project"`
	}
	if err := toml.Unmarshal(data, &pyproj); err != nil {
		return nil, fmt.Errorf("parse pyproject.toml: %w", err)
	}

	var deps []DepEntry
	for _, dep := range pyproj.Project.Dependencies {
		name, version := parsePEP508(dep)
		deps = append(deps, DepEntry{Name: name, Version: version})
	}
	return deps, nil
}

func parsePEP508(spec string) (string, string) {
	// PEP 508: "name>=version", "name==version", "name~=version", "name<version", "name!=version", "name"
	spec = strings.TrimSpace(spec)
	// Try splitting on version operators in order of length
	for _, op := range []string{"~=", ">=", "<=", "!=", "==", ">", "<"} {
		if idx := strings.Index(spec, op); idx > 0 {
			return strings.TrimSpace(spec[:idx]), op + strings.TrimSpace(spec[idx+len(op):])
		}
	}
	return spec, "" // no version constraint
}
