package teststat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePytestOutput(t *testing.T) {
	output := `FAILED tests/test_api.py::test_create_user - AssertionError: 404 != 201
FAILED tests/test_api.py::test_delete_user - KeyError: 'id'
3 passed, 2 failed, 1 skipped in 2.45s
`

	result := parsePytestOutput(output)

	assert.Equal(t, 6, result.TotalTests)
	assert.Equal(t, 3, result.Passed)
	assert.Equal(t, 2, result.Failed)
	assert.Equal(t, 1, result.Skipped)
	assert.Equal(t, "2.45s", result.Duration)
	require.Len(t, result.Failures, 2)
	assert.Equal(t, "tests/test_api.py::test_create_user", result.Failures[0].Name)
}
