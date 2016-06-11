package tdgw

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func testPgSetup(t *testing.T) {
	conn, err := sql.Open("postgres", "")
	if err != nil {
		t.Fatal(err)
	}
}
