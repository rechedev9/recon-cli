package depgraph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePythonRequirements(t *testing.T) {
	root := t.TempDir()

	reqs := "flask==2.3.0\nrequests==2.31.0\n# comment\nnumpy>=1.24\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "requirements.txt"), []byte(reqs), 0o644))

	b := &PythonGraphBuilder{}
	report, err := b.Build(root)
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Equal(t, "requirements.txt", report.Manager)
	assert.Len(t, report.Nodes, 3) // flat list, no children
}
