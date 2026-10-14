// Package sqlite implements the persistence ports on top of SQLite.
package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (no CGO)
)

//go:embed migrations.sql
var migrationsSQL string

// singleUserID is the only user in the system. The schema already carries
// user_id everywhere so multi-user support will not need a migration.
const singleUserID int64 = 1

const timeLayout = time.RFC3339

// Open connects to the SQLite database at path and applies the schema.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}
	// SQLite does not handle concurrent writers well; a single connection
	// avoids "database is locked" errors.
	db.SetMaxOpenConns(1)

	for _, stmt := range strings.Split(migrationsSQL, ";") {
		if !hasSQL(stmt) {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("running migration: %w (statement: %s)", err, strings.TrimSpace(stmt))
		}
	}
	return db, nil
}

// hasSQL reports whether stmt contains anything besides whitespace and
// "--" comment lines.
func hasSQL(stmt string) bool {
	for _, line := range strings.Split(stmt, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			return true
		}
	}
	return false
}

func formatTime(t time.Time) string {
	return t.UTC().Format(timeLayout)
}

func formatNullableTime(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: formatTime(*t), Valid: true}
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(timeLayout, s)
}

func parseNullableTime(ns sql.NullString) (*time.Time, error) {
	if !ns.Valid || ns.String == "" {
		return nil, nil
	}
	t, err := parseTime(ns.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
