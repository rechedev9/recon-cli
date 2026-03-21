package scanner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type gitScanner struct{}

func (s *gitScanner) Scan(ctx context.Context, root string) (*GitReport, error) {
	if _, err := s.git(ctx, root, "rev-parse", "--git-dir"); err != nil {
		return nil, nil
	}

	report := &GitReport{}

	branch, err := s.git(ctx, root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("git branch: %w", err)
	}
	report.Branch = strings.TrimSpace(branch)

	status, err := s.git(ctx, root, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		report.DirtyFiles = len(strings.Split(trimmed, "\n"))
	}

	log, err := s.git(ctx, root, "log", "-5", "--format=%h|%s|%as")
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(log), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		report.LastCommits = append(report.LastCommits, CommitInfo{
			Hash:    parts[0],
			Subject: parts[1],
			Date:    parts[2],
		})
	}

	return report, nil
}

func (s *gitScanner) git(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
