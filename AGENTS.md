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

## Development Workflow (cubrid-labs org standard)

All non-trivial work across cubrid-labs repositories MUST follow this 4-phase cycle:

1. **Oracle Design Review** — Consult Oracle before implementation to validate architecture, API surface, and approach. Raise concerns early.
2. **Implementation** — Build the feature/fix with tests. Follow existing codebase patterns.
3. **Documentation Update** — Update ALL affected docs (README, CHANGELOG, ROADMAP, API docs, SUPPORT_MATRIX, PRD, etc.) in the same PR or as an immediate follow-up. Code without doc updates is incomplete.
4. **Oracle Post-Implementation Review** — Consult Oracle to review the completed work for correctness, edge cases, and consistency before merging.

Skipping any phase requires explicit justification. Trivial changes (typos, single-line fixes) may skip phases 1 and 4.

## Validation
- `go test ./...`
- `make vet`
- `make test-integration` when integration paths are touched
