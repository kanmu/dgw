// go:generate go-bindata -o bindata.go template mapconfig
package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"go/format"
	"io/ioutil"
	"regexp"
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
		return nil, errors.Wrap(err, "failed to connect to database")
	}
	return conn, nil
}

const pgLoadEnumDef = `
SELECT n.nspname AS schema,
       pg_catalog.format_type ( t.oid, NULL ) AS name,
       ARRAY( SELECT e.enumlabel
                  FROM pg_catalog.pg_enum e
                  WHERE e.enumtypid = t.oid
                  ORDER BY e.oid )
         AS elements
FROM pg_catalog.pg_type t
       LEFT JOIN pg_catalog.pg_namespace n
                 ON n.oid = t.typnamespace
WHERE ( t.typrelid = 0
  OR ( SELECT c.relkind = 'c'
       FROM pg_catalog.pg_class c
       WHERE c.oid = t.typrelid
        )
  )
  AND NOT EXISTS
  ( SELECT 1
    FROM pg_catalog.pg_type el
    WHERE el.oid = t.typelem
      AND el.typarray = t.oid
  )
  AND n.nspname = $1
  AND pg_catalog.pg_type_is_visible ( t.oid )
ORDER BY 1, 2;
`

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
    END AS data_type
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

	compiled   bool
	rePatterns []*regexp.Regexp
}

func (t *TypeMap) Match(s string) bool {
	if !t.compiled {
		for _, v := range t.DBTypes {
			if strings.HasPrefix(v, "re/") {
				t.rePatterns = append(t.rePatterns, regexp.MustCompile(v[3:]))
			}
		}
	}
	if contains(s, t.DBTypes) {
		return true
	}
	for _, v := range t.rePatterns {
		if v.MatchString(s) {
			return true
		}
	}
	return false
}

// AutoKeyMap auto generating key config
type AutoKeyMap struct {
	Types []string `toml:"db_types"`
}

// PgTypeMapConfig go/db type map struct toml config
type PgTypeMapConfig map[string]*TypeMap

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
				if c.DDLType == typ {
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
		return nil, errors.Wrap(err, "faild to parse config file")
	}
	return &conf, nil
}

type PgEnum struct {
	Schema string
	Name   string
	Values []string
}

type EnumValue struct {
	Type  *EnumType
	Name  string
	Value string
}

type EnumType struct {
	Name    string
	Enum    *PgEnum
	Comment string
	Values  []EnumValue
}

func PgLoadEnumDef(db Queryer, schema string) ([]*PgEnum, error) {
	enumDefs, err := db.Query(pgLoadEnumDef, schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load enum def")
	}

	enums := []*PgEnum{}
	for enumDefs.Next() {
		e := &PgEnum{}
		var vals pq.StringArray
		err := enumDefs.Scan(
			&e.Schema,
			&e.Name,
			&vals,
		)
		e.Values = vals
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}
		enums = append(enums, e)
	}
	return enums, nil
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
			&c.DDLType,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan")
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
		return nil, errors.Wrap(err, "failed to load table def")
	}
	tbs := []*PgTable{}
	for tbDefs.Next() {
		t := &PgTable{Schema: schema}
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

func contains(v string, l []string) bool {
	sort.Strings(l)
	i := sort.SearchStrings(l, v)
	if i < len(l) && l[i] == v {
		return true
	}
	return false
}

// PgConvertType converts type
func PgConvertType(col *PgColumn, typeCfg PgTypeMapConfig) string {
	typ := typeCfg["default"].NotNullGoType
	for _, v := range typeCfg {
		if v.Match(col.DataType) {
			if col.NotNull {
				return v.NotNullGoType
			}
			return v.NullableGoType
		}
	}
	return typ
}

// PgColToField converts pg column to go struct field
func PgColToField(col *PgColumn, typeCfg PgTypeMapConfig) (*StructField, error) {
	stfType := PgConvertType(col, typeCfg)
	stf := &StructField{
		Name:   varfmt.PublicVarName(col.Name),
		Type:   stfType,
		Column: col,
	}
	return stf, nil
}

// PgTableToStruct converts table def to go struct
func PgTableToStruct(t *PgTable, typeCfg PgTypeMapConfig, keyConfig *AutoKeyMap) (*Struct, error) {
	t.setPrimaryKeyInfo(keyConfig)
	s := &Struct{
		Name:  varfmt.PublicVarName(t.Name),
		Table: t,
	}
	var fs []*StructField
	for _, c := range t.Columns {
		f, err := PgColToField(c, typeCfg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert col to field")
		}
		fs = append(fs, f)
	}
	s.Fields = fs
	return s, nil
}

// PgExecuteDefaultTmpl execute struct template with *Struct
func PgExecuteDefaultTmpl(st interface{}, path string) ([]byte, error) {
	var src []byte
	d, err := Asset(path)
	if err != nil {
		return src, errors.Wrap(err, "failed to load asset")
	}
	tpl, err := template.New("struct").Funcs(tmplFuncMap).Parse(string(d))
	if err != nil {
		return src, errors.Wrap(err, "failed to parse template")
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
func PgExecuteCustomTmpl(st interface{}, customTmpl string) ([]byte, error) {
	var src []byte
	tpl, err := template.New("struct").Funcs(tmplFuncMap).Parse(customTmpl)
	if err != nil {
		return src, errors.Wrap(err, "failed to parse template")
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

func getPgTypeMapConfig(typeMapPath string) (PgTypeMapConfig, error) {
	cfg := make(PgTypeMapConfig)
	if typeMapPath == "" {
		if _, err := toml.Decode(typeMap, &cfg); err != nil {
			return nil, errors.Wrap(err, "failed to read type map")
		}
	} else {
		if _, err := toml.DecodeFile(typeMapPath, &cfg); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode type map file %s", typeMapPath))
		}
	}
	return cfg, nil
}

func PgEnumToType(e *PgEnum, typeCfg PgTypeMapConfig, keyConfig *AutoKeyMap) (*EnumType, error) {
	en := &EnumType{
		Name: varfmt.PublicVarName(e.Name),
		Enum: e,
	}
	for _, v := range e.Values {
		en.Values = append(en.Values, EnumValue{
			Type:  en,
			Name:  en.Name + "_" + varfmt.PublicVarName(v),
			Value: v,
		})
	}
	if _,ok := typeCfg[e.Name]; !ok {
		typeCfg[e.Name] = &TypeMap{
			DBTypes:        []string{e.Name},
			NotNullGoType:  en.Name,
			NullableGoType: "Null"+en.Name,

			compiled:       true,
			rePatterns:     nil,
		}
	}

	return en, nil
}

func PgCreateEnums(db Queryer, schema string, cfg PgTypeMapConfig, customTmpl string) ([]byte, error) {
	var src []byte

	enums, err := PgLoadEnumDef(db, schema)
	if err != nil {
		return src, errors.Wrap(err, "failed to load enum definitions")
	}

	for _, pgEnum := range enums {
		enum, err := PgEnumToType(pgEnum, cfg, autoGenKeyCfg)
		if err != nil {
			return src, errors.Wrap(err, "failed to convert enum definition to type")
		}

		if customTmpl != "" {
			tmpl, err := ioutil.ReadFile(customTmpl)
			if err != nil {
				return nil, err
			}
			s, err := PgExecuteCustomTmpl(enum, string(tmpl))
			if err != nil {
				return nil, errors.Wrap(err, "PgExecuteCustomTmpl failed")
			}
			src = append(src, s...)
		} else {
			s, err := PgExecuteDefaultTmpl(enum, "template/enum.tmpl")
			if err != nil {
				return src, errors.Wrap(err, "failed to execute template")
			}
			src = append(src, s...)
		}
	}
	return src, nil
}

// PgCreateStruct creates struct from given schema
func PgCreateStruct(
	db Queryer, schema string, cfg PgTypeMapConfig, pkgName, customTmpl string, exTbls []string) ([]byte, error) {
	var src []byte
	pkgDef := []byte(fmt.Sprintf("package %s\n\n", pkgName))
	src = append(src, pkgDef...)

	tbls, err := PgLoadTableDef(db, schema)
	if err != nil {
		return src, errors.Wrap(err, "failed to load table definitions")
	}

	for _, tbl := range tbls {
		if contains(tbl.Name, exTbls) {
			continue
		}
		st, err := PgTableToStruct(tbl, cfg, autoGenKeyCfg)
		if err != nil {
			return src, errors.Wrap(err, "failed to convert table definition to struct")
		}
		if customTmpl != "" {
			tmpl, err := ioutil.ReadFile(customTmpl)
			if err != nil {
				return nil, err
			}
			s, err := PgExecuteCustomTmpl(&StructTmpl{Struct: st}, string(tmpl))
			if err != nil {
				return nil, errors.Wrap(err, "PgExecuteCustomTmpl failed")
			}
			src = append(src, s...)
		} else {
			s, err := PgExecuteDefaultTmpl(&StructTmpl{Struct: st}, "template/struct.tmpl")
			if err != nil {
				return src, errors.Wrap(err, "failed to execute template")
			}
			m, err := PgExecuteDefaultTmpl(&StructTmpl{Struct: st}, "template/method.tmpl")
			if err != nil {
				return src, errors.Wrap(err, "failed to execute template")
			}
			src = append(src, s...)
			src = append(src, m...)
		}
	}
	return src, nil
}
