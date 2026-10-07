# golang-rest-api

A simple REST API written in Go for managing audiobooks, tags, and languages. It uses [gorilla/mux](https://github.com/gorilla/mux) for routing, [GORM](https://gorm.io/) with PostgreSQL for persistence, and [zerolog](https://github.com/rs/zerolog) for structured logging.

## Features

- Audiobook CRUD endpoints, including multipart file upload
- Tag listing and many-to-many association with audiobooks
- Language / gender-voice listing
- Structured JSON logging per request
- Config via CLI flags or environment variables
- Automatic DB schema migration on startup
- Docker and docker-compose setup (API + Postgres + pgAdmin)
- Kubernetes deployment manifests

## Tech stack

- Go 1.26
- [gorilla/mux](https://github.com/gorilla/mux) — HTTP routing
- [gorilla/schema](https://github.com/gorilla/schema) — query string decoding
- [GORM](https://gorm.io/) + [gorm.io/driver/postgres](https://github.com/go-gorm/postgres) — ORM / PostgreSQL
- [go-playground/validator](https://github.com/go-playground/validator) — struct validation
- [rs/zerolog](https://github.com/rs/zerolog) — logging
- [peterbourgon/ff](https://github.com/peterbourgon/ff) — flag/env config parsing
- [justinas/alice](https://github.com/justinas/alice) — middleware chaining

## Project structure

```
cmd/go-rest/   main entrypoint
cmd/           CLI flag parsing and app wiring
server/        HTTP server, routes, handlers, middleware
repositories/  data access layer (audiobooks, tags, languages)
models/        GORM models
db/            database connection and migrations
errs/          error types and HTTP error handling
logger/        zerolog setup
kubernetes/    k8s deployment manifests
```

## Getting started

### Prerequisites

- Go 1.26+
- PostgreSQL (or use the provided `docker-compose.yml`)

### Run with Docker Compose

1. Create a `.env` file in the project root:

   ```env
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=golang_rest_api
   PGADMIN_DEFAULT_EMAIL=admin@example.com
   ```

2. Start the stack:

   ```bash
   docker compose up --build
   ```

   This brings up the API on `localhost:8080`, Postgres on `localhost:5432`, and pgAdmin on `localhost:5050`.

### Run locally

```bash
go run ./cmd/go-rest \
  --db-host localhost \
  --db-port 5432 \
  --db-name golang_rest_api \
  --db-user postgres \
  --db-password postgres
```

The server listens on `localhost:8080` by default and automatically migrates the database schema on startup.

## Configuration

All settings can be passed as CLI flags or environment variables:

| Flag                | Environment variable | Default     | Description                          |
|----------------------|-----------------------|-------------|---------------------------------------|
| `--log-level`        | `LOG_LEVEL`            | `info`      | Log level (trace, debug, info, warn, error, fatal, panic, disabled) |
| `--log-error-stack`  | `LOG_ERROR_STACK`      | `true`      | Log full error stack traces           |
| `--http-server-host` | `HTTP_SERVER_HOST`     | `localhost` | HTTP server bind host                 |
| `--http-server-port` | `HTTP_SERVER_PORT`     | `8080`      | HTTP server bind port                 |
| `--db-host`          | `DB_HOST`              | `localhost` | PostgreSQL host                       |
| `--db-port`          | `DB_PORT`              | `5432`      | PostgreSQL port                       |
| `--db-name`          | `DB_NAME`              | —           | PostgreSQL database name              |
| `--db-user`          | `DB_USER`              | —           | PostgreSQL user                       |
| `--db-password`      | `DB_PASSWORD`          | —           | PostgreSQL password                   |
| `--db-sslmode`       | `DB_SSLMODE`           | `disable`   | PostgreSQL SSL mode                   |

## Design patterns

- **Repository pattern** — each domain (`AudiobookRepository`, `TagRepository`, `LanguagesRepository` in `repositories/`) is defined as an interface with a concrete GORM-backed implementation, decoupling handlers from persistence details.
- **Dependency injection** — `server.New` wires in the router, driver, and logger, and `cmd.Run` constructs the repositories and injects them into the `Server.Repositories` struct, rather than having handlers reach for globals.
- **Adapter** — `server/driver` defines a small `Server` interface (`ListenAndServe`/`Shutdown`) that the concrete `Driver` (wrapping `http.Server`) satisfies, isolating the app from the standard library's HTTP server type.
- **Decorator / chain of responsibility** — HTTP middleware (`JsonContentTypeResponseHandler`, the `LoggerChain` built with `justinas/alice`) wraps handlers to layer cross-cutting concerns like logging and content-type headers without touching handler logic.
- **DTO / data transfer object (model-view separation)** — `server/json.go` defines `*JSON` types (`AudiobookJSON`, `TagJSON`, `LanguageJSON`, ...) with `New*Response` constructor functions that map GORM models to API response shapes, keeping persistence models decoupled from the wire format.
- **Unit of work** — `doTransactionSlice` in `repositories/audiobook.go` runs a slice of operations inside a single DB transaction, committing only if all succeed and rolling back otherwise.
- **Centralized/wrapped error handling** — `errs.E(...)` builds a chain of typed `*errs.Error` values (op, code, wrapped error) that is unwound by `errs.HTTPErrorResponse`, similar to the Upspin-style error wrapping pattern, giving consistent error codes and stack context across layers.

## API endpoints

| Method | Path                      | Description                          |
|--------|---------------------------|---------------------------------------|
| GET    | `/v1/audiobook`           | List audiobooks (paginated)          |
| POST   | `/v1/audiobook`           | Create an audiobook (multipart form) |
| GET    | `/v1/audiobook/{id}`      | Get an audiobook by ID               |
| PUT    | `/v1/audiobook/{id}`      | Edit an audiobook                    |
| DELETE | `/v1/audiobook/{id}`      | Delete an audiobook                  |
| GET    | `/v1/tags`                | List tags                            |

Requests are scoped to an account via the request headers set by the caller's auth layer.

## License

Licensed under the terms of the [GNU General Public License v3.0](LICENSE).
