# Article Service

Small REST API in Go. It exposes endpoints to create and fetch articles and persists them in PostgreSQL.

## Quick start (local)

1) Set env vars (minimum):
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/articles_service?sslmode=disable"
# optional
export HTTP_PORT=8080
```

2) Run migrations (requires `migrate` CLI):
```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

3) Start the API:
```bash
go run ./cmd/api
# or
make run
```

Health check: `GET /healthz` (pings DB).

## Quick start (Docker Compose)

```bash
make up 💄
# or: docker compose up --build
```

This starts Postgres and the API, runs migrations via the `migrate` image, and exposes the service on `http://localhost:8080`.

## API

- `POST /article` – create an article.
  - Body: `{"title":"I'm NARUTO UZUMAKI"}`
  - 201 response: `{"id":1,"title":"...","created_at":"2025-12-17T19:38:28.991780128Z"}`
- `GET /article/{id}` – fetch a single article by ID.
  - 200 response: same response as above.

Errors are returned as `{"error":"message"}` with the status codes.

Example:

```bash
curl -X POST http://localhost:8080/article \
  -H "Content-Type: application/json" \
  -d '{"title":"Minecraft OneLove"}'

curl http://localhost:8080/article/<returned-id>
```

## Testing

```bash
go test ./...
# or
make test
```

## Architecture

- **Domain**: core entity and error definitions (`internal/domain`).
- **Use case**: business rules and validation (`internal/usecase`).
- **Adapters**: HTTP transport and storage implementations (`internal/adapter/http`, `internal/adapter/storage/postgres`).
- **Framework/driver**: server wiring (`cmd/api`, `internal/server`), plus migrations in `db/migrations`.
