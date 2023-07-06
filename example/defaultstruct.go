package dgwexample

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

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
		`INSERT INTO t1 (i, str, nullable_str, t_with_tz, t_without_tz, tm) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		&r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm).Scan(&r.ID)
	if err != nil {
		return errors.Wrap(err, "failed to insert t1")
	}
	return nil
}

// GetT1ByPkContext select the T1 from the database.
func GetT1ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T1, error) {
	var r T1
	err := db.QueryRowContext(ctx,
		`SELECT id, i, str, nullable_str, t_with_tz, t_without_tz, tm FROM t1 WHERE id = $1`,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.Tm)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select t1")
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
		`INSERT INTO t2 (i, str, t_with_tz, t_without_tz) VALUES ($1, $2, $3, $4) RETURNING id`,
		&r.I, &r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID)
	if err != nil {
		return errors.Wrap(err, "failed to insert t2")
	}
	return nil
}

// GetT2ByPkContext select the T2 from the database.
func GetT2ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T2, error) {
	var r T2
	err := db.QueryRowContext(ctx,
		`SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1`,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select t2")
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
		`INSERT INTO t3 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) RETURNING id, i`,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.Wrap(err, "failed to insert t3")
	}
	return nil
}

// GetT3ByPkContext select the T3 from the database.
func GetT3ByPkContext(ctx context.Context, db Queryer, pk0 int64, pk1 int) (*T3, error) {
	var r T3
	err := db.QueryRowContext(ctx,
		`SELECT id, i, str, t_with_tz, t_without_tz FROM t3 WHERE id = $1 AND i = $2`,
		pk0, pk1).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select t3")
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
		`INSERT INTO t4 (id, i) VALUES ($1, $2)`,
		&r.ID, &r.I)
	if err != nil {
		return errors.Wrap(err, "failed to insert t4")
	}
	return nil
}

// GetT4ByPkContext select the T4 from the database.
func GetT4ByPkContext(ctx context.Context, db Queryer, pk0 int, pk1 int) (*T4, error) {
	var r T4
	err := db.QueryRowContext(ctx,
		`SELECT id, i FROM t4 WHERE id = $1 AND i = $2`,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select t4")
	}
	return &r, nil
}

// T5 represents public.t5
type T5 struct {
	ID     int64  // id
	Select int    // select
	From   string // from
}

// Create inserts the T5 to the database.
func (r *T5) Create(db Queryer) error {
	return r.CreateContext(context.Background(), db)
}

// GetT5ByPk select the T5 from the database.
func GetT5ByPk(db Queryer, pk0 int64) (*T5, error) {
	return GetT5ByPkContext(context.Background(), db, pk0)
}

// CreateContext inserts the T5 to the database.
func (r *T5) CreateContext(ctx context.Context, db Queryer) error {
	err := db.QueryRowContext(ctx,
		`INSERT INTO t5 (select, from) VALUES ($1, $2) RETURNING id`,
		&r.Select, &r.From).Scan(&r.ID)
	if err != nil {
		return errors.Wrap(err, "failed to insert t5")
	}
	return nil
}

// GetT5ByPkContext select the T5 from the database.
func GetT5ByPkContext(ctx context.Context, db Queryer, pk0 int64) (*T5, error) {
	var r T5
	err := db.QueryRowContext(ctx,
		`SELECT id, select, from FROM t5 WHERE id = $1`,
		pk0).Scan(&r.ID, &r.Select, &r.From)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select t5")
	}
	return &r, nil
}
