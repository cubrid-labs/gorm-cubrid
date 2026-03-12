package cubrid

import (
	"database/sql"
	"strings"
	"testing"

	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// --- Dialector ---

func TestName(t *testing.T) {
	d := Open("cci:CUBRID:localhost:33000:demodb:public::")
	if d.Name() != "cubrid" {
		t.Errorf("Name() = %q, want %q", d.Name(), "cubrid")
	}
}

func TestOpen(t *testing.T) {
	d := Open("cci:CUBRID:localhost:33000:demodb:public::")
	if d == nil {
		t.Fatal("Open() returned nil")
	}
	if d.Name() != "cubrid" {
		t.Errorf("Name() = %q, want %q", d.Name(), "cubrid")
	}
}

func TestNew(t *testing.T) {
	d := New(Config{DSN: "cci:CUBRID:localhost:33000:demodb:public::"})
	if d == nil {
		t.Fatal("New() returned nil")
	}
	if d.Name() != "cubrid" {
		t.Errorf("Name() = %q, want %q", d.Name(), "cubrid")
	}
}

func TestNew_CustomDriverName(t *testing.T) {
	d := New(Config{DriverName: "my_cubrid", DSN: "cci:CUBRID:localhost:33000:demodb:public::"})
	inner := d.(*Dialector)
	if inner.DriverName != "my_cubrid" {
		t.Errorf("DriverName = %q, want %q", inner.DriverName, "my_cubrid")
	}
}

// --- DataTypeOf ---

func TestDataTypeOf(t *testing.T) {
	d := &Dialector{Config: &Config{}}

	tests := []struct {
		name     string
		field    *schema.Field
		expected string
	}{
		// Bool
		{"bool", &schema.Field{DataType: schema.Bool}, "tinyint(1)"},

		// Signed integers by bit size
		{"int8", &schema.Field{DataType: schema.Int, Size: 8}, "tinyint"},
		{"int16", &schema.Field{DataType: schema.Int, Size: 16}, "smallint"},
		{"int32", &schema.Field{DataType: schema.Int, Size: 32}, "int"},
		{"int64", &schema.Field{DataType: schema.Int, Size: 64}, "bigint"},

		// Unsigned integers: CUBRID has no unsigned types, mapped to signed equivalents
		{"uint8", &schema.Field{DataType: schema.Uint, Size: 8}, "tinyint"},
		{"uint16", &schema.Field{DataType: schema.Uint, Size: 16}, "smallint"},
		{"uint32", &schema.Field{DataType: schema.Uint, Size: 32}, "int"},
		{"uint64", &schema.Field{DataType: schema.Uint, Size: 64}, "bigint"},

		// AutoIncrement
		{"int_autoinc", &schema.Field{DataType: schema.Int, Size: 32, AutoIncrement: true}, "int auto_increment"},
		{"bigint_autoinc", &schema.Field{DataType: schema.Int, Size: 64, AutoIncrement: true}, "bigint auto_increment"},

		// Float
		{"float32", &schema.Field{DataType: schema.Float, Size: 32}, "float"},
		{"float64", &schema.Field{DataType: schema.Float, Size: 64}, "double"},
		{"numeric", &schema.Field{DataType: schema.Float, Precision: 10, Scale: 2}, "numeric(10, 2)"},

		// String
		{"string_default", &schema.Field{DataType: schema.String}, "varchar(256)"},
		{"string_100", &schema.Field{DataType: schema.String, Size: 100}, "varchar(100)"},
		{"string_at_limit", &schema.Field{DataType: schema.String, Size: 65535}, "varchar(65535)"},
		{"string_clob", &schema.Field{DataType: schema.String, Size: 65536}, "clob"},
		{"string_clob_large", &schema.Field{DataType: schema.String, Size: 1000000}, "clob"},

		// Time and Bytes
		{"time", &schema.Field{DataType: schema.Time}, "datetime"},
		{"bytes", &schema.Field{DataType: schema.Bytes}, "blob"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.DataTypeOf(tt.field)
			if got != tt.expected {
				t.Errorf("DataTypeOf() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDataTypeOf_CustomDefaultStringSize(t *testing.T) {
	d := &Dialector{Config: &Config{DefaultStringSize: 512}}
	got := d.DataTypeOf(&schema.Field{DataType: schema.String})
	if got != "varchar(512)" {
		t.Errorf("DataTypeOf() = %q, want %q", got, "varchar(512)")
	}
}

func TestDataTypeOf_CustomType(t *testing.T) {
	d := &Dialector{Config: &Config{}}
	// A custom database type tag (gorm:"type:json") passes through as-is.
	got := d.DataTypeOf(&schema.Field{DataType: "json"})
	if got != "json" {
		t.Errorf("DataTypeOf(json) = %q, want %q", got, "json")
	}
}

func TestDataTypeOf_CustomType_AutoIncrement(t *testing.T) {
	d := &Dialector{Config: &Config{}}
	got := d.DataTypeOf(&schema.Field{DataType: "serial", AutoIncrement: true})
	// "serial" already implies auto-increment; but if it doesn't contain the keyword
	// our code appends it.
	if !strings.Contains(got, "serial") {
		t.Errorf("DataTypeOf(serial) = %q, should contain %q", got, "serial")
	}
}

// --- QuoteTo ---

func TestQuoteTo(t *testing.T) {
	d := &Dialector{Config: &Config{}}

	tests := []struct {
		input    string
		expected string
	}{
		{"users", "`users`"},
		{"order", "`order`"},               // reserved word
		{"schema.table", "`schema`.`table`"}, // dot-separated
		{"back`tick", "`back``tick`"},        // escaped backtick
		{"a.b.c", "`a`.`b.c`"},              // only first dot is a separator
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var w strings.Builder
			d.QuoteTo(&w, tt.input)
			if got := w.String(); got != tt.expected {
				t.Errorf("QuoteTo(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// --- BindVarTo ---

func TestBindVarTo(t *testing.T) {
	d := &Dialector{Config: &Config{}}
	var w strings.Builder
	d.BindVarTo(&w, nil, nil)
	if got := w.String(); got != "?" {
		t.Errorf("BindVarTo() = %q, want %q", got, "?")
	}
}

// BindVarTo should always produce "?" regardless of the value passed.
func TestBindVarTo_IgnoresValue(t *testing.T) {
	d := &Dialector{Config: &Config{}}
	for _, v := range []interface{}{nil, 42, "hello", true, 3.14} {
		var w strings.Builder
		d.BindVarTo(&w, nil, v)
		if got := w.String(); got != "?" {
			t.Errorf("BindVarTo(%v) = %q, want %q", v, got, "?")
		}
	}
}

// --- DefaultValueOf ---

func TestDefaultValueOf(t *testing.T) {
	d := &Dialector{Config: &Config{}}
	expr := d.DefaultValueOf(&schema.Field{})
	e, ok := expr.(clause.Expr)
	if !ok {
		t.Fatalf("DefaultValueOf() returned %T, want clause.Expr", expr)
	}
	if e.SQL != "DEFAULT" {
		t.Errorf("DefaultValueOf().SQL = %q, want %q", e.SQL, "DEFAULT")
	}
}

// --- Explain ---

func TestExplain(t *testing.T) {
	d := &Dialector{Config: &Config{}}

	tests := []struct {
		sql      string
		vars     []interface{}
		contains string
	}{
		{
			sql:      "SELECT * FROM users WHERE id = ?",
			vars:     []interface{}{42},
			contains: "42",
		},
		{
			sql:      "SELECT * FROM users WHERE name = ?",
			vars:     []interface{}{"alice"},
			contains: "'alice'",
		},
		{
			sql:      "UPDATE users SET name = ? WHERE id = ?",
			vars:     []interface{}{"bob", 1},
			contains: "'bob'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.sql, func(t *testing.T) {
			got := d.Explain(tt.sql, tt.vars...)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("Explain() = %q, should contain %q", got, tt.contains)
			}
		})
	}
}

// --- buildColumnType (migrator helper) ---

func TestBuildColumnType(t *testing.T) {
	tests := []struct {
		name      string
		dataType  string
		charMax   sql.NullInt64
		numPrec   sql.NullInt64
		numScale  sql.NullInt64
		expected  string
	}{
		{
			name:     "int - no size components",
			dataType: "int",
			expected: "int",
		},
		{
			name:     "bigint",
			dataType: "bigint",
			expected: "bigint",
		},
		{
			name:     "datetime",
			dataType: "datetime",
			expected: "datetime",
		},
		{
			name:     "blob",
			dataType: "blob",
			expected: "blob",
		},
		{
			name:     "varchar with length",
			dataType: "varchar",
			charMax:  sql.NullInt64{Int64: 100, Valid: true},
			expected: "varchar(100)",
		},
		{
			name:     "varchar without length falls back to type name",
			dataType: "varchar",
			expected: "varchar",
		},
		{
			name:     "char with length",
			dataType: "char",
			charMax:  sql.NullInt64{Int64: 50, Valid: true},
			expected: "char(50)",
		},
		{
			name:     "nchar with length",
			dataType: "nchar",
			charMax:  sql.NullInt64{Int64: 30, Valid: true},
			expected: "nchar(30)",
		},
		{
			name:     "varnchar with length",
			dataType: "varnchar",
			charMax:  sql.NullInt64{Int64: 200, Valid: true},
			expected: "varnchar(200)",
		},
		{
			name:     "numeric with precision and scale",
			dataType: "numeric",
			numPrec:  sql.NullInt64{Int64: 10, Valid: true},
			numScale: sql.NullInt64{Int64: 2, Valid: true},
			expected: "numeric(10,2)",
		},
		{
			name:     "numeric with precision only",
			dataType: "numeric",
			numPrec:  sql.NullInt64{Int64: 8, Valid: true},
			expected: "numeric(8)",
		},
		{
			name:     "decimal with precision and scale",
			dataType: "decimal",
			numPrec:  sql.NullInt64{Int64: 15, Valid: true},
			numScale: sql.NullInt64{Int64: 3, Valid: true},
			expected: "decimal(15,3)",
		},
		{
			name:     "NUMERIC uppercase is case-insensitive",
			dataType: "NUMERIC",
			numPrec:  sql.NullInt64{Int64: 6, Valid: true},
			numScale: sql.NullInt64{Int64: 0, Valid: true},
			expected: "NUMERIC(6,0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildColumnType(tt.dataType, tt.charMax, tt.numPrec, tt.numScale)
			if got != tt.expected {
				t.Errorf("buildColumnType(%q, ...) = %q, want %q", tt.dataType, got, tt.expected)
			}
		})
	}
}
