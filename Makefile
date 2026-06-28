BINARY      := phpvm
MODULE      := github.com/mahtdy/phpvm
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE        ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS     := -s -w \
	-X '$(MODULE)/pkg/version.Version=$(VERSION)' \
	-X '$(MODULE)/pkg/version.Commit=$(COMMIT)' \
	-X '$(MODULE)/pkg/version.Date=$(DATE)' \
	-X '$(MODULE)/pkg/version.BuiltBy=make'

GO          := go
GOFLAGS     :=
BUILD_DIR   := dist

.PHONY: all build test lint clean release fmt vet help

## build: compile binary for the current platform
build:
	@echo "Building $(BINARY) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) .

## install: build and install to GOPATH/bin
install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" .

## test: run all unit tests
test:
	$(GO) test -v -race -count=1 ./...

## test-cover: run tests with coverage report
test-cover:
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## fmt: format all Go source files
fmt:
	$(GO) fmt ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## lint: run golangci-lint (must be installed)
lint:
	golangci-lint run ./...

## clean: remove build artifacts
clean:
	@rm -rf $(BUILD_DIR) coverage.out coverage.html
	@echo "Cleaned"

## release: build with GoReleaser
release:
	goreleaser release --clean

## release-snapshot: local test release (no publish)
release-snapshot:
	goreleaser release --snapshot --clean

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
