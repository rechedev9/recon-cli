MODULE   := github.com/rechedev9/recon-cli
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X $(MODULE)/internal/cli.version=$(VERSION)

LDFLAGS_FC := -s -w -X $(MODULE)/internal/filechunkcli.version=$(VERSION)
LDFLAGS_DG := -s -w -X $(MODULE)/internal/depgraphcli.version=$(VERSION)
LDFLAGS_TS := -s -w -X $(MODULE)/internal/teststatcli.version=$(VERSION)

.PHONY: fmt lint test check build build-filechunk build-depgraph build-teststat build-all install install-filechunk install-depgraph install-teststat install-all

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
