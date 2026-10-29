# Homework 05: Ledger HTTP API

The Gateway exposes an HTTP API for transactions and category budgets. It uses
the local `ledger` module for validation, storage, and budget checks.

## Run

Requires Go 1.27.0 or later. From the `hw-05` directory:

```bash
cd gateway
go run .
```

The server listens on `http://127.0.0.1:8080`. Data is stored in memory and is
cleared when the server restarts. Each request logs its method, path, and duration.

## Endpoints

API responses use `Content-Type: application/json; charset=utf-8`.

| Method | Path | Description | Success status |
| --- | --- | --- | --- |
| GET | `/ping` | Check that the server is running | 200 |
| GET | `/api/transactions` | List transactions; returns `[]` when empty | 200 |
| POST | `/api/transaction` | Create a transaction and return it with its assigned ID | 201 |
| GET | `/api/budgets` | List budgets; returns `[]` when empty | 200 |
| POST | `/api/budget` | Create or replace a budget for a category | 201 |

## cURL examples

Run these commands in Bash in a second terminal. Start with a fresh server and
run the examples in order.

Check the server:

```bash
curl -i http://127.0.0.1:8080/ping
```

Expected body: `{"message":"pong"}`.

Create a Food budget of 1000:

```bash
curl -i -X POST http://127.0.0.1:8080/api/budget \
  -H 'Content-Type: application/json' \
  -d '{"Category":"Food","Limit":1000}'
```

Expected body: `{"Category":"Food","Limit":1000,"Period":""}`.
The optional `Period` field supports `monthly` and `yearly`; when omitted, the
budget covers all transactions in the category.

List budgets:

```bash
curl -i http://127.0.0.1:8080/api/budgets
```

Create a transaction:

```bash
curl -i -X POST http://127.0.0.1:8080/api/transaction \
  -H 'Content-Type: application/json' \
  -d '{"Amount":999,"Category":"Food","Description":"Groceries","Date":"2026-09-27"}'
```

Expected body on a fresh server:

```json
{"ID":1,"Amount":999,"Category":"Food","Description":"Groceries","Date":"2026-09-27"}
```

List transactions:

```bash
curl -i http://127.0.0.1:8080/api/transactions
```

Check the budget limit by adding another expense of 2:

```bash
curl -i -X POST http://127.0.0.1:8080/api/transaction \
  -H 'Content-Type: application/json' \
  -d '{"Amount":2,"Category":"Food","Description":"Snack","Date":"2026-09-27"}'
```

Expected status: `409 Conflict`. Body: `{"error":"budget exceeded"}`.

## Errors

Errors use the JSON format `{"error":"description"}`.

| Status | Meaning |
| --- | --- |
| 400 Bad Request | Invalid JSON or validation failure |
| 404 Not Found | Unknown path |
| 405 Method Not Allowed | Unsupported method for an endpoint |
| 409 Conflict | Budget exceeded |
| 500 Internal Server Error | Unexpected internal error |
