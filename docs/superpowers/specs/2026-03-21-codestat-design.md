---
summary: Design spec for codestat — a CLI tool that summarizes project structure in one call
read_when: implementing or modifying codestat
---

# codestat — Project Structure Summarizer

## Problem

AI coding assistants (Claude Code, etc.) burn tokens making many sequential calls to understand a codebase: glob for files, grep for patterns, read for contents, git for history, cat for configs. A single CLI call that returns a compact project summary would replace 5-10 tool invocations.

## Solution

`codestat` — a Go CLI that scans a project directory and outputs a structured JSON report covering: directory tree, language breakdown, entry points, git summary, dependencies, and documentation index.

## CLI Interface

```
codestat [path] [flags]

Flags:
  --format json|md    Output format (default: json)
  --depth N           Directory tree depth (default: 4)
  --no-git            Skip git summary
  --no-deps           Skip dependency analysis
  --no-docs           Skip doc index
  --version           Print version
```

Default `path` is `.` (current directory). Single command, no subcommands. Sections disabled by flags are omitted from output entirely. All `exec.Command` calls use `exec.CommandContext` with the parent context; the CLI sets a 30s default timeout via `--timeout`.

## Output Structure

```json
{
  "path": "/absolute/path/to/project",
  "tree": {
    "total_files": 42,
    "total_dirs": 8,
    "entries": [
      {
        "path": "cmd",
        "is_dir": true,
        "files": 2,
        "children": [
          { "path": "cmd/codestat", "is_dir": true, "files": 1 }
        ]
      },
      { "path": "go.mod", "is_dir": false }
    ]
  },
  "languages": {
    "total_files": 42,
    "total_loc": 3200,
    "languages": [
      { "name": "Go", "files": 30, "loc": 2800 },
      { "name": "Markdown", "files": 8, "loc": 300 },
      { "name": "YAML", "files": 4, "loc": 100 }
    ]
  },
  "entry_points": [
    { "type": "main", "path": "cmd/codestat/main.go" },
    { "type": "makefile", "path": "Makefile" },
    { "type": "script", "path": "scripts/committer" }
  ],
  "git": {
    "branch": "main",
    "dirty_files": 3,
    "last_commits": [
      { "hash": "abc1234", "subject": "feat(scanner): add language detection", "date": "2026-03-21" },
      { "hash": "def5678", "subject": "chore: init project", "date": "2026-03-20" }
    ]
  },
  "dependencies": {
    "manifests": [
      {
        "manager": "go.mod",
        "path": "go.mod",
        "deps": [
          { "name": "github.com/spf13/cobra", "version": "v1.8.1" }
        ]
      },
      {
        "manager": "package.json",
        "path": "frontend/package.json",
        "deps": [
          { "name": "react", "version": "^18.2.0" }
        ]
      }
    ]
  },
  "docs": {
    "has_readme": true,
    "has_claude_md": true,
    "docs_dir": true,
    "files": ["README.md", "CLAUDE.md", "docs/handoff.md", "docs/pickup.md"]
  }
}
```

## Entry Point

```go
// cmd/codestat/main.go — thin shell (REQUIRED by go-skill)
package main

import (
    "context"
    "os"

    "github.com/rechedev9/CLIClaudeCode/internal/cli"
)

func main() {
    ctx := context.Background()
    if err := cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
        os.Exit(cli.ExitCode(err))
    }
}
```

Exit codes: `0` success, `1` runtime error, `2` usage error. `cli.ExitCode(err)` maps error types to codes. Scanner errors propagate up via `fmt.Errorf` wrapping; `cli.Run` prints the error to stderr and returns it for `main` to exit.

## Project Structure

```
cmd/codestat/main.go         # 5-15 lines, delegates to internal/cli
internal/
  cli/
    root.go                   # Cobra root command, flag parsing, orchestration
    exitcode.go               # ExitCode(error) int — maps errors to exit codes
    version.go                # Version var for ldflags injection
  scanner/
    scanner.go                # Orchestrator — defines consumer interfaces, assembles Report
    tree.go                   # Directory tree walker (hardcoded skip list)
    languages.go              # Language detection by extension + LOC counting
    entrypoints.go            # Entry point detection (main files, Makefiles, scripts)
    git.go                    # Git summary (branch, commits, dirty count) via git CLI
    deps.go                   # Dependency manifest parser (go.mod, package.json, requirements.txt)
    docs.go                   # Doc index (README, CLAUDE.md, docs/ presence)
    types.go                  # All output structs
  output/
    format.go                 # JSON and markdown formatters
Makefile
.goreleaser.yaml
.golangci.yml
go.mod
```

## Consumer-Defined Interfaces

The scanner orchestrator defines interfaces for each analyzer (REQUIRED by go-skill). Concrete implementations are in separate files. This enables testing the orchestrator with fakes.

```go
// scanner.go — consumer defines what it needs
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
```

## Data Structures

```go
type Report struct {
    Path         string        `json:"path"`
    Tree         *TreeReport   `json:"tree,omitempty"`
    Languages    *LangReport   `json:"languages,omitempty"`
    EntryPoints  []EntryPoint  `json:"entry_points,omitempty"`
    Git          *GitReport    `json:"git,omitempty"`
    Dependencies *DepsReport   `json:"dependencies,omitempty"`
    Docs         *DocsReport   `json:"docs,omitempty"`
}

type TreeReport struct {
    TotalFiles int        `json:"total_files"`
    TotalDirs  int        `json:"total_dirs"`
    Entries    []TreeNode `json:"entries"`
}

type TreeNode struct {
    Path     string     `json:"path"`
    IsDir    bool       `json:"is_dir"`
    Files    int        `json:"files,omitempty"`
    Children []TreeNode `json:"children,omitempty"`
}

type LangReport struct {
    TotalFiles int         `json:"total_files"`
    TotalLOC   int         `json:"total_loc"`
    Languages  []LangEntry `json:"languages"`
}

type LangEntry struct {
    Name  string `json:"name"`
    Files int    `json:"files"`
    LOC   int    `json:"loc"`
}

// EntryPoint.Type is a closed set: "main", "makefile", "dockerfile", "script", "package.json", "pyproject"
type EntryPoint struct {
    Type string `json:"type"`
    Path string `json:"path"`
}

type GitReport struct {
    Branch      string       `json:"branch"`
    DirtyFiles  int          `json:"dirty_files"`
    LastCommits []CommitInfo `json:"last_commits"`
}

type CommitInfo struct {
    Hash    string `json:"hash"`
    Subject string `json:"subject"`
    Date    string `json:"date"`
}

type DepsReport struct {
    Manifests []ManifestReport `json:"manifests"`
}

type ManifestReport struct {
    Manager string     `json:"manager"`
    Path    string     `json:"path"`
    Deps    []DepEntry `json:"deps"`
}

type DepEntry struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

type DocsReport struct {
    HasReadme   bool     `json:"has_readme"`
    HasClaudeMD bool     `json:"has_claude_md"`
    DocsDir     bool     `json:"docs_dir"`
    Files       []string `json:"files,omitempty"`
}
```

## Scanner Behavior

### Tree Scanner
- Walks directory up to `--depth` levels (default 4)
- Hardcoded skip list: `.git`, `node_modules`, `vendor`, `__pycache__`, `.next`, `dist`, `build`, hidden dirs (`.foo`)
- v1 uses hardcoded skip list only; `.gitignore` respect deferred to v2 (correct `.gitignore` parsing including nested files, negation patterns, and `**` globs is substantial complexity)
- Reports file counts per directory

### Language Detection
- Extension-based: `.go` → Go, `.ts`/`.tsx` → TypeScript, `.py` → Python, `.rs` → Rust, `.js`/`.jsx` → JavaScript, `.md` → Markdown, `.yaml`/`.yml` → YAML, `.json` → JSON, `.sh` → Shell, `.sql` → SQL, `.html` → HTML, `.css` → CSS, `.toml` → TOML
- LOC counts all lines containing at least one non-whitespace character. No language-specific comment detection.
- Sorted by LOC descending

### Entry Point Detection
- Go: files with `package main` + `func main()`
- Node: `package.json` with `scripts` or `main`/`bin` fields
- Python: `__main__.py`, `setup.py`, `pyproject.toml`
- Rust: `src/main.rs`
- Generic: `Makefile`, `Dockerfile`, `docker-compose.yml`, `scripts/*` (executable files)

### Git Summary
- Shells out to `git` CLI via `exec.CommandContext` (not a Go git library — keeps deps minimal)
- Branch: `git rev-parse --abbrev-ref HEAD`
- Dirty count: `git status --porcelain | wc -l`
- Last 5 commits: `git log -5 --format="%h|%s|%as"`
- Gracefully returns nil if not a git repo
- All commands respect parent context for cancellation/timeout

### Dependency Parser
- Scans for all known manifests recursively (respects same skip list as tree)
- v1 supported manifests (formats parseable without TOML/YAML libs):
  - `go.mod`: parses `require` blocks for direct deps (custom format, line-based)
  - `package.json`: parses `dependencies` + `devDependencies` (via `encoding/json`)
  - `requirements.txt`: line-based `name==version`
- v2 deferred (requires TOML library): `Cargo.toml`, `pyproject.toml`
- Only direct dependencies — no transitive resolution

### Doc Index
- Checks root for: `README.md` (case-insensitive), `CLAUDE.md`, `AGENTS.md`
- Checks for `docs/` directory existence
- Lists all `.md` files in root + `docs/`

## Dependencies

```
github.com/spf13/cobra       # CLI framework
github.com/stretchr/testify   # Testing (dev only)
```

Standard library for everything else: `os`, `path/filepath`, `encoding/json`, `os/exec`, `bufio`, `strings`, `fmt`, `errors`, `sort`.

Standard library for everything else. No TOML/YAML parsing libraries in v1 — manifest parsers use line-based parsing (`go.mod`, `requirements.txt`) or `encoding/json` (`package.json`).

## Build

```makefile
MODULE   := github.com/rechedev9/CLIClaudeCode
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X $(MODULE)/internal/cli.version=$(VERSION)

.PHONY: fmt lint test check build install

fmt:
	gofumpt -w .

lint:
	golangci-lint run

test:
	go test ./...

check: fmt lint test

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/codestat ./cmd/codestat

install: build
	cp bin/codestat ~/.local/bin/codestat
```

## GoReleaser

```yaml
version: 2
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    main: ./cmd/codestat
    ldflags: -s -w -X github.com/rechedev9/CLIClaudeCode/internal/cli.version={{ .Version }}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
```

## golangci-lint

```yaml
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
```

## Testing Strategy

- Table-driven tests per scanner (tree, languages, git, deps, docs)
- Each test creates its own temp directory with `t.TempDir()`
- Git tests init a real git repo in temp dir
- Hand-written fakes where needed (no gomock)
- `testify/assert` and `testify/require` for assertions
- Integration test: run full `codestat` on a fixture project, validate JSON output

## Markdown Output

When `--format md` is used, output a compact human-readable summary:

```markdown
# /path/to/project

## Structure (42 files, 8 dirs)
cmd/codestat/ (1 file)
internal/cli/ (2 files)
internal/scanner/ (8 files)
internal/output/ (1 file)

## Languages (3200 LOC)
Go: 30 files, 2800 LOC
Markdown: 8 files, 300 LOC
YAML: 4 files, 100 LOC

## Entry Points
main: cmd/codestat/main.go
makefile: Makefile

## Git (main, 3 dirty)
abc1234 feat(scanner): add language detection (2026-03-21)
def5678 chore: init project (2026-03-20)

## Dependencies
go.mod: cobra v1.8.1

## Docs
README.md, CLAUDE.md, docs/ (2 files)
```

## Future Tools (Same Monorepo)

This project is structured as a monorepo. Future tools will live alongside codestat:
- `cmd/filechunk/` — smart file reader returning only signatures/exports
- `cmd/depgraph/` — dependency tree analyzer
- `cmd/teststat/` — test runner + result summarizer

Shared code in `internal/output/` (formatters) and potentially `internal/scanner/` (language detection, gitignore handling).
