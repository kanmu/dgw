package dgwexample

import (
	"database/sql"
	"io/ioutil"
	"testing"
	"time"
)

func testPgSetup(t *testing.T) (*sql.DB, func()) {
	conn, err := sql.Open("postgres", "user=dgw_test dbname=dgw_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	setupSQL, err := ioutil.ReadFile("./ddl.sql")
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

func TestT1(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	now := time.Now()
	t1 := T1{
		I:           100,
		NullableStr: sql.NullString{String: "test"},
		Str:         "test",
		TWithTz:     now.AddDate(0, 0, 7),
		TWithoutTz:  now.AddDate(0, 0, 7),
	}
	if err := t1.Create(conn); err != nil {
		t.Fatal(err)
	}
	if t1.ID == 0 {
		t.Errorf("want other than zero")
	}
	target, err := GetT1ByPk(conn, t1.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", target)
}
