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
	setupSQL, err := ioutil.ReadFile("./dgw_test.sql")
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

func TestPgLoadColumnDef(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	table := "t1"
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
	path := "./mapconfig/typemap.toml"
	cfg, err := PgLoadTypeMapFromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cols {
		f, err := PgColToField(c, cfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", f)
	}
}

func TestPgLoadTypeMap(t *testing.T) {
	path := "./mapconfig/typemap.toml"
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
	path := "./mapconfig/typemap.toml"
	cfg, err := PgLoadTypeMapFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tbl := range tbls {
		st, err := PgTableToStruct(tbl, cfg)
		if err != nil {
			t.Fatal(err)
		}
		src, err := PgExecuteStructTmpl(st, "template/struct.tmpl")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", src)
	}
}
