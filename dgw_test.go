package main

import (
	"database/sql"
	"io/ioutil"
	"testing"

	_ "github.com/lib/pq"
)

// before running test, create user and database
// CREATE USER dgw_test;
// CREATE DATABASE  dgw_test OWNER dgw_test;

func testPgSetup(t *testing.T) (*sql.DB, func()) {
	conn, err := sql.Open("postgres", "user=dgw_test dbname=dgw_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	setupSQL, err := ioutil.ReadFile("./test.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(string(setupSQL))
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		conn.Close()
	}
	return conn, cleanup
}

func testSetupStruct(t *testing.T, conn *sql.DB) []*Struct {
	schema := "public"
	tbls, err := PgLoadTableDef(conn, schema)
	if err != nil {
		t.Fatal(err)
	}

	var sts []*Struct
	for _, tbl := range tbls {
		st, err := PgTableToStruct(tbl, defaultTypeMapCfg, autoGenKeyCfg)
		if err != nil {
			t.Fatal(err)
		}
		sts = append(sts, st)
	}
	return sts
}

func TestPgLoadColumnDef(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	table := "user_account_uuid_address"
	cols, err := PgLoadColumnDef(conn, schema, table)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cols {
		t.Logf("%+v", c)
	}
}

func TestPgLoadTableDef(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	tbls, err := PgLoadTableDef(conn, schema)
	if err != nil {
		t.Fatal(err)
	}
	for _, tbl := range tbls {
		t.Logf("%+v", tbl)
	}
}

func TestPgColToField(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	table := "t1"
	cols, err := PgLoadColumnDef(conn, schema, table)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cols {
		f, err := PgColToField(c, defaultTypeMapCfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", f)
	}
}

func TestPgLoadTypeMap(t *testing.T) {
	path := "./typemap.toml"
	c, err := PgLoadTypeMapFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range *c {
		t.Logf("%+v, %+v", k, v)
	}
}

func TestPgTableToStruct(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	tbls, err := PgLoadTableDef(conn, schema)
	if err != nil {
		t.Fatal(err)
	}

	for _, tbl := range tbls {
		st, err := PgTableToStruct(tbl, defaultTypeMapCfg, autoGenKeyCfg)
		if err != nil {
			t.Fatal(err)
		}
		src, err := PgExecuteDefaultTmpl(&StructTmpl{Struct: st}, "template/struct.tmpl")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", src)
	}
}

func TestPgTableToMethod(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	tbls, err := PgLoadTableDef(conn, schema)
	if err != nil {
		t.Fatal(err)
	}
	for _, tbl := range tbls {
		st, err := PgTableToStruct(tbl, defaultTypeMapCfg, autoGenKeyCfg)
		if err != nil {
			t.Fatal(err)
		}
		src, err := PgExecuteDefaultTmpl(&StructTmpl{Struct: st}, "template/method.tmpl")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", src)
	}
}

var testTmpl = `// {{ .Struct.Name }}  {{ .Struct.Table.Schema }}.{{ .Struct.Table.Name }}
// this is custom template with "Tbl" suffix
type {{ .Struct.Name }}Tbl struct {
{{- range .Struct.Fields }}
	{{ .Name }} {{ .Type }} // {{ .Column.Name }}
{{- end }}
}
`

func TestPgExecuteCustomTemplate(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	tbls, err := PgLoadTableDef(conn, schema)
	if err != nil {
		t.Fatal(err)
	}
	for _, tbl := range tbls {
		st, err := PgTableToStruct(tbl, defaultTypeMapCfg, autoGenKeyCfg)
		if err != nil {
			t.Fatal(err)
		}
		src, err := PgExecuteCustomTmpl(&StructTmpl{Struct: st}, testTmpl)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", src)
	}
}
