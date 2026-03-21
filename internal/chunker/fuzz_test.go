package chunker

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzGoParser(f *testing.F) {
	f.Add([]byte("package main\n\nfunc main() {}\n"))
	f.Add([]byte("package foo\n\ntype Config struct {\n\tHost string\n}\n"))
	f.Add([]byte(""))
	f.Add([]byte("not go code"))
	f.Add([]byte("package main\n\nfunc ("))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "fuzz.go")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		p := &GoParser{}
		p.Parse(path) // must not panic
	})
}

func FuzzTSParser(f *testing.F) {
	f.Add([]byte("export function foo() {}\n"))
	f.Add([]byte("export default class Bar {}\n"))
	f.Add([]byte(""))
	f.Add([]byte("const x = () => {}"))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "fuzz.ts")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		p := &TSParser{}
		p.Parse(path) // must not panic
	})
}

func FuzzPythonParser(f *testing.F) {
	f.Add([]byte("class Foo:\n    pass\n"))
	f.Add([]byte("def bar():\n    pass\n"))
	f.Add([]byte("MAX_RETRIES = 3\n"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "fuzz.py")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		p := &PythonParser{}
		p.Parse(path) // must not panic
	})
}

func FuzzRustParser(f *testing.F) {
	f.Add([]byte("pub fn foo() {}\n"))
	f.Add([]byte("pub struct Bar {\n    pub x: i32,\n}\n"))
	f.Add([]byte(""))
	f.Add([]byte("impl Foo for Bar {}"))

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "fuzz.rs")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Skip()
		}
		p := &RustParser{}
		p.Parse(path) // must not panic
	})
}
