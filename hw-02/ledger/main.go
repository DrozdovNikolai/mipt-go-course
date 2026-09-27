package main

import (
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

var transactions = make([]Transaction, 0)

func AddTransaction(tx Transaction) error {
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
}
