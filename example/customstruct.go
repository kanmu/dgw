package dgwexample

import (
	"database/sql"
	"time"
)

// T1Table represents public.t1
type T1Table struct {
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
func (r *T1Table) Create(db Queryer) error {
	err := db.QueryRow(
		`INSERT INTO t1 (i, str, num_float, nullable_str, t_with_tz, t_without_tz, nullable_tz, json_data, xml_data) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		&r.I, &r.Str, &r.NumFloat, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.NullableTz, &r.JSONData, &r.XMLData).Scan(&r.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetT1TableByPk select the T1 from the database.
func GetT1TableByPk(db Queryer, pk0 int64) (*T1, error) {
	var r T1
	err := db.QueryRow(
		`SELECT id, i, str, num_float, nullable_str, t_with_tz, t_without_tz, nullable_tz, json_data, xml_data FROM t1 WHERE id = $1`,
		pk0).Scan(&r.ID, &r.I, &r.Str, &r.NumFloat, &r.NullableStr, &r.TWithTz, &r.TWithoutTz, &r.NullableTz, &r.JSONData, &r.XMLData)
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
		`INSERT INTO t2 (str, t_with_tz, t_without_tz) VALUES ($1, $2, $3) RETURNING id, i`,
		&r.Str, &r.TWithTz, &r.TWithoutTz).Scan(&r.ID, &r.I)
	if err != nil {
		return err
	}
	return nil
}

// GetT2TableByPk select the T2 from the database.
func GetT2TableByPk(db Queryer, pk0 int64, pk1 int) (*T2, error) {
	var r T2
	err := db.QueryRow(
		`SELECT id, i, str, t_with_tz, t_without_tz FROM t2 WHERE id = $1 AND i = $2`,
		pk0, pk1).Scan(&r.ID, &r.I, &r.Str, &r.TWithTz, &r.TWithoutTz)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// T3Table represents public.t3
type T3Table struct {
	ID int // id
	I  int // i
}

// Create inserts the T3 to the database.
func (r *T3Table) Create(db Queryer) error {
	_, err := db.Exec(
		`INSERT INTO t3 (id, i) VALUES ($1, $2)`,
		&r.ID, &r.I)
	if err != nil {
		return err
	}
	return nil
}

// GetT3TableByPk select the T3 from the database.
func GetT3TableByPk(db Queryer, pk0 int, pk1 int) (*T3, error) {
	var r T3
	err := db.QueryRow(
		`SELECT id, i FROM t3 WHERE id = $1 AND i = $2`,
		pk0, pk1).Scan(&r.ID, &r.I)
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

// UserAccountTable represents public.user_account
type UserAccountTable struct {
	ID        int64  // id
	Email     string // email
	LastName  string // last_name
	FirstName string // first_name
}

// Create inserts the UserAccount to the database.
func (r *UserAccountTable) Create(db Queryer) error {
	err := db.QueryRow(
		`INSERT INTO user_account (email, last_name, first_name) VALUES ($1, $2, $3) RETURNING id`,
		&r.Email, &r.LastName, &r.FirstName).Scan(&r.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetUserAccountTableByPk select the UserAccount from the database.
func GetUserAccountTableByPk(db Queryer, pk0 int64) (*UserAccount, error) {
	var r UserAccount
	err := db.QueryRow(
		`SELECT id, email, last_name, first_name FROM user_account WHERE id = $1`,
		pk0).Scan(&r.ID, &r.Email, &r.LastName, &r.FirstName)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// UserAccountCompositePkTable represents public.user_account_composite_pk
type UserAccountCompositePkTable struct {
	ID        int64  // id
	Email     string // email
	LastName  string // last_name
	FirstName string // first_name
}

// Create inserts the UserAccountCompositePk to the database.
func (r *UserAccountCompositePkTable) Create(db Queryer) error {
	_, err := db.Exec(
		`INSERT INTO user_account_composite_pk (id, email, last_name, first_name) VALUES ($1, $2, $3, $4)`,
		&r.ID, &r.Email, &r.LastName, &r.FirstName)
	if err != nil {
		return err
	}
	return nil
}

// GetUserAccountCompositePkTableByPk select the UserAccountCompositePk from the database.
func GetUserAccountCompositePkTableByPk(db Queryer, pk0 int64, pk1 string) (*UserAccountCompositePk, error) {
	var r UserAccountCompositePk
	err := db.QueryRow(
		`SELECT id, email, last_name, first_name FROM user_account_composite_pk WHERE id = $1 AND email = $2`,
		pk0, pk1).Scan(&r.ID, &r.Email, &r.LastName, &r.FirstName)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// UserAccountUUIDTable represents public.user_account_uuid
type UserAccountUUIDTable struct {
	UUID      string // uuid
	Email     string // email
	LastName  string // last_name
	FirstName string // first_name
}

// Create inserts the UserAccountUUID to the database.
func (r *UserAccountUUIDTable) Create(db Queryer) error {
	err := db.QueryRow(
		`INSERT INTO user_account_uuid (email, last_name, first_name) VALUES ($1, $2, $3) RETURNING uuid`,
		&r.Email, &r.LastName, &r.FirstName).Scan(&r.UUID)
	if err != nil {
		return err
	}
	return nil
}

// GetUserAccountUUIDTableByPk select the UserAccountUUID from the database.
func GetUserAccountUUIDTableByPk(db Queryer, pk0 string) (*UserAccountUUID, error) {
	var r UserAccountUUID
	err := db.QueryRow(
		`SELECT uuid, email, last_name, first_name FROM user_account_uuid WHERE uuid = $1`,
		pk0).Scan(&r.UUID, &r.Email, &r.LastName, &r.FirstName)
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
}
