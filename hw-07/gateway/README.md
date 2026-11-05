# Gateway

Gateway keeps the handlers and request/response types from homework 06. It calls
Ledger directly; Ledger stores data in PostgreSQL and caches reports in Redis.
The budget and transaction routes, JSON fields, and status codes stay the same.

## Run

Create the database and Redis and apply Goose migrations using the
[root setup instructions](../README.md#run). From `hw-07/gateway`, in PowerShell:

```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:15432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "127.0.0.1:16379"
$env:REDIS_DB = "0"
$env:REDIS_PASSWORD = ""
go run .
```

Gateway initializes Ledger and listens on `:8080`. There is one application process.

## Check the API

Run these cURL commands in Bash. Start with a new category if `food` already has expenses.

```bash
curl -i -X POST http://localhost:8080/api/budgets \
  -H 'Content-Type: application/json' \
  -d '{"Category":"food","Limit":5000}'

curl -i -X POST http://localhost:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"Amount":3000,"Category":"food","Description":"Groceries","Date":"2026-11-05"}'
```

Both return `201`; the transaction response includes its database-generated `ID`.
The next request exceeds the total limit and returns `409` with
`{"error":"budget exceeded"}`:

```bash
curl -i -X POST http://localhost:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"Amount":2500,"Category":"food","Date":"2026-11-05"}'

curl -i http://localhost:8080/api/budgets
curl -i http://localhost:8080/api/transactions
```

GET requests return `200` and JSON arrays read from PostgreSQL. The rejected
transaction is absent. Handler responses use `application/json; charset=utf-8`.

Invalid JSON returns `400` and a JSON error:

```bash
curl -i -X POST http://localhost:8080/api/transactions \
  -H 'Content-Type: application/json' -d '{"Amount":'
```

Call the report twice within 30 seconds:

```bash
curl 'http://localhost:8080/api/reports/summary?from=2026-11-01&to=2026-11-30'
curl 'http://localhost:8080/api/reports/summary?from=2026-11-01&to=2026-11-30'
```

On an otherwise empty database the report is `{"food":3000}`. Ledger logs a cache
miss, then a cache hit. Other existing transactions contribute to the totals.

Stop Gateway with Ctrl+C and start it again with the same database settings.
This also restarts Ledger. Repeat the GET requests: the budget and transaction,
including its ID, remain in PostgreSQL.

The HTTP tests from homework 06 use a separate test database;
see [test instructions](../README.md#tests).
