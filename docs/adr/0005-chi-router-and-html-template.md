# ADR 0005: chi router + html/template rendering

- **Status**: accepted
- **Date**: 2026-07-08
- **Deciders**: Marco Mondini

## Context

The old backend used Gin to serve a JSON API. The rewrite serves HTML directly.
We need an HTTP router and a rendering approach for a server-rendered app driven
by htmx.

## Decision

We will use **`github.com/go-chi/chi/v5`** over `net/http`, and Go's standard
**`html/template`** for rendering.

Templates are embedded via `//go:embed` and parsed once into a single set by a
small `view.Renderer`. It renders either **full pages** (a content template
injected into `base.html` through a `partial` template func) or **fragments**
(standalone named templates) for htmx swaps. Chart SVGs (radar, donut) are
produced by template funcs, so no client-side chart library is needed. All
templates and static assets are embedded, so the server is a single binary.

## Consequences

- Idiomatic stdlib HTTP; chi adds only routing, URL params, and middleware
  composition, which compose cleanly with the auth middleware.
- `html/template` gives contextual auto-escaping by default — a security win
  for user-generated feedback content.
- One deployable artifact with no external asset directory.
- We give up Gin conveniences (binding, JSON helpers) we no longer need, and the
  richer client interactivity of a SPA; htmx covers this app's interactions.

## Alternatives considered

- **Keep Gin** — fine, but heavier than needed once we're not serving JSON; chi
  is closer to stdlib for an HTML app.
- **A templating engine (templ, quicktemplate)** — compile-time typed templates
  are attractive, but add a codegen/build step; stdlib `html/template` keeps the
  toolchain minimal and the auto-escaping guarantees are well understood.
