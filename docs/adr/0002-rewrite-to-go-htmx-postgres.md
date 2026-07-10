# ADR 0002: Rewrite to a server-rendered Go + htmx + Postgres stack

- **Status**: accepted
- **Date**: 2026-07-08
- **Deciders**: Marco Mondini

## Context

The original Smart 360 is two applications: a Go/Gin JSON API backed by MongoDB
and Google Gemini, and a separate Vue 3 SPA (Vite + Pinia + TypeScript + SASS).
This split imposes ongoing cost disproportionate to the size of the product:

- Two languages, two toolchains, two dependency trees, two test frameworks.
- A hand-maintained JSON API contract and duplicated types (Go structs ↔ TS
  interfaces) that drift.
- Client-side routing, state stores, and an auth token dance (JWT in
  `localStorage`) for a UI that is almost entirely load-on-mount + form submit,
  with no real-time collaboration and no offline needs.

The product is a good fit for server rendering: mostly forms and read views, one
genuinely long-running operation (AI consolidation), and a single small team.

## Decision

We will rewrite Smart 360 as a **single server-rendered Go web application**:

- **`html/template` + htmx** for the UI, with **SSE** for the one long-running
  operation (feedback consolidation progress).
- **Postgres** as the datastore, replacing MongoDB.
- Feature parity with the existing app is the target for the first release.

Three existing Go layers are stack-agnostic and are carried over largely intact:
the Gemini moderation + synthesis passes, the `fpdf` PDF renderer, and the
competency aggregation math. The Vue SVG charts (radar, donut) are re-expressed
as server-rendered SVG in Go templates, removing the need for any JS chart
library.

The rewrite happens on a branch and replaces the `backend/` and `frontend/`
trees once verified locally.

## Consequences

- One language, one build, one `go test ./...`, one deployable binary (templates
  and static assets embedded). No Node/Vite in the toolchain.
- No API contract to keep in sync; the server owns rendering end to end.
- We give up the rich client-side interactivity a SPA affords. htmx covers the
  interactions this product actually uses; anything needing heavy client state
  would be a deliberate, localized exception.
- A data migration and a rebuild of every screen are required — a large one-time
  cost, taken knowingly for lower long-run maintenance.

Follow-on decisions are recorded separately: pgx + hand-written SQL (ADR-0003),
session-cookie auth (ADR-0004), chi + `html/template` (ADR-0005), SSE for
consolidation (ADR-0006), and the normalized relational schema (ADR-0007).

## Alternatives considered

- **Keep Vue, swap only Mongo→Postgres** — leaves the largest source of
  complexity (the two-app split) untouched.
- **Keep the SPA but adopt a typed RPC layer (e.g. tRPC-style/codegen)** — still
  two runtimes and toolchains; more machinery, not less.
- **Rails/Django server-rendered rewrite** — would discard the reusable Go AI,
  PDF, and aggregation code and introduce a new language/runtime to operate.
