package teststat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJestJSON(t *testing.T) {
	output := `{
  "numPassedTests": 5,
  "numFailedTests": 1,
  "numPendingTests": 2,
  "numTotalTests": 8,
  "testResults": [
    {
      "name": "/app/src/utils.test.ts",
      "assertionResults": [
        {"fullName": "utils > parseDate should work", "status": "passed"},
        {"fullName": "utils > formatDate should fail", "status": "failed", "failureMessages": ["Expected 'foo' got 'bar'"]}
      ]
    }
  ]
}`

	result := parseJestJSON(output)

	assert.Equal(t, 8, result.TotalTests)
	assert.Equal(t, 5, result.Passed)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 2, result.Skipped)
	require.Len(t, result.Failures, 1)
	assert.Equal(t, "utils > formatDate should fail", result.Failures[0].Name)
	assert.Contains(t, result.Failures[0].Message, "Expected 'foo' got 'bar'")
}
