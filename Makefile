MODULE   := github.com/rechedev9/recon-cli
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X $(MODULE)/internal/cli.version=$(VERSION)

LDFLAGS_FC := -s -w -X $(MODULE)/internal/filechunkcli.version=$(VERSION)
LDFLAGS_DG := -s -w -X $(MODULE)/internal/depgraphcli.version=$(VERSION)
LDFLAGS_TS := -s -w -X $(MODULE)/internal/teststatcli.version=$(VERSION)

.PHONY: fmt lint test check build build-filechunk build-depgraph build-teststat build-all install install-filechunk install-depgraph install-teststat install-all fuzz-quick

fmt:
	gofumpt -w .

lint:
	golangci-lint run

test:
	go test ./...

check: fmt lint test

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/recon ./cmd/recon

install: build
	cp bin/recon ~/.local/bin/recon

build-filechunk:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS_FC)" -o bin/filechunk ./cmd/filechunk

build-depgraph:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS_DG)" -o bin/depgraph ./cmd/depgraph

build-teststat:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS_TS)" -o bin/teststat ./cmd/teststat

build-all: build build-filechunk build-depgraph build-teststat

install-filechunk: build-filechunk
	cp bin/filechunk ~/.local/bin/filechunk

install-depgraph: build-depgraph
	cp bin/depgraph ~/.local/bin/depgraph

install-teststat: build-teststat
	cp bin/teststat ~/.local/bin/teststat

install-all: install install-filechunk install-depgraph install-teststat

fuzz-quick:
	@echo "Fuzzing parsers (5s each)..."
	@go test ./internal/scanner -fuzz=FuzzParseGoMod -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/scanner -fuzz=FuzzParsePackageJSON -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/scanner -fuzz=FuzzParseRequirementsTxt -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/scanner -fuzz=FuzzParseCargoToml -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/scanner -fuzz=FuzzParsePyprojectToml -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/scanner -fuzz=FuzzParsePEP508 -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/scanner -fuzz=FuzzParseGitLog -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/chunker -fuzz=FuzzGoParser -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/chunker -fuzz=FuzzTSParser -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/chunker -fuzz=FuzzPythonParser -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/chunker -fuzz=FuzzRustParser -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/teststat -fuzz=FuzzParseGoTestJSON -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/teststat -fuzz=FuzzParseJestJSON -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/teststat -fuzz=FuzzParsePytestOutput -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@go test ./internal/depgraph -fuzz=FuzzParseGoModGraphOutput -fuzztime=5s -run=^$$ 2>&1 | tail -1
	@echo "Fuzz complete."
