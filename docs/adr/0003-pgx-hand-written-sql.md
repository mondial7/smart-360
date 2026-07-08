# ADR 0003: Data access with pgx and hand-written SQL

- **Status**: accepted
- **Date**: 2026-07-08
- **Deciders**: Marco Mondini

## Context

The rewrite (ADR-0002) moves the datastore from MongoDB to Postgres. We need a
Go data-access approach. Options ranged from a full ORM (GORM) to a codegen
layer (sqlc) to hand-written SQL over a driver (pgx). The project values
simplicity and reviewable queries, and the schema is small (roughly ten tables).

## Decision

We will use **`github.com/jackc/pgx/v5`** with **hand-written SQL** in the
`internal/repo` package. Each aggregate has a repository interface with a pgx
implementation and an in-memory fake. Queries are plain SQL string constants;
rows are scanned into domain structs. jsonb columns are marshaled/unmarshaled
with `encoding/json` via small helpers (`mustJSON`/`decodeJSON`).

IDs are UUID strings end to end; a `querier` interface lets a repository method
run against either the pool or a transaction.

## Consequences

- Every query is visible and greppable; no hidden N+1s or ORM query DSL to learn.
- No codegen step in the build (unlike sqlc): fewer moving parts, at the cost of
  hand-writing scan code and column lists.
- The fakes make handler tests fast and DB-free; a `testcontainers`-backed
  gateway suite validates the real SQL against Postgres.
- We take on manual discipline for column-list/scan alignment. The compile-time
  interface assertions and gateway tests catch drift.

## Alternatives considered

- **sqlc** — type-safe and pleasant, but adds a codegen tool and generated code
  to the repo for a schema small enough to hand-write.
- **GORM** — least SQL to write, but reflection/magic and hidden queries cut
  against the "simpler, reviewable" goal that motivated the rewrite.
