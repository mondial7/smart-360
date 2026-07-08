# ADR 0006: Server-Sent Events for consolidation progress

- **Status**: accepted
- **Date**: 2026-07-08
- **Deciders**: Marco Mondini

## Context

Generating a consolidation is the one genuinely long-running operation in the
app: a per-submission Gemini moderation pass (fanned out, up to ~25s each)
followed by a synthesis call (up to ~60s). In the old SPA this was a single
synchronous POST — the user stared at a spinner with no feedback and a fragile
long request. We want live progress in the server-rendered app.

## Decision

We will stream consolidation progress over **Server-Sent Events (SSE)**,
consumed by the htmx SSE extension.

- A CSRF-protected `POST /rounds/{id}/consolidate` returns a small panel that
  opens an SSE connection to `GET /rounds/{id}/consolidate/stream`.
- The stream handler runs `ai.Consolidate`, which reports progress via a
  `Progress` callback. Because moderation runs across several goroutines, the
  callback pushes events onto a channel that the single stream goroutine drains
  and writes as SSE `progress` events — keeping exactly one writer on the
  `ResponseWriter`.
- On completion the handler persists the consolidation and moderation logs and
  emits a final `done` event that navigates the browser to the result.

SSE (not WebSockets) because the data flow is one-directional (server→client)
and SSE is plain HTTP with trivial client wiring via htmx.

## Consequences

- Users see live progress ("Screening submissions 3/5", "Synthesizing…") instead
  of an opaque wait.
- The pattern is single-instance: generation happens inside the streaming
  request. That's fine for this app's scale; a multi-instance deployment would
  need a shared job/queue and is explicitly out of scope now.
- The single-writer-via-channel design is required for correctness — writing to
  an `http.ResponseWriter` from multiple goroutines is unsafe.

## Alternatives considered

- **WebSockets** — bidirectional and heavier; unnecessary for one-way progress.
- **Client polling a status endpoint** — simpler transport but needs a durable
  job record and more round-trips; SSE gives push updates for free over HTTP.
- **Keep the synchronous POST** — leaves the original bad UX and fragile long
  request unaddressed.
