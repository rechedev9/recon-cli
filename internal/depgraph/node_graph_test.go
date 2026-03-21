package depgraph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePackageLock(t *testing.T) {
	root := t.TempDir()

	lockfile := `{
  "name": "myapp",
  "version": "1.0.0",
  "lockfileVersion": 3,
  "packages": {
    "": {
      "name": "myapp",
      "dependencies": {
        "express": "^4.18.0"
      }
    },
    "node_modules/express": {
      "version": "4.18.2",
      "dependencies": {
        "body-parser": "1.20.1"
      }
    },
    "node_modules/body-parser": {
      "version": "1.20.1"
    }
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(root, "package-lock.json"), []byte(lockfile), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"myapp"}`), 0o644))

	b := &NodeGraphBuilder{}
	report, err := b.Build(root)
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Equal(t, "package-lock.json", report.Manager)
	require.Len(t, report.Nodes, 1) // express
	assert.Equal(t, "express", report.Nodes[0].Name)
	assert.Equal(t, "4.18.2", report.Nodes[0].Version)
	require.Len(t, report.Nodes[0].Children, 1) // body-parser
}
