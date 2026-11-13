//go:build integration

package ledger

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
)

func setup(t *testing.T) Service {
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
	svc, closeFn, err := New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closeFn)
	return svc
}

func TestAddTransaction(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		svc := setup(t)

		if err := svc.SetBudget(Budget{Limit: 200, Category: "Food"}); err != nil {
			t.Fatalf("cannot set budget: %v", err)
		}

		tx := Transaction{Amount: 150, Category: "Food", Date: "2026-11-02"}
		if _, err := svc.AddTransaction(tx); err != nil {
			t.Fatalf("unexpected error when adding transaction within budget: %v", err)
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		svc := setup(t)

		if err := svc.SetBudget(Budget{Limit: 100, Category: "Food"}); err != nil {
			t.Fatalf("cannot set budget: %v", err)
		}

		before, err := svc.ListTransactions()
		if err != nil {
			t.Fatal(err)
		}
		tx := Transaction{Amount: 150, Category: "Food", Date: "2026-11-02"}
		_, err = svc.AddTransaction(tx)

		if !errors.Is(err, ErrBudgetExceeded) {
			t.Errorf("want error %v, got %v", ErrBudgetExceeded, err)
		}
		after, err := svc.ListTransactions()
		if err != nil {
			t.Fatal(err)
		}
		if len(after) != len(before) {
			t.Errorf("want %d transactions after rejection, got %d", len(before), len(after))
		}
	})
}
