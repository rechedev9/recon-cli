package teststat

import "context"

type TestResult struct {
	Framework  string        `json:"framework"`
	TotalTests int           `json:"total_tests"`
	Passed     int           `json:"passed"`
	Failed     int           `json:"failed"`
	Skipped    int           `json:"skipped"`
	Duration   string        `json:"duration"`
	Command    string        `json:"command"`
	Failures   []TestFailure `json:"failures,omitempty"`
}

type TestFailure struct {
	Name    string `json:"name"`
	Package string `json:"package,omitempty"`
	Message string `json:"message"`
}

type TestRunner interface {
	Run(ctx context.Context, root string) (*TestResult, error)
	Command(root string) string // for --dry-run
}
