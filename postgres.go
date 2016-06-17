package tdgw

import (
	"database/sql"
	"fmt"

	"github.com/achiku/varfmt"
	"github.com/pkg/errors"
)

const pgLoadColumnDef = `
SELECT
    a.attnum AS field_ordinal,
    a.attname AS column_name,
    format_type(a.atttypid, a.atttypmod) AS data_type,
    a.attnotnull AS not_null,
    COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') AS default_value,
    COALESCE(ct.contype = 'p', false) AS  is_primary_key
FROM pg_attribute a
JOIN ONLY pg_class c ON c.oid = a.attrelid
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid
AND a.attnum = ANY(ct.conkey) AND ct.contype IN ('p', 'u')
LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
WHERE a.attisdropped = false
AND n.nspname = $1
AND c.relname = $2
AND a.attnum > 0
ORDER BY a.attnum
`

const pgLoadTableDef = `
SELECT
c.relkind AS type,
c.relname AS table_name
FROM pg_class c
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1
AND c.relkind = 'r'
`

// PgTable postgres table
type PgTable struct {
	Name     string
	DataType string
	Columns  []*PgColumn
}

// PgColumn postgres columns
type PgColumn struct {
	FieldOrdinal int
	Name         string
	DataType     string
	NotNull      bool
	DefaultValue sql.NullString
	IsPrimaryKey bool
}

// PgLoadColumnDef load Postgres column definition
func PgLoadColumnDef(db Queryer, schema string, table string) ([]*PgColumn, error) {
	colDefs, err := db.Query(pgLoadColumnDef, schema, table)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load table def")
	}

	cols := []*PgColumn{}
	for colDefs.Next() {
		c := &PgColumn{}
		err := colDefs.Scan(
			&c.FieldOrdinal,
			&c.Name,
			&c.DataType,
			&c.NotNull,
			&c.DefaultValue,
			&c.IsPrimaryKey,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}
		cols = append(cols, c)
	}
	return cols, nil
}

// PgLoadTableDef load Postgres table definition
func PgLoadTableDef(db Queryer, schema string) ([]*PgTable, error) {
	tbDefs, err := db.Query(pgLoadTableDef, schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load table def")
	}
	tbs := []*PgTable{}
	for tbDefs.Next() {
		t := &PgTable{}
		err := tbDefs.Scan(
			&t.DataType,
			&t.Name,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}
		cols, err := PgLoadColumnDef(db, schema, t.Name)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to get columns of %s", t.Name))
		}
		t.Columns = cols
		tbs = append(tbs, t)
	}
	return tbs, nil
}

// StructField go struct field
type StructField struct {
	Name   string
	Type   string
	Tag    string
	NilVal string
}

// Struct go struct
type Struct struct {
	Fields []*StructField
}

// PgColToField converts pg column to go struct field
func PgColToField(col *PgColumn) (*StructField, error) {
	stfName := varfmt.PublicVarName(col.Name)
	stfType, nilVal := PgConvertType(col)
	stf := &StructField{Name: stfName, Type: stfType, NilVal: nilVal}
	return stf, nil
}

// PgConvertType converts type
func PgConvertType(col *PgColumn) (string, string) {

	nilVal := "nil"
	var typ string
	switch col.DataType {
	case "boolean":
		nilVal = "false"
		typ = "bool"
		if !col.NotNull {
			nilVal = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "character", "character varying", "text", "money":
		nilVal = `""`
		typ = "string"
		if !col.NotNull {
			nilVal = "sql.NullString{}"
			typ = "sql.NullString"
		}

	case "smallint":
		nilVal = "0"
		typ = "int16"
		if !col.NotNull {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "integer":
		nilVal = "0"
		typ = "int"
		if !col.NotNull {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigint":
		nilVal = "0"
		typ = "int64"
		if !col.NotNull {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "smallserial":
		nilVal = "0"
		typ = "uint16"
		if !col.NotNull {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "serial":
		nilVal = "0"
		typ = "uint32"
		if !col.NotNull {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigserial":
		nilVal = "0"
		typ = "uint64"
		if !col.NotNull {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "real":
		nilVal = "0.0"
		typ = "float32"
		if !col.NotNull {
			nilVal = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}
	case "numeric", "double precision":
		nilVal = "0.0"
		typ = "float64"
		if !col.NotNull {
			nilVal = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "bytea":
		typ = "byte"

	case "timestamp with time zone":
		typ = "time.Time"
		if !col.NotNull {
			nilVal = "pq.NullTime{}"
			typ = "pq.NullTime"
		}

	case "date":
		typ = "time.Time"
		if !col.NotNull {
			nilVal = "pq.NullTime{}"
			typ = "pq.NullTime"
		}

	case "time with time zone", "time without time zone", "timestamp without time zone":
		nilVal = "0"
		typ = "int64"
		if !col.NotNull {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "interval":
		typ = "*time.Duration"

	case `"char"`, "bit":
		// FIXME: this needs to actually be tested ...
		// i think this should be 'rune' but I don't think database/sql
		// supports 'rune' as a type?
		//
		// this is mainly here because postgres's pg_catalog.* meta tables have
		// this as a type.
		//typ = "rune"
		nilVal = `uint8(0)`
		typ = "uint8"
	case `"any"`, "bit varying":
		typ = "byte"
	default:
		typ = "interface{}"
	}
	return typ, nilVal
}

func pgLoadTypeMapp(col *PgColumn) (string, string) {
	return "", ""
}
