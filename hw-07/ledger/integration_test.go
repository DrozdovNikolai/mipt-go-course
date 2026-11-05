//go:build integration

package ledger

import (
	"context"
	"errors"
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
	t.Setenv("DATABASE_URL", dsn)
	if err := Open(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(Close)

	var name string
	if err := database.QueryRow("SELECT current_database()").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(name, "_test") {
		t.Fatal("integration tests only clear databases with names ending in _test")
	}
	if _, err := database.Exec("TRUNCATE expenses, budgets RESTART IDENTITY"); err != nil {
		t.Fatal(err)
	}
}

func TestAddTransaction(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		setup(t)

		if err := SetBudget(Budget{Limit: 200, Category: "Food"}); err != nil {
			t.Fatalf("cannot set budget: %v", err)
		}

		tx := Transaction{Amount: 150, Category: "Food", Date: "2026-11-02"}
		if _, err := AddTransaction(tx); err != nil {
			t.Fatalf("unexpected error when adding transaction within budget: %v", err)
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		setup(t)

		if err := SetBudget(Budget{Limit: 100, Category: "Food"}); err != nil {
			t.Fatalf("cannot set budget: %v", err)
		}

		before, err := ListTransactions()
		if err != nil {
			t.Fatal(err)
		}
		tx := Transaction{Amount: 150, Category: "Food", Date: "2026-11-02"}
		_, err = AddTransaction(tx)

		if !errors.Is(err, ErrBudgetExceeded) {
			t.Errorf("want error %v, got %v", ErrBudgetExceeded, err)
		}
		after, err := ListTransactions()
		if err != nil {
			t.Fatal(err)
		}
		if len(after) != len(before) {
			t.Errorf("want %d transactions after rejection, got %d", len(before), len(after))
		}
	})
}
