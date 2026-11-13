# Ledger: Homework 08

Ledger exposes a service interface and a factory. The existing entities and
validation moved to `internal/domain`; SQL moved to `internal/repository/pg`;
validation and budget checks are called from `internal/service`.

`ledger.New(ctx)` returns `(Service, closeFn, error)`. It creates PostgreSQL and
Redis connections, builds the repositories, and injects them into the service.
The service uses only domain interfaces. The factory passes the caller's context
to PostgreSQL's startup ping and returns a function that closes both connections.
There is no package-level database or Redis client.

Public aliases in `main.go` allow Gateway to keep using `ledger.Transaction`,
`ledger.Budget`, and `ledger.Service` without accessing internal packages.
See the [layer guide](../README.md#structure-and-purpose).

## Database and startup

From `hw-08`, `docker compose up -d --wait` creates PostgreSQL, the `cashapp`
database, and Redis. See the [setup commands](../README.md#run) for installing Goose.
After setting `DATABASE_URL`, apply or roll back migrations from `hw-08`:

```powershell
goose -dir ./ledger/migrations postgres "$env:DATABASE_URL" up
goose -dir ./ledger/migrations postgres "$env:DATABASE_URL" down
```

Use `"$DATABASE_URL"` in Bash. Each numbered migration has an `Up` and a `Down`
section. Apply `up` before starting the application; `down` drops the latest table.

The tables use the columns and constraints from the assignment. The additional
`budgets.period` column preserves the monthly/yearly budget behavior from homework 06.

Start from `hw-08/gateway`, in PowerShell:

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
`sql.Open("pgx", dsn)` and `PingContext(ctx)`, then uses a pool of 10 open and 5 idle
connections. Redis implements the domain report-cache interface with the same key,
JSON encoding, and TTL as homework 07. Both connections are logged.

## Storage and reports

The service validates budgets and transactions. `AddTransaction` asks its budget
repository for the category limit and its transaction repository for the sum,
then rejects an exceeded budget or asks the repository to insert the expense.
A category without a budget has no limit. Repositories keep the original UPSERT,
list ordering, generated IDs, and monthly/yearly filtering.
The queries run sequentially without a SQL transaction, as allowed by the task;
concurrent budget enforcement is outside this simple implementation.

`GetReportSummary(ctx, from, to)` returns the same map of category totals. The
service asks the report-cache interface first. On a miss, it asks the transaction
repository for category sums and writes them through the cache interface. The
Redis implementation uses `report:summary:<from>:<to>` and a 30-second TTL. Cache hits and misses are logged. An existing cached report may
show the previous values until its TTL expires.

Call the report twice through Gateway:

```sh
curl 'http://localhost:8080/api/reports/summary?from=2026-11-01&to=2026-11-30'
curl 'http://localhost:8080/api/reports/summary?from=2026-11-01&to=2026-11-30'
```

The terminal shows `Report cache miss`, then `Report cache hit`.
See [test instructions](../README.md#tests) for the original validation and storage scenarios.
