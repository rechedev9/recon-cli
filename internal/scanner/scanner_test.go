package scanner

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTree struct{ report *TreeReport }
func (f *fakeTree) Scan(_ context.Context, _ string, _ int) (*TreeReport, error) {
	return f.report, nil
}

type fakeLang struct{ report *LangReport }
func (f *fakeLang) Scan(_ context.Context, _ string) (*LangReport, error) {
	return f.report, nil
}

type fakeEntryPoints struct{ points []EntryPoint }
func (f *fakeEntryPoints) Scan(_ context.Context, _ string) ([]EntryPoint, error) {
	return f.points, nil
}

type fakeGit struct{ report *GitReport }
func (f *fakeGit) Scan(_ context.Context, _ string) (*GitReport, error) {
	return f.report, nil
}

type fakeDeps struct{ report *DepsReport }
func (f *fakeDeps) Scan(_ context.Context, _ string) (*DepsReport, error) {
	return f.report, nil
}

type fakeDocs struct{ report *DocsReport }
func (f *fakeDocs) Scan(_ context.Context, _ string) (*DocsReport, error) {
	return f.report, nil
}

func TestScannerRun(t *testing.T) {
	s := &Scanner{
		Tree:        &fakeTree{report: &TreeReport{TotalFiles: 10}},
		Lang:        &fakeLang{report: &LangReport{TotalFiles: 10}},
		EntryPoints: &fakeEntryPoints{points: []EntryPoint{{Type: "main", Path: "main.go"}}},
		Git:         &fakeGit{report: &GitReport{Branch: "main"}},
		Deps:        &fakeDeps{report: &DepsReport{}},
		Docs:        &fakeDocs{report: &DocsReport{HasReadme: true}},
	}

	report, err := s.Run(context.Background(), "/tmp/test", 4)
	require.NoError(t, err)

	assert.Equal(t, "/tmp/test", report.Path)
	assert.Equal(t, 10, report.Tree.TotalFiles)
	assert.Equal(t, "main", report.Git.Branch)
	assert.True(t, report.Docs.HasReadme)
}

func TestScannerRunSkipSections(t *testing.T) {
	s := &Scanner{
		Tree:        &fakeTree{report: &TreeReport{TotalFiles: 5}},
		Lang:        &fakeLang{report: &LangReport{TotalFiles: 5}},
		EntryPoints: &fakeEntryPoints{points: nil},
	}

	report, err := s.Run(context.Background(), "/tmp/test", 4)
	require.NoError(t, err)

	assert.NotNil(t, report.Tree)
	assert.Nil(t, report.Git)
	assert.Nil(t, report.Dependencies)
	assert.Nil(t, report.Docs)
}

func TestScannerRunConcurrentError(t *testing.T) {
	errFake := &fakeTreeErr{}
	s := &Scanner{
		Tree:        errFake,
		Lang:        &fakeLang{report: &LangReport{TotalFiles: 5}},
		EntryPoints: &fakeEntryPoints{points: nil},
		Git:         &fakeGit{report: &GitReport{Branch: "main"}},
	}

	_, err := s.Run(context.Background(), "/tmp/test", 4)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tree scan")
}

type fakeTreeErr struct{}

func (f *fakeTreeErr) Scan(_ context.Context, _ string, _ int) (*TreeReport, error) {
	return nil, fmt.Errorf("fake tree error")
}
