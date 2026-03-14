# Agent Playbook

## Source Of Truth
- `README.md` for public usage and supported features.
- `CONTRIBUTING.md` for development and integration test workflow.
- `go.mod` and `Makefile` for language/runtime expectations.

## Repository Map
- Root Go packages implement the GORM dialector.
- `docs/` holds supplemental agent guidance.
- Tests live alongside the Go packages.

## Change Workflow
1. Check whether the change affects configuration, SQL generation, migration, or docs only.
2. Keep examples in README current when public usage changes.
3. Preserve default DSN assumptions unless the change explicitly updates the contract.
4. Run integration coverage when touching connection or migrator paths and a local CUBRID server is available.

## Validation
- `go test ./...`
- `make vet`
- `make test-integration`
