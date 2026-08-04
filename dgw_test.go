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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg, []string{}, "", nil)
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg, []string{}, "", nil)
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg, []string{}, "", nil)
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
		st, err := PgTableToStruct(tbl, &defaultTypeMapCfg, autoGenKeyCfg, []string{}, "", nil)
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
// %EMPTY_COMMENT%
// Deprecated: Use CreateContext instead.
func (r *T1) Create(db Queryer) error {
        return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
//
// Deprecated: Use GetT1ByPkContext instead.
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
// %EMPTY_COMMENT%
// Deprecated: Use CreateContext instead.
func (r *T2) Create(db Queryer) error {
        return r.CreateContext(context.Background(), db)
}

// GetT2ByPk select the T2 from the database.
//
// Deprecated: Use GetT2ByPkContext instead.
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
// %EMPTY_COMMENT%
// Deprecated: Use CreateContext instead.
func (r *T3) Create(db Queryer) error {
        return r.CreateContext(context.Background(), db)
}

// GetT3ByPk select the T3 from the database.
//
// Deprecated: Use GetT3ByPkContext instead.
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
// %EMPTY_COMMENT%
// Deprecated: Use CreateContext instead.
func (r *T4) Create(db Queryer) error {
        return r.CreateContext(context.Background(), db)
}

// GetT4ByPk select the T4 from the database.
//
// Deprecated: Use GetT4ByPkContext instead.
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
			st, err := PgTableToStruct(tt.table, &defaultTypeMapCfg, autoGenKeyCfg, []string{}, "", nil)
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
	src, err := PgCreateStruct(conn, schema, "", "mypkg", "", []string{}, []string{}, []string{}, []string{}, "")
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
//
// Deprecated: Use CreateContext instead.
func (r *T1) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
//
// Deprecated: Use GetT1ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T2) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT2ByPk select the T2 from the database.
//
// Deprecated: Use GetT2ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T3) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT3ByPk select the T3 from the database.
//
// Deprecated: Use GetT3ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T4) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT4ByPk select the T4 from the database.
//
// Deprecated: Use GetT4ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T5) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT5ByPk select the T5 from the database.
//
// Deprecated: Use GetT5ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T6) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT6ByPk select the T6 from the database.
//
// Deprecated: Use GetT6ByPkContext instead.
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
	src, err := PgCreateStruct(conn, schema, "", "mypkg", "", []string{}, []string{}, []string{"smallserial", "serial", "bigserial", "autogenuuid", "integer"}, []string{}, "")
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
//
// Deprecated: Use CreateContext instead.
func (r *T1) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
//
// Deprecated: Use GetT1ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T2) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT2ByPk select the T2 from the database.
//
// Deprecated: Use GetT2ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T3) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT3ByPk select the T3 from the database.
//
// Deprecated: Use GetT3ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T4) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT4ByPk select the T4 from the database.
//
// Deprecated: Use GetT4ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T5) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT5ByPk select the T5 from the database.
//
// Deprecated: Use GetT5ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T6) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT6ByPk select the T6 from the database.
//
// Deprecated: Use GetT6ByPkContext instead.
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
	src, err := PgCreateStruct(conn, schema, "", "mypkg", "", []string{}, []string{}, []string{}, deprecated, "")
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
//
// Deprecated: Use CreateContext instead.
func (r *T1) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
//
// Deprecated: Use GetT1ByPkContext instead.
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
//
// Deprecated: T2 is no longer maintained
type T2 struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}
// Create inserts the T2 to the database.
//
// Deprecated: T2 is no longer maintained
func (r *T2) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT2ByPk select the T2 from the database.
//
// Deprecated: T2 is no longer maintained
func GetT2ByPk(db Queryer, pk0 int64) (*T2, error) {
	return GetT2ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T2 to the database.
//
// Deprecated: T2 is no longer maintained
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
//
// Deprecated: T2 is no longer maintained
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
//
// Deprecated: T2 is no longer maintained
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
//
// Deprecated: Use CreateContext instead.
func (r *T3) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT3ByPk select the T3 from the database.
//
// Deprecated: Use GetT3ByPkContext instead.
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
//
// Deprecated: Use CreateContext instead.
func (r *T4) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT4ByPk select the T4 from the database.
//
// Deprecated: Use GetT4ByPkContext instead.
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
//
// Deprecated: T5 is no longer maintained
type T5 struct {
	ID int // id
	I  int // i
}
// Create inserts the T5 to the database.
//
// Deprecated: T5 is no longer maintained
func (r *T5) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT5ByPk select the T5 from the database.
//
// Deprecated: T5 is no longer maintained
func GetT5ByPk(db Queryer, pk0 int, pk1 int) (*T5, error) {
	return GetT5ByPkContext(context.Background(), db, pk0, pk1)
}

// CreateContext inserts the T5 to the database.
//
// Deprecated: T5 is no longer maintained
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
//
// Deprecated: T5 is no longer maintained
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
//
// Deprecated: T5 is no longer maintained
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
//
// Deprecated: Use CreateContext instead.
func (r *T6) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT6ByPk select the T6 from the database.
//
// Deprecated: Use GetT6ByPkContext instead.
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

func TestPgCreateStructWithQueryer(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()
	assert := assert.New(t)

	schema := "public"
	deprecated := []string{"t2", "t5"}
	src, err := PgCreateStruct(conn, schema, "", "mypkg", "", []string{}, []string{}, []string{}, deprecated, "MyQueryer")
	if err != nil {
		t.Fatal(err)
	}

	expected := `
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
//
// Deprecated: Use CreateContext instead.
func (r *T1) Create(db MyQueryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT1ByPk select the T1 from the database.
//
// Deprecated: Use GetT1ByPkContext instead.
func GetT1ByPk(db MyQueryer, pk0 int64) (*T1, error) {
	return GetT1ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T1 to the database.
func (r *T1) CreateContext(ctx context.Context, db MyQueryer) error {
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
func (r *T1) CreateOnConflictDoNothing(ctx context.Context, db MyQueryer) (bool, error) {
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
func GetT1ByPkContext(ctx context.Context, db MyQueryer, pk0 int64) (*T1, error) {
	var r T1
	err := db.QueryRowContext(ctx,
		` + "`SELECT id, i, str, nullable_str, t_with_tz, t_without_tz, tm FROM t1 WHERE id = $1`" + `,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &r, nil
}
`

	assert.Contains(string(src), expected)
}

func TestNewExcludeColumns(t *testing.T) {
	assert := assert.New(t)

	ex, err := NewExcludeColumns([]string{"t1.str", "t1.tm", "t2.str"})
	if err != nil {
		t.Fatal(err)
	}
	assert.True(ex.Contains("t1", "str"))
	assert.True(ex.Contains("t1", "tm"))
	assert.True(ex.Contains("t2", "str"))
	// exclusion must not cross the table boundary
	assert.False(ex.Contains("t2", "tm"))
	assert.False(ex.Contains("t3", "str"))
	assert.False(ex.Contains("t1", "id"))

	// no exclusion at all
	ex, err = NewExcludeColumns([]string{})
	if err != nil {
		t.Fatal(err)
	}
	assert.False(ex.Contains("t1", "str"))

	var nilEx *ExcludeColumns
	assert.False(nilEx.Contains("t1", "str"))

	// a column name without a table name is not allowed
	for _, spec := range []string{"str", "t1.", ".str", "", "."} {
		if _, err := NewExcludeColumns([]string{spec}); err == nil {
			t.Errorf("expected error for spec %q, got nil", spec)
		}
	}
}

func TestExcludeColumnsValidate(t *testing.T) {
	tbls := []*PgTable{
		{
			Schema: "public",
			Name:   "t1",
			Columns: []*PgColumn{
				{Name: "id", IsPrimaryKey: true},
				{Name: "str"},
				{Name: "tm"},
			},
		},
		{
			Schema: "public",
			Name:   "t_nopk",
			Columns: []*PgColumn{
				{Name: "a"},
				{Name: "b"},
			},
		},
	}

	tests := []struct {
		name   string
		specs  []string
		errStr string
	}{
		{"valid", []string{"t1.str", "t1.tm", "t_nopk.a"}, ""},
		{"unknown table", []string{"t2.str"}, "no such table t2"},
		{"unknown column", []string{"t1.nosuch"}, "no such column t1.nosuch"},
		{"primary key", []string{"t1.id"}, "cannot exclude primary key column t1.id"},
		{"all columns", []string{"t_nopk.a", "t_nopk.b"}, "all columns of t_nopk are excluded"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex, err := NewExcludeColumns(tt.specs)
			if err != nil {
				t.Fatal(err)
			}
			err = ex.Validate(tbls)
			if tt.errStr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.errStr)
		})
	}

	var nilEx *ExcludeColumns
	assert.NoError(t, nilEx.Validate(tbls))
}

func TestPgCreateStructWithExcludeColumn(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()
	assert := assert.New(t)

	schema := "public"
	exCols := []string{"t1.nullable_str", "t1.tm"}
	src, err := PgCreateStruct(conn, schema, "", "mypkg", "", []string{}, exCols, []string{}, []string{}, "")
	if err != nil {
		t.Fatal(err)
	}
	srcStr := string(src)

	// compare ignoring gofmt alignment
	re := regexp.MustCompile(`\s`)
	expectedStruct := `// T1 represents public.t1
type T1 struct {
	ID int64 // id
	I int // i
	Str string // str
	TWithTz time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}`
	assert.Contains(re.ReplaceAllString(srcStr, ""), re.ReplaceAllString(expectedStruct, ""))

	// the excluded columns are gone from the generated SQL and the scan/param lists
	assert.Contains(srcStr, "INSERT INTO t1 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) RETURNING id")
	assert.Contains(srcStr, "SELECT id, i, str, t_with_tz, t_without_tz FROM t1 WHERE id = $1")
	assert.Contains(srcStr, "&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)")
	assert.Contains(srcStr, "pk0).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)")
	assert.NotContains(srcStr, "nullable_str")
	assert.NotContains(srcStr, "NullableStr")
	assert.NotContains(srcStr, "&r.Tm")

	// other tables are not affected
	assert.Contains(srcStr, "INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) RETURNING id")
	assert.Contains(srcStr, "SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1")
}

func TestPgCreateStructWithInvalidExcludeColumn(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	tests := []struct {
		name   string
		exCols []string
		errStr string
	}{
		{"primary key", []string{"t1.id"}, "cannot exclude primary key column t1.id"},
		{"unknown column", []string{"t1.nosuch"}, "no such column t1.nosuch"},
		{"unknown table", []string{"nosuch.str"}, "no such table nosuch"},
		{"without table name", []string{"nullable_str"}, `invalid exclude column "nullable_str"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PgCreateStruct(conn, "public", "", "mypkg", "", []string{}, tt.exCols, []string{}, []string{}, "")
			assert.ErrorContains(t, err, tt.errStr)
		})
	}
}
