# Trading-Insights

A quick, human-readable map of the codebase so you can see how the pieces fit together.

## What this service is trying to do
- Periodically fetch foreign-exchange rates, cache them for quick reads, and persist them for time-series analytics.
- Expose an HTTP API for users to sign up/login, manage a watchlist, record trades in a ledger, and run lightweight analytics (cross rates, correlations, portfolio value).
- Keep the stack small: Go + Chi for HTTP, PostgreSQL/TimescaleDB for storage, Redis for caching, and JWT for auth.

## How it starts up
1. `main.go` loads `.env`, opens PostgreSQL and Redis connections, auto-migrates the main models, and tries to enable TimescaleDB if available.
2. Repositories (`repository/`) wrap DB/Redis for currencies, users, ledger, and watchlists.
3. Services (`services/`) hold the business logic (auth, currency fetching, analytics, ledger bookkeeping, watchlists, ingestion).
4. HTTP handlers (`handlers/`) translate requests/responses and plug into the router defined in `server/routes.go`.
5. A scheduler (`codnect.io/chrono`) runs every 30s to call the ingestion service, which pulls rates from the `EXCONVERT_URL` API, caches them in Redis, and stores them in Postgres.
6. The Chi server listens on `:8000` with logging/recovery/timeout middleware, plus JWT auth middleware on protected routes.

## Directory tour
- `main.go`: wiring for env loading, DB clients, service construction, router setup, and the scheduler.
- `server/`: Chi router factory and all route registrations.
- `handlers/`: One file per feature area (auth, currencies/history/candles, ledger, watchlist, analytics). They parse inputs, call services, and shape HTTP responses.
- `services/`: Business rules:
  - `ingestion.go`: fetch live FX rates from `EXCONVERT_URL` and fan them out to cache + Postgres.
  - `currency.go`: read cached or stored rates, normalize bucket sizes for candles.
  - `user.go`: signup/login, password hashing, JWT issuance.
  - `watchlist.go`: CRUD for user watchlists.
  - `ledger.go`: double-entry-ish storage of trades/fees per user.
  - `analytics.go`: cross-rates, correlations, and portfolio valuation over time.
- `repository/`: Data access for each domain. Notable bits include Timescale-friendly candle queries and paired-rate joins, ledger balance queries, and Redis-backed snapshot caching.
- `models/`: GORM models for users, currencies (snapshots), ledger entries, transactions, and watch items.
- `middleware/`: JWT auth middleware that decorates the request context with user claims.
- `authentication/`: Token generation/validation helpers (reads `JWT_SECRET`).
- `database/`: Connection helpers for Postgres/Redis and optional TimescaleDB setup.
- `docker-compose.yml`: Local infra (TimescaleDB/Postgres, Redis, Kafka/ZooKeeper for future streaming ideas).
- `trading-insights`: Built binary artifact currently in the repo.

## Request lifecycle (happy path)
- A request hits Chi in `server/routes.go`; logging/recovery/timeout middleware run first.
- Protected endpoints add `middleware.AuthMiddleware`, which validates the `Authorization: Bearer <token>` header and injects user claims.
- Handlers map inputs to service calls; services rely on repositories to talk to Postgres/Redis.
- Responses are JSON across the board, with simple error messages on validation/auth failures.

## Background ingestion loop
- Every 30 seconds the scheduler calls `services.ingestion.FetchRates`, which:
  - Pulls a snapshot from `EXCONVERT_URL`.
  - Caches the snapshot in Redis for fast `/currencies/latest` reads.
  - Persists rates in Postgres (and converts the table to a Timescale hypertable when available) so history/candles/analytics can query efficiently.

## Endpoints at a glance
- `POST /auth/signup`, `POST /auth/login`: user creation and JWT login.
- `GET /currencies/latest`, `GET /currencies/{ticker}/history`, `GET /currencies/{ticker}/candles`: public market data reads.
- `POST /watchlist/add`, `POST /watchlist/remove`, `GET /watchlist/`: auth-required watchlist operations.
- `POST /ledger/exchange`, `GET /ledger/`, `GET /ledger/trade/{tradeID}`: record and inspect trades; ledger rows are signed (+inflow, -outflow).
- `GET /analytics/cross|chart|convert|correlation`: public cross-rate analytics; `GET /analytics/portfolio/value|history`: auth-required portfolio valuation using ledger balances + stored FX.

## Running it locally (short version)
- `docker-compose up -d` to start Postgres/TimescaleDB and Redis (Kafka is optional right now).
- Copy `.env` (already present) or set equivalent env vars: DB host/port/user/pass/name, `EXCONVERT_URL`, `JWT_SECRET`, `REDIS_ADDR`.
- `go run main.go` (or build) to launch the API on port `8000`.

## Notes on current state
- Minimal automated tests exist (see `database/user_test.go`), and many endpoints assume happy-path data; hardening, validation, and error handling still need work.
- The project is not finished.
