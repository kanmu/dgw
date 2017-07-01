package dgwexample

import (
	"database/sql"

	_ "github.com/lib/pq" // postgres
	"github.com/pkg/errors"
)

// OpenDB opens database connection
func OpenDB(connStr string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}
	return conn, nil
}
