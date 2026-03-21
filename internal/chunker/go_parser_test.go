package chunker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoParser(t *testing.T) {
	root := t.TempDir()
	src := `package example

import "fmt"

const MaxRetries = 3

var DefaultTimeout = 30

type Config struct {
	Host string
	Port int
}

type Runner interface {
	Run() error
}

func NewConfig(host string, port int) *Config {
	return &Config{Host: host, Port: port}
}

func (c *Config) String() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
`
	path := filepath.Join(root, "example.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	p := &GoParser{}
	symbols, err := p.Parse(path)
	require.NoError(t, err)

	kinds := make(map[string][]string)
	for _, s := range symbols {
		kinds[s.Kind] = append(kinds[s.Kind], s.Name)
	}

	assert.Contains(t, kinds["const"], "MaxRetries")
	assert.Contains(t, kinds["var"], "DefaultTimeout")
	assert.Contains(t, kinds["struct"], "Config")
	assert.Contains(t, kinds["interface"], "Runner")
	assert.Contains(t, kinds["function"], "NewConfig")
	assert.Contains(t, kinds["function"], "Config.String")
}
