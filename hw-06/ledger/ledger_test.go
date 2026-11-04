package ledger

import (
	"errors"
	"testing"
)

func TestTransactionValidate(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool // true — expected error, false — expected nil
	}{
		{
			name:    "Valid transaction",
			tx:      Transaction{Amount: 100, Category: "Food", Date: "2026-11-02"},
			wantErr: false,
		},
		{
			name:    "Zero amount",
			tx:      Transaction{Amount: 0, Category: "Food", Date: "2026-11-02"},
			wantErr: true,
		},
		{
			name:    "Negative amount",
			tx:      Transaction{Amount: -100, Category: "Food", Date: "2026-11-02"},
			wantErr: true,
		},
		{
			name:    "Empty category",
			tx:      Transaction{Amount: 100, Category: "", Date: "2026-11-02"},
			wantErr: true,
		},
		{
			name:    "Empty date",
			tx:      Transaction{Amount: 100, Category: "Food", Date: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.tx.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() returned %v; wanted err: %v", err, tt.wantErr)
			}
		})
	}
}

func TestBudgetValidate(t *testing.T) {
	tests := []struct {
		name    string
		budget  Budget
		wantErr bool // true — expected error, false — expected nil
	}{
		{
			name:    "Valid budget",
			budget:  Budget{Limit: 1000, Category: "Food"},
			wantErr: false,
		},
		{
			name:    "Zero amount",
			budget:  Budget{Limit: 0, Category: "Food"},
			wantErr: true,
		},
		{
			name:    "Negative amount",
			budget:  Budget{Limit: -1000, Category: "Food"},
			wantErr: true,
		},
		{
			name:    "Empty category",
			budget:  Budget{Limit: 1000, Category: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.budget.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() returned %v; wanted err: %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddTransaction(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		if err := SetBudget(Budget{Limit: 200, Category: "Food"}); err != nil {
			t.Fatalf("cannot set budget: %v", err)
		}

		tx := Transaction{Amount: 150, Category: "Food", Date: "2026-11-02"}
		if _, err := AddTransaction(tx); err != nil {
			t.Fatalf("unexpected error when adding transaction within budget: %v", err)
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		if err := SetBudget(Budget{Limit: 100, Category: "Food"}); err != nil {
			t.Fatalf("cannot set budget: %v", err)
		}

		before := len(ListTransactions())
		tx := Transaction{Amount: 150, Category: "Food", Date: "2026-11-02"}
		_, err := AddTransaction(tx)

		if !errors.Is(err, ErrBudgetExceeded) {
			t.Errorf("want error %v, got %v", ErrBudgetExceeded, err)
		}
		if after := len(ListTransactions()); after != before {
			t.Errorf("want %d transactions after rejection, got %d", before, after)
		}
	})
}
