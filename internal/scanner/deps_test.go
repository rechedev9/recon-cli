package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDepsScanGoMod(t *testing.T) {
	root := t.TempDir()

	gomod := "module github.com/example/app\n\ngo 1.24\n\nrequire (\n\tgithub.com/spf13/cobra v1.8.1\n\tgithub.com/stretchr/testify v1.9.0\n)\n\nrequire (\n\tgithub.com/inconshreveable/mousetrap v1.1.0 // indirect\n)\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte(gomod), 0o644))

	ds := &depsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	require.Len(t, report.Manifests, 1)
	m := report.Manifests[0]
	assert.Equal(t, "go.mod", m.Manager)
	assert.Len(t, m.Deps, 2)
	assert.Equal(t, "github.com/spf13/cobra", m.Deps[0].Name)
	assert.Equal(t, "v1.8.1", m.Deps[0].Version)
}

func TestDepsScanPackageJSON(t *testing.T) {
	root := t.TempDir()

	pkgjson := `{"name":"myapp","dependencies":{"react":"^18.2.0","next":"^14.0.0"},"devDependencies":{"typescript":"^5.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(root, "package.json"), []byte(pkgjson), 0o644))

	ds := &depsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	require.Len(t, report.Manifests, 1)
	m := report.Manifests[0]
	assert.Equal(t, "package.json", m.Manager)
	assert.Len(t, m.Deps, 3)
}

func TestDepsScanRequirementsTxt(t *testing.T) {
	root := t.TempDir()

	reqs := "flask==2.3.0\nrequests==2.31.0\n# comment\n\n-e git+https://...\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "requirements.txt"), []byte(reqs), 0o644))

	ds := &depsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	require.Len(t, report.Manifests, 1)
	m := report.Manifests[0]
	assert.Equal(t, "requirements.txt", m.Manager)
	assert.Len(t, m.Deps, 2)
	assert.Equal(t, "flask", m.Deps[0].Name)
	assert.Equal(t, "2.3.0", m.Deps[0].Version)
}

func TestDepsScanCargoToml(t *testing.T) {
	root := t.TempDir()

	cargo := `[package]
name = "myapp"
version = "0.1.0"

[dependencies]
serde = "1.0"
tokio = { version = "1.0", features = ["full"] }

[dev-dependencies]
criterion = "0.5"
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "Cargo.toml"), []byte(cargo), 0o644))

	ds := &depsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	require.Len(t, report.Manifests, 1)
	m := report.Manifests[0]
	assert.Equal(t, "Cargo.toml", m.Manager)
	assert.Len(t, m.Deps, 3) // serde, tokio, criterion
}

func TestDepsScanPyprojectToml(t *testing.T) {
	root := t.TempDir()

	pyproj := `[project]
name = "myapp"
version = "1.0.0"
dependencies = [
    "flask>=2.3.0",
    "requests==2.31.0",
    "numpy",
    "pandas~=2.0",
]
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte(pyproj), 0o644))

	ds := &depsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	require.Len(t, report.Manifests, 1)
	m := report.Manifests[0]
	assert.Equal(t, "pyproject.toml", m.Manager)
	assert.Len(t, m.Deps, 4)

	// Check specific parsing
	depMap := make(map[string]string)
	for _, d := range m.Deps {
		depMap[d.Name] = d.Version
	}
	assert.Equal(t, ">=2.3.0", depMap["flask"])
	assert.Equal(t, "==2.31.0", depMap["requests"])
	assert.Equal(t, "", depMap["numpy"]) // no version constraint
	assert.Equal(t, "~=2.0", depMap["pandas"])
}

func TestDepsScanMultipleManifests(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"),
		[]byte("module app\n\ngo 1.24\n\nrequire github.com/foo/bar v1.0.0\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "frontend"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "frontend", "package.json"),
		[]byte(`{"dependencies":{"react":"^18"}}`), 0o644))

	ds := &depsScanner{}
	report, err := ds.Scan(context.Background(), root)
	require.NoError(t, err)

	assert.Len(t, report.Manifests, 2)
}
