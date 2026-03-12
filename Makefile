.PHONY: all build test test-verbose test-integration vet fmt lint tidy clean

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
