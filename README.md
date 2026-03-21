<p align="center">
  <h1 align="center">recon — Project Reconnaissance CLI</h1>
</p>

<p align="center">
  <strong>One CLI call replaces 10 tool invocations.<br>Structured project summaries for AI coding assistants.</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go" alt="Go 1.24">
  <img src="https://img.shields.io/badge/Scanners-6-7C3AED?style=for-the-badge" alt="Scanners: 6">
  <img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge" alt="MIT License">
</p>

---

AI coding assistants burn tokens understanding codebases — glob for files, grep for patterns, read configs, check git, parse dependencies. Each call costs tokens and context. `recon` replaces all of that with a single structured output.

```
Without recon        With recon
─────────────        ──────────
glob **/*.go         recon .       ← one call
grep "func main"
read go.mod
read package.json
git status
git log
ls docs/
read README.md
───────────────
8+ tool calls        1 tool call
```

---

## Install

```bash
git clone https://github.com/rechedev9/recon-cli
cd recon-cli
make build && make install
```

Requires [Go 1.24+](https://go.dev/dl/).

Or download a pre-built binary from [Releases](https://github.com/rechedev9/recon-cli/releases).

## Quick start

```bash
recon .                         # JSON output (default)
recon . --format md             # Compact markdown
recon ~/other/project           # Any path
recon . --no-git --no-deps      # Skip sections
recon . --depth 6               # Deeper tree
```

---

## What it reports

| Section | What it scans | Key data |
|---------|--------------|----------|
| **Tree** | Directory structure | Files, dirs, depth-limited traversal |
| **Languages** | File extensions + LOC | Per-language file counts and lines of code |
| **Entry Points** | Main files, Makefiles, scripts | Go mains, Dockerfiles, package.json, scripts/* |
| **Git** | Repository state | Branch, dirty files, last 5 commits |
| **Dependencies** | Manifest files | go.mod, package.json, requirements.txt |
| **Docs** | Documentation index | README, CLAUDE.md, docs/ contents |

---

## Output formats

### JSON (default)

```bash
recon .
```

```json
{
  "path": "/home/user/myproject",
  "tree": { "total_files": 42, "total_dirs": 8, "entries": [...] },
  "languages": { "total_files": 42, "total_loc": 3200, "languages": [...] },
  "entry_points": [{ "type": "main", "path": "cmd/app/main.go" }],
  "git": { "branch": "main", "dirty_files": 0, "last_commits": [...] },
  "dependencies": { "manifests": [{ "manager": "go.mod", "deps": [...] }] },
  "docs": { "has_readme": true, "has_claude_md": true, "files": [...] }
}
```

### Markdown

```bash
recon . --format md
```

```
# /home/user/myproject

## Structure (42 files, 8 dirs)
cmd/app/ (1 file)
internal/cli/ (3 files)

## Languages (3200 LOC)
Go: 30 files, 2800 LOC
Markdown: 8 files, 300 LOC

## Entry Points
main: cmd/app/main.go
makefile: Makefile

## Git (main, 0 dirty)
abc1234 feat: add feature (2026-03-21)

## Dependencies
go.mod: github.com/spf13/cobra v1.10.2

## Docs
README.md, CLAUDE.md, docs/ (2 files)
```

---

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `json` | Output format: `json` or `md` |
| `--depth` | `4` | Directory tree depth limit |
| `--no-git` | `false` | Skip git summary |
| `--no-deps` | `false` | Skip dependency analysis |
| `--no-docs` | `false` | Skip doc index |
| `--timeout` | `30` | Timeout in seconds for external commands |
| `--version` | | Print version |

Sections disabled by flags are omitted from output entirely.

---

## Claude Code integration

Add a `SessionStart` hook to `~/.claude/settings.json` so Claude gets project context automatically:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "RECON=$(recon . --format md 2>/dev/null) && jq -n --arg ctx \"$RECON\" '{hookSpecificOutput:{hookEventName:\"SessionStart\",additionalContext:$ctx}}'",
            "timeout": 10,
            "statusMessage": "Running recon..."
          }
        ]
      }
    ]
  }
}
```

Every session starts with a full project summary in Claude's context — zero manual exploration needed.

---

## Architecture

```
cmd/recon/main.go              # Thin shell — 15 lines, delegates to internal/cli
internal/
  cli/
    root.go                    # Cobra command, flag parsing, orchestration
    exitcode.go                # Error → exit code mapping (0/1/2)
    version.go                 # Version injection via ldflags
  scanner/
    scanner.go                 # Consumer-defined interfaces + orchestrator
    tree.go                    # Directory walker with skip list
    languages.go               # Extension-based detection + LOC counting
    entrypoints.go             # Main files, Makefiles, scripts, Dockerfiles
    git.go                     # Branch, dirty count, commits via exec.CommandContext
    deps.go                    # go.mod, package.json, requirements.txt parsers
    docs.go                    # README, CLAUDE.md, docs/ index
    types.go                   # All output structs
  output/
    format.go                  # JSON and markdown formatters
```

### Design principles

- **Thin shell, fat core** — `main.go` is 15 lines. All logic in `internal/`.
- **Consumer-defined interfaces** — 6 analyzer interfaces defined in `scanner.go` where consumed, not at provider.
- **Error wrapping** — every error wrapped with `fmt.Errorf("context: %w", err)`.
- **Zero external deps** — only Cobra (CLI) and testify (tests). Everything else is stdlib.
- **CGO_ENABLED=0** — pure Go, cross-compiles to linux/darwin/windows.

### Scanner skip list

The tree walker skips: `.git`, `node_modules`, `vendor`, `__pycache__`, `.next`, `dist`, `build`, and all hidden directories.

---

## Build

```bash
make build          # Build to bin/recon
make install        # Copy to ~/.local/bin/recon
make test           # Run all tests
make lint           # golangci-lint
make check          # fmt + lint + test
```

### Cross-platform releases

```bash
goreleaser release --snapshot --clean
```

Builds for linux/darwin/windows on amd64/arm64.

---

## Testing

- **25 tests** across 3 packages (cli, output, scanner)
- Table-driven tests with `t.TempDir()` — each test owns its setup
- Hand-written fakes for orchestrator testing (no gomock)
- Integration tests with real git repos in temp dirs
- `testify/assert` and `testify/require` for assertions

```bash
go test ./... -v
```

---

## Roadmap

- [ ] `.gitignore` respect (v2) — proper parsing including nested files and negation
- [ ] `Cargo.toml` / `pyproject.toml` parsing (v2) — requires TOML library
- [ ] Concurrent scanners — run 6 analyzers in parallel with errgroup
- [ ] `filechunk` — smart file reader returning only signatures/exports
- [ ] `depgraph` — full dependency tree analyzer
- [ ] `teststat` — test runner + result summarizer

---

## Contributing

- **Scanners** — new language detection, manifest parsers, entry point types
- **Output** — new formats (YAML, TOML, custom templates)
- **Performance** — concurrent scanning, buffer reuse
- **Tests** — edge cases, large repo benchmarks

---

## License

MIT
