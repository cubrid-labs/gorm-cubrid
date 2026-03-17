# GORM CUBRID Driver

<!-- BADGES:START -->
[![Go Reference](https://pkg.go.dev/badge/github.com/cubrid-labs/gorm-cubrid.svg)](https://pkg.go.dev/github.com/cubrid-labs/gorm-cubrid)
[![v0.1.0](https://img.shields.io/badge/version-v0.1.0-blue.svg)](https://github.com/cubrid-labs/gorm-cubrid/releases)
[![Go 1.21+](https://img.shields.io/badge/go-1.21%2B-00ADD8.svg)](https://go.dev)
[![ci workflow](https://github.com/cubrid-labs/gorm-cubrid/actions/workflows/ci.yml/badge.svg)](https://github.com/cubrid-labs/gorm-cubrid/actions/workflows/ci.yml)
[![license](https://img.shields.io/github/license/cubrid-labs/gorm-cubrid)](https://github.com/cubrid-labs/gorm-cubrid/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/cubrid-labs/gorm-cubrid)](https://github.com/cubrid-labs/gorm-cubrid)
<!-- BADGES:END -->


CUBRID database driver for [GORM](https://gorm.io/).

## Requirements

- Go 1.21+
- CUBRID 10.2+ database server

## Installation

```bash
go get github.com/cubrid-labs/gorm-cubrid
```

The CUBRID Go SQL driver must also be installed:

```bash
go get github.com/cubrid-labs/cubrid-go
```

`cubrid-go` is a pure Go driver — no CGO or native libraries required.

## Usage

```go
package main

import (
    _ "github.com/cubrid-labs/cubrid-go" // register CUBRID SQL driver

    cubrid "github.com/cubrid-labs/gorm-cubrid"
    "gorm.io/gorm"
)

func main() {
    // DSN format: cci:CUBRID:<host>:<port>:<dbname>:<user>:<password>:
    dsn := "cci:CUBRID:localhost:33000:demodb:public::"

    db, err := gorm.Open(cubrid.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    // AutoMigrate example
    type Product struct {
        gorm.Model
        Name  string
        Price float64
    }
    db.AutoMigrate(&Product{})
}
```

### Advanced Configuration

```go
db, err := gorm.Open(cubrid.New(cubrid.Config{
    DSN:               "cci:CUBRID:localhost:33000:demodb:public::",
    DefaultStringSize: 512, // default VARCHAR size (default: 256)
}), &gorm.Config{})
```

### Using an Existing Connection

```go
import "database/sql"

sqlDB, err := sql.Open("cubrid", "cci:CUBRID:localhost:33000:demodb:public::")
if err != nil {
    panic(err)
}

db, err := gorm.Open(cubrid.New(cubrid.Config{Conn: sqlDB}), &gorm.Config{})
```

## Data Type Mapping

| Go / GORM Type | CUBRID Type |
|---|---|
| `bool` | `TINYINT(1)` |
| `int8`, `uint8` | `TINYINT` |
| `int16`, `uint16` | `SMALLINT` |
| `int32`, `uint32` | `INT` |
| `int64`, `uint64` | `BIGINT` |
| `float32` | `FLOAT` |
| `float64` | `DOUBLE` |
| `float` with precision | `NUMERIC(p, s)` |
| `string` (size < 65536) | `VARCHAR(n)` |
| `string` (size ≥ 65536) | `CLOB` |
| `time.Time` | `DATETIME` |
| `[]byte` | `BLOB` |

## DSN Format

```
cci:CUBRID:<host>:<port>:<dbname>:<user>:<password>:
```

| Field | Description | Example |
|---|---|---|
| `host` | CUBRID server hostname or IP | `localhost` |
| `port` | CUBRID broker port | `33000` |
| `dbname` | Database name | `demodb` |
| `user` | Database user | `public` |
| `password` | User password (empty if none) | `` |

Example:
```
cci:CUBRID:localhost:33000:demodb:dba:password:
```

## Supported Features

- `AutoMigrate` — create and alter tables automatically
- `CreateTable` / `DropTable`
- `HasTable` / `HasColumn` / `HasIndex`
- `GetTables` / `ColumnTypes`
- Standard GORM CRUD operations
- Associations (HasOne, HasMany, BelongsTo, ManyToMany)
- Transactions
- Hooks

## AutoMigrate with CUBRID

Basic usage:

```go
type User struct {
    ID   uint
    Name string
}

type Product struct {
    ID    uint
    Name  string
    Price float64
}

if err := db.AutoMigrate(&User{}, &Product{}); err != nil {
    panic(err)
}
```

Best practices and CUBRID-specific notes:

- Treat `bool` as an integer-backed type (`SMALLINT` semantics; emitted as `TINYINT(1)`), because CUBRID has no native `BOOLEAN` type.
- CUBRID has no native JSON type. Store JSON payloads as `VARCHAR`/`CLOB` and validate in the application layer.
- CUBRID DDL statements auto-commit. Schema changes from `AutoMigrate` cannot be rolled back inside a transaction.
- Keep index definitions explicit (`gorm:"index"`, `gorm:"uniqueIndex"`) and review generated indexes for large tables before production rollout.
- `AutoMigrate` handles table creation and additive changes well, but complex column type changes and some `ALTER TABLE` patterns may require manual SQL.
- For production-critical systems, prefer versioned/manual migration files (or a dedicated migration tool) for deterministic review and rollback planning.

## Soft Delete Support

GORM soft delete works with CUBRID when a model includes `gorm.DeletedAt`.

```go
import "gorm.io/gorm"

type User struct {
    ID        uint
    Name      string
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Soft delete: sets deleted_at instead of physical delete.
db.Delete(&user)

// Normal queries automatically exclude rows where deleted_at is not NULL.
var users []User
db.Find(&users)

// Include soft-deleted rows.
db.Unscoped().Find(&users)

// Permanently delete rows.
db.Unscoped().Delete(&user)
```

Notes:

- This works because CUBRID supports `DATETIME` columns and `WHERE ... IS NULL` filters used by GORM soft delete.
- Soft delete in GORM is application-level behavior (query/update conventions), not a database-native soft-delete feature.

## Notes

- CUBRID does not support **unsigned** integer types. `uint` fields are mapped
  to their signed equivalents of the next larger size.
- Schema introspection (`HasTable`, `HasColumn`, `ColumnTypes`, etc.) requires
  `INFORMATION_SCHEMA` support, available in CUBRID 11.2+.
- `AUTO_INCREMENT` columns use CUBRID's native `AUTO_INCREMENT` attribute.

## Benchmark

Performance benchmarks comparing CUBRID drivers against MySQL are tracked in the [cubrid-benchmark](https://github.com/cubrid-labs/cubrid-benchmark) suite.

[![Benchmark](https://github.com/cubrid-labs/cubrid-benchmark/actions/workflows/bench.yml/badge.svg)](https://cubrid-labs.github.io/cubrid-benchmark/)

- **Tier 0** — Functional smoke tests (connect + CRUD)
- **Tier 1** — Driver throughput: 10K INSERT/SELECT, 1K UPDATE/DELETE
- Same schema, same seed data, same CI hardware per run
- Results published to [GitHub Pages dashboard](https://cubrid-labs.github.io/cubrid-benchmark/)


## Ecosystem

| Package | Description | Language |
|:--------|:------------|:---------|
| [cubrid-go](https://github.com/cubrid-labs/cubrid-go) | database/sql driver + GORM dialector | Go |
| [pycubrid](https://github.com/cubrid-labs/pycubrid) | Python DB-API 2.0 driver | Python |
| [sqlalchemy-cubrid](https://github.com/cubrid-labs/sqlalchemy-cubrid) | SQLAlchemy 2.0 dialect | Python |
| [cubrid-client](https://github.com/cubrid-labs/cubrid-client) | TypeScript CAS client | TypeScript |
| [drizzle-cubrid](https://github.com/cubrid-labs/drizzle-cubrid) | Drizzle ORM dialect | TypeScript |
| [cubrid-rs](https://github.com/cubrid-labs/cubrid-rs) | Native Rust driver (sync + async) | Rust |
| [sea-orm-cubrid](https://github.com/cubrid-labs/sea-orm-cubrid) | SeaORM backend for CUBRID | Rust |
| [cubrid-cookbook](https://github.com/cubrid-labs/cubrid-cookbook) | Practical examples across ecosystems | Multi |
| [cubrid-benchmark](https://github.com/cubrid-labs/cubrid-benchmark) | Multi-language benchmark suite | Multi |

## License

MIT
