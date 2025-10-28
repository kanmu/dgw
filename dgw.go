package main

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/achiku/varfmt"
	_ "github.com/lib/pq" // postgres
	"github.com/pkg/errors"
)

// Queryer database/sql compatible query interface
type Queryer interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// OpenDB opens database connection
func OpenDB(connStr string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return conn, nil
}

const queryInterface = `
// Queryer database/sql compatible query interface
type Queryer interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
`

const pgLoadColumnDef = `
SELECT
    a.attnum AS field_ordinal,
    a.attname AS column_name,
    format_type(a.atttypid, a.atttypmod) AS data_type,
    a.attnotnull AS not_null,
    COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') AS default_value,
    COALESCE(ct.contype = 'p', false) AS  is_primary_key,
    CASE
        WHEN a.atttypid = ANY ('{int,int8,int2}'::regtype[])
          AND EXISTS (
             SELECT 1 FROM pg_attrdef ad
             WHERE  ad.adrelid = a.attrelid
             AND    ad.adnum   = a.attnum
             AND    pg_get_expr(ad.adbin, ad.adrelid) = 'nextval('''
                || (pg_get_serial_sequence (a.attrelid::regclass::text
                                          , a.attname))::regclass
                || '''::regclass)'
             )
            THEN CASE a.atttypid
                    WHEN 'int'::regtype  THEN 'serial'
                    WHEN 'int8'::regtype THEN 'bigserial'
                    WHEN 'int2'::regtype THEN 'smallserial'
                 END
        WHEN a.atttypid = ANY ('{uuid}'::regtype[]) AND COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') != ''
            THEN 'autogenuuid'
        ELSE format_type(a.atttypid, a.atttypmod)
    END AS data_type,
    a.attidentity AS identity
FROM pg_attribute a
JOIN ONLY pg_class c ON c.oid = a.attrelid
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid
AND a.attnum = ANY(ct.conkey) AND ct.contype = 'p'
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
ORDER BY c.relname
`

// TypeMap go/db type map struct
type TypeMap struct {
	DBTypes        []string `toml:"db_types"`
	NotNullGoType  string   `toml:"notnull_go_type"`
	NullableGoType string   `toml:"nullable_go_type"`
}

// AutoKeyMap auto generating key config
type AutoKeyMap struct {
	Types        []string `toml:"db_types"`
	IdentityAsPk bool
}

// PgTypeMapConfig go/db type map struct toml config
type PgTypeMapConfig map[string]TypeMap

// PgTable postgres table
type PgTable struct {
	Schema      string
	Name        string
	DataType    string
	AutoGenPk   bool
	PrimaryKeys []*PgColumn
	Columns     []*PgColumn
}

var autoGenKeyCfg = &AutoKeyMap{
	Types: []string{"smallserial", "serial", "bigserial", "autogenuuid"},
}

func (t *PgTable) setPrimaryKeyInfo(cfg *AutoKeyMap) {
	t.AutoGenPk = false
	for _, c := range t.Columns {
		if c.IsPrimaryKey {
			t.PrimaryKeys = append(t.PrimaryKeys, c)
			for _, typ := range cfg.Types {
				// https://www.postgresql.jp/docs/16/catalog-pg-attribute.html
				if c.DDLType == typ || (cfg.IdentityAsPk && (c.Identity == "a" || c.Identity == "d")) {
					t.AutoGenPk = true
				}
			}
		}
	}
}

// PgColumn postgres columns
type PgColumn struct {
	FieldOrdinal int
	Name         string
	DataType     string
	DDLType      string
	NotNull      bool
	DefaultValue sql.NullString
	IsPrimaryKey bool
	Identity     string
}

// Struct go struct
type Struct struct {
	Name    string
	Table   *PgTable
	Comment string
	Fields  []*StructField
}

// StructTmpl go struct passed to template
type StructTmpl struct {
	Struct *Struct
}

// StructField go struct field
type StructField struct {
	Name   string
	Type   string
	Tag    string
	Column *PgColumn
}

// PgLoadTypeMapFromFile load type map from toml file
func PgLoadTypeMapFromFile(filePath string) (*PgTypeMapConfig, error) {
	var conf PgTypeMapConfig
	if _, err := toml.DecodeFile(filePath, &conf); err != nil {
		return nil, errors.WithStack(err)
	}
	return &conf, nil
}

// PgLoadColumnDef load Postgres column definition
func PgLoadColumnDef(db Queryer, schema string, table string) ([]*PgColumn, error) {
	colDefs, err := db.Query(pgLoadColumnDef, schema, table)
	if err != nil {
		return nil, errors.WithStack(err)
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
			&c.DDLType,
			&c.Identity,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Some data types have an extra part e.g, "character varying(16)" and
		// "numeric(10, 5)". We want to drop the extra part.
		if i := strings.Index(c.DataType, "("); i > 0 {
			c.DataType = c.DataType[0:i]
		}

		cols = append(cols, c)
	}
	return cols, nil
}

// PgLoadTableDef load Postgres table definition
func PgLoadTableDef(db Queryer, schema string) ([]*PgTable, error) {
	tbDefs, err := db.Query(pgLoadTableDef, schema)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tbs := []*PgTable{}
	for tbDefs.Next() {
		t := &PgTable{Schema: schema}
		err := tbDefs.Scan(
			&t.DataType,
			&t.Name,
		)
		if err != nil {
			return nil, errors.WithStack(err)
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

func contains(v string, l []string) bool {
	sort.Strings(l)
	i := sort.SearchStrings(l, v)
	if i < len(l) && l[i] == v {
		return true
	}
	return false
}

// PgConvertType converts type
func PgConvertType(col *PgColumn, typeCfg *PgTypeMapConfig) string {
	cfg := map[string]TypeMap(*typeCfg)
	typ := cfg["default"].NotNullGoType
	for _, v := range cfg {
		if contains(col.DataType, v.DBTypes) {
			if col.NotNull {
				return v.NotNullGoType
			}
			return v.NullableGoType
		}
	}
	return typ
}

// PgColToField converts pg column to go struct field
func PgColToField(col *PgColumn, typeCfg *PgTypeMapConfig) (*StructField, error) {
	stfType := PgConvertType(col, typeCfg)
	stf := &StructField{
		Name:   varfmt.PublicVarName(col.Name),
		Type:   stfType,
		Column: col,
	}
	return stf, nil
}

// PgTableToStruct converts table def to go struct
func PgTableToStruct(t *PgTable, typeCfg *PgTypeMapConfig, keyConfig *AutoKeyMap) (*Struct, error) {
	t.setPrimaryKeyInfo(keyConfig)
	s := &Struct{
		Name:  varfmt.PublicVarName(t.Name),
		Table: t,
	}
	var fs []*StructField
	for _, c := range t.Columns {
		f, err := PgColToField(c, typeCfg)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		fs = append(fs, f)
	}
	s.Fields = fs
	return s, nil
}

//go:embed template/struct.tmpl
var structTemplate string

// PgExecuteDefaultStructTmpl execute struct template with *Struct
func PgExecuteDefaultStructTmpl(st *StructTmpl) ([]byte, error) {
	var src []byte

	tpl, err := template.New("struct").Funcs(tmplFuncMap).Parse(structTemplate)
	if err != nil {
		return src, errors.WithStack(err)
	}
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, st); err != nil {
		return src, errors.Wrap(err, fmt.Sprintf("failed to execute template:\n%s", src))
	}
	src, err = format.Source(buf.Bytes())
	if err != nil {
		return src, errors.Wrap(err, fmt.Sprintf("failed to format code:\n%s", src))
	}
	return src, nil
}

//go:embed template/method.tmpl
var methodTemplate string

// PgExecuteDefaultMethodTmpl execute method template with *Struct
func PgExecuteDefaultMethodTmpl(st *StructTmpl) ([]byte, error) {
	var src []byte

	tpl, err := template.New("struct").Funcs(tmplFuncMap).Parse(methodTemplate)
	if err != nil {
		return src, errors.WithStack(err)
	}
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, st); err != nil {
		return src, errors.Wrap(err, fmt.Sprintf("failed to execute template:\n%s", src))
	}
	src, err = format.Source(buf.Bytes())
	if err != nil {
		return src, errors.Wrap(err, fmt.Sprintf("failed to format code:\n%s", src))
	}
	return src, nil
}

// PgExecuteCustomTmpl execute custom template
func PgExecuteCustomTmpl(st *StructTmpl, customTmpl string) ([]byte, error) {
	var src []byte
	tpl, err := template.New("struct").Funcs(tmplFuncMap).Parse(customTmpl)
	if err != nil {
		return src, errors.WithStack(err)
	}
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, st); err != nil {
		return src, errors.Wrap(err, fmt.Sprintf("failed to execute custom template:\n%s", src))
	}
	src, err = format.Source(buf.Bytes())
	if err != nil {
		return src, errors.Wrap(err, fmt.Sprintf("failed to format code:\n%s", src))
	}
	return src, nil
}

// PgCreateStruct creates struct from given schema
func PgCreateStruct(
	db Queryer, schema, typeMapPath, pkgName, customTmpl string, exTbls []string, autoGenKeyList []string, identityAsPk bool) ([]byte, error) {
	src := []byte("// Code generated by dgw. DO NOT EDIT.\n\n")
	pkgDef := []byte(fmt.Sprintf("package %s\n\n", pkgName))
	src = append(src, pkgDef...)

	tbls, err := PgLoadTableDef(db, schema)
	if err != nil {
		return src, errors.WithStack(err)
	}
	cfg := &PgTypeMapConfig{}
	if typeMapPath == "" {
		if _, err := toml.Decode(typeMap, cfg); err != nil {
			return src, errors.WithStack(err)
		}
	} else {
		if _, err := toml.DecodeFile(typeMapPath, cfg); err != nil {
			return src, errors.Wrap(err, fmt.Sprintf("failed to decode type map file %s", typeMapPath))
		}
	}
	agkCfg := autoGenKeyCfg
	if len(autoGenKeyList) > 0 {
		agkCfg = &AutoKeyMap{
			Types: autoGenKeyList,
		}
	}
	agkCfg.IdentityAsPk = identityAsPk
	for _, tbl := range tbls {
		if contains(tbl.Name, exTbls) {
			continue
		}
		st, err := PgTableToStruct(tbl, cfg, agkCfg)
		if err != nil {
			return src, errors.WithStack(err)
		}
		if customTmpl != "" {
			tmpl, err := os.ReadFile(customTmpl)
			if err != nil {
				return nil, err
			}
			s, err := PgExecuteCustomTmpl(&StructTmpl{Struct: st}, string(tmpl))
			if err != nil {
				return nil, errors.WithStack(err)
			}
			src = append(src, s...)
		} else {
			s, err := PgExecuteDefaultStructTmpl(&StructTmpl{Struct: st})
			if err != nil {
				return src, errors.WithStack(err)
			}
			m, err := PgExecuteDefaultMethodTmpl(&StructTmpl{Struct: st})
			if err != nil {
				return src, errors.WithStack(err)
			}
			src = append(src, s...)
			src = append(src, m...)
		}
	}
	return src, nil
}
