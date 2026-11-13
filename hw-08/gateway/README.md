# Gateway: Homework 08

Gateway keeps the routes, DTOs, JSON fields, and status codes from homework 07.
It obtains a `ledger.Service` from `ledger.New(ctx)` and passes it to
`newHandlers(svc)`. Handlers call service methods; validation and budget decisions
are made by the service. Gateway has no SQL, pgx, or repository imports.

## Run

Create the database and Redis and apply Goose migrations using the
[root setup instructions](../README.md#run). From `hw-08/gateway`, in PowerShell:

```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:15432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "127.0.0.1:16379"
$env:REDIS_DB = "0"
$env:REDIS_PASSWORD = ""
go run .
```

The factory returns the service, `closeFn`, and an error. Gateway stops on startup
errors and calls `defer closeFn()` after successful initialization. It still
listens on `:8080`; there is one application process.

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

Run `go test ./...` in this module for the HTTP scenarios from homework 06. They
inject a small service stub to check the same responses without importing database
code. Ledger integration tests use the real factory, service, and PostgreSQL
repositories; see [test instructions](../README.md#tests).
