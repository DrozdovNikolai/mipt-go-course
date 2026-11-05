package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/internal/api"
	"ledger"
	"log"
	"net/http"
	"time"
)

func main() {
	if err := ledger.Open(context.Background()); err != nil {
		log.Fatalf("connect Ledger: %v", err)
	}
	defer ledger.Close()

	http.HandleFunc("/ping", pingHandler)

	http.HandleFunc("POST /api/transactions", createTransactionHandler)
	http.HandleFunc("GET /api/transactions", listTransactionsHandler)

	http.HandleFunc("POST /api/budgets", createBudgetHandler)
	http.HandleFunc("GET /api/budgets", listBudgetsHandler)

	http.HandleFunc("GET /api/reports/summary", reportSummaryHandler)

	http.HandleFunc("/", notFoundHandler)

	fmt.Println("Gateway service started on :8080")

	err := http.ListenAndServe(":8080", loggingMiddleware(http.DefaultServeMux))
	if err != nil {
		fmt.Println("server error:", err)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("Processed %s request for %s in %v\n", r.Method, r.URL.Path, time.Since(timeStart))
	})
}

func listTransactionsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	transactions, err := ledger.ListTransactions()
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	transactionResponses := make([]api.TransactionResponse, len(transactions))
	for i, t := range transactions {
		transactionResponses[i] = api.TransactionToResponse(&t)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(transactionResponses)
}

func createTransactionHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	jsonDecoder := json.NewDecoder(r.Body)
	var req api.CreateTransactionRequest
	err := jsonDecoder.Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request body"})
		return
	}

	transaction := api.TransactionFromRequest(&req)

	err = transaction.Validate()
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	transaction, err = ledger.AddTransaction(transaction)
	if err != nil {
		if err == ledger.ErrBudgetExceeded {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(api.TransactionToResponse(&transaction))

}

func listBudgetsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	budgets, err := ledger.ListBudgets()
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	budgetResponses := make([]api.BudgetResponse, len(budgets))
	for i, b := range budgets {
		budgetResponses[i] = api.BudgetToResponse(&b)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(budgetResponses)
}

func createBudgetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	jsonDecoder := json.NewDecoder(r.Body)
	var req api.CreateBudgetRequest
	err := jsonDecoder.Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request body"})
		return
	}

	budget := api.BudgetFromRequest(&req)

	err = budget.Validate()
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if err := ledger.SetBudget(budget); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(api.BudgetToResponse(&budget))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "pong"})
}

func reportSummaryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	from, to := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	report, err := ledger.GetReportSummary(r.Context(), from, to)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(report)
}
