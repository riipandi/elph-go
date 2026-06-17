# Datastore

SQLite database layer for Elph, backed by [Turso](https://turso.tech) (`turso.tech/database/tursogo`). Provides stateful persistence for sessions, messages, todos, goals, skill cache, and snip command tracking.

## Purpose

Elph was originally stateless — session data lived only in JSONL log files and in-memory structures. The datastore adds a proper relational layer so that:

1. **Sessions survive restarts** — conversation history, todos, and goals persist across TUI sessions.
2. **Querying is possible** — aggregate stats, search history, filter by date/command without parsing JSONL.
3. **Multiple packages share one DB** — core, snip, and future packages register migrations independently.
4. **Turso Cloud is optional** — local SQLite works out of the box; Turso sync is a config toggle.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Application (cmd/elph/root.go PersistentPreRun)       │
│    calls datastore.Open(ctx) on startup                 │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│  internal/datastore/                                    │
│    datastore.go   — Open, DB, Close, buildDSN           │
│    migrator.go    — Register, RunMigrations, splitSQL   │
│    migrations.go  — core migrations 1-5                 │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│  turso.tech/database/tursogo  (registered as "turso")   │
│    Local file  → direct SQLite                          │
│    libsql://   → Turso Cloud (with authToken)           │
└─────────────────────────────────────────────────────────┘
```

## Database location

| Mode            | Path / URL                                                      |
|-----------------|-----------------------------------------------------------------|
| Default (local) | `~/.local/share/elph/metadata.db`                               |
| Custom local    | `database.url` in `settings.json` (file path)                   |
| Turso Cloud     | `database.url` = `libsql://your-db.turso.io` + `database.token` |

Controlled by `internal/appdir` (`DataDir()` → `~/.local/share/elph/`).

## Configuration

Add to `~/.elph/settings.json`:

```json
{
  "database": {
    "url": "libsql://your-db.turso.io",
    "token": "your-auth-token"
  }
}
```

When both fields are empty, Elph uses the default local SQLite file.

## Migration system

### How it works

1. On startup, `datastore.Open()` calls `RunMigrations()`.
2. `RunMigrations()` creates the `app_migrations` tracking table if missing.
3. It reads `MAX(version)` from `app_migrations` to find the current level.
4. Each registered `Migration` with `version > current` is applied in a transaction.
5. The version is recorded in `app_migrations` on success.

### Registering migrations

Any package can register migrations via `init()`:

```go
package mypackage

import "github.com/riipandi/elph/internal/datastore"

func init() {
    datastore.Register(datastore.MigrationSet{
        Source: "mypackage",
        Migrations: []datastore.Migration{
            {
                Version: 7,
                Name:    "create_my_table",
                Up: `CREATE TABLE IF NOT EXISTS my_table (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    data TEXT NOT NULL
                );`,
            },
        },
    })
}
```

**Rules:**
- Use the next available version number (check existing migrations first).
- Version numbers are globally unique across all sources — they determine execution order.
- Each migration runs in a transaction; if any statement fails, the whole migration rolls back.
- The `Up` SQL is split on `;` — multiple statements are supported.

### Current migrations

| Version | Source | Name                       | Tables created  |
|---------|--------|----------------------------|-----------------|
| 1       | core   | create_sessions_table      | `sessions`      |
| 2       | core   | create_messages_table      | `messages`      |
| 3       | core   | create_todos_table         | `todos`         |
| 4       | core   | create_goals_table         | `goals`         |
| 5       | core   | create_skill_cache_table   | `skill_cache`   |
| 6       | snip   | create_snip_commands_table | `snip_commands` |

### Schema details

#### `sessions`

| Column          | Type     | Notes                             |
|-----------------|----------|-----------------------------------|
| `id`            | TEXT PK  | TypeID (`sess_...`)               |
| `created_at`    | DATETIME |                                   |
| `updated_at`    | DATETIME |                                   |
| `work_dir`      | TEXT     | Working directory for the session |
| `provider_id`   | TEXT     | Active provider                   |
| `model_id`      | TEXT     | Active model                      |
| `agent_mode`    | TEXT     | `build`, `plan`, `ask`, `brave`   |
| `system_prompt` | TEXT     | Assembled system prompt           |
| `metadata`      | TEXT     | JSON blob for extensible metadata |

#### `messages`

| Column         | Type       | Notes                                       |
|----------------|------------|---------------------------------------------|
| `id`           | INTEGER PK | Auto-increment                              |
| `session_id`   | TEXT FK    | References `sessions(id)` ON DELETE CASCADE |
| `role`         | TEXT       | `user`, `assistant`, `system`, `tool`       |
| `content`      | TEXT       | Message text                                |
| `tool_call_id` | TEXT       | For tool result messages                    |
| `tool_calls`   | TEXT       | JSON array of tool calls                    |
| `created_at`   | DATETIME   |                                             |

#### `todos`

| Column       | Type       | Notes                                       |
|--------------|------------|---------------------------------------------|
| `id`         | INTEGER PK | Auto-increment                              |
| `session_id` | TEXT FK    | References `sessions(id)` ON DELETE CASCADE |
| `content`    | TEXT       | Todo text                                   |
| `completed`  | BOOLEAN    | Default `0`                                 |
| `position`   | INTEGER    | Sort order                                  |
| `created_at` | DATETIME   |                                             |
| `updated_at` | DATETIME   |                                             |

#### `goals`

| Column                 | Type       | Notes                                       |
|------------------------|------------|---------------------------------------------|
| `id`                   | INTEGER PK | Auto-increment                              |
| `session_id`           | TEXT FK    | References `sessions(id)` ON DELETE CASCADE |
| `objective`            | TEXT       | Goal description                            |
| `completion_criterion` | TEXT       | How to determine completion                 |
| `status`               | TEXT       | `active`, `complete`, `blocked`             |
| `turns_used`           | INTEGER    | Agent turns consumed                        |
| `tokens_used`          | INTEGER    | Tokens consumed                             |
| `wall_clock_ms`        | INTEGER    | Elapsed time                                |
| `wall_clock_budget_ms` | INTEGER    | Time limit                                  |
| `turn_budget`          | INTEGER    | Turn limit                                  |
| `token_budget`         | INTEGER    | Token limit                                 |
| `created_at`           | DATETIME   |                                             |
| `completed_at`         | DATETIME   | NULL if not completed                       |

#### `skill_cache`

| Column       | Type       | Notes                         |
|--------------|------------|-------------------------------|
| `id`         | INTEGER PK | Auto-increment                |
| `skill_name` | TEXT       | Skill identifier              |
| `skill_hash` | TEXT       | Content hash for invalidation |
| `content`    | TEXT       | Cached skill content          |
| `created_at` | DATETIME   |                               |
| `expires_at` | DATETIME   | Optional TTL                  |

UNIQUE constraint on `(skill_name, skill_hash)`.

#### `snip_commands` (snip package)

| Column          | Type       | Notes                        |
|-----------------|------------|------------------------------|
| `id`            | INTEGER PK | Auto-increment               |
| `timestamp`     | DATETIME   | When the command was tracked |
| `original_cmd`  | TEXT       | Original command             |
| `snip_cmd`      | TEXT       | Compressed/snipped version   |
| `input_tokens`  | INTEGER    | Original token count         |
| `output_tokens` | INTEGER    | Output token count           |
| `saved_tokens`  | INTEGER    | Tokens saved                 |
| `savings_pct`   | REAL       | Savings percentage           |
| `exec_time_ms`  | INTEGER    | Execution time               |

## API reference

### `datastore.Open(ctx) (*sql.DB, error)`

Opens the database (once per process). Runs migrations automatically. Returns the global connection.

### `datastore.DB() *sql.DB`

Returns the global connection. Panics if `Open` was not called.

### `datastore.Close() error`

Closes the database connection.

### `datastore.Register(set MigrationSet)`

Registers a migration set. Call in `init()`.

### `datastore.RunMigrations(ctx, db) error`

Applies all pending migrations. Called automatically by `Open`.

## Testing

Run datastore tests:

```bash
go test ./internal/datastore/... -v
```

Tests cover:
- DSN building (local, libsql, https, with/without tokens)
- Directory creation for local databases
- SQL statement splitting (quotes, semicolons in strings)
- Migration registry
- Full open + migrate lifecycle

## Dependencies

- `turso.tech/database/tursogo` — SQLite-compatible driver (no CGO, uses purego)
- `github.com/riipandi/elph/internal/settings` — reads `database.url` and `database.token`
- `github.com/riipandi/elph/internal/appdir` — XDG-compliant data directory paths
