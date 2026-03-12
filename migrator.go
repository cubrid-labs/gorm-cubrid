package cubrid

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
)

// Migrator implements gorm.Migrator for CUBRID databases.
// It embeds migrator.Migrator (GORM's common base implementation) and overrides
// methods that require CUBRID-specific SQL.
//
// Schema introspection relies on INFORMATION_SCHEMA (available in CUBRID 11.2+)
// and CUBRID system catalog tables (db_class, db_index).
type Migrator struct {
	migrator.Migrator
}

// CurrentDatabase returns the name of the currently selected database.
func (m Migrator) CurrentDatabase() string {
	var name string
	m.DB.Raw("SELECT database()").Row().Scan(&name)
	return name
}

// GetTables returns all user (non-system) table names in the current database.
func (m Migrator) GetTables() (tableList []string, err error) {
	err = m.DB.Raw(
		"SELECT table_name FROM information_schema.tables "+
			"WHERE table_schema = ? AND table_type = 'BASE TABLE' "+
			"ORDER BY table_name",
		m.CurrentDatabase(),
	).Scan(&tableList).Error
	return
}

// HasTable reports whether the table for value exists in the database.
// value may be a struct, pointer to struct, or table name string.
func (m Migrator) HasTable(value interface{}) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.DB.Raw(
			"SELECT count(*) FROM information_schema.tables "+
				"WHERE table_schema = ? AND table_name = ? AND table_type = 'BASE TABLE'",
			m.CurrentDatabase(),
			strings.ToLower(stmt.Table),
		).Row().Scan(&count)
	})
	return count > 0
}

// HasColumn reports whether the column field exists in the table for value.
// field may be the Go struct field name or the DB column name.
func (m Migrator) HasColumn(value interface{}, field string) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		columnName := field
		if stmt.Schema != nil {
			if f := stmt.Schema.LookUpField(field); f != nil {
				columnName = f.DBName
			}
		}
		return m.DB.Raw(
			"SELECT count(*) FROM information_schema.columns "+
				"WHERE table_schema = ? AND table_name = ? AND column_name = ?",
			m.CurrentDatabase(),
			strings.ToLower(stmt.Table),
			columnName,
		).Row().Scan(&count)
	})
	return count > 0
}

// HasIndex reports whether an index with the given name exists on the table for value.
func (m Migrator) HasIndex(value interface{}, name string) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}
		// db_index is available in all CUBRID versions.
		return m.DB.Raw(
			"SELECT count(*) FROM db_index WHERE class_name = ? AND index_name = ?",
			strings.ToLower(stmt.Table),
			name,
		).Row().Scan(&count)
	})
	return count > 0
}

// RenameTable renames a table.
// CUBRID uses "RENAME TABLE old TO new" rather than "ALTER TABLE old RENAME TO new".
func (m Migrator) RenameTable(oldName, newName interface{}) error {
	return m.RunWithValue(oldName, func(oldStmt *gorm.Statement) error {
		return m.RunWithValue(newName, func(newStmt *gorm.Statement) error {
			return m.DB.Exec(
				"RENAME TABLE ? TO ?",
				m.CurrentTable(oldStmt),
				clause.Table{Name: newStmt.Table},
			).Error
		})
	})
}

// DropIndex drops the named index from the table associated with value.
// CUBRID requires "DROP INDEX idx ON table", unlike the base "DROP INDEX idx".
func (m Migrator) DropIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}
		return m.DB.Exec(
			"DROP INDEX ? ON ?",
			clause.Column{Name: name},
			m.CurrentTable(stmt),
		).Error
	})
}

// AlterColumn changes an existing column's definition.
// CUBRID uses "ALTER TABLE t MODIFY COLUMN col type" instead of the PostgreSQL-style
// "ALTER TABLE t ALTER COLUMN col TYPE type" that the base migrator generates.
func (m Migrator) AlterColumn(value interface{}, field string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if f := stmt.Schema.LookUpField(field); f != nil {
			fileType := m.DB.Migrator().FullDataTypeOf(f)
			return m.DB.Exec(
				"ALTER TABLE ? MODIFY COLUMN ? ?",
				m.CurrentTable(stmt),
				clause.Column{Name: f.DBName},
				fileType,
			).Error
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	})
}

// ColumnTypes returns detailed column type information for the table
// associated with value, in ordinal position order.
// Requires CUBRID 11.2+ (INFORMATION_SCHEMA support).
func (m Migrator) ColumnTypes(value interface{}) (columnTypes []gorm.ColumnType, err error) {
	currentDatabase := m.CurrentDatabase()
	err = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		const query = `
			SELECT
				column_name,
				data_type,
				is_nullable,
				column_default,
				character_maximum_length,
				numeric_precision,
				numeric_scale,
				column_key,
				extra
			FROM information_schema.columns
			WHERE table_schema = ? AND table_name = ?
			ORDER BY ordinal_position`

		rows, rowErr := m.DB.Raw(query, currentDatabase, strings.ToLower(stmt.Table)).Rows()
		if rowErr != nil {
			return rowErr
		}
		defer rows.Close()

		for rows.Next() {
			ct, scanErr := scanColumnType(rows)
			if scanErr != nil {
				return scanErr
			}
			columnTypes = append(columnTypes, ct)
		}
		return rows.Err()
	})
	return
}

// scanColumnType reads a single row from INFORMATION_SCHEMA.COLUMNS and
// returns a populated *migrator.ColumnType.
func scanColumnType(rows *sql.Rows) (*migrator.ColumnType, error) {
	var (
		ct           = &migrator.ColumnType{}
		name         string
		dataType     string
		isNullable   string
		defaultValue sql.NullString
		charMaxLen   sql.NullInt64
		numPrecision sql.NullInt64
		numScale     sql.NullInt64
		columnKey    sql.NullString
		extra        sql.NullString
	)

	if err := rows.Scan(
		&name, &dataType, &isNullable, &defaultValue,
		&charMaxLen, &numPrecision, &numScale, &columnKey, &extra,
	); err != nil {
		return nil, err
	}

	ct.NameValue = sql.NullString{String: name, Valid: true}
	ct.DataTypeValue = sql.NullString{String: dataType, Valid: true}
	ct.NullableValue = sql.NullBool{Bool: strings.EqualFold(isNullable, "YES"), Valid: true}

	if columnKey.Valid {
		ct.PrimaryKeyValue = sql.NullBool{Bool: columnKey.String == "PRI", Valid: true}
		ct.UniqueValue = sql.NullBool{Bool: columnKey.String == "UNI", Valid: true}
	}

	if extra.Valid && strings.Contains(strings.ToLower(extra.String), "auto_increment") {
		ct.AutoIncrementValue = sql.NullBool{Bool: true, Valid: true}
	}

	if defaultValue.Valid {
		ct.DefaultValueValue = defaultValue
	}

	if charMaxLen.Valid {
		ct.LengthValue = charMaxLen
	} else if numPrecision.Valid {
		ct.LengthValue = numPrecision
		ct.DecimalSizeValue = numPrecision
		if numScale.Valid {
			ct.ScaleValue = numScale
		}
	}

	ct.ColumnTypeValue = sql.NullString{
		String: buildColumnType(dataType, charMaxLen, numPrecision, numScale),
		Valid:  true,
	}

	return ct, nil
}

// buildColumnType constructs the full SQL column type string, e.g. "varchar(100)"
// or "numeric(10,2)", from INFORMATION_SCHEMA component columns.
func buildColumnType(dataType string, charMaxLen, numPrecision, numScale sql.NullInt64) string {
	switch strings.ToLower(dataType) {
	case "varchar", "char", "nchar", "varnchar":
		if charMaxLen.Valid {
			return fmt.Sprintf("%s(%d)", dataType, charMaxLen.Int64)
		}
	case "numeric", "decimal":
		if numPrecision.Valid && numScale.Valid {
			return fmt.Sprintf("%s(%d,%d)", dataType, numPrecision.Int64, numScale.Int64)
		} else if numPrecision.Valid {
			return fmt.Sprintf("%s(%d)", dataType, numPrecision.Int64)
		}
	}
	return dataType
}
