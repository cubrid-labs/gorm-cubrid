# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-12

### Added

- `Open(dsn string)` constructor — creates a Dialector from a CUBRID DSN string
- `New(config Config)` constructor — creates a Dialector from a `Config` struct,
  supporting custom driver name, existing `*sql.DB` connection pool, and
  default string size override
- `gorm.Dialector` interface implementation for CUBRID 11.2+:
  - `DataTypeOf` — maps GORM schema types to CUBRID SQL types
    (`TINYINT`, `SMALLINT`, `INT`, `BIGINT`, `FLOAT`, `DOUBLE`, `NUMERIC`,
    `VARCHAR`, `CLOB`, `DATETIME`, `BLOB`)
  - `BindVarTo` — MySQL-compatible `?` positional placeholders
  - `QuoteTo` — backtick identifier quoting with dot-notation and escape support
  - `Explain` — variable substitution for SQL logging
  - `DefaultValueOf` — `DEFAULT` keyword expression
- `Migrator` — CUBRID-specific schema management:
  - `CurrentDatabase` — queries current database name via `SELECT database()`
  - `GetTables` / `HasTable` — `INFORMATION_SCHEMA.TABLES`-based introspection
  - `HasColumn` — `INFORMATION_SCHEMA.COLUMNS`-based introspection
  - `HasIndex` — `db_index` system catalog-based introspection
  - `ColumnTypes` — full column metadata from `INFORMATION_SCHEMA.COLUMNS`
  - `RenameTable` — CUBRID-native `RENAME TABLE old TO new`
  - `DropIndex` — CUBRID-required `DROP INDEX idx ON table` syntax
  - `AlterColumn` — CUBRID-native `ALTER TABLE t MODIFY COLUMN col type`
- Connection validation via `Ping()` on `Initialize`
- Unit test suite (38 tests, no CUBRID server required)

[Unreleased]: https://github.com/cubrid-labs/gorm-cubrid/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/cubrid-labs/gorm-cubrid/releases/tag/v0.1.0
