MODULE   := github.com/rechedev9/recon-cli
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X $(MODULE)/internal/cli.version=$(VERSION)

LDFLAGS_FC := -s -w -X $(MODULE)/internal/filechunkcli.version=$(VERSION)

.PHONY: fmt lint test check build build-filechunk build-all install install-filechunk install-all

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

build-all: build build-filechunk

install-filechunk: build-filechunk
	cp bin/filechunk ~/.local/bin/filechunk

install-all: install install-filechunk
