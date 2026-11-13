package domain

import (
	"context"
	"errors"
	"fmt"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

type Validatable interface {
	Validate() error
}

var ErrBudgetExceeded = errors.New("budget exceeded")

func (t Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("transaction amount cannot be zero or negative")
	}
	if t.Category == "" {
		return errors.New("transaction category cannot be empty")
	}
	if t.Date == "" {
		return errors.New("transaction date cannot be empty")
	}
	return nil
}

func (b Budget) Validate() error {
	if b.Category == "" {
		return errors.New("budget category cannot be empty")
	}
	if b.Limit <= 0 {
		return errors.New("budget limit must be positive")
	}

	return nil
}

func CheckValid(v Validatable) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}

type BudgetRepository interface {
	SetBudget(Budget) error
	GetBudget(category string) (Budget, bool, error)
	ListBudgets() ([]Budget, error)
}

type TransactionRepository interface {
	AddTransaction(Transaction) (Transaction, error)
	ListTransactions() ([]Transaction, error)
	SumByCategory(category, period, date string) (float64, error)
	GetReportSummary(ctx context.Context, from, to string) (map[string]float64, error)
}

type ReportCache interface {
	GetReportSummary(ctx context.Context, from, to string) (map[string]float64, bool, error)
	SetReportSummary(ctx context.Context, from, to string, report map[string]float64) error
}
