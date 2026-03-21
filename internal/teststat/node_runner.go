package teststat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NodeRunner struct{}

func (r *NodeRunner) Command(root string) string {
	return "npx jest --json --forceExit"
}

func (r *NodeRunner) Run(ctx context.Context, root string) (*TestResult, error) {
	pkgPath := filepath.Join(root, "package.json")
	if _, err := os.Stat(pkgPath); err != nil {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx, "npx", "jest", "--json", "--forceExit")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("jest: %w", err)
	}

	result := parseJestJSON(string(out))
	result.Framework = "jest"
	result.Command = r.Command(root)

	return result, nil
}

type jestOutput struct {
	NumPassedTests  int `json:"numPassedTests"`
	NumFailedTests  int `json:"numFailedTests"`
	NumPendingTests int `json:"numPendingTests"`
	NumTotalTests   int `json:"numTotalTests"`
	TestResults     []struct {
		Name             string `json:"name"`
		AssertionResults []struct {
			FullName        string   `json:"fullName"`
			Status          string   `json:"status"`
			FailureMessages []string `json:"failureMessages"`
		} `json:"assertionResults"`
	} `json:"testResults"`
}

func parseJestJSON(output string) *TestResult {
	var jest jestOutput
	if err := json.Unmarshal([]byte(output), &jest); err != nil {
		return &TestResult{}
	}

	result := &TestResult{
		TotalTests: jest.NumTotalTests,
		Passed:     jest.NumPassedTests,
		Failed:     jest.NumFailedTests,
		Skipped:    jest.NumPendingTests,
	}

	for _, suite := range jest.TestResults {
		for _, test := range suite.AssertionResults {
			if test.Status == "failed" {
				msg := ""
				if len(test.FailureMessages) > 0 {
					msg = strings.Join(test.FailureMessages, "\n")
				}
				result.Failures = append(result.Failures, TestFailure{
					Name:    test.FullName,
					Package: suite.Name,
					Message: msg,
				})
			}
		}
	}

	return result
}
