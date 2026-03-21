package scanner

import (
	"context"
	"fmt"
)

type TreeAnalyzer interface {
	Scan(ctx context.Context, root string, depth int) (*TreeReport, error)
}

type LangAnalyzer interface {
	Scan(ctx context.Context, root string) (*LangReport, error)
}

type EntryPointAnalyzer interface {
	Scan(ctx context.Context, root string) ([]EntryPoint, error)
}

type GitAnalyzer interface {
	Scan(ctx context.Context, root string) (*GitReport, error)
}

type DepsAnalyzer interface {
	Scan(ctx context.Context, root string) (*DepsReport, error)
}

type DocsAnalyzer interface {
	Scan(ctx context.Context, root string) (*DocsReport, error)
}

type Scanner struct {
	Tree        TreeAnalyzer
	Lang        LangAnalyzer
	EntryPoints EntryPointAnalyzer
	Git         GitAnalyzer
	Deps        DepsAnalyzer
	Docs        DocsAnalyzer
}

func (s *Scanner) Run(ctx context.Context, root string, depth int) (*Report, error) {
	report := &Report{Path: root}

	if s.Tree != nil {
		tree, err := s.Tree.Scan(ctx, root, depth)
		if err != nil {
			return nil, fmt.Errorf("tree scan: %w", err)
		}
		report.Tree = tree
	}

	if s.Lang != nil {
		lang, err := s.Lang.Scan(ctx, root)
		if err != nil {
			return nil, fmt.Errorf("language scan: %w", err)
		}
		report.Languages = lang
	}

	if s.EntryPoints != nil {
		points, err := s.EntryPoints.Scan(ctx, root)
		if err != nil {
			return nil, fmt.Errorf("entry point scan: %w", err)
		}
		report.EntryPoints = points
	}

	if s.Git != nil {
		git, err := s.Git.Scan(ctx, root)
		if err != nil {
			return nil, fmt.Errorf("git scan: %w", err)
		}
		report.Git = git
	}

	if s.Deps != nil {
		deps, err := s.Deps.Scan(ctx, root)
		if err != nil {
			return nil, fmt.Errorf("deps scan: %w", err)
		}
		report.Dependencies = deps
	}

	if s.Docs != nil {
		docs, err := s.Docs.Scan(ctx, root)
		if err != nil {
			return nil, fmt.Errorf("docs scan: %w", err)
		}
		report.Docs = docs
	}

	return report, nil
}

func NewDefault() *Scanner {
	return &Scanner{
		Tree:        &treeScanner{},
		Lang:        &langScanner{},
		EntryPoints: &entryPointScanner{},
		Git:         &gitScanner{},
		Deps:        &depsScanner{},
		Docs:        &docsScanner{},
	}
}
