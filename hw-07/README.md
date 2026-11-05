# Homework 07: PostgreSQL and Redis

The structure is the same as in homework 06: Gateway handles HTTP and calls the
Ledger package directly. Ledger now stores budgets and transactions in PostgreSQL
and caches report summaries in Redis for 30 seconds.

- `ledger/main.go`: the original types, validation, budget and transaction functions,
  plus connection initialization and `GetReportSummary`.
- `ledger/internal/db` and `ledger/internal/cache`: connection settings from env.
- `ledger/migrations`: Goose SQL migrations.
- `gateway/main.go`: the original handlers, with database initialization, errors
  from the list functions, and the new report route.

## Run

Requires Go 1.27 or later and Docker with Compose. From `hw-07`, in PowerShell:

```powershell
docker compose up -d --wait
go install github.com/pressly/goose/v3/cmd/goose@v3.29.0
$env:Path += ";$(go env GOPATH)\bin"

$env:DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:15432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "127.0.0.1:16379"
$env:REDIS_DB = "0"
$env:REDIS_PASSWORD = ""

goose -dir ./ledger/migrations postgres "$env:DATABASE_URL" up
cd gateway
go run .
```

Compose creates the `cashapp` database with `postgres` / `postgres` credentials.
Its local ports are 15432 for PostgreSQL and 16379 for Redis to avoid conflicts
with existing services. The application reads environment variables from the
terminal; `.env.example` is a reference, not an automatically loaded Go config.

Gateway listens on `:8080`. Ledger is a package in the same process, so starting
Gateway also connects Ledger to PostgreSQL and Redis. Both connections are logged;
a connection failure stops startup.

To roll back the latest migration, run from `hw-07`:

```powershell
goose -dir ./ledger/migrations postgres "$env:DATABASE_URL" down
```

Each SQL file contains Goose `Up` and `Down` sections. One rollback drops
`expenses`; the next drops `budgets`. Apply `up` before running the application.
In Bash, use `export NAME=value` for the variables and `"$DATABASE_URL"` in Goose commands.

See [Ledger configuration](ledger/README.md) and the
[Gateway cURL examples](gateway/README.md). Create a budget and an expense, restart
Gateway with the same database settings, then repeat the GET requests: their
values and IDs remain in PostgreSQL. The Docker volume also survives container restarts.

## Tests

Run the validation tests from `hw-07/ledger`:

```sh
go test ./...
```

The storage and HTTP scenarios from homework 06 now need a test database.
Create it once and apply the same migrations, from `hw-07` in PowerShell:

```powershell
docker compose exec postgres createdb -U postgres cashapp_test
$env:TEST_DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:15432/cashapp_test?sslmode=disable"
goose -dir ./ledger/migrations postgres "$env:TEST_DATABASE_URL" up
$env:REDIS_ADDR = "127.0.0.1:16379"
$env:REDIS_DB = "1"

cd ledger
go test -tags=integration ./...
cd ../gateway
go test -tags=integration ./...
```

Run these module tests sequentially. Each scenario clears the tables in the test
database; its name must end in `_test`. The normal `cashapp` database is not used
by these tests. Migrations are applied by the Goose CLI before the tests.
