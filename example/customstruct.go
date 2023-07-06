package dgwexample

import (
	"context"
	"database/sql"
	"time"
)

// T1Table represents public.t1
type T1Table struct {
	ID          int64          // id
	I           int            // i
	Str         string         // str
	NullableStr sql.NullString // nullable_str
	TWithTz     time.Time      // t_with_tz
	TWithoutTz  time.Time      // t_without_tz
	Tm          *time.Time     // tm
}

// Create inserts the T1 to the database.
func (r *T1Table) Create(db Queryer) error {
	err := db.QueryRow(
		`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		&r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetT1TableByPk select the T1 from the database.
func GetT1TableByPk(db Queryer, pk0 int64) (*T1, error) {
	var r T1
	err := db.QueryRow(
		`SELECT id, i, str, nullable_str, t_with_tz, t_without_tz, tm FROM t1 WHERE id = $1`,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// T2Table represents public.t2
type T2Table struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}

// Create inserts the T2 to the database.
func (r *T2Table) Create(db Queryer) error {
	err := db.QueryRow(
		`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) RETURNING id`,
		&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetT2TableByPk select the T2 from the database.
func GetT2TableByPk(db Queryer, pk0 int64) (*T2, error) {
	var r T2
	err := db.QueryRow(
		`SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1`,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// T3Table represents public.t3
type T3Table struct {
	ID         int64     // id
	I          int       // i
	Str        string    // str
	TWithTz    time.Time // t_with_tz
	TWithoutTz time.Time // t_without_tz
}

// Create inserts the T3 to the database.
func (r *T3Table) Create(db Queryer) error {
	err := db.QueryRow(
		`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) RETURNING id, i`,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		return err
	}
	return nil
}

// GetT3TableByPk select the T3 from the database.
func GetT3TableByPk(db Queryer, pk0 int64, pk1 int) (*T3, error) {
	var r T3
	err := db.QueryRow(
		`SELECT id, i, str, t_with_tz, t_without_tz FROM t3 WHERE id = $1 AND i = $2`,
		pk0, pk1).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// T4Table represents public.t4
type T4Table struct {
	ID int // id
	I  int // i
}

// Create inserts the T4 to the database.
func (r *T4Table) Create(db Queryer) error {
	_, err := db.Exec(
		`INSERT INTO t4 (id, i) VALUES ($1, $2)`,
		&r.ID, &r.I)
	if err != nil {
		return err
	}
	return nil
}

// GetT4TableByPk select the T4 from the database.
func GetT4TableByPk(db Queryer, pk0 int, pk1 int) (*T4, error) {
	var r T4
	err := db.QueryRow(
		`SELECT id, i FROM t4 WHERE id = $1 AND i = $2`,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// T5Table represents public.t5
type T5Table struct {
	ID     int64  // id
	Select int    // select
	From   string // from
}

// Create inserts the T5 to the database.
func (r *T5Table) Create(db Queryer) error {
	err := db.QueryRow(
		`INSERT INTO t5 (select, from) VALUES ($1, $2) RETURNING id`,
		&r.Select, &r.From).Scan(&r.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetT5TableByPk select the T5 from the database.
func GetT5TableByPk(db Queryer, pk0 int64) (*T5, error) {
	var r T5
	err := db.QueryRow(
		`SELECT id, select, from FROM t5 WHERE id = $1`,
		pk0).Scan(&r.ID, &r.Select, &r.From)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Queryer database/sql compatible query interface
type Queryer interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
