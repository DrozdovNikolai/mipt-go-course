//go:build integration

package main

import (
	"context"
	"database/sql"
	"ledger"
	"os"
	"strings"
	"testing"
)

func setup(t *testing.T) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("set TEST_DATABASE_URL to a migrated database with a name ending in _test")
	}
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var name string
	if err := conn.QueryRow("SELECT current_database()").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(name, "_test") {
		t.Fatal("integration tests only clear databases with names ending in _test")
	}
	if _, err := conn.Exec("TRUNCATE expenses, budgets RESTART IDENTITY"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATABASE_URL", dsn)
	if err := ledger.Open(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ledger.Close)
}
