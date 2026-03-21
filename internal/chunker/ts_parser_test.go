package chunker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTSParser(t *testing.T) {
	root := t.TempDir()
	src := `export function fetchData(url: string): Promise<Response> {
  return fetch(url);
}

export default class UserService {
  constructor(private db: Database) {}
}

export interface Config {
  host: string;
  port: number;
}

export type UserID = string;

export const MAX_RETRIES = 3;

function privateHelper() {}

class InternalClass {}
`
	path := filepath.Join(root, "service.ts")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	p := &TSParser{}
	symbols, err := p.Parse(path)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, s := range symbols {
		names[s.Name] = true
	}

	assert.True(t, names["fetchData"])
	assert.True(t, names["UserService"])
	assert.True(t, names["Config"])
	assert.True(t, names["UserID"])
	assert.True(t, names["MAX_RETRIES"])
	assert.True(t, names["privateHelper"])
	assert.True(t, names["InternalClass"])
}
