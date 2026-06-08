# [003] Account Management

> **Feature ID:** 003 · **Area:** Core (user-facing) · **Status:** refactored; not yet wired in a compiling entrypoint
>
> **Backend packages:** `internal/account` · `transport/http/handlers/account` · `transport/messaging/handlers/{portfolio,assets}` · `internal/storage/sqlx_account_store.go`
>
> **Related features:** [002] CSV imports · [009] Cashflow · [013] Portfolio · [024] Assets · [017] Event-driven messaging

## Overview

The `account` aggregate is the canonical scope every other feature hangs off. An account
is a named entity with an optional external identity (e.g. Google / Entra ID); cashflow,
portfolio, assets, and importer all reference it by ID. The feature is deliberately small:

- **Write:** create an account (`POST /accounts`).
- **Read:** look one up / test existence (`Exists`, `GetByID`) — consumed by other features
  as a collaborator, not (currently) exposed as a list endpoint.

Account creation is the system's central **fan-out trigger**: it publishes `account.created`,
and each feature creates its own per-account projection row in response (see [Events](#events)).

## Domain model

```mermaid
classDiagram
    class Commands {
        -CommandStore c
        -eventbus.Bus b
        +Create(ctx, id, externalID, name) UUID
    }
    class Queries {
        -QueryStore qs
        +Exists(ctx, id) bool
        +GetByID(ctx, id) Account
    }
    class CommandStore {
        <<interface>>
        +Create(ctx, acc) error
    }
    class QueryStore {
        <<interface>>
        +Exists(ctx, id) bool
        +GetByID(ctx, id) Account
    }
    class Account {
        +UUID ID
        +string ExternalID
        +string Name
        +time CreatedAt
        +time UpdatedAt
    }
    class SQLXAccountStore

    Commands ..> CommandStore : persists via
    Commands ..> Account : creates
    Commands ..> eventbus.Bus : publishes account.created
    Queries ..> QueryStore : reads via
    CommandStore <|.. SQLXAccountStore
    QueryStore <|.. SQLXAccountStore

    note for Account "NewAccount validates name (required), generates ID if absent"
    note for SQLXAccountStore "maps unique-violation to ErrAccountAlreadyExists"
```

## Data model

`account.accounts` is a root table (no FKs). It is referenced 1:1 by each feature's
per-account projection, all of which cascade-delete with the account.

```mermaid
erDiagram
    ACCOUNTS ||--o| CASHFLOW_ACCOUNTS  : "account_id (CASCADE)"
    ACCOUNTS ||--o| PORTFOLIO_ACCOUNTS : "account_id (CASCADE)"
    ACCOUNTS ||--o| ASSETS_ACCOUNTS    : "account_id (CASCADE)"
    ACCOUNTS ||--o{ LISTINGS           : "account_id (SET NULL)"

    ACCOUNTS {
        uuid id PK
        string external_id "nullable"
        string name UK "NOT NULL, unique"
        timestamptz created_at
        timestamptz updated_at
    }
    CASHFLOW_ACCOUNTS {
        uuid id PK
        uuid account_id FK "NOT NULL, unique"
    }
    PORTFOLIO_ACCOUNTS {
        uuid id PK
        uuid account_id FK "NOT NULL, unique"
        bool building "rebuild lock"
    }
    ASSETS_ACCOUNTS {
        uuid id PK
        uuid account_id FK "NOT NULL, unique"
    }
    LISTINGS {
        uuid id PK
        uuid account_id FK "nullable"
    }
```

Physical names: Postgres `account.accounts`; SQLite flattens to `accounts`. Uniqueness is
enforced on **`name`** (not `external_id`). There is no status column — an account has no
lifecycle/state machine.

## Create + fan-out flow

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant H as HTTP handler (Create)
    participant C as account.Commands
    participant S as SQLXAccountStore
    participant B as eventbus.Bus

    Client->>H: POST /accounts {name, external_id?}
    H->>H: validate (name required, <=255)
    H->>C: Create(name, externalID)
    C->>S: Create(account)
    S-->>C: ok (or ErrAccountAlreadyExists on name clash)
    C->>B: Publish(account.created, {AccID})
    C-->>H: account ID
    H-->>Client: 201 {id}

    Note over B: async — each feature projects the new account
    participant PP as portfolio AccountCreatedHandler
    participant AP as assets AccountCreatedHandler
    B->>PP: account.created
    PP->>PP: portfolio.Commands.CreateAccount (projection + build lock)
    B->>AP: account.created
    AP->>AP: assets.Commands.CreateAccount (projection)
```

## Validation & error mapping

| Condition | Error | HTTP status |
| --- | --- | --- |
| Empty/whitespace name | `ErrAccountNameRequired` (domain) / transport `isValid` | 400 |
| Name > 255 chars | transport `isValid` | 400 |
| Name already taken | `ErrAccountAlreadyExists` (unique violation on `name`) | 409 |
| Anything else | — | 500 |

`external_id` is free-form and optional; it is not format-validated and not unique.

## Events

| Topic | Constant | Payload | Published when | Consumers (current code) |
| --- | --- | --- | --- | --- |
| `account.created` | `account.TopicCreated` | `Created{AccID uuid.UUID}` | after a successful insert | portfolio projection, assets projection |

**Intended vs. implemented fan-out.** `FEATURES.md` describes the fan-out reaching
portfolio, cashflow, assets, **and** importer. In the refactored tree only **portfolio**
and **assets** `account.created` handlers exist; the `cashflow.accounts` projection table
exists but nothing populates it yet, and the importer has no account-projection handler in
the new structure. See [Refactor state](#refactor-state).

## Code map

| Path | Responsibility |
| --- | --- |
| `internal/account/account.go` | `Account` aggregate + `NewAccount` validation |
| `internal/account/commands.go` | `Commands.Create` (publishes `account.created`) + `CommandStore` |
| `internal/account/queries.go` | `Queries.Exists`/`GetByID` + `QueryStore` |
| `internal/account/events.go` | `TopicCreated` + `Created` payload |
| `internal/account/errors.go` | `ErrAccountNotFound` / `ErrAccountAlreadyExists` / `ErrAccountNameRequired` |
| `transport/http/handlers/account/create.go` | `POST /accounts` handler + DTOs + error mapping |
| `internal/storage/sqlx_account_store.go` | `SQLXAccountStore` implementing both store interfaces |
| `internal/bootstrap/accounts.go` | Seeds a default account on startup (via the store, bypassing `Commands`) |
| `transport/messaging/handlers/{portfolio,assets}/account_created.go` | Projection subscribers |

## Refactor state

- **No `account.NewCommands` constructor** exists and `Commands` has unexported fields, so
  the write side cannot yet be composed outside the package; the only route registration is
  in the stale `cmd/server/main.go`. No compiling entrypoint currently serves `POST /accounts`.
- **Bootstrap seeding** inserts the default account directly through the store, so it does
  **not** emit `account.created` — the seeded account's projections are not created via the
  event path.
- **Not implemented:** update/delete/list account operations; a `GET /accounts` read endpoint;
  cashflow and importer `account.created` projections.
