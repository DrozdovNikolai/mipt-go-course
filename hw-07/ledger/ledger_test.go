package ledger

import (
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
