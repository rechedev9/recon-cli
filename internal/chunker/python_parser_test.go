package chunker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPythonParser(t *testing.T) {
	root := t.TempDir()
	src := `import os

MAX_RETRIES = 3

class UserService:
    def __init__(self, db):
        self.db = db

    def get_user(self, user_id: str) -> dict:
        pass

def fetch_data(url: str) -> dict:
    pass

async def async_fetch(url: str) -> dict:
    pass
`
	path := filepath.Join(root, "service.py")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	p := &PythonParser{}
	symbols, err := p.Parse(path)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, s := range symbols {
		names[s.Name] = true
	}

	assert.True(t, names["MAX_RETRIES"])
	assert.True(t, names["UserService"])
	assert.True(t, names["fetch_data"])
	assert.True(t, names["async_fetch"])
}
