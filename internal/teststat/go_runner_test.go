package teststat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGoTestJSON(t *testing.T) {
	// Simulated output of `go test -json ./...`
	output := `{"Time":"2026-03-21T12:00:00Z","Action":"start","Package":"example.com/app"}
{"Time":"2026-03-21T12:00:00Z","Action":"run","Package":"example.com/app","Test":"TestFoo"}
{"Time":"2026-03-21T12:00:00Z","Action":"output","Package":"example.com/app","Test":"TestFoo","Output":"--- PASS: TestFoo (0.00s)\n"}
{"Time":"2026-03-21T12:00:00Z","Action":"pass","Package":"example.com/app","Test":"TestFoo","Elapsed":0.001}
{"Time":"2026-03-21T12:00:00Z","Action":"run","Package":"example.com/app","Test":"TestBar"}
{"Time":"2026-03-21T12:00:00Z","Action":"output","Package":"example.com/app","Test":"TestBar","Output":"    bar_test.go:10: expected 1 got 2\n"}
{"Time":"2026-03-21T12:00:00Z","Action":"output","Package":"example.com/app","Test":"TestBar","Output":"--- FAIL: TestBar (0.00s)\n"}
{"Time":"2026-03-21T12:00:00Z","Action":"fail","Package":"example.com/app","Test":"TestBar","Elapsed":0.001}
{"Time":"2026-03-21T12:00:00Z","Action":"run","Package":"example.com/app","Test":"TestSkipped"}
{"Time":"2026-03-21T12:00:00Z","Action":"output","Package":"example.com/app","Test":"TestSkipped","Output":"--- SKIP: TestSkipped (0.00s)\n"}
{"Time":"2026-03-21T12:00:00Z","Action":"skip","Package":"example.com/app","Test":"TestSkipped","Elapsed":0}
{"Time":"2026-03-21T12:00:01Z","Action":"fail","Package":"example.com/app","Elapsed":1.5}
`

	result := parseGoTestJSON(output)

	assert.Equal(t, 3, result.TotalTests)
	assert.Equal(t, 1, result.Passed)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 1, result.Skipped)
	require.Len(t, result.Failures, 1)
	assert.Equal(t, "TestBar", result.Failures[0].Name)
	assert.Equal(t, "example.com/app", result.Failures[0].Package)
	assert.Contains(t, result.Failures[0].Message, "expected 1 got 2")
}

func TestParseGoTestJSONAllPass(t *testing.T) {
	output := `{"Time":"2026-03-21T12:00:00Z","Action":"run","Package":"example.com/app","Test":"TestA"}
{"Time":"2026-03-21T12:00:00Z","Action":"pass","Package":"example.com/app","Test":"TestA","Elapsed":0.001}
{"Time":"2026-03-21T12:00:00Z","Action":"run","Package":"example.com/app","Test":"TestB"}
{"Time":"2026-03-21T12:00:00Z","Action":"pass","Package":"example.com/app","Test":"TestB","Elapsed":0.002}
{"Time":"2026-03-21T12:00:01Z","Action":"pass","Package":"example.com/app","Elapsed":0.5}
`

	result := parseGoTestJSON(output)

	assert.Equal(t, 2, result.TotalTests)
	assert.Equal(t, 2, result.Passed)
	assert.Equal(t, 0, result.Failed)
	assert.Empty(t, result.Failures)
	assert.Equal(t, "0.5s", result.Duration)
}
