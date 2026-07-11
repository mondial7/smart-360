# Smart 360 Feedback — Project Guide

This guide is the entry point for anyone (humans or AI assistants) working on the
codebase. It covers architecture, conventions, and common tasks.

> Product context: see [`README.md`](README.md).
> Security & AI data flow: see [`SECURITY.md`](SECURITY.md).
> Architecture decisions & their rationale: see [`docs/adr/`](docs/adr/README.md).
> Roadmap: tracked in [GitHub Issues](https://github.com/mondial7/smart-360/issues).

---

## Stack

Smart 360 is a **single server-rendered Go web application**. There is no
separate frontend build and no JSON API.

- **Language**: Go 1.26
- **HTTP**: `net/http` + [chi](https://github.com/go-chi/chi) router
- **UI**: Go `html/template` + [htmx](https://htmx.org) + **SSE** (no SPA, no JS
  framework, no chart library — charts are server-rendered SVG)
- **Database**: PostgreSQL 16, via [pgx](https://github.com/jackc/pgx) with
  hand-written SQL (no ORM)
- **Auth**: Google OAuth 2.0 → server-side session in an HttpOnly cookie
- **AI**: Google Gemini (`generative-ai-go`) for moderation + synthesis
- **PDF**: `github.com/go-pdf/fpdf` (server-side rendering)
- **Assets**: templates and static files embedded via `//go:embed` → one binary
- **Deploy**: Docker + Docker Compose

The rationale for each of these is recorded in the ADR log; start at
[ADR-0002](docs/adr/0002-rewrite-to-go-htmx-postgres.md).

---

## Repository Layout

```
cmd/server/            # entrypoint: config → pgx pool → migrate → seed → router → serve
  main.go
  router.go            # chi routes + middleware wiring
internal/
  config/              # env parsing
  db/                  # pgx pool, migration runner, migrations/*.sql, seed
  models/              # domain structs (UUID string IDs, json tags)
  repo/                # repository interfaces + pgx impls (pg_*.go) + fakes + gateway tests
  ai/                  # Gemini moderation + synthesis + competency aggregation (pure, no HTTP/DB)
  pdf/                 # fpdf consolidation renderer (pure)
  auth/                # Google OAuth, sessions, role middleware, CSRF
  view/                # template Renderer + SVG chart generators
  handlers/            # HTTP handlers (one file per domain) + routes.go
web/
  embed.go             # //go:embed of templates + static
  templates/           # base.html, <page>.html, partials/
  static/              # htmx.min.js, sse.js, css
docs/adr/              # architecture decision records
docker-compose.yml     # postgres + app
Dockerfile
```

---

## Architecture in one paragraph

`cmd/server` wires everything and mounts routes. Requests hit **handlers**
(`internal/handlers`), which pull data through **repositories**
(`internal/repo`) and render HTML via the **view Renderer** (`internal/view`).
Three packages are deliberately free of HTTP/DB concerns and are unit-tested in
isolation: **`ai`** (Gemini moderation + synthesis + the deterministic
competency math), **`pdf`** (fpdf), and the SVG chart generators in **`view`**.
**`auth`** provides OAuth login, session cookies, role middleware, and CSRF.

---

## Data flow: rendering

- The `view.Renderer` parses all templates once. `Page` renders a content
  template inside `base.html`; `Fragment` renders a standalone named template for
  htmx swaps.
- Handlers build a `view.PageData` envelope via `h.page(r, title, active,
  contentName, data)`, which injects the current user and CSRF token.
- Charts are server-rendered SVG: `view.RadarSVG` / `view.DonutSVG`, exposed as
  the `radarSVG` / `donutSVG` template funcs. **Do not** add a JS chart library.

## Data flow: auth

- All app routes sit behind `auth.RequireAuth` (loads the session user into the
  request context) and `auth.ProtectCSRF`. Retrieve the user with
  `auth.UserFrom(r.Context())`.
- Role gates: `RequireAdmin`, `RequireTeamAdminOrAdmin`, and `RequireTeamScope`
  (team admin limited to their own `:id` team).
- The session cookie is `<sessionID>.<HMAC>`, HMAC-signed with `SESSION_SECRET`.
  CSRF token = `HMAC("csrf:"+sessionID)`, sent by htmx via `X-CSRF-Token` (the
  `base.html` `hx-headers` attribute) or a hidden `csrf_token` form field.

---

## Database

Schema lives in `internal/db/migrations/*.sql` and is applied on boot by
`db.Migrate` (forward-only; tracked in `schema_migrations`). IDs are `uuid`.

| Table | Notes |
|-------|-------|
| `users` | `email` unique, `role` (admin/team_admin/member), `team_id` (denormalized single-team pointer) |
| `teams` | `team_admin_id` |
| `team_members` | join table (replaces the old embedded member array) |
| `templates` | `slug` unique; `questions`/`competencies` as `jsonb`; seeded on every boot by `db.Seed` |
| `feedback_rounds` | `subject_id`, `created_by_id`, `template_id`, `status` (draft/active/closed/shared), `deadline` |
| `round_reviewers` | join table; `UNIQUE(round_id, reviewer_id)` |
| `submissions` | `responses` jsonb, `ratings` jsonb, peer `relationship`/`interaction_frequency`, `private_notes`; `UNIQUE(round_id, reviewer_id)` |
| `consolidations` | typed `jsonb` for the list/map/sub-doc fields; `manager_only_channel` stripped from anything the subject sees |
| `audit_logs` | cached display fields, **not** FK-constrained (survives deletes) |
| `moderation_logs` | per-submission scrub audit trail |
| `sessions` | server-side login sessions (opaque id in the cookie) |

See [ADR-0007](docs/adr/0007-normalize-schema-reviewers-and-members.md) for the
normalization rationale.

---

## Round lifecycle

`draft` → `active` → `closed` → `shared`

- An admin / team admin creates a round (draft) with a subject, reviewers, and a
  template (defaults to the `default` slug).
- Starting it opens submissions. Both peers **and** the subject submit — the
  subject's submission is `is_self=true` and feeds the self-vs-peers delta.
- Closing freezes submissions and unlocks consolidation.
- Consolidation runs two Gemini passes — a per-submission **moderation scrub**
  (persists `moderation_logs`) then **synthesis** — streamed to the browser over
  **SSE** ([ADR-0006](docs/adr/0006-sse-for-consolidation.md)). Deterministic
  aggregates (competency ratings, voice counts) are computed server-side. With no
  `GEMINI_API_KEY`, moderation is a no-op and synthesis falls back to a non-AI
  combine.
- Sharing flips the round to `shared` and exposes the consolidation to the subject
  (in-app + PDF) — but the `manager_only_channel` is stripped from the subject.

Status changes are recorded in `audit_logs`; scrubs in `moderation_logs`.

---

## Test strategy

Same three-layer pyramid, one command: `go test ./...`.

1. **Unit** — pure logic in `ai` (validation, competency math, fallback synthesis),
   `view` (chart geometry), `auth` (session/CSRF crypto).
2. **In-memory integration** — handlers against `repo` fakes via a real
   `httptest` server (`internal/handlers/*_test.go`); asserts rendered HTML,
   redirects, and status codes. No database needed.
3. **Gateway** — pgx repositories against a real Postgres via
   `testcontainers-go` (`internal/repo/gateway_*_test.go`). Skipped under
   `-short` or when Docker is unavailable.

```bash
go test ./...            # full pyramid (gateway tests need Docker)
go test -short ./...     # skip the container-backed gateway tests
```

There is intentionally **no browser/E2E layer**.

---

## Quick start (development)

```bash
docker compose up -d postgres      # start Postgres
cp .env.example .env               # set SESSION_SECRET (or rely on DEV_MODE), Google/Gemini creds
go run ./cmd/server                # migrates + seeds + serves :8080
```

Open <http://localhost:8080>. With `DEV_MODE=true`, sign in via the dev-login on
the login page (or `GET /auth/dev-login?email=you@example.com`). The user whose
email matches **`ADMIN_EMAIL`** is the admin; everyone else starts as a member.

---

## Config (env)

| Var | Default | Notes |
|-----|---------|-------|
| `PORT` | `8080` | listen port |
| `APP_URL` | `http://localhost:8080` | external base URL |
| `DATABASE_URL` | local compose DSN | Postgres connection string |
| `SESSION_SECRET` | dev fallback if `DEV_MODE` | required in prod; signs cookies |
| `ADMIN_EMAIL` | empty | the user with this email is (and stays) the global admin; deterministic bootstrap |
| `GOOGLE_CLIENT_ID` / `_SECRET` / `_REDIRECT_URL` | — | OAuth |
| `GEMINI_API_KEY` | empty | empty → moderation no-op + non-AI fallback |
| `LOG_FORMAT` | `text` | `text` or `json` (structured logs for shippers) |
| `DEV_MODE` | `false` | enables `/auth/dev-login`, relaxes `Secure`, dev fallbacks. **Never true in prod.** |

---

## Common tasks

### Add a page / route
1. Add a handler method in `internal/handlers/<domain>.go` that builds a
   `view.PageData` and calls `h.View.Page` (or `Fragment` for an htmx swap).
2. Register it in `internal/handlers/routes.go` (`MountAppRoutes`) with the right
   role middleware. Public/auth routes live in `cmd/server/router.go`.
3. Add a content template `web/templates/<page>.html` defining `{{define
   "<name>_content"}}` and reference it by name from the handler.
4. Add a handler test in `internal/handlers` against the fakes.

### Change the schema
1. Add a new `internal/db/migrations/000N_*.sql` (forward-only — never edit an
   applied migration).
2. Update the struct in `internal/models` and the affected repo(s) in
   `internal/repo` (pg impl **and** fake).
3. Extend the gateway test in `internal/repo`.

### Add a repository method
Add it to the interface in `internal/repo/repo.go`, then implement in the pg type
(`pg_*.go`) and the fake (`fakes.go`). The compile-time assertions in
`asserts.go` keep them in sync.

---

## Coding style (Go)

- `gofmt` / `go vet` clean; CI fails on either.
- Keep handlers thin: pull data through repos, push logic into `ai`/`pdf`/helpers.
- Check errors and return early; user-facing messages are generic, details logged.
- `internal/ai` and `internal/pdf` are intentionally dependency-light and free of
  HTTP/DB. Keep them that way.
- Don't reintroduce a JS framework or client-side chart library; htmx + SSE +
  server-rendered SVG is the deliberate UI strategy.

---

## Notes for AI assistants

- Prefer the smallest change that solves the problem; this project values
  simplicity over premature optimization.
- Record any significant architecture decision as a new ADR in `docs/adr/`.
- `dev-login` must stay gated behind `DEV_MODE`.
- When touching the consolidation flow, preserve two invariants: the
  `manager_only_channel` is **never** shown to the subject, and the Gemini API
  key is scrubbed from any logged error (`ai.sanitiseErr`).
- Conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
  Do **not** add Claude as a commit co-author.
