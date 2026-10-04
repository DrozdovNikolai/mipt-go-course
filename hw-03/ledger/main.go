package main

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

var transactions = make([]Transaction, 0)

var budgets = make(map[string]Budget)

func AddTransaction(tx Transaction) error {

	el, exists := budgets[tx.Category]

	if exists {

		total := 0.0
		for _, t := range transactions {
			if t.Category == tx.Category &&
				((el.Period == "monthly" && t.Date[:7] == tx.Date[:7]) || (el.Period == "yearly" && t.Date[:4] == tx.Date[:4])) {
				total += t.Amount
			}

		}
		if total+tx.Amount > budgets[tx.Category].Limit {
			return errors.New("transaction exceeds budget limit for category: " + tx.Category)
		}
	}

	if tx.Amount == 0 {
		return errors.New("transaction amount cannot be zero")
	}

	tx.ID = len(transactions) + 1

	transactions = append(transactions, tx)

	return nil
}

func ListTransactions() []Transaction {
	result := make([]Transaction, len(transactions))
	copy(result, transactions)

	return result
}

func SetBudget(budget Budget) {
	budgets[budget.Category] = budget
}

func LoadBudgets(r io.Reader) error {
	fmt.Println("Loading budgets from JSON file")
	var b []Budget
	err := json.NewDecoder(r).Decode(&b)
	if err != nil {
		return fmt.Errorf("failed to decode budgets JSON: %w", err)
	}
	for _, budget := range b {
		SetBudget(budget)
	}
	return nil
}

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

	err := AddTransaction(tx1)
	if err != nil {
		fmt.Println("error:", err)
	}

	err = AddTransaction(tx2)
	if err != nil {
		fmt.Println("error:", err)
	}

	err = AddTransaction(tx3)
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

	err = AddTransaction(Transaction{
		Amount:      400,
		Category:    "Food",
		Description: "Restaurant",
		Date:        "2026-09-28",
	})
	if err != nil {
		fmt.Println("error:", err)
	}
	err = AddTransaction(Transaction{
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

}
