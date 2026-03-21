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

type GoRunner struct{}

func (r *GoRunner) Command(root string) string {
	return "go test -json ./..."
}

func (r *GoRunner) Run(ctx context.Context, root string) (*TestResult, error) {
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return nil, nil
	}

	goBin := findGoBin()
	cmd := exec.CommandContext(ctx, goBin, "test", "-json", "./...")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	// go test returns exit code 1 on test failure, but still produces valid JSON
	// Only treat as error if no output was produced
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("go test: %w", err)
	}

	result := parseGoTestJSON(string(out))
	result.Framework = "go"
	result.Command = r.Command(root)

	return result, nil
}

func findGoBin() string {
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}
	return "/usr/local/go/bin/go"
}

type goTestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

func parseGoTestJSON(output string) *TestResult {
	result := &TestResult{}
	failureOutputs := make(map[string][]string) // test name -> output lines

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}

		var event goTestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Only count test-level events (not package-level)
		if event.Test == "" {
			// Package-level pass/fail gives us duration
			if (event.Action == "pass" || event.Action == "fail") && event.Elapsed > 0 {
				result.Duration = fmt.Sprintf("%.1fs", event.Elapsed)
			}
			continue
		}

		switch event.Action {
		case "pass":
			result.TotalTests++
			result.Passed++
		case "fail":
			result.TotalTests++
			result.Failed++
			// Collect failure output
			msg := strings.Join(failureOutputs[event.Test], "")
			result.Failures = append(result.Failures, TestFailure{
				Name:    event.Test,
				Package: event.Package,
				Message: strings.TrimSpace(msg),
			})
		case "skip":
			result.TotalTests++
			result.Skipped++
		case "output":
			if event.Test != "" {
				failureOutputs[event.Test] = append(failureOutputs[event.Test], event.Output)
			}
		}
	}

	return result
}
