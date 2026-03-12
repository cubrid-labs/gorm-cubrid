# GORM CUBRID Driver

CUBRID database driver for [GORM](https://gorm.io/).

## Requirements

- Go 1.21+
- CUBRID 11.2+ (or CCI 11.0.0+)
- CUBRID CCI library installed on the system

## Installation

```bash
go get github.com/cubrid-labs/gorm-cubrid
```

The CUBRID Go SQL driver must also be installed:

```bash
go get github.com/CUBRID/cubrid-go
```

> **Note:** `github.com/CUBRID/cubrid-go` uses CGO and requires the CUBRID CCI
> library. See the [cubrid-go README](https://github.com/CUBRID/cubrid-go) for
> installation instructions.

## Usage

```go
package main

import (
    _ "github.com/CUBRID/cubrid-go" // register CUBRID SQL driver

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

## Notes

- CUBRID does not support **unsigned** integer types. `uint` fields are mapped
  to their signed equivalents of the next larger size.
- Schema introspection (`HasTable`, `HasColumn`, `ColumnTypes`, etc.) requires
  `INFORMATION_SCHEMA` support, available in CUBRID 11.2+.
- `AUTO_INCREMENT` columns use CUBRID's native `AUTO_INCREMENT` attribute.

## License

MIT
