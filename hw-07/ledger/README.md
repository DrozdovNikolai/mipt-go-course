# Ledger

Ledger remains a Go package used by Gateway. Its functions in `main.go` now use
PostgreSQL instead of the in-memory map and slice. The `Validate()` methods are
unchanged from homework 06.

## Database and startup

From `hw-07`, `docker compose up -d --wait` creates PostgreSQL, the `cashapp`
database, and Redis. See the [setup commands](../README.md#run) for installing Goose.
After setting `DATABASE_URL`, apply or roll back migrations from `hw-07`:

```powershell
goose -dir ./ledger/migrations postgres "$env:DATABASE_URL" up
goose -dir ./ledger/migrations postgres "$env:DATABASE_URL" down
```

Use `"$DATABASE_URL"` in Bash. Each numbered migration has an `Up` and a `Down`
section. Apply `up` before starting the application; `down` drops the latest table.

The tables use the columns and constraints from the assignment. The additional
`budgets.period` column preserves the monthly/yearly budget behavior from homework 06.

Start from `hw-07/gateway`, in PowerShell:

```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:15432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "127.0.0.1:16379"
$env:REDIS_DB = "0"
$env:REDIS_PASSWORD = ""
go run .
```

| Variable | Default |
| --- | --- |
| `DATABASE_URL` | Takes priority over the separate DB settings |
| `DB_HOST`, `DB_PORT` | `localhost`, `5432` |
| `DB_USER`, `DB_PASS`, `DB_NAME` | `postgres`, `postgres`, `cashapp` |
| `REDIS_ADDR` | `localhost:6379` |
| `REDIS_DB` | `0` |
| `REDIS_PASSWORD` | Empty |

The separate DB settings use `sslmode=disable`. Connection setup calls
`sql.Open("pgx", dsn)` and `Ping()`, then uses a pool of 10 open and 5 idle
connections. Redis uses the standard `*redis.Client`. Both connections are logged.

## Storage and reports

`SetBudget` validates and upserts by category. `ListBudgets` sorts by category.
`AddTransaction` validates, reads the budget, calculates spending with `SUM`, and
inserts only if the limit permits it. It returns the generated ID. A category
without a budget has no limit. `ListTransactions` sorts by date and ID descending.
The queries run sequentially without a SQL transaction, as allowed by the task;
concurrent budget enforcement is outside this simple implementation.

`GetReportSummary(ctx, from, to)` returns a map of category totals for the inclusive
`YYYY-MM-DD` interval. It reads `report:summary:<from>:<to>` from Redis. On a miss,
it calculates `SUM(amount) GROUP BY category` in PostgreSQL and stores JSON with a
30-second TTL. Cache hits and misses are logged. An existing cached report may
show the previous values until its TTL expires.

Call the report twice through Gateway:

```sh
curl 'http://localhost:8080/api/reports/summary?from=2026-11-01&to=2026-11-30'
curl 'http://localhost:8080/api/reports/summary?from=2026-11-01&to=2026-11-30'
```

The terminal shows `Report cache miss`, then `Report cache hit`.
See [test instructions](../README.md#tests) for the original validation and storage scenarios.
