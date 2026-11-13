package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"ledger/internal/domain"
	"log"
)

type Service interface {
	SetBudget(domain.Budget) error
	ListBudgets() ([]domain.Budget, error)
	AddTransaction(domain.Transaction) (domain.Transaction, error)
	ListTransactions() ([]domain.Transaction, error)
	GetReportSummary(ctx context.Context, from, to string) (map[string]float64, error)
	LoadBudgets(io.Reader) error
}

// ValidationError lets Gateway preserve its existing 400 responses.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type ledgerService struct {
	budgets      domain.BudgetRepository
	transactions domain.TransactionRepository
	reports      domain.ReportCache
}

func New(budgets domain.BudgetRepository, transactions domain.TransactionRepository, reports domain.ReportCache) Service {
	return &ledgerService{budgets: budgets, transactions: transactions, reports: reports}
}

func (s *ledgerService) SetBudget(budget domain.Budget) error {
	if err := budget.Validate(); err != nil {
		return &ValidationError{Message: err.Error()}
	}
	return s.budgets.SetBudget(budget)
}

func (s *ledgerService) ListBudgets() ([]domain.Budget, error) {
	return s.budgets.ListBudgets()
}

func (s *ledgerService) AddTransaction(tx domain.Transaction) (domain.Transaction, error) {
	if err := tx.Validate(); err != nil {
		return domain.Transaction{}, &ValidationError{Message: err.Error()}
	}
	budget, exists, err := s.budgets.GetBudget(tx.Category)
	if err != nil {
		return domain.Transaction{}, err
	}
	if exists {
		spent, err := s.transactions.SumByCategory(tx.Category, budget.Period, tx.Date)
		if err != nil {
			return domain.Transaction{}, err
		}
		if spent+tx.Amount > budget.Limit {
			return domain.Transaction{}, domain.ErrBudgetExceeded
		}
	}
	return s.transactions.AddTransaction(tx)
}

func (s *ledgerService) ListTransactions() ([]domain.Transaction, error) {
	return s.transactions.ListTransactions()
}

func (s *ledgerService) GetReportSummary(ctx context.Context, from, to string) (map[string]float64, error) {
	report, found, err := s.reports.GetReportSummary(ctx, from, to)
	if err != nil {
		return nil, err
	}
	if found {
		log.Println("Report cache hit:", from, to)
		return report, nil
	}
	log.Println("Report cache miss:", from, to)
	report, err = s.transactions.GetReportSummary(ctx, from, to)
	if err != nil {
		return nil, err
	}
	if err := s.reports.SetReportSummary(ctx, from, to, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (s *ledgerService) LoadBudgets(r io.Reader) error {
	fmt.Println("Loading budgets from JSON file")
	var b []domain.Budget
	err := json.NewDecoder(r).Decode(&b)
	if err != nil {
		return fmt.Errorf("failed to decode budgets JSON: %w", err)
	}
	for _, budget := range b {
		if err := budget.Validate(); err != nil {
			return fmt.Errorf("invalid budget: %w", err)
		}
		if err := s.SetBudget(budget); err != nil {
			return fmt.Errorf("failed to set budget: %w", err)
		}
	}
	return nil
}
