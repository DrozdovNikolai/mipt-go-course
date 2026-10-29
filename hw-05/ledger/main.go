package ledger

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

var transactions = make([]Transaction, 0)

var budgets = make(map[string]Budget)

func AddTransaction(tx Transaction) (Transaction, error) {

	el, exists := budgets[tx.Category]

	if err := tx.Validate(); err != nil {
		return Transaction{}, fmt.Errorf("invalid transaction: %w", err)
	}

	if exists {

		total := 0.0
		for _, t := range transactions {
			if t.Category == tx.Category &&
				(el.Period == "" || (el.Period == "monthly" && t.Date[:7] == tx.Date[:7]) || (el.Period == "yearly" && t.Date[:4] == tx.Date[:4])) {
				total += t.Amount
			}

		}
		if total+tx.Amount > budgets[tx.Category].Limit {
			return Transaction{}, ErrBudgetExceeded
		}
	}

	tx.ID = len(transactions) + 1

	transactions = append(transactions, tx)

	return tx, nil
}

func ListTransactions() []Transaction {
	result := make([]Transaction, len(transactions))
	copy(result, transactions)

	return result
}

func ListBudgets() []Budget {
	result := make([]Budget, 0, len(budgets))
	for _, budget := range budgets {
		result = append(result, budget)
	}
	return result
}

func SetBudget(budget Budget) error {
	if err := budget.Validate(); err != nil {
		return fmt.Errorf("invalid budget: %w", err)
	}
	budgets[budget.Category] = budget
	return nil
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

var ErrBudgetExceeded = errors.New("budget exceeded")

func main() {
	fmt.Println("Ledger service started")

	tx1 := Transaction{
		Amount:      1000,
		Category:    "Food",
		Description: "Groceries",
		Date:        "2026-09-27",
	}

	tx2 := Transaction{
		Amount:      2500,
		Category:    "Transport",
		Description: "Taxi",
		Date:        "2026-09-27",
	}

	tx3 := Transaction{
		Amount:      500,
		Category:    "Entertainment",
		Description: "Cinema",
		Date:        "2026-09-27",
	}

	tx1, err := AddTransaction(tx1)
	if err != nil {
		fmt.Println("error:", err)
	}

	tx2, err = AddTransaction(tx2)
	if err != nil {
		fmt.Println("error:", err)
	}

	tx3, err = AddTransaction(tx3)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("Transactions:")

	for _, tx := range ListTransactions() {
		fmt.Printf(
			"ID: %d, Amount: %.2f, Category: %s, Description: %s, Date: %s\n",
			tx.ID,
			tx.Amount,
			tx.Category,
			tx.Description,
			tx.Date,
		)
	}

	f, err := os.Open("budgets.json")
	if err != nil {
		fmt.Println("error opening budgets file:", err)
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	err = LoadBudgets(reader)
	if err != nil {
		fmt.Println("error loading budgets:", err)
		return
	}

	_, err = AddTransaction(Transaction{
		Amount:      400,
		Category:    "Food",
		Description: "Restaurant",
		Date:        "2026-09-28",
	})
	if err != nil {
		fmt.Println("error:", err)
	}
	_, err = AddTransaction(Transaction{
		Amount:      400,
		Category:    "Food",
		Description: "Restaurant",
		Date:        "2026-09-28",
	})

	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("Transactions:")

	for _, tx := range ListTransactions() {
		fmt.Printf(
			"ID: %d, Amount: %.2f, Category: %s, Description: %s, Date: %s\n",
			tx.ID,
			tx.Amount,
			tx.Category,
			tx.Description,
			tx.Date,
		)
	}

	err = CheckValid(Transaction{
		Amount:      0,
		Category:    "Food",
		Description: "Restaurant",
		Date:        "2026-09-28",
	})
	if err != nil {
		fmt.Println("error:", err)
	}
	err = CheckValid(Budget{
		Category: "Food",
		Limit:    -100,
	})
	if err != nil {
		fmt.Println("error:", err)
	}

}
