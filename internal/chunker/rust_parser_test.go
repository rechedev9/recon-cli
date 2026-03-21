package chunker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRustParser(t *testing.T) {
	root := t.TempDir()
	src := `use std::io;

pub const MAX_RETRIES: u32 = 3;

pub struct Config {
    pub host: String,
    pub port: u16,
}

pub trait Runner {
    fn run(&self) -> Result<(), Error>;
}

impl Config {
    pub fn new(host: String, port: u16) -> Self {
        Config { host, port }
    }
}

pub fn fetch(url: &str) -> Result<String, Error> {
    Ok(String::new())
}

fn private_helper() {}

pub enum Status {
    Active,
    Inactive,
}
`
	path := filepath.Join(root, "lib.rs")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	p := &RustParser{}
	symbols, err := p.Parse(path)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, s := range symbols {
		names[s.Name] = true
	}

	assert.True(t, names["MAX_RETRIES"])
	assert.True(t, names["Config"])
	assert.True(t, names["Runner"])
	assert.True(t, names["fetch"])
	assert.True(t, names["private_helper"])
	assert.True(t, names["Status"])
}
