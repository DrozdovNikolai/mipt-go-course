package ledger

import (
	"context"
	"ledger/internal/cache"
	"ledger/internal/db"
	"ledger/internal/domain"
	"ledger/internal/repository/pg"
	"ledger/internal/service"
	"log"
)

// Public aliases let Gateway use Ledger without importing its internal packages.
type Transaction = domain.Transaction
type Budget = domain.Budget
type Service = service.Service
type ValidationError = service.ValidationError

var ErrBudgetExceeded = domain.ErrBudgetExceeded

func New(ctx context.Context) (Service, func(), error) {
	conn, err := db.Open(ctx)
	if err != nil {
		return nil, nil, err
	}
	reports, err := cache.Open(ctx)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	budgets := pg.NewBudgetRepository(conn)
	transactions := pg.NewTransactionRepository(conn)
	svc := service.New(budgets, transactions, reports)
	closeFn := func() {
		reports.Close()
		conn.Close()
	}
	log.Println("Ledger connected to PostgreSQL")
	log.Println("Ledger connected to Redis")
	return svc, closeFn, nil
}
