# Contributing to gorm-cubrid

Thank you for your interest in contributing!

## Requirements

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.21+ | Build and test |
| Git | any | Version control |
| CUBRID | 11.2+ | Integration tests only |
| Docker | any | Integration tests via `docker-compose` |

## Getting Started

```bash
git clone https://github.com/cubrid-labs/gorm-cubrid.git
cd gorm-cubrid
go mod download
```

## Running Tests

### Unit tests (no CUBRID server needed)

```bash
go test ./...
# or, with verbose output:
go test -v -race ./...
```

All offline tests must pass before submitting a pull request.
Target coverage: **90%+** for non-integration code.

### Integration tests (CUBRID server required)

Integration tests are tagged with `//go:build integration` and require a running
CUBRID 11.2+ instance.

```bash
# Start CUBRID with Docker Compose
docker-compose up -d

# Run integration tests
go test -tags integration ./...
```

Set the connection DSN via environment variable:

```bash
export CUBRID_DSN="cci:CUBRID:localhost:33000:demodb:dba::"
go test -tags integration ./...
```

## Code Style

- Format: `gofmt` / `goimports` (run before every commit)
- Lint: `go vet ./...`
- No external linters are required, but the CI gate enforces `go vet` clean output

```bash
make lint   # runs go vet
make fmt    # runs gofmt -w
```

## Commit Messages

Follow the conventional commits style used in this repository:

```
<type>: <short summary>

[optional body]
```

Types: `feat`, `fix`, `test`, `docs`, `refactor`, `chore`

Examples:
```
feat: add support for CUBRID ENUM type
fix: correct MODIFY COLUMN syntax for nullable columns
test: add integration test for AutoMigrate with associations
docs: update DSN format in README
```

## Pull Request Checklist

Before opening a PR, please verify:

- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` is clean
- [ ] `go test -race ./...` passes (all unit tests)
- [ ] New behaviour is covered by tests
- [ ] `CHANGELOG.md` is updated under `## [Unreleased]`
- [ ] If adding a CUBRID-specific DDL behaviour, the reason is documented
      in a code comment (CUBRID does not always match MySQL/PostgreSQL syntax)

## Reporting Issues

When reporting a bug, please include:

- Go version (`go version`)
- CUBRID server version
- GORM version (`go list -m gorm.io/gorm`)
- Minimal reproduction snippet
- Full error output or stack trace

## Architecture Notes

See [CLAUDE.md](CLAUDE.md) for a detailed breakdown of the package architecture,
CUBRID-specific gotchas, and overridden Migrator methods.
