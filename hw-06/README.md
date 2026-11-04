# Homework 06: Tests and Coverage for the Ledger HTTP API

This homework tests an in-memory ledger and its HTTP gateway. It includes
table-driven domain tests, HTTP integration tests, named subtests, and statement
coverage reports. The tests use the Go standard library.

## Project layout

| Path | Purpose |
| --- | --- |
| `ledger/main.go` | Transaction and budget validation, storage, and budget enforcement |
| `ledger/ledger_test.go` | Domain unit tests |
| `gateway/main.go` | HTTP routes, handlers, and request logging |
| `gateway/http_test.go` | HTTP handler integration tests |
| `gateway/internal/api/http.go` | Request/response types and conversion functions |

Requires **Go 1.27.0 or later**. The `ledger` and `gateway` directories are separate
Go modules. The gateway uses the local ledger through a `replace` directive in
its `go.mod`.

## Run the tests

Starting in `hw-06`, run both modules:

```sh
cd ledger
go test -v ./...

cd ../gateway
go test -v ./...
```

The `-v` flag shows individual subtest names and results. The suites cover:

| Test group | Scenarios |
| --- | --- |
| `TestTransactionValidate` | Valid input, zero/negative amount, empty category, empty date |
| `TestBudgetValidate` | Valid input, zero/negative limit, empty category |
| `TestAddTransaction` | Successful addition and budget rejection without storing the transaction |
| `TestBudgetsHTTP` | `ok`: POST then GET; `invalid_limit`: HTTP 400 with a JSON error |
| `TestTransactionsHTTP` | `ok`: create a budget, POST a transaction, then GET it; `exceeded`: HTTP 409 and no extra transaction; `bad_json`: HTTP 400 |

Validation subtests use `t.Parallel()` because they only inspect their own input.
Tests that change ledger state run sequentially. Each such subtest starts with
`Reset()` and registers `t.Cleanup(Reset)` to clear the state after it finishes
(`ledger.Reset` from the gateway package). Resetting shared data does not make
parallel access safe.

HTTP tests use `httptest.NewRequest` and `httptest.NewRecorder` to call the
handlers directly. They verify status codes, JSON fields, and
`Content-Type: application/json; charset=utf-8`. These tests run in-process and
exercise the handlers together with the ledger; route registration and the
logging middleware are outside their coverage.

## Measure coverage

Run the following commands in **each module directory**:
`hw-06/ledger` and `hw-06/gateway`.

```sh
go test -cover "-coverprofile=cover.out" ./...
go tool cover "-func=cover.out"
go tool cover "-html=cover.out" -o cover.html
```

Each directory gets its own `cover.out` profile and `cover.html` report.
Keep the complete filename flags in quotes, especially in Windows PowerShell
5.1, to avoid argument parsing errors.

The command with `-o cover.html` saves the report. To open it in a browser
directly, use:

```sh
go tool cover "-html=cover.out"
```

Coverage measured on **2026-11-03**, using Go 1.27.0 on Windows/amd64:

| Package | Statement coverage reported by `go test` |
| --- | ---: |
| `ledger` | **29.7%** |
| `gateway` | **36.1%** |
| `gateway/internal/api` | **0.0%** |

These results use default per-package instrumentation. The `internal/api`
package has no tests of its own; calls from the gateway tests are not counted
toward that package's separate result with this command.

The total shown by `go tool cover -func` is **29.7%** for the ledger profile and
**35.3%** for the gateway profile. The latter combines `gateway` and
`gateway/internal/api`, so it differs from the root package's 36.1%.

## Run the HTTP server

Starting in `hw-06`:

```sh
cd gateway
go run .
```

The server listens on `http://127.0.0.1:8080`. It uses the ledger in the same
process. Data is stored in memory and is cleared when the server restarts.
Requests are logged with their method, path, and duration.

## HTTP API

Handlers return JSON with
`Content-Type: application/json; charset=utf-8`.

| Method | Path | Description | Success status |
| --- | --- | --- | --- |
| GET | `/ping` | Health check: `{"message":"pong"}` | 200 |
| GET | `/api/budgets` | List budgets; returns `[]` when empty | 200 |
| POST | `/api/budgets` | Create or replace a category budget | 201 |
| GET | `/api/transactions` | List transactions; returns `[]` when empty | 200 |
| POST | `/api/transactions` | Create a transaction and return its assigned ID | 201 |

Transaction amounts and budget limits must be positive. Categories must be
nonempty; transactions also require a nonempty date. Use dates in
`YYYY-MM-DD` format, as in the tests.

Budget checks sum transactions in the same category. An optional budget
`Period` of `monthly` or `yearly` limits that sum to the matching month or year.
Omitting `Period` includes all transactions in the category. Category names
are case-sensitive.

## Manual examples

Run these commands in **Bash** in a second terminal. Start a fresh server and
execute them in order.

Create a budget of 5000:

```bash
curl -i -X POST http://127.0.0.1:8080/api/budgets \
  -H 'Content-Type: application/json' \
  -d '{"Category":"food","Limit":5000}'
```

Expected status: `201 Created`.
Body: `{"Category":"food","Limit":5000,"Period":""}`.

Create a transaction within the budget:

```bash
curl -i -X POST http://127.0.0.1:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"Amount":3000,"Category":"food","Description":"Groceries","Date":"2026-11-03"}'
```

Expected status: `201 Created`. Body on a fresh server:

```json
{"ID":1,"Amount":3000,"Category":"food","Description":"Groceries","Date":"2026-11-03"}
```

Retrieve the stored budget and transaction:

```bash
curl -i http://127.0.0.1:8080/api/budgets
curl -i http://127.0.0.1:8080/api/transactions
```

Both responses have status `200 OK` and contain an array with the created item.

Try to add another transaction that would bring the total to 5500:

```bash
curl -i -X POST http://127.0.0.1:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"Amount":2500,"Category":"food","Description":"Dinner","Date":"2026-11-03"}'
```

Expected status: `409 Conflict`. Body: `{"error":"budget exceeded"}`.
The rejected transaction is not stored.

Send malformed JSON:

```bash
curl -i -X POST http://127.0.0.1:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"Amount":'
```

Expected status: `400 Bad Request`.
Body: `{"error":"invalid request body"}`.

## Error responses

Handler errors use the JSON format `{"error":"description"}`.

| Status | Meaning |
| --- | --- |
| 400 Bad Request | Invalid JSON or validation failure |
| 404 Not Found | A request reaches the fallback handler |
| 405 Method Not Allowed | A handler rejects the method, for example POST to `/ping` |
| 409 Conflict | Adding a transaction would exceed its category budget |
| 500 Internal Server Error | An unexpected ledger error |

Collection routes are registered with explicit GET and POST methods. An
unsupported method/path combination can reach the fallback handler and return
404.
