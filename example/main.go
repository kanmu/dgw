package main

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// T1 represents public.t1
type T1 struct {
	ID          int64          // id
	I           int            // i
	Str         string         // str
	NumFloat    float64        // num_float
	NullableStr sql.NullString // nullable_str
	TWithTz     time.Time      // t_with_tz
	TWithoutTz  time.Time      // t_without_tz
	NullableTz  *time.Time     // nullable_tz
	JSONData    []byte         // json_data
	XMLData     []byte         // xml_data
}

// Create inserts the T1 to the database.
func (r *T1) Create(db Queryer) error {
	err := db.QueryRow(
		`INSERT INTO t1 (i, str, num_float, nullable_str, t_with_tz, t_without_tz, nullable_tz, json_data, xml_data) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		&r.I, &r.Str, &r.NumFloat, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.NullableTz, &r.JSONData, &r.XMLData).Scan(&r.ID)
	if err != nil {
		return errors.Wrap(err, "failed to insert t1")
	}
	return nil
}

// GetT1ByPk select the T1 from the database.
func GetT1ByPk(db Queryer, pk0 int64) (*T1, error) {
	var r T1
	err := db.QueryRow(
		`SELECT id, i, str, num_float, nullable_str, t_with_tz, t_without_tz, nullable_tz, json_data, xml_data FROM t1 WHERE id = $1`,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.NumFloat, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.NullableTz, &r.JSONData, &r.XMLData)
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
	err := db.QueryRow(
		`INSERT INTO t2 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) RETURNING id, i`,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		return errors.Wrap(err, "failed to insert t2")
	}
	return nil
}

// GetT2ByPk select the T2 from the database.
func GetT2ByPk(db Queryer, pk0 int64, pk1 int) (*T2, error) {
	var r T2
	err := db.QueryRow(
		`SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1 AND i = $2`,
		pk0, pk1).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select t2")
	}
	return &r, nil
}

// T3 represents public.t3
type T3 struct {
	ID int // id
	I  int // i
}

// Create inserts the T3 to the database.
func (r *T3) Create(db Queryer) error {
	_, err := db.Exec(
		`INSERT INTO t3 (id, i) VALUES ($1, $2)`,
		&r.ID, &r.I)
	if err != nil {
		return errors.Wrap(err, "failed to insert t3")
	}
	return nil
}

// GetT3ByPk select the T3 from the database.
func GetT3ByPk(db Queryer, pk0 int, pk1 int) (*T3, error) {
	var r T3
	err := db.QueryRow(
		`SELECT id, i FROM t3 WHERE id = $1 AND i = $2`,
		pk0, pk1).Scan(&r.ID, &r.I)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select t3")
	}
	return &r, nil
}
