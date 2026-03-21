package depgraph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGoModGraph(t *testing.T) {
	// Simulated output of `go mod graph`
	output := `example.com/app github.com/spf13/cobra@v1.8.1
example.com/app github.com/stretchr/testify@v1.9.0
github.com/spf13/cobra@v1.8.1 github.com/spf13/pflag@v1.0.5
github.com/stretchr/testify@v1.9.0 github.com/davecgh/go-spew@v1.1.1
github.com/stretchr/testify@v1.9.0 github.com/pmezard/go-difflib@v1.0.0
`

	nodes := parseGoModGraphOutput(output, "example.com/app")

	require.Len(t, nodes, 2) // cobra and testify at top level

	cobraNode := findNode(nodes, "github.com/spf13/cobra")
	require.NotNil(t, cobraNode)
	assert.Equal(t, "v1.8.1", cobraNode.Version)
	require.Len(t, cobraNode.Children, 1)
	assert.Equal(t, "github.com/spf13/pflag", cobraNode.Children[0].Name)

	testifyNode := findNode(nodes, "github.com/stretchr/testify")
	require.NotNil(t, testifyNode)
	require.Len(t, testifyNode.Children, 2)
}

func findNode(nodes []DepNode, name string) *DepNode {
	for i := range nodes {
		if nodes[i].Name == name {
			return &nodes[i]
		}
	}
	return nil
}
