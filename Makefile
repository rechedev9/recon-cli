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
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/recon ./cmd/recon

install: build
	cp bin/recon ~/.local/bin/recon
