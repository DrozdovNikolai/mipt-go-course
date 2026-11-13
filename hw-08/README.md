# Homework 08: Domain, Repositories, and Service

This homework starts from homework 07 and separates the existing code into the
layers required by the assignment. SQL, migrations, validation rules, HTTP DTOs,
status codes, and the 30-second report cache keep the same behavior.

## Structure and purpose

| Location | Responsibility |
| --- | --- |
| `ledger/internal/domain/domain.go` | Existing entities and validation; interfaces describing what the service needs from storage |
| `ledger/internal/repository/pg/pg.go` | Existing PostgreSQL queries, implemented as repository methods |
| `ledger/internal/service/service.go` | Application interface, validation, budget limit checks, and report flow |
| `ledger/internal/cache/cache.go` | Existing Redis key, JSON serialization, and TTL behind the report-cache interface |
| `ledger/main.go` | Public types and the factory that creates connections, repositories, and the service |
| `gateway/main.go` | Existing HTTP handlers, now receiving a `ledger.Service` in their constructor |

The domain imports no storage or HTTP code. The service depends on domain
interfaces, not PostgreSQL or Redis clients. PostgreSQL repositories use domain
types. Only the Ledger factory knows which concrete implementations to create.
Gateway imports the public `ledger` API; it does not import SQL, pgx, or repositories.

The aliases in `ledger/main.go` expose domain types and the service interface to
Gateway. This is necessary because Go does not allow Gateway to import Ledger's
`internal` packages. Existing DTO conversion functions stay unchanged.

## How initialization works

Gateway starts with:

```go
svc, closeFn, err := ledger.New(context.Background())
if err != nil {
    log.Fatalf("connect Ledger: %v", err)
}
defer closeFn()
h := newHandlers(svc)
```

`ledger.New` opens PostgreSQL with `PingContext(ctx)`, connects Redis, constructs
both repositories, and passes them into `service.New`. It returns the ready
`Service`, a function that closes both connections, and an error. Failed startup
returns no service and closes resources opened before the failure. The factory
does not run migrations.

Passing the repositories into the service and the service into the handlers is
**dependency injection**: each object receives the dependencies it needs instead
of reading a global database variable. No DI framework is used.

On POST, Gateway decodes the unchanged DTO and calls the service. The service
validates the entity and checks the budget using repository methods. A typed
`ValidationError` lets Gateway keep the same `400` message; `ErrBudgetExceeded`
still maps to `409`, and database errors to `500`.

Read `gateway/main.go`, then `ledger/main.go`, then the domain, service, and
PostgreSQL repository files to follow the request from HTTP to SQL.

## Run

Requires Go 1.27 or later and Docker with Compose. From `hw-08`, in PowerShell:

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

Gateway still listens on `:8080` and runs in one process with Ledger. Compose
intentionally reuses the `mipt-go-hw07` project and its existing volume: this
refactor needs no new database schema or data migration. Both connections are
logged, and a connection failure stops startup.

To roll back the latest migration, run from `hw-08`:

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

Run the original validation and HTTP scenarios, plus the startup-context check:

```sh
cd ledger
go test ./...
cd ../gateway
go test ./...
```

Validation tests moved to the domain package. Gateway HTTP tests receive a small
stub implementing `Service`, so they check DTOs and error mapping without knowing
about SQL. The actual budget rule and PostgreSQL writes are checked through the
real factory by Ledger's integration tests.

For those tests, use the existing `cashapp_test` database from homework 07. If it
does not exist, create it once from `hw-08`:

```sh
docker compose exec postgres createdb -U postgres cashapp_test
```

Apply the unchanged migrations and run Ledger integration tests, from `hw-08` in
PowerShell:

```powershell
$env:TEST_DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:15432/cashapp_test?sslmode=disable"
goose -dir ./ledger/migrations postgres "$env:TEST_DATABASE_URL" up
$env:REDIS_ADDR = "127.0.0.1:16379"
$env:REDIS_DB = "1"
cd ledger
go test -tags=integration ./...
```

Each scenario clears the tables in this dedicated test database; its name must
end in `_test`. The factory itself never clears tables or applies migrations.
Use the [cURL walkthrough](gateway/README.md#check-the-api) for the full HTTP chain.
