package ledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"ledger/internal/cache"
	"ledger/internal/db"
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

var database *sql.DB
var reportCache *redis.Client

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

func AddTransaction(tx Transaction) (Transaction, error) {
	if err := tx.Validate(); err != nil {
		return Transaction{}, fmt.Errorf("invalid transaction: %w", err)
	}

	var budget Budget
	err := database.QueryRow("SELECT limit_amount, period FROM budgets WHERE category=$1", tx.Category).Scan(&budget.Limit, &budget.Period)
	if err != nil && err != sql.ErrNoRows {
		return Transaction{}, err
	}
	if err == nil {
		query := "SELECT COALESCE(SUM(amount),0) FROM expenses WHERE category=$1"
		args := []any{tx.Category}
		// Keep the monthly and yearly rules from homework 06.
		if budget.Period == "monthly" {
			query += " AND date_trunc('month', date) = date_trunc('month', $2::date)"
			args = append(args, tx.Date)
		} else if budget.Period == "yearly" {
			query += " AND date_trunc('year', date) = date_trunc('year', $2::date)"
			args = append(args, tx.Date)
		} else if budget.Period != "" {
			query += " AND FALSE"
		}
		var spent float64
		if err := database.QueryRow(query, args...).Scan(&spent); err != nil {
			return Transaction{}, err
		}
		if spent+tx.Amount > budget.Limit {
			return Transaction{}, ErrBudgetExceeded
		}
	}

	err = database.QueryRow(`INSERT INTO expenses(amount, category, description, date)
		VALUES($1,$2,$3,$4) RETURNING id`, tx.Amount, tx.Category, tx.Description, tx.Date).Scan(&tx.ID)
	if err != nil {
		return Transaction{}, err
	}
	return tx, nil
}

func ListTransactions() ([]Transaction, error) {
	rows, err := database.Query(`SELECT id, amount, category, COALESCE(description,''), date
		FROM expenses ORDER BY date DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var tx Transaction
		var date time.Time
		if err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &date); err != nil {
			return nil, err
		}
		tx.Date = date.Format("2006-01-02")
		transactions = append(transactions, tx)
	}
	return transactions, rows.Err()
}

func ListBudgets() ([]Budget, error) {
	rows, err := database.Query("SELECT category, limit_amount, period FROM budgets ORDER BY category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	budgets := make([]Budget, 0)
	for rows.Next() {
		var budget Budget
		if err := rows.Scan(&budget.Category, &budget.Limit, &budget.Period); err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}
	return budgets, rows.Err()
}

func SetBudget(budget Budget) error {
	if err := budget.Validate(); err != nil {
		return fmt.Errorf("invalid budget: %w", err)
	}
	_, err := database.Exec(`INSERT INTO budgets(category, limit_amount, period) VALUES($1,$2,$3)
		ON CONFLICT(category) DO UPDATE SET limit_amount=EXCLUDED.limit_amount, period=EXCLUDED.period`,
		budget.Category, budget.Limit, budget.Period)
	return err
}

func LoadBudgets(r io.Reader) error {
	fmt.Println("Loading budgets from JSON file")
	var b []Budget
	err := json.NewDecoder(r).Decode(&b)
	if err != nil {
		return fmt.Errorf("failed to decode budgets JSON: %w", err)
	}
	for _, budget := range b {
		if err := budget.Validate(); err != nil {
			return fmt.Errorf("invalid budget: %w", err)
		}
		if err := SetBudget(budget); err != nil {
			return fmt.Errorf("failed to set budget: %w", err)
		}
	}
	return nil
}

func CheckValid(v Validatable) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}

func Open(ctx context.Context) error {
	conn, err := db.Open()
	if err != nil {
		return err
	}
	client, err := cache.Open(ctx)
	if err != nil {
		conn.Close()
		return err
	}
	database = conn
	reportCache = client
	log.Println("Ledger connected to PostgreSQL")
	log.Println("Ledger connected to Redis")
	return nil
}

func Close() {
	reportCache.Close()
	database.Close()
}

func GetReportSummary(ctx context.Context, from, to string) (map[string]float64, error) {
	key := "report:summary:" + from + ":" + to
	data, err := reportCache.Get(ctx, key).Bytes()
	if err == nil {
		var report map[string]float64
		if err := json.Unmarshal(data, &report); err != nil {
			return nil, err
		}
		log.Println("Report cache hit:", key)
		return report, nil
	}
	if err != redis.Nil {
		return nil, err
	}

	log.Println("Report cache miss:", key)
	rows, err := database.QueryContext(ctx, `SELECT category, SUM(amount) FROM expenses
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	data, err = json.Marshal(report)
	if err != nil {
		return nil, err
	}
	if err := reportCache.Set(ctx, key, data, 30*time.Second).Err(); err != nil {
		return nil, err
	}
	return report, nil
}
