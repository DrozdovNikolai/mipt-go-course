package api

import "ledger"

type CreateTransactionRequest struct {
	Amount      float64
	Category    string
	Description string
	Date        string
}

type TransactionResponse struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

type CreateBudgetRequest struct {
	Category string
	Limit    float64
	Period   string
}

type BudgetResponse struct {
	Category string
	Limit    float64
	Period   string
}

func TransactionToResponse(t *ledger.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          t.ID,
		Amount:      t.Amount,
		Category:    t.Category,
		Description: t.Description,
		Date:        t.Date,
	}
}

func TransactionFromRequest(r *CreateTransactionRequest) ledger.Transaction {
	return ledger.Transaction{
		Amount:      r.Amount,
		Category:    r.Category,
		Description: r.Description,
		Date:        r.Date,
	}
}

func BudgetFromRequest(r *CreateBudgetRequest) ledger.Budget {
	return ledger.Budget{
		Category: r.Category,
		Limit:    r.Limit,
		Period:   r.Period,
	}
}

func BudgetToResponse(b *ledger.Budget) BudgetResponse {
	return BudgetResponse{
		Category: b.Category,
		Limit:    b.Limit,
		Period:   b.Period,
	}
}
