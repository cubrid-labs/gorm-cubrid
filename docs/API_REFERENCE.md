# API Reference

Complete API documentation for the GORM CUBRID dialect (`gorm-cubrid`).

## Architecture

```mermaid
flowchart TD
    App[Application Code] --> GORM[GORM Framework]
    GORM --> Dialector[gorm-cubrid Dialector]
    Dialector --> Driver[cubrid-go Driver]
    Driver --> CAS[CAS Protocol]
    CAS --> CUBRID[(CUBRID Database)]

    subgraph "gorm-cubrid"
        Dialector --> Migrator[Migrator]
        Dialector --> DataTypeOf[DataTypeOf]
        Dialector --> QuoteTo[QuoteTo]
        Dialector --> BindVarTo[BindVarTo]
    end
```

## Quick Start

```go
import (
    "gorm.io/gorm"
    cubrid "github.com/cubrid-labs/gorm-cubrid"
    _ "github.com/cubrid-labs/cubrid-go"
)

db, err := gorm.Open(cubrid.Open("cubrid://dba@localhost:33000/demodb"), &gorm.Config{})
```

## Types

### `Config`

Configuration struct for the CUBRID dialect.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DriverName` | `string` | `"cubrid"` | SQL driver name. Override for custom-registered drivers. |
| `DSN` | `string` | — | Data source name. Format: `cubrid://[user[:password]]@host[:port]/database` |
| `Conn` | `gorm.ConnPool` | `nil` | Existing connection pool. When set, `DSN` is ignored. |
| `DefaultStringSize` | `uint` | `256` | Default VARCHAR length when no explicit size is provided. |
| `SkipPing` | `bool` | `false` | Skip `Ping()` on initialization. Useful for lazy connections. |

### `Dialector`

Implements `gorm.Dialector` for CUBRID databases.

```go
type Dialector struct {
    *Config
}
```

### `Migrator`

CUBRID-specific migrator for schema management. Embeds `gorm.Migrator` with CUBRID customizations.

## Functions

### `Open(dsn string) gorm.Dialector`

Creates a new CUBRID Dialector from a DSN string.

```go
db, err := gorm.Open(cubrid.Open("cubrid://dba@localhost:33000/demodb"), &gorm.Config{})
```

**DSN format**: `cubrid://[user[:password]]@host[:port]/database[?autocommit=true&timeout=30s]`

### `New(config Config) gorm.Dialector`

Creates a new CUBRID Dialector from a `Config` struct. Use this for advanced configuration.

```go
db, err := gorm.Open(cubrid.New(cubrid.Config{
    DSN:               "cubrid://dba@localhost:33000/demodb",
    DefaultStringSize: 512,
    SkipPing:          true,
}), &gorm.Config{})
```

## Dialector Methods

### `Name() string`

Returns `"cubrid"`.

### `Initialize(db *gorm.DB) error`

Sets up the connection pool and registers GORM callbacks. Opens a new connection via `DSN` unless `Config.Conn` is provided. Pings the database unless `SkipPing` is true.

### `DataTypeOf(field *schema.Field) string`

Maps GORM schema types to CUBRID SQL types:

| GORM Type | CUBRID SQL Type | Notes |
|-----------|----------------|-------|
| `Bool` | `TINYINT(1)` | CUBRID has no native boolean |
| `Int` (≤8 bit) | `TINYINT` | |
| `Int` (≤16 bit) | `SMALLINT` | |
| `Int` (≤32 bit) | `INT` | |
| `Int` (>32 bit) | `BIGINT` | |
| `Uint` | Same as `Int` | CUBRID has no unsigned types |
| `Float` (≤32 bit) | `FLOAT` | |
| `Float` (>32 bit) | `DOUBLE` | |
| `Float` (precision>0) | `NUMERIC(p,s)` | |
| `String` (size<65536) | `VARCHAR(n)` | Default size: 256 |
| `String` (size≥65536) | `CLOB` | |
| `Time` | `DATETIME` | |
| `Bytes` | `BLOB` | |

### `Migrator(db *gorm.DB) gorm.Migrator`

Returns a CUBRID-specific migrator. Creates indexes after table creation (`CreateIndexAfterCreateTable: true`).

### `QuoteTo(writer clause.Writer, str string)`

Writes backtick-quoted identifiers. Handles dot-separated `schema.table` names with proper escaping.

### `BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{})`

Writes `?` positional placeholder (MySQL-style).

### `DefaultValueOf(field *schema.Field) clause.Expression`

Returns `DEFAULT` keyword expression.

### `Explain(sql string, vars ...interface{}) string`

Formats SQL with bind variables substituted for debugging/logging.

## Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `DefaultStringSize` | `256` | Default VARCHAR length |

## Connection Examples

```go
// Basic connection
db, err := gorm.Open(cubrid.Open("cubrid://dba@localhost:33000/demodb"), &gorm.Config{})

// With password
db, err := gorm.Open(cubrid.Open("cubrid://dba:secret@localhost:33000/demodb"), &gorm.Config{})

// With query parameters
db, err := gorm.Open(cubrid.Open("cubrid://dba@localhost:33000/demodb?autocommit=true&timeout=30s"), &gorm.Config{})

// Using existing connection pool
sqlDB, _ := sql.Open("cubrid", "cubrid://dba@localhost:33000/demodb")
db, err := gorm.Open(cubrid.New(cubrid.Config{Conn: sqlDB}), &gorm.Config{})
```

## CRUD Operations

```go
// Create
db.Create(&User{Name: "Alice", Age: 30})

// Read
var user User
db.First(&user, 1)
db.Where("name = ?", "Alice").First(&user)

// Update
db.Model(&user).Update("age", 31)

// Delete
db.Delete(&user)
```

## Auto-Migration

```go
db.AutoMigrate(&User{}, &Product{}, &Order{})
```

## Limitations

- No `RETURNING` clause support
- No JSON data type
- No native boolean (uses `TINYINT(1)`)
- No unsigned integer types
- DDL auto-commits (not transactional)
- No sequence/serial — uses `AUTO_INCREMENT`
