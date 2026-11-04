package main

import (
	"encoding/json"
	"gateway/internal/api"
	"ledger"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBudgetsHTTP(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ledger.Reset()
		t.Cleanup(ledger.Reset)

		// Create a budget via POST.
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/budgets",
			strings.NewReader(`{"category":"food","limit":5000}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		createBudgetHandler(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var created api.BudgetResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("cannot unmarshal budget response: %v", err)
		}

		want := api.BudgetResponse{Category: "food", Limit: 5000}
		if created != want {
			t.Fatalf("got %+v, want %+v", created, want)
		}

		// Retrieve the list of budgets via GET.
		req = httptest.NewRequest(http.MethodGet, "/api/budgets", nil)
		rec = httptest.NewRecorder()
		listBudgetsHandler(rec, req)

		resp = rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusOK, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var budgets []api.BudgetResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &budgets); err != nil {
			t.Fatalf("cannot unmarshal budgets response: %v", err)
		}

		if len(budgets) != 1 {
			t.Fatalf("want 1 budget, got %d", len(budgets))
		}
		if budgets[0] != want {
			t.Errorf("got budget %+v, want %+v", budgets[0], want)
		}
	})

	t.Run("invalid_limit", func(t *testing.T) {
		ledger.Reset()
		t.Cleanup(ledger.Reset)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/budgets",
			strings.NewReader(`{"category":"food","limit":0}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		createBudgetHandler(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusBadRequest, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var failure struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil {
			t.Fatalf("cannot unmarshal error response: %v", err)
		}

		if failure.Error == "" {
			t.Error("want error message, got empty string")
		}
	})
}

func TestTransactionsHTTP(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ledger.Reset()
		t.Cleanup(ledger.Reset)

		// Create a budget before adding a transaction.
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/budgets",
			strings.NewReader(`{"category":"food","limit":5000}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		createBudgetHandler(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		// Create a transaction within the budget.
		req = httptest.NewRequest(
			http.MethodPost,
			"/api/transactions",
			strings.NewReader(`{"amount":3000,"category":"food","description":"Groceries","date":"2026-11-03"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec = httptest.NewRecorder()
		createTransactionHandler(rec, req)

		resp = rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var created api.TransactionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("cannot unmarshal transaction response: %v", err)
		}
		if created.ID <= 0 {
			t.Fatalf("want a positive transaction ID, got %d", created.ID)
		}

		want := api.TransactionResponse{
			ID:          created.ID,
			Amount:      3000,
			Category:    "food",
			Description: "Groceries",
			Date:        "2026-11-03",
		}
		if created != want {
			t.Fatalf("got transaction %+v, want %+v", created, want)
		}

		// Retrieve the transaction through the HTTP handler.
		req = httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
		rec = httptest.NewRecorder()
		listTransactionsHandler(rec, req)

		resp = rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusOK, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var transactions []api.TransactionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &transactions); err != nil {
			t.Fatalf("cannot unmarshal transactions response: %v", err)
		}
		if len(transactions) != 1 {
			t.Fatalf("want 1 transaction, got %d", len(transactions))
		}
		if transactions[0] != created {
			t.Errorf("got transaction %+v, want %+v", transactions[0], created)
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		ledger.Reset()
		t.Cleanup(ledger.Reset)

		// Set up this scenario independently of the successful scenario.
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/budgets",
			strings.NewReader(`{"category":"food","limit":5000}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		createBudgetHandler(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		req = httptest.NewRequest(
			http.MethodPost,
			"/api/transactions",
			strings.NewReader(`{"amount":3000,"category":"food","description":"Groceries","date":"2026-11-03"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec = httptest.NewRecorder()
		createTransactionHandler(rec, req)

		resp = rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var created api.TransactionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("cannot unmarshal transaction response: %v", err)
		}

		// Both amounts fit individually, but their sum exceeds the budget.
		req = httptest.NewRequest(
			http.MethodPost,
			"/api/transactions",
			strings.NewReader(`{"amount":2500,"category":"food","description":"Dinner","date":"2026-11-03"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec = httptest.NewRecorder()
		createTransactionHandler(rec, req)

		resp = rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusConflict, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var failure struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil {
			t.Fatalf("cannot unmarshal error response: %v", err)
		}
		if failure.Error != "budget exceeded" {
			t.Errorf("want error %q, got %q", "budget exceeded", failure.Error)
		}

		// The rejected transaction must not appear in the list.
		req = httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
		rec = httptest.NewRecorder()
		listTransactionsHandler(rec, req)

		resp = rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusOK, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var transactions []api.TransactionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &transactions); err != nil {
			t.Fatalf("cannot unmarshal transactions response: %v", err)
		}
		if len(transactions) != 1 {
			t.Fatalf("want 1 transaction after rejection, got %d", len(transactions))
		}
		if transactions[0] != created {
			t.Errorf("got transaction %+v, want %+v", transactions[0], created)
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		ledger.Reset()
		t.Cleanup(ledger.Reset)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/transactions",
			strings.NewReader(`{"amount":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		createTransactionHandler(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("want status %d, got %d; body: %s", http.StatusBadRequest, resp.StatusCode, rec.Body.String())
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("want Content-Type %q, got %q", "application/json; charset=utf-8", got)
		}

		var failure struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil {
			t.Fatalf("cannot unmarshal error response: %v", err)
		}
		if failure.Error == "" {
			t.Error("want error message, got empty string")
		}
	})
}
