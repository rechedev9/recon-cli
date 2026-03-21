package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzParseGoMod(f *testing.F) {
	f.Add([]byte("module example\n\ngo 1.24\n\nrequire (\n\tgithub.com/foo/bar v1.0.0\n)\n"))
	f.Add([]byte("module example\n\nrequire github.com/foo v1.0.0\n"))
	f.Add([]byte(""))
	f.Add([]byte("not a go.mod"))
	f.Add([]byte("require (\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "go.mod")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		parseGoMod(path) // must not panic
	})
}

func FuzzParsePackageJSON(f *testing.F) {
	f.Add([]byte(`{"dependencies":{"react":"^18"}}`))
	f.Add([]byte(`{"devDependencies":{"jest":"^29"}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "package.json")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		parsePackageJSON(path) // must not panic
	})
}

func FuzzParseRequirementsTxt(f *testing.F) {
	f.Add([]byte("flask==2.3.0\nrequests==2.31.0\n"))
	f.Add([]byte("# comment\n\n-e git+https://foo\n"))
	f.Add([]byte(""))
	f.Add([]byte("broken===line"))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "requirements.txt")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		parseRequirementsTxt(path) // must not panic
	})
}

func FuzzParseCargoToml(f *testing.F) {
	f.Add([]byte(`[dependencies]
serde = "1.0"
tokio = { version = "1.0", features = ["full"] }
`))
	f.Add([]byte(`[dev-dependencies]
criterion = "0.5"
`))
	f.Add([]byte(``))
	f.Add([]byte(`not toml`))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "Cargo.toml")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		parseCargoToml(path) // must not panic
	})
}

func FuzzParsePyprojectToml(f *testing.F) {
	f.Add([]byte(`[project]
dependencies = ["flask>=2.3", "requests==2.31"]
`))
	f.Add([]byte(``))
	f.Add([]byte(`not toml`))
	f.Add([]byte(`[project]
dependencies = []
`))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "pyproject.toml")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		parsePyprojectToml(path) // must not panic
	})
}

func FuzzParsePEP508(f *testing.F) {
	f.Add("flask>=2.3.0")
	f.Add("requests==2.31.0")
	f.Add("numpy")
	f.Add("pandas~=2.0")
	f.Add("foo!=1.0")
	f.Add("bar<3,>=2")
	f.Add("")

	f.Fuzz(func(t *testing.T, spec string) {
		parsePEP508(spec) // must not panic
	})
}
