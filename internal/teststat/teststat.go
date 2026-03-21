package teststat

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DetectRunner(root string) TestRunner {
	// Check for Go
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		return &GoRunner{}
	}

	// Check for Node (jest)
	if _, err := os.Stat(filepath.Join(root, "package.json")); err == nil {
		return &NodeRunner{}
	}

	// Check for Python (pytest)
	pyIndicators := []string{"conftest.py", "pytest.ini", "setup.cfg"}
	for _, f := range pyIndicators {
		if _, err := os.Stat(filepath.Join(root, f)); err == nil {
			return &PythonRunner{}
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "pyproject.toml")); err == nil {
		if strings.Contains(string(data), "[tool.pytest") {
			return &PythonRunner{}
		}
	}

	return nil
}

func RunTests(ctx context.Context, root string) (*TestResult, error) {
	runner := DetectRunner(root)
	if runner == nil {
		return nil, fmt.Errorf("no supported test framework detected in %s", root)
	}

	result, err := runner.Run(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("run tests: %w", err)
	}

	return result, nil
}
