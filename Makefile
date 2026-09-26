.PHONY: help build test lint fmt vet clean docker

.DEFAULT_GOAL := help

BINARY_NAME := vault-tool
VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell LC_ALL=C date -u +%Y-%m-%dT%H:%M:%SZ)

help:
	@echo "Makefile targets:"
	@echo "  make build           - Build the binary"
	@echo "  make test            - Run tests"
	@echo "  make lint            - Run linters"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Run go vet"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make deb             - Build Debian package"

build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	go build -ldflags "-X github.com/rossigee/vault-tool/internal/version.Version=$(VERSION) -X github.com/rossigee/vault-tool/internal/version.Commit=$(COMMIT) -X github.com/rossigee/vault-tool/internal/version.BuildTime=$(BUILD_TIME)" -o $(BINARY_NAME) .
	@echo "✅ Build complete: $(BINARY_NAME)"

test:
	@echo "Running tests..."
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	@echo "✅ Tests passed"

lint:
	@echo "Running linters..."
	golangci-lint run ./... --config .golangci.yml
	@echo "✅ Linting passed"

fmt:
	@echo "Formatting code..."
	gofmt -w .
	@echo "✅ Formatting complete"

vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✅ Vet passed"

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	@echo "✅ Clean complete"

deb:
	@echo "Building Debian package..."
	dpkg-buildpackage -us -uc -b
	@echo "✅ Debian package built"
