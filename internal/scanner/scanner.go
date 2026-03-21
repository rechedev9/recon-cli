package scanner

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
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
	g, ctx := errgroup.WithContext(ctx)

	if s.Tree != nil {
		g.Go(func() error {
			tree, err := s.Tree.Scan(ctx, root, depth)
			if err != nil {
				return fmt.Errorf("tree scan: %w", err)
			}
			report.Tree = tree
			return nil
		})
	}

	if s.Lang != nil {
		g.Go(func() error {
			lang, err := s.Lang.Scan(ctx, root)
			if err != nil {
				return fmt.Errorf("language scan: %w", err)
			}
			report.Languages = lang
			return nil
		})
	}

	if s.EntryPoints != nil {
		g.Go(func() error {
			points, err := s.EntryPoints.Scan(ctx, root)
			if err != nil {
				return fmt.Errorf("entry point scan: %w", err)
			}
			report.EntryPoints = points
			return nil
		})
	}

	if s.Git != nil {
		g.Go(func() error {
			git, err := s.Git.Scan(ctx, root)
			if err != nil {
				return fmt.Errorf("git scan: %w", err)
			}
			report.Git = git
			return nil
		})
	}

	if s.Deps != nil {
		g.Go(func() error {
			deps, err := s.Deps.Scan(ctx, root)
			if err != nil {
				return fmt.Errorf("deps scan: %w", err)
			}
			report.Dependencies = deps
			return nil
		})
	}

	if s.Docs != nil {
		g.Go(func() error {
			docs, err := s.Docs.Scan(ctx, root)
			if err != nil {
				return fmt.Errorf("docs scan: %w", err)
			}
			report.Docs = docs
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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
