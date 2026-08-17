# Task Manager API

A REST API for managing personal todos, written in Go. Users register, log in
for a JWT, and manage todos scoped to their own account.

## Stack

| Concern | Library |
| --- | --- |
| Routing | [chi](https://github.com/go-chi/chi) v5 |
| Database | PostgreSQL 13 via [sqlx](https://github.com/jmoiron/sqlx) + [lib/pq](https://github.com/lib/pq) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Auth | [golang-jwt](https://github.com/golang-jwt/jwt) v5, HS256 |
| Validation | [validator](https://github.com/go-playground/validator) v10 |
| Passwords | bcrypt |
| Logging | [logrus](https://github.com/sirupsen/logrus) |

Requires Go 1.22 or newer.

## Getting started

Start Postgres:

```bash
docker compose up -d
```

This runs `postgres:13` on host port **5434**, with database `todo` and user
`local` / `local`. Data persists in `./pgdata`.

Set the environment and run the server:

```bash
export DB_HOST=localhost
export DB_PORT=5434
export DB_NAME=todo
export DB_USER=local
export DB_PASS=local
export JWT_SECRET_KEY=change-me

go run ./cmd
```

The server listens on **:8080**. Migrations from `database/migrations` run
automatically at startup, so there is no separate migrate step.

Run it from the repository root — the migration source is the relative path
`file://database/migrations`, so starting from elsewhere fails to find it.

### Environment variables

| Variable | Description |
| --- | --- |
| `DB_HOST` | Postgres host |
| `DB_PORT` | Postgres port (`5434` with the bundled compose file) |
| `DB_NAME` | Database name |
| `DB_USER` | Database user |
| `DB_PASS` | Database password |
| `JWT_SECRET_KEY` | Secret used to sign and verify tokens |

## Authentication

`POST /v1/login` returns a JWT. Send it on protected routes in a **`token`**
header — not `Authorization: Bearer`:

```bash
curl -H "token: <jwt>" http://localhost:8080/v1/todo
```

Tokens expire 10 minutes after issue. Each token also carries a session id;
logging out or deleting the account archives the session, which invalidates the
token immediately even though it has not expired yet.

## API

Base path `/v1`. Everything below `/user` and `/todo` requires a token.

### Public

| Method | Path | Body | Success |
| --- | --- | --- | --- |
| `POST` | `/register` | `name`, `email`, `password` | `201` |
| `POST` | `/login` | `email`, `password` | `200` + token |

`password` must be 6–15 characters.

### User

| Method | Path | Success | Notes |
| --- | --- | --- | --- |
| `GET` | `/user/profile` | `200` | Returns id, name, email |
| `POST` | `/user/logout` | `200` | Archives the current session |
| `DELETE` | `/user/delete` | `200` | Archives the account, its todos, and every session |

Account deletion runs in a single transaction, so it either fully succeeds or
leaves nothing changed.

### Todos

| Method | Path | Success | Notes |
| --- | --- | --- | --- |
| `POST` | `/todo` | `201` | Body: `name`, `description` |
| `GET` | `/todo` | `200` | Supports `?keyword=` and `?completed=true\|false` |
| `PUT` | `/todo/{todoId}/mark-completed` | `200` | |
| `DELETE` | `/todo/{todoId}/` | `200` | |
| `DELETE` | `/todo/delete-all` | `200` | Archives every todo of the caller |

Todo names are unique per user among active todos.

### Status codes

| Code | Meaning |
| --- | --- |
| `200` | Success |
| `201` | User or todo created |
| `400` | Malformed body, failed validation, or a bad query parameter |
| `401` | Missing, invalid, or expired token; wrong email or password |
| `404` | Todo or user does not exist, is archived, or belongs to someone else |
| `409` | Email already registered, or a todo with that name already exists |
| `500` | Something failed server-side |

Errors come back as:

```json
{
  "id": "aB3xY9z",
  "messageToUser": "todo not found",
  "developerInfo": "todo not found",
  "error": "sql: no rows in result set",
  "statusCode": 404,
  "isClientError": true
}
```

The `id` is generated per error and written to the server log, so quoting it in
a bug report points straight at the matching log line.

Note that `developerInfo` and `error` carry internal detail — the raw driver
message above names the failure exactly. That is useful in development, but
before this runs anywhere public those two fields should be logged only, not
serialized to the client.

## Data model

Three tables, created by `database/migrations`:

- **users** — id, name, email, password hash
- **todos** — id, user_id, name, description, is_completed
- **user_session** — id, user_id, created_at

Nothing is ever hard deleted. Rows carry `archived_at`, and every query filters
on `archived_at IS NULL`. Unique constraints are partial indexes limited to
active rows, so an email or todo name frees up once the row is archived.

## Layout

```
cmd/                  entrypoint, graceful shutdown
server/               router, timeouts
handlers/             HTTP handlers: parse, validate, respond
middlewares/          auth, CORS, panic recovery
database/             connection, migrations, transaction helper
database/dbHelper/    SQL queries
models/               request, response, and database structs
utils/                JSON, JWT, bcrypt, validation, error responses
```

Handlers own HTTP concerns and `dbHelper` owns SQL; helpers return errors and
never write to the response themselves. Handlers that need more than one
statement to be atomic wrap them in `database.Tx`, which commits on success and
rolls back on any returned error.

## Testing with Postman

`todo.postman_environment.json` defines `v1` (`localhost:8080/v1`) and an empty
`token` variable. Import it, log in, and paste the returned token into `token`.
