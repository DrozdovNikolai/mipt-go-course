package pg

import (
	"context"
	"database/sql"
	"ledger/internal/domain"
	"time"
)

type BudgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) SetBudget(budget domain.Budget) error {
	_, err := r.db.Exec(`INSERT INTO budgets(category, limit_amount, period) VALUES($1,$2,$3)
		ON CONFLICT(category) DO UPDATE SET limit_amount=EXCLUDED.limit_amount, period=EXCLUDED.period`,
		budget.Category, budget.Limit, budget.Period)
	return err
}

func (r *BudgetRepository) GetBudget(category string) (domain.Budget, bool, error) {
	budget := domain.Budget{Category: category}
	err := r.db.QueryRow("SELECT limit_amount, period FROM budgets WHERE category=$1", category).Scan(&budget.Limit, &budget.Period)
	if err == sql.ErrNoRows {
		return domain.Budget{}, false, nil
	}
	if err != nil {
		return domain.Budget{}, false, err
	}
	return budget, true, nil
}

func (r *BudgetRepository) ListBudgets() ([]domain.Budget, error) {
	rows, err := r.db.Query("SELECT category, limit_amount, period FROM budgets ORDER BY category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	budgets := make([]domain.Budget, 0)
	for rows.Next() {
		var budget domain.Budget
		if err := rows.Scan(&budget.Category, &budget.Limit, &budget.Period); err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}
	return budgets, rows.Err()
}

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) AddTransaction(tx domain.Transaction) (domain.Transaction, error) {
	err := r.db.QueryRow(`INSERT INTO expenses(amount, category, description, date)
		VALUES($1,$2,$3,$4) RETURNING id`, tx.Amount, tx.Category, tx.Description, tx.Date).Scan(&tx.ID)
	if err != nil {
		return domain.Transaction{}, err
	}
	return tx, nil
}

func (r *TransactionRepository) ListTransactions() ([]domain.Transaction, error) {
	rows, err := r.db.Query(`SELECT id, amount, category, COALESCE(description,''), date
		FROM expenses ORDER BY date DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]domain.Transaction, 0)
	for rows.Next() {
		var tx domain.Transaction
		var date time.Time
		if err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &date); err != nil {
			return nil, err
		}
		tx.Date = date.Format("2006-01-02")
		transactions = append(transactions, tx)
	}
	return transactions, rows.Err()
}

func (r *TransactionRepository) SumByCategory(category, period, date string) (float64, error) {
	query := "SELECT COALESCE(SUM(amount),0) FROM expenses WHERE category=$1"
	args := []any{category}
	// Keep the monthly and yearly rules from homework 06.
	if period == "monthly" {
		query += " AND date_trunc('month', date) = date_trunc('month', $2::date)"
		args = append(args, date)
	} else if period == "yearly" {
		query += " AND date_trunc('year', date) = date_trunc('year', $2::date)"
		args = append(args, date)
	} else if period != "" {
		query += " AND FALSE"
	}
	var spent float64
	err := r.db.QueryRow(query, args...).Scan(&spent)
	return spent, err
}

func (r *TransactionRepository) GetReportSummary(ctx context.Context, from, to string) (map[string]float64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT category, SUM(amount) FROM expenses
		WHERE date >= $1::date AND date <= $2::date GROUP BY category`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := make(map[string]float64)
	for rows.Next() {
		var category string
		var amount float64
		if err := rows.Scan(&category, &amount); err != nil {
			return nil, err
		}
		report[category] = amount
	}
	return report, rows.Err()
}
