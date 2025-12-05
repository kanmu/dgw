package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg, []string{})
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg, []string{})
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg, []string{})
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg, []string{})
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

	if len(structs) != 6 {
		t.Fatalf("Expected the number of testing structs is 6, got: %d", len(structs))
	}

	tests := []struct {
		tableStruct *Struct
		expectSQL   string
	}{
		{
			tableStruct: structs[0],
			expectSQL:   "INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING RETURNING id",
		},
		{
			tableStruct: structs[1],
			expectSQL:   "INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id",
		},
		{
			tableStruct: structs[2],
			expectSQL:   "INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id, i",
		},
		{
			tableStruct: structs[3],
			expectSQL:   "INSERT INTO t4 (id, i) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		},
	}
	for _, tt := range tests {
		t.Run(tt.tableStruct.Table.Name, func(t *testing.T) {
			sql := createInsertOnConflictDoNothingSQL(tt.tableStruct)
			if sql != tt.expectSQL {
				t.Errorf("Expected SQL: %s, got: %s", tt.expectSQL, sql)
			}
			t.Logf("Table: %s, Generated SQL: %s", tt.tableStruct.Name, sql)
		})
	}
}

func TestMethodGeneration(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	schema := "public"
	tbls, err := PgLoadTableDef(conn, schema)
	if err != nil {
		t.Fatal(err)
	}

	if len(tbls) != 6 {
		t.Fatalf("Expected the number of testing PgTable is 6, got: %d", len(tbls))
	}

	tests := []struct {
		table  *PgTable
		expect string
	}{
		{
			table: tbls[0],
			expect: `// Create inserts the T1 to the database.
func (r *T1) Create(db Queryer) error {
        return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
func GetT1ByPk(db Queryer, pk0 int64) (*T1, error) {
        return GetT1ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T1 to the database.
func (r *T1) CreateContext(ctx context.Context, db Queryer) error {
        err := db.QueryRowContext(ctx,
                ` + "`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`" + `,
                &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
        if err != nil {
                return errors.WithStack(err)
        }
        return nil
}

// CreateOnConflictDoNothing inserts the T1 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T1) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
        err := db.QueryRowContext(ctx,
                ` + "`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING RETURNING id`" + `,
                &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
        if err != nil {
                if err == sql.ErrNoRows {
                        return false, nil
                }
                return false, errors.WithStack(err)
        }
        // Row was successfully inserted
        return true, nil
}

// GetT1ByPkContext select the T1 from the database.
func GetT1ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T1, error) {
        var r T1
        err := db.QueryRowContext(ctx,
                ` + "`SELECT id, i, str, nullable_str, t_with_tz, t_without_tz, tm FROM t1 WHERE id = $1`" + `,
                        pk0).Scan(&r.ID, &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm)
        if err != nil {
                return nil, errors.WithStack(err)
        }
        return &r, nil
}

`,
		},
		{
			table: tbls[1],
			expect: `// Create inserts the T2 to the database.
func (r *T2) Create(db Queryer) error {
        return r.CreateContext(context.Background(), db)
}

// GetT2ByPk select the T2 from the database.
func GetT2ByPk(db Queryer, pk0 int64) (*T2, error) {
        return GetT2ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T2 to the database.
func (r *T2) CreateContext(ctx context.Context, db Queryer) error {
        err := db.QueryRowContext(ctx,
                ` + "`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) RETURNING id`" + `,
                &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
        if err != nil {
                return errors.WithStack(err)
        }
        return nil
}

// CreateOnConflictDoNothing inserts the T2 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T2) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
        err := db.QueryRowContext(ctx,
                ` + "`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id`" + `,
                &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
        if err != nil {
                if err == sql.ErrNoRows {
                        return false, nil
                }
                return false, errors.WithStack(err)
        }
        // Row was successfully inserted
        return true, nil
}

// GetT2ByPkContext select the T2 from the database.
func GetT2ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T2, error) {
        var r T2
        err := db.QueryRowContext(ctx,
                ` + "`SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1`" + `,
                pk0).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
        if err != nil {
                return nil, errors.WithStack(err)
        }
        return &r, nil
}
`,
		},
		{
			table: tbls[2],
			expect: `// Create inserts the T3 to the database.
func (r *T3) Create(db Queryer) error {
        return r.CreateContext(context.Background(), db)
}

// GetT3ByPk select the T3 from the database.
func GetT3ByPk(db Queryer, pk0 int64, pk1 int) (*T3, error) {
        return GetT3ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T3 to the database.
func (r *T3) CreateContext(ctx context.Context, db Queryer) error {
        err := db.QueryRowContext(ctx,
                ` + "`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) RETURNING id, i`" + `,
                &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
        if err != nil {
                return errors.WithStack(err)
        }
        return nil
}

// CreateOnConflictDoNothing inserts the T3 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T3) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
        err := db.QueryRowContext(ctx,
                ` + "`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id, i`" + `,
                &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
        if err != nil {
                if err == sql.ErrNoRows {
                        return false, nil
                }
                return false, errors.WithStack(err)
        }
        // Row was successfully inserted
        return true, nil
}

// GetT3ByPkContext select the T3 from the database.
func GetT3ByPkContext(ctx context.Context, db Queryer, pk0 int64, pk1 int) (*T3, error) {
        var r T3
        err := db.QueryRowContext(ctx,
                ` + "`SELECT id, i, str, t_with_tz, t_without_tz FROM t3 WHERE id = $1 AND i = $2`" + `,
                pk0, pk1).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
        if err != nil {
                return nil, errors.WithStack(err)
        }
        return &r, nil
}
`,
		},
		{
			table: tbls[3],
			expect: `// Create inserts the T4 to the database.
func (r *T4) Create(db Queryer) error {
        return r.CreateContext(context.Background(), db)
}

// GetT4ByPk select the T4 from the database.
func GetT4ByPk(db Queryer, pk0 int, pk1 int) (*T4, error) {
        return GetT4ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T4 to the database.
func (r *T4) CreateContext(ctx context.Context, db Queryer) error {
        _, err := db.ExecContext(ctx,
                ` + "`INSERT INTO t4 (id, i) VALUES ($1, $2)`" + `,
                &r.ID, &r.I)
        if err != nil {
                return errors.WithStack(err)
        }
        return nil
}

// CreateOnConflictDoNothing inserts the T4 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T4) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
        result, err := db.ExecContext(ctx,
                ` + "`INSERT INTO t4 (id, i) VALUES ($1, $2) ON CONFLICT DO NOTHING`" + `,
                &r.ID, &r.I)
        if err != nil {
                return false, errors.WithStack(err)
        }
        rowsAffected, err := result.RowsAffected()
        if err != nil {
                return false, errors.WithStack(err)
        }
        return rowsAffected > 0, nil
}

// GetT4ByPkContext select the T4 from the database.
func GetT4ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T4, error) {
        var r T4
        err := db.QueryRowContext(ctx,
                ` + "`SELECT id, i FROM t4 WHERE id = $1 AND i = $2`" + `,
                pk0, pk1).Scan(&r.ID, &r.I)
        if err != nil {
                return nil, errors.WithStack(err)
        }
        return &r, nil
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.table.Name, func(t *testing.T) {
			st, err := PgTableToStruct(tt.table, &defaultTypeMapCfg, autoGenKeyCfg, []string{})
			if err != nil {
				t.Fatal(err)
			}
			src, err := PgExecuteDefaultMethodTmpl(&StructTmpl{Struct: st})
			if err != nil {
				t.Fatal(err)
			}

			re1 := regexp.MustCompile(`\s`)

			srcStr := string(src)
			if re1.ReplaceAllString(srcStr, "") != re1.ReplaceAllString(tt.expect, "") {
				t.Errorf("Expected generated code: %s, got: %s", tt.expect, srcStr)
			}
		})
	}
}

func TestPgCreateStruct(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()
	assert := assert.New(t)

	schema := "public"
	src, err := PgCreateStruct(conn, schema, "", "mypkg", "", []string{}, []string{}, []string{})
	if err != nil {
		t.Fatal(err)
	}

	expected := `// Code generated by dgw. DO NOT EDIT.

package mypkg

// T1 represents public.t1
type T1 struct {
	ID          int64          // id
	I           int            // i
	Str         string         // str
	NullableStr sql.NullString // nullable_str
	TWithTz     time.Time      // t_with_tz
	TWithoutTz  time.Time      // t_without_tz
	Tm          *time.Time     // tm
}
// Create inserts the T1 to the database.
func (r *T1) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
func GetT1ByPk(db Queryer, pk0 int64) (*T1, error) {
	return GetT1ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T1 to the database.
func (r *T1) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`" + `,
		&r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T1 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T1) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING RETURNING id`" + `,
		&r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT1ByPkContext select the T1 from the database.
func GetT1ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T1, error) {
	var r T1
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, nullable_str, t_with_tz, t_without_tz, tm FROM t1 WHERE id = $1`" + `,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T2 represents public.t2
type T2 struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}
// Create inserts the T2 to the database.
func (r *T2) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT2ByPk select the T2 from the database.
func GetT2ByPk(db Queryer, pk0 int64) (*T2, error) {
	return GetT2ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T2 to the database.
func (r *T2) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) RETURNING id`" + `,
		&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T2 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T2) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id`" + `,
		&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT2ByPkContext select the T2 from the database.
func GetT2ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T2, error) {
	var r T2
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1`" + `,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T3 represents public.t3
type T3 struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}
// Create inserts the T3 to the database.
func (r *T3) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT3ByPk select the T3 from the database.
func GetT3ByPk(db Queryer, pk0 int64, pk1 int) (*T3, error) {
	return GetT3ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T3 to the database.
func (r *T3) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) RETURNING id, i`" + `,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T3 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T3) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id, i`" + `,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT3ByPkContext select the T3 from the database.
func GetT3ByPkContext(ctx context.Context, db Queryer, pk0 int64, pk1 int) (*T3, error) {
	var r T3
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, t_with_tz, t_without_tz FROM t3 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T4 represents public.t4
type T4 struct {
	ID int // id
	I  int // i
}
// Create inserts the T4 to the database.
func (r *T4) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT4ByPk select the T4 from the database.
func GetT4ByPk(db Queryer, pk0 int, pk1 int) (*T4, error) {
	return GetT4ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T4 to the database.
func (r *T4) CreateContext(ctx context.Context, db Queryer) error {
	_, err := db.ExecContext(ctx,
		` + "`INSERT INTO t4 (id, i) VALUES ($1, $2)`" + `,
		&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T4 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T4) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	result, err := db.ExecContext(ctx,
		` + "`INSERT INTO t4 (id, i) VALUES ($1, $2) ON CONFLICT DO NOTHING`" + `,
		&r.ID, &r.I)
	if err != nil {
		return false, errors.WithStack(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, errors.WithStack(err)
	}
	return rowsAffected > 0, nil
}

// GetT4ByPkContext select the T4 from the database.
func GetT4ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T4, error) {
	var r T4
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t4 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T5 represents public.t5
type T5 struct {
	ID int // id
	I  int // i
}
// Create inserts the T5 to the database.
func (r *T5) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT5ByPk select the T5 from the database.
func GetT5ByPk(db Queryer, pk0 int, pk1 int) (*T5, error) {
	return GetT5ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T5 to the database.
func (r *T5) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t5 () VALUES () RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T5 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T5) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t5 () VALUES () ON CONFLICT DO NOTHING RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT5ByPkContext select the T5 from the database.
func GetT5ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T5, error) {
	var r T5
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t5 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T6 represents public.t6
type T6 struct {
	ID int // id
	I  int // i
}
// Create inserts the T6 to the database.
func (r *T6) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT6ByPk select the T6 from the database.
func GetT6ByPk(db Queryer, pk0 int, pk1 int) (*T6, error) {
	return GetT6ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T6 to the database.
func (r *T6) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t6 () VALUES () RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T6 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T6) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t6 () VALUES () ON CONFLICT DO NOTHING RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT6ByPkContext select the T6 from the database.
func GetT6ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T6, error) {
	var r T6
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t6 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
`

	assert.Equal(expected, string(src))
}

func TestPgCreateStructWithAutoGenKey(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()
	assert := assert.New(t)

	schema := "public"
	src, err := PgCreateStruct(conn, schema, "", "mypkg", "", []string{}, []string{"smallserial", "serial", "bigserial", "autogenuuid", "integer"}, []string{})
	if err != nil {
		t.Fatal(err)
	}

	expected := `// Code generated by dgw. DO NOT EDIT.

package mypkg

// T1 represents public.t1
type T1 struct {
	ID          int64          // id
	I           int            // i
	Str         string         // str
	NullableStr sql.NullString // nullable_str
	TWithTz     time.Time      // t_with_tz
	TWithoutTz  time.Time      // t_without_tz
	Tm          *time.Time     // tm
}
// Create inserts the T1 to the database.
func (r *T1) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
func GetT1ByPk(db Queryer, pk0 int64) (*T1, error) {
	return GetT1ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T1 to the database.
func (r *T1) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`" + `,
		&r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T1 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T1) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING RETURNING id`" + `,
		&r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT1ByPkContext select the T1 from the database.
func GetT1ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T1, error) {
	var r T1
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, nullable_str, t_with_tz, t_without_tz, tm FROM t1 WHERE id = $1`" + `,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T2 represents public.t2
type T2 struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}
// Create inserts the T2 to the database.
func (r *T2) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT2ByPk select the T2 from the database.
func GetT2ByPk(db Queryer, pk0 int64) (*T2, error) {
	return GetT2ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T2 to the database.
func (r *T2) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) RETURNING id`" + `,
		&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T2 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T2) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id`" + `,
		&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT2ByPkContext select the T2 from the database.
func GetT2ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T2, error) {
	var r T2
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1`" + `,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T3 represents public.t3
type T3 struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}
// Create inserts the T3 to the database.
func (r *T3) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT3ByPk select the T3 from the database.
func GetT3ByPk(db Queryer, pk0 int64, pk1 int) (*T3, error) {
	return GetT3ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T3 to the database.
func (r *T3) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) RETURNING id, i`" + `,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T3 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T3) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id, i`" + `,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT3ByPkContext select the T3 from the database.
func GetT3ByPkContext(ctx context.Context, db Queryer, pk0 int64, pk1 int) (*T3, error) {
	var r T3
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, t_with_tz, t_without_tz FROM t3 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T4 represents public.t4
type T4 struct {
	ID int // id
	I  int // i
}
// Create inserts the T4 to the database.
func (r *T4) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT4ByPk select the T4 from the database.
func GetT4ByPk(db Queryer, pk0 int, pk1 int) (*T4, error) {
	return GetT4ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T4 to the database.
func (r *T4) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t4 () VALUES () RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T4 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T4) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t4 () VALUES () ON CONFLICT DO NOTHING RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT4ByPkContext select the T4 from the database.
func GetT4ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T4, error) {
	var r T4
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t4 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T5 represents public.t5
type T5 struct {
	ID int // id
	I  int // i
}
// Create inserts the T5 to the database.
func (r *T5) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT5ByPk select the T5 from the database.
func GetT5ByPk(db Queryer, pk0 int, pk1 int) (*T5, error) {
	return GetT5ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T5 to the database.
func (r *T5) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t5 () VALUES () RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T5 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T5) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t5 () VALUES () ON CONFLICT DO NOTHING RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT5ByPkContext select the T5 from the database.
func GetT5ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T5, error) {
	var r T5
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t5 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T6 represents public.t6
type T6 struct {
	ID int // id
	I  int // i
}
// Create inserts the T6 to the database.
func (r *T6) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT6ByPk select the T6 from the database.
func GetT6ByPk(db Queryer, pk0 int, pk1 int) (*T6, error) {
	return GetT6ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T6 to the database.
func (r *T6) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t6 () VALUES () RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T6 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T6) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t6 () VALUES () ON CONFLICT DO NOTHING RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT6ByPkContext select the T6 from the database.
func GetT6ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T6, error) {
	var r T6
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t6 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
`

	assert.Equal(expected, string(src))
}

func TestPgCreateStructWithDeprecated(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()
	assert := assert.New(t)

	schema := "public"
	deprecated := []string{"t2", "t5"}
	src, err := PgCreateStruct(conn, schema, "", "mypkg", "", []string{}, []string{}, deprecated)
	if err != nil {
		t.Fatal(err)
	}

	expected := `// Code generated by dgw. DO NOT EDIT.

package mypkg

// T1 represents public.t1
type T1 struct {
	ID          int64          // id
	I           int            // i
	Str         string         // str
	NullableStr sql.NullString // nullable_str
	TWithTz     time.Time      // t_with_tz
	TWithoutTz  time.Time      // t_without_tz
	Tm          *time.Time     // tm
}
// Create inserts the T1 to the database.
func (r *T1) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
func GetT1ByPk(db Queryer, pk0 int64) (*T1, error) {
	return GetT1ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T1 to the database.
func (r *T1) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`" + `,
		&r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T1 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T1) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING RETURNING id`" + `,
		&r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT1ByPkContext select the T1 from the database.
func GetT1ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T1, error) {
	var r T1
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, nullable_str, t_with_tz, t_without_tz, tm FROM t1 WHERE id = $1`" + `,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T2 represents public.t2
// Deprecated: T2 is deprecated
type T2 struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}
// Create inserts the T2 to the database.
// Deprecated: T2 is deprecated
func (r *T2) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT2ByPk select the T2 from the database.
// Deprecated: T2 is deprecated
func GetT2ByPk(db Queryer, pk0 int64) (*T2, error) {
	return GetT2ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T2 to the database.
// Deprecated: T2 is deprecated
func (r *T2) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) RETURNING id`" + `,
		&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T2 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
// Deprecated: T2 is deprecated
func (r *T2) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id`" + `,
		&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT2ByPkContext select the T2 from the database.
// Deprecated: T2 is deprecated
func GetT2ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T2, error) {
	var r T2
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1`" + `,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T3 represents public.t3
type T3 struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}
// Create inserts the T3 to the database.
func (r *T3) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT3ByPk select the T3 from the database.
func GetT3ByPk(db Queryer, pk0 int64, pk1 int) (*T3, error) {
	return GetT3ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T3 to the database.
func (r *T3) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) RETURNING id, i`" + `,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T3 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T3) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id, i`" + `,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT3ByPkContext select the T3 from the database.
func GetT3ByPkContext(ctx context.Context, db Queryer, pk0 int64, pk1 int) (*T3, error) {
	var r T3
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, t_with_tz, t_without_tz FROM t3 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T4 represents public.t4
type T4 struct {
	ID int // id
	I  int // i
}
// Create inserts the T4 to the database.
func (r *T4) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT4ByPk select the T4 from the database.
func GetT4ByPk(db Queryer, pk0 int, pk1 int) (*T4, error) {
	return GetT4ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T4 to the database.
func (r *T4) CreateContext(ctx context.Context, db Queryer) error {
	_, err := db.ExecContext(ctx,
		` + "`INSERT INTO t4 (id, i) VALUES ($1, $2)`" + `,
		&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T4 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T4) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	result, err := db.ExecContext(ctx,
		` + "`INSERT INTO t4 (id, i) VALUES ($1, $2) ON CONFLICT DO NOTHING`" + `,
		&r.ID, &r.I)
	if err != nil {
		return false, errors.WithStack(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, errors.WithStack(err)
	}
	return rowsAffected > 0, nil
}

// GetT4ByPkContext select the T4 from the database.
func GetT4ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T4, error) {
	var r T4
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t4 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T5 represents public.t5
// Deprecated: T5 is deprecated
type T5 struct {
	ID int // id
	I  int // i
}
// Create inserts the T5 to the database.
// Deprecated: T5 is deprecated
func (r *T5) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT5ByPk select the T5 from the database.
// Deprecated: T5 is deprecated
func GetT5ByPk(db Queryer, pk0 int, pk1 int) (*T5, error) {
	return GetT5ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T5 to the database.
// Deprecated: T5 is deprecated
func (r *T5) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t5 () VALUES () RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T5 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
// Deprecated: T5 is deprecated
func (r *T5) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t5 () VALUES () ON CONFLICT DO NOTHING RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT5ByPkContext select the T5 from the database.
// Deprecated: T5 is deprecated
func GetT5ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T5, error) {
	var r T5
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t5 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
// T6 represents public.t6
type T6 struct {
	ID int // id
	I  int // i
}
// Create inserts the T6 to the database.
func (r *T6) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT6ByPk select the T6 from the database.
func GetT6ByPk(db Queryer, pk0 int, pk1 int) (*T6, error) {
	return GetT6ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T6 to the database.
func (r *T6) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t6 () VALUES () RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateOnConflictDoNothing inserts the T6 to the database.
// If a conflict occurs (e.g., unique constraint violation), the insert is skipped without error.
// Returns true if the row was inserted, false if it was skipped due to conflict.
func (r *T6) CreateOnConflictDoNothing(ctx context.Context, db Queryer) (bool, error) {
	err := db.QueryRowContext(ctx,
		` + "`INSERT INTO t6 () VALUES () ON CONFLICT DO NOTHING RETURNING id, i`" + `,
	).Scan(&r.ID, &r.I)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	// Row was successfully inserted
	return true, nil
}

// GetT6ByPkContext select the T6 from the database.
func GetT6ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T6, error) {
	var r T6
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i FROM t6 WHERE id = $1 AND i = $2`" + `,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
`

	assert.Equal(expected, string(src))
}
