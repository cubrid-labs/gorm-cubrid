.PHONY: all build test test-verbose test-integration vet fmt lint tidy clean security check-all changelog clean-all doctor

# Default target
all: build test

## Build
build:
	go build ./...

## Unit tests (no CUBRID server required)
test:
	go test -race -count=1 ./...

test-verbose:
	go test -v -race -count=1 ./...

## Integration tests (requires CUBRID 11.2+ running on localhost:33000)
## Set CUBRID_DSN to override the default connection string.
CUBRID_DSN ?= cci:CUBRID:localhost:33000:demodb:dba::
test-integration:
	CUBRID_DSN=$(CUBRID_DSN) go test -tags integration -race -count=1 ./...

## Coverage report (opens in browser on macOS/Linux)
coverage:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## Code quality
vet:
	go vet ./...

fmt:
	gofmt -w .

lint: vet

## Dependency management
tidy:
	go mod tidy

## Clean generated files
clean:
	rm -f coverage.out

# Security scan
security:
	@which govulncheck > /dev/null 2>&1 && govulncheck ./... || echo "govulncheck not installed, skipping"

# Full check (all quality gates)
check-all: fmt vet lint security test

# Generate changelog (requires git-cliff)
changelog:
	@which git-cliff > /dev/null 2>&1 && git-cliff -o CHANGELOG.md || echo "git-cliff not installed, skipping"

# Clean all artifacts
clean-all: clean
	rm -rf dist/ bin/

# Doctor check (verify tool availability)
doctor:
	@echo "=== Go Environment ==="
	@go version
	@echo ""
	@echo "=== Tools ==="
	@which golangci-lint > /dev/null 2>&1 && echo "✓ golangci-lint" || echo "✗ golangci-lint (optional)"
	@which govulncheck > /dev/null 2>&1 && echo "✓ govulncheck" || echo "✗ govulncheck (optional: go install golang.org/x/vuln/cmd/govulncheck@latest)"
	@which git-cliff > /dev/null 2>&1 && echo "✓ git-cliff" || echo "✗ git-cliff (optional)"
