package teststat

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type PythonRunner struct{}

func (r *PythonRunner) Command(root string) string {
	return "python -m pytest --tb=short -q"
}

func (r *PythonRunner) Run(ctx context.Context, root string) (*TestResult, error) {
	// Check for pytest indicators
	indicators := []string{"conftest.py", "pytest.ini", "setup.cfg"}
	found := false
	for _, f := range indicators {
		if _, err := os.Stat(filepath.Join(root, f)); err == nil {
			found = true
			break
		}
	}
	if !found {
		// Check pyproject.toml for pytest config
		if data, err := os.ReadFile(filepath.Join(root, "pyproject.toml")); err == nil {
			if strings.Contains(string(data), "[tool.pytest") {
				found = true
			}
		}
	}
	if !found {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx, "python", "-m", "pytest", "--tb=short", "-q")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("pytest: %w", err)
	}

	result := parsePytestOutput(string(out))
	result.Framework = "pytest"
	result.Command = r.Command(root)

	return result, nil
}

var pytestSummaryRe = regexp.MustCompile(`(\d+) passed(?:.*?(\d+) failed)?(?:.*?(\d+) skipped)? in ([\d.]+s)`)
var pytestFailedRe = regexp.MustCompile(`^FAILED (.+?) - (.+)$`)

func parsePytestOutput(output string) *TestResult {
	result := &TestResult{}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		// Parse FAILED lines
		if matches := pytestFailedRe.FindStringSubmatch(line); matches != nil {
			result.Failures = append(result.Failures, TestFailure{
				Name:    matches[1],
				Message: matches[2],
			})
			continue
		}

		// Parse summary line: "3 passed, 2 failed, 1 skipped in 2.45s"
		if matches := pytestSummaryRe.FindStringSubmatch(line); matches != nil {
			result.Passed, _ = strconv.Atoi(matches[1])
			if matches[2] != "" {
				result.Failed, _ = strconv.Atoi(matches[2])
			}
			if matches[3] != "" {
				result.Skipped, _ = strconv.Atoi(matches[3])
			}
			result.Duration = matches[4]
			result.TotalTests = result.Passed + result.Failed + result.Skipped
		}
	}

	return result
}
