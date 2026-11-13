package main

import (
	"context"
	"io"
	"ledger"
	"testing"
)

// This stub controls service responses so HTTP tests can check status codes and DTOs.
// The real budget rule is checked by Ledger's tests.
type stubService struct {
	budget            *ledger.Budget
	transactions      []ledger.Transaction
	setBudgetErr      error
	addTransactionErr error
}

func setup(t *testing.T) (*handlers, *stubService) {
	t.Helper()
	svc := &stubService{}
	return newHandlers(svc), svc
}

func (s *stubService) SetBudget(b ledger.Budget) error {
	if s.setBudgetErr != nil {
		return s.setBudgetErr
	}
	s.budget = &b
	return nil
}

func (s *stubService) ListBudgets() ([]ledger.Budget, error) {
	if s.budget == nil {
		return []ledger.Budget{}, nil
	}
	return []ledger.Budget{*s.budget}, nil
}

func (s *stubService) AddTransaction(tx ledger.Transaction) (ledger.Transaction, error) {
	if s.addTransactionErr != nil {
		return ledger.Transaction{}, s.addTransactionErr
	}
	tx.ID = len(s.transactions) + 1
	s.transactions = append(s.transactions, tx)
	return tx, nil
}

func (s *stubService) ListTransactions() ([]ledger.Transaction, error) {
	return s.transactions, nil
}

func (s *stubService) GetReportSummary(context.Context, string, string) (map[string]float64, error) {
	return map[string]float64{}, nil
}

func (s *stubService) LoadBudgets(io.Reader) error {
	return nil
}
