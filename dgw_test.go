package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	setupSQL, err := os.ReadFile(filepath.Join("sql", "test.sql"))
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg)
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
		f, err := PgColToField(c, &defaultTypeMapCfg)
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg)
		if err != nil {
			t.Fatal(err)
		}
		src, err := PgExecuteDefaultStructTmpl(&StructTmpl{Struct: st})
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg)
		if err != nil {
			t.Fatal(err)
		}
		src, err := PgExecuteDefaultMethodTmpl(&StructTmpl{Struct: st})
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg)
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

func TestCreateInsertOnConflictDoNothingSQL(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	structs := testSetupStruct(t, conn)

	for _, st := range structs {
		sql := createInsertOnConflictDoNothingSQL(st)

		// Check that SQL contains ON CONFLICT DO NOTHING
		if !strings.Contains(sql, "ON CONFLICT DO NOTHING") {
			t.Errorf("Expected SQL to contain 'ON CONFLICT DO NOTHING', got: %s", sql)
		}

		// Check that SQL starts with INSERT INTO
		if !strings.HasPrefix(sql, "INSERT INTO") {
			t.Errorf("Expected SQL to start with 'INSERT INTO', got: %s", sql)
		}

		// Log the generated SQL for manual inspection
		t.Logf("Table: %s, Generated SQL: %s", st.Table.Name, sql)

		// For tables with auto-generated primary keys, check RETURNING clause
		if st.Table.AutoGenPk {
			if !strings.Contains(sql, "RETURNING") {
				t.Errorf("Expected SQL for table %s with AutoGenPk to contain 'RETURNING', got: %s", st.Table.Name, sql)
			}
		}
	}
}

func TestCreateOnConflictDoNothingMethodGeneration(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	tbls, err := PgLoadTableDef(conn, schema)
	if err != nil {
		t.Fatal(err)
	}

	for _, tbl := range tbls {
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg)
		if err != nil {
			t.Fatal(err)
		}

		src, err := PgExecuteDefaultMethodTmpl(&StructTmpl{Struct: st})
		if err != nil {
			t.Fatal(err)
		}

		srcStr := string(src)

		// Check that CreateOnConflictDoNothing method is generated
		if !strings.Contains(srcStr, "CreateOnConflictDoNothing") {
			t.Errorf("Expected generated code to contain CreateOnConflictDoNothing method for %s", st.Name)
		}

		// Check that the method has the correct signature
		expectedSignature := fmt.Sprintf("func (r *%s) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error)", st.Name)
		if !strings.Contains(srcStr, expectedSignature) {
			t.Errorf("Expected method signature '%s' not found in generated code", expectedSignature)
		}

		// Check that the method contains the ON CONFLICT DO NOTHING SQL
		if !strings.Contains(srcStr, "ON CONFLICT DO NOTHING") {
			t.Errorf("Expected generated method to contain 'ON CONFLICT DO NOTHING' SQL")
		}

		// Check for the comment explaining the behavior
		if !strings.Contains(srcStr, "If a conflict occurs") {
			t.Errorf("Expected generated method to contain comment explaining conflict behavior")
		}

		t.Logf("Generated CreateOnConflictDoNothing for %s", st.Name)
	}
}
