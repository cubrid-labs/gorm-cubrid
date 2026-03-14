# AGENTS.md

## Purpose
`gorm-cubrid` is a GORM dialector and driver integration for CUBRID.

## Read First
- `README.md`
- `CONTRIBUTING.md`
- `docs/agent-playbook.md`

## Working Rules
- Preserve GORM conventions and keep examples aligned with actual dialect behavior.
- Minimize breaking changes to migrator, clause, and DSN behavior.
- Update docs when data type mappings or supported features change.
- Add tests for dialect behavior changes, especially schema and SQL generation.

## Validation
- `go test ./...`
- `make vet`
- `make test-integration` when integration paths are touched
