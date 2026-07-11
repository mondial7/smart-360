# Smart 360 Feedback

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> AI-powered anonymous 360° peer feedback for professional development.

Smart 360 collects structured anonymous feedback from your team and uses
Google Gemini to consolidate it into clear strengths, growth areas, and
concrete next steps. It is a **single server-rendered Go application** —
one binary, backed by Postgres, with no separate frontend build.

- **Source:** <https://github.com/mondial7/smart-360>
- **Architecture decisions:** [`docs/adr/`](docs/adr/README.md)
- **Contributor guide:** [`CLAUDE.md`](CLAUDE.md)

---

## Table of Contents

- [What it does](#what-it-does)
- [How a round works](#how-a-round-works)
- [Who can do what](#who-can-do-what)
- [Tech stack](#tech-stack)
- [Run it locally](#run-it-locally)
- [Configuration](#configuration)
- [Google OAuth setup](#google-oauth-setup)
- [Gemini API setup](#gemini-api-setup)
- [Testing](#testing)
- [Architecture](#architecture)
- [Security & privacy](#security--privacy)
- [Contributing & license](#contributing--license)

---

## What it does

Smart 360 runs structured peer-feedback rounds and uses Google Gemini to turn
raw answers into something a person can act on. Designed for scaleup
engineering / product teams where calibrated, growth-oriented feedback matters
more than a pile of platitudes.

- **Growth-framed, behaviourally anchored questions** — the default template
  uses Continue / Block / Amplify / Experiment framing. Question text lives in a
  `templates` table, so you can ship one template per role family without code
  changes.
- **Self-assessment + delta** — the subject answers the same questions in
  first-person; the AI surfaces blind spots, hidden strengths, and aligned
  themes (the highest-signal output of a 360).
- **Voice breakdown** — manager, peer, and direct-report inputs are synthesised
  as distinct streams so a manager's "ready for next level" doesn't average into
  a peer's "easy to work with".
- **Competency rubric (Likert)** — templates can declare competencies; reviewers
  rate 1–5 with a one-line justification. The consolidation shows per-voice
  averages, spread, and a wide-spread warning for calibration gaps.
- **Private-to-manager channel** — an optional note on peer submissions feeds a
  manager-only synthesis, stripped from the subject's view and PDF.
- **Dedicated moderation pass** — a separate Gemini call scrubs identity-targeted
  / personality-attack / off-topic content before synthesis, recorded in
  `moderation_logs`.
- **Live consolidation** — moderation + synthesis progress is streamed to the
  browser over Server-Sent Events.
- **PDF export** — a print-ready PDF of any shared consolidation (manager-only
  section omitted from the subject's copy).
- **Server-rendered analytics** — radar and donut charts drawn as SVG on the
  server, no JS chart library.
- **Role-based access + audit log** — Admin, Team Admin, Member; every status
  transition (`draft → active → closed → shared`) is recorded.
- **Single binary** — templates and assets are embedded via `//go:embed`.

Reviewer identity is anonymised to the subject (the breakdown shows only
relationship + frequency tags); the admin / round creator can see whose
submission is whose for accountability and audit.

## How a round works

1. **Admin (or Team Admin) creates a round** — picks the subject, a template,
   reviewers, and a deadline. The subject participates via self-assessment, not
   as a reviewer.
2. **The subject self-assesses** — same questions, first-person wording.
3. **Reviewers submit** — each declares their relationship (manager / report /
   peer / cross-functional) and interaction frequency (daily / weekly / monthly /
   rarely), answers the questions, rates competencies 1–5 with justifications,
   and can leave a private manager-only note.
4. **Round closes** — no further submissions.
5. **Moderation scrub** — every submission goes through a Gemini call that strips
   identity-targeted / personality-attack / off-topic content (logged).
6. **AI synthesis** — a second Gemini call produces the executive summary, top
   1–3 focus areas, growth opportunities, per-question summaries, the self-vs-
   peers delta, the manager / peer / report voice streams, and the manager-only
   synthesis. Progress streams live over SSE.
7. **Server-side aggregates** — competency averages per voice, spread, and
   reviewer counts are computed deterministically (not via the AI).
8. **Admin reviews and shares** — sharing flips the round to `shared` and unlocks
   the consolidation for the subject.
9. **Subject reads** — sees everything **except** the manager-only channel.

Without a `GEMINI_API_KEY`, moderation is a no-op and synthesis falls back to a
non-AI combine that stitches the raw answers together honestly.

## Who can do what

| Role | Can do |
|------|--------|
| **Admin** | Everything — manage users + teams, create any round, view all rounds, generate and share any consolidation. Bootstrapped via `ADMIN_EMAIL`. |
| **Team Admin** | Manage their team's members; create and manage rounds. |
| **Member** | Submit feedback when assigned as reviewer, read consolidations shared with them (in-app + PDF). |

---

## Tech stack

- **Go 1.26** · `net/http` + [chi](https://github.com/go-chi/chi)
- **UI:** Go `html/template` + [htmx](https://htmx.org) + SSE (server-rendered
  SVG charts — no SPA, no JS framework)
- **PostgreSQL 16** via [pgx](https://github.com/jackc/pgx) (hand-written SQL)
- **Auth:** Google OAuth 2.0 → HttpOnly session cookie
- **AI:** Google Gemini (`generative-ai-go`)
- **PDF:** `github.com/go-pdf/fpdf`

See the [ADR log](docs/adr/README.md) for why.

---

## Run it locally

```bash
git clone https://github.com/mondial7/smart-360.git
cd smart-360

docker compose up -d postgres      # start Postgres
cp .env.example .env               # DEV_MODE=true is fine for local
go run ./cmd/server                # migrates + seeds + serves :8080
```

Open <http://localhost:8080>. With `DEV_MODE=true` you can sign in from the login
page's dev-login (or `GET /auth/dev-login?email=you@example.com`) without Google
credentials. **Set `ADMIN_EMAIL` to your email** — that account is the admin.

### Full stack in Docker

```bash
cp .env.example .env               # set SESSION_SECRET; DEV_MODE=false for a real run
docker compose up -d --build       # Postgres + the app on :8080
```

The app is a single binary (`go build ./cmd/server`) with all templates and
static assets embedded.

---

## Configuration

All configuration is read from the environment (`.env` in the repo root for local
runs). See [`.env.example`](.env.example) for the annotated list.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | Listen port. |
| `APP_URL` | No | `http://localhost:8080` | External base URL. |
| `DATABASE_URL` | Yes | local compose DSN | Postgres connection string. |
| `SESSION_SECRET` | **Yes** (prod) | dev fallback if `DEV_MODE` | Signs session cookies. Generate with `openssl rand -hex 32`. |
| `ADMIN_EMAIL` | Recommended | — | The user who signs in with this email is the global admin (deterministic bootstrap). |
| `LOG_FORMAT` | No | `text` | `text` or `json` (structured logs for a log shipper). |
| `GOOGLE_CLIENT_ID` | For Google login | — | OAuth client ID. |
| `GOOGLE_CLIENT_SECRET` | For Google login | — | OAuth client secret. |
| `GOOGLE_REDIRECT_URL` | No | `http://localhost:8080/auth/callback` | Must match the Google Cloud Console redirect URI. |
| `GEMINI_API_KEY` | No | — | Empty → moderation no-op + non-AI consolidation fallback. |
| `DEV_MODE` | No | `false` | Enables `/auth/dev-login`, relaxes the `Secure` cookie flag. **Never true in production.** |

**Security notes**

- Generate a real `SESSION_SECRET` for production (`openssl rand -hex 32`).
- Never commit `.env` (already gitignored).
- OAuth in production requires HTTPS — terminate TLS at a reverse proxy and set
  `GOOGLE_REDIRECT_URL` / `APP_URL` accordingly.

---

## Google OAuth setup

1. Open [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials).
2. **Create credentials → OAuth client ID → Web application**.
3. Add an **Authorized redirect URI**:
   - Local: `http://localhost:8080/auth/callback`
   - Production: `https://yourdomain.com/auth/callback`
4. Copy the **Client ID** and **Client Secret** into your `.env`.

## Gemini API setup

1. Open [Google AI Studio](https://aistudio.google.com/app/apikey).
2. **Create API Key** and paste it into `GEMINI_API_KEY`.

Consolidation works without a key — it just uses the non-AI fallback.

---

## Testing

```bash
go test ./...            # full pyramid (gateway tests need Docker)
go test -short ./...     # unit + in-memory handler tests only (no Docker)
```

Three layers: unit (`ai`, `view`, `auth`), in-memory handler integration against
repo fakes, and pgx gateway tests against a real Postgres via `testcontainers`.
See [`CLAUDE.md`](CLAUDE.md) for detail.

---

## Architecture

```
Browser ──HTML/htmx/SSE──▶  Go server (single binary)
                             ├─ chi router + html/template + embedded assets
                             ├─ auth: Google OAuth → session cookie + CSRF
                             ├─ ai:  Gemini moderation + synthesis (SSE progress)
                             ├─ pdf: fpdf consolidation renderer
                             └─ repo: pgx ──▶  PostgreSQL
```

Requests hit handlers (`internal/handlers`) → repositories (`internal/repo`) →
render via the view Renderer (`internal/view`). The `ai`, `pdf`, and chart
packages are pure (no HTTP/DB) and unit-tested in isolation. See
[`CLAUDE.md`](CLAUDE.md) for conventions and [`docs/adr/`](docs/adr/README.md) for
the decision log.

---

## Security & privacy

The product handles sensitive workplace feedback. How data is stored, who can
see it, and what is sent to Google Gemini are documented in
[`SECURITY.md`](SECURITY.md).

---

## Contributing & license

1. Fork and create a feature branch (`feat/your-thing`).
2. Run `go test ./...` and `go vet ./...`.
3. Use conventional commit prefixes (`feat:`, `fix:`, `docs:`, …).
4. Record significant architecture decisions as a new ADR in `docs/adr/`.

Licensed under the [MIT License](LICENSE).
