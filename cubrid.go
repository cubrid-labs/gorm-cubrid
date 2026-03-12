// Package cubrid implements a GORM dialector for CUBRID database.
//
// # Usage
//
// Users must import the CUBRID SQL driver separately:
//
//	import _ "github.com/CUBRID/cubrid-go"
//
// Then open a connection using the CUBRID DSN format:
//
//	cci:CUBRID:<host>:<port>:<dbname>:<user>:<password>:
//
// Example:
//
//	db, err := gorm.Open(cubrid.Open("cci:CUBRID:localhost:33000:demodb:public::"), &gorm.Config{})
package cubrid

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

// DefaultStringSize is the default length used for string (VARCHAR) fields
// when no explicit size is provided.
const DefaultStringSize uint = 256

// Config holds CUBRID-specific configuration for the Dialector.
type Config struct {
	// DriverName overrides the default SQL driver name ("cubrid").
	// Useful when using a custom-registered driver.
	DriverName string

	// DSN is the CUBRID data source name.
	// Format: cci:CUBRID:<host>:<port>:<dbname>:<user>:<password>:
	// Example: cci:CUBRID:localhost:33000:demodb:public::
	DSN string

	// Conn allows providing an existing *sql.DB connection pool instead of
	// opening a new connection via DSN.
	Conn gorm.ConnPool

	// DefaultStringSize sets the default length for VARCHAR fields.
	// Defaults to 256 if not set.
	DefaultStringSize uint
}

// Dialector implements gorm.Dialector for CUBRID databases.
type Dialector struct {
	*Config
}

// Open creates a new CUBRID Dialector from a DSN string.
//
// DSN format: cci:CUBRID:<host>:<port>:<dbname>:<user>:<password>:
//
// Note: The CUBRID SQL driver must be imported separately:
//
//	import _ "github.com/CUBRID/cubrid-go"
func Open(dsn string) gorm.Dialector {
	return &Dialector{Config: &Config{DSN: dsn}}
}

// New creates a new CUBRID Dialector from a Config struct.
func New(config Config) gorm.Dialector {
	return &Dialector{Config: &config}
}

// Name returns the dialect name ("cubrid").
func (dialector Dialector) Name() string {
	return "cubrid"
}

// Initialize sets up the CUBRID connection pool and registers GORM callbacks.
// If Config.Conn is provided, it is used directly; otherwise a new connection
// is opened using Config.DSN.
func (dialector Dialector) Initialize(db *gorm.DB) (err error) {
	if dialector.DriverName == "" {
		dialector.DriverName = "cubrid"
	}

	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else {
		db.ConnPool, err = sql.Open(dialector.DriverName, dialector.DSN)
		if err != nil {
			return err
		}
	}

	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{})
	return nil
}

// Migrator returns a CUBRID-specific Migrator for schema management.
func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:                          db,
				Dialector:                   dialector,
				CreateIndexAfterCreateTable: true,
			},
		},
	}
}

// DataTypeOf returns the CUBRID SQL data type for the given schema field.
//
// Supported GORM schema types and their CUBRID equivalents:
//   - Bool      → TINYINT(1)
//   - Int/Uint  → TINYINT / SMALLINT / INT / BIGINT (by size)
//   - Float     → FLOAT / DOUBLE / NUMERIC(p,s)
//   - String    → VARCHAR(n) / CLOB
//   - Time      → DATETIME
//   - Bytes     → BLOB
func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "tinyint(1)"
	case schema.Int, schema.Uint:
		return dialector.intSQLType(field)
	case schema.Float:
		return dialector.floatSQLType(field)
	case schema.String:
		return dialector.stringSQLType(field)
	case schema.Time:
		return "datetime"
	case schema.Bytes:
		return "blob"
	default:
		sqlType := string(field.DataType)
		if field.AutoIncrement && !strings.Contains(strings.ToLower(sqlType), "auto_increment") {
			sqlType += " auto_increment"
		}
		return sqlType
	}
}

func (dialector Dialector) intSQLType(field *schema.Field) string {
	var sqlType string
	switch {
	case field.Size <= 8:
		sqlType = "tinyint"
	case field.Size <= 16:
		sqlType = "smallint"
	case field.Size <= 32:
		sqlType = "int"
	default:
		sqlType = "bigint"
	}
	if field.AutoIncrement {
		sqlType += " auto_increment"
	}
	return sqlType
}

func (dialector Dialector) floatSQLType(field *schema.Field) string {
	if field.Precision > 0 {
		return fmt.Sprintf("numeric(%d, %d)", field.Precision, field.Scale)
	}
	if field.Size <= 32 {
		return "float"
	}
	return "double"
}

func (dialector Dialector) stringSQLType(field *schema.Field) string {
	size := field.Size
	if size == 0 {
		defaultSize := dialector.DefaultStringSize
		if defaultSize == 0 {
			defaultSize = DefaultStringSize
		}
		size = int(defaultSize)
	}
	if size >= 65536 {
		return "clob"
	}
	return fmt.Sprintf("varchar(%d)", size)
}

// DefaultValueOf returns the DEFAULT keyword expression for a field's default value.
func (dialector Dialector) DefaultValueOf(_ *schema.Field) clause.Expression {
	return clause.Expr{SQL: "DEFAULT"}
}

// BindVarTo writes a "?" positional bind variable placeholder.
// CUBRID uses MySQL-style "?" placeholders.
func (dialector Dialector) BindVarTo(writer clause.Writer, _ *gorm.Statement, _ interface{}) {
	writer.WriteByte('?')
}

// QuoteTo writes a backtick-quoted identifier to writer.
// Handles dot-separated schema.table identifiers and backtick escaping.
func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	writer.WriteByte('`')
	if strings.Contains(str, ".") {
		parts := strings.SplitN(str, ".", 2)
		writer.WriteString(strings.ReplaceAll(parts[0], "`", "``"))
		writer.WriteString("`.`")
		writer.WriteString(strings.ReplaceAll(parts[1], "`", "``"))
	} else {
		writer.WriteString(strings.ReplaceAll(str, "`", "``"))
	}
	writer.WriteByte('`')
}

// Explain formats an SQL statement with its bind variables substituted,
// for use in logging and debugging.
func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `'`, vars...)
}
