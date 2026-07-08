# Smart 360 Feedback — Project Guide

This guide is the entry point for anyone (humans or AI assistants) working on the codebase. It covers architecture, conventions, and common tasks.

> Product context: see [`README.md`](README.md) ("What it does", "How a round works", "Who can do what").
> Security & AI data flow: see [`SECURITY.md`](SECURITY.md).
> Roadmap: tracked in [GitHub Issues](https://github.com/mondial7/smart-360/issues).

---

## Stack

- **Backend**: Go 1.25 + Gin + MongoDB driver + Google Gemini (`generative-ai-go`)
- **Frontend**: Vue 3 (Composition API) + TypeScript + Vite + Pinia + SASS
- **Auth**: Google OAuth 2.0 → JWT
- **PDF**: `github.com/go-pdf/fpdf` (server-side rendering)
- **Database**: MongoDB 8.0
- **Deploy**: Docker + Docker Compose

---

## Repository Layout

```
backend/
├── cmd/seed/             # CLI seeding utility
├── database/             # DB init + indexes + seed_dev
├── handlers/             # HTTP handlers (auth, rounds, submissions,
│                         #   consolidation, pdf, analytics, dashboard, …)
├── middleware/           # AuthMiddleware, AdminOnly, TeamAdminOrGlobalAdmin
├── models/               # Mongo document structs
├── repositories/         # Repository interfaces + Mongo implementations + fakes
├── testutil/             # Test fixtures and helpers (real MongoDB via memongo)
├── scripts/mongo-init.js # Mongo init script
├── main.go               # Server entrypoint + routes
└── reseed-dev.sh         # Reset DB + reseed for local dev

frontend/
├── src/
│   ├── api/client.ts     # Axios with JWT interceptor
│   ├── components/       # NavBar, RadarChart, MyAnalyticsCard
│   ├── stores/auth.ts    # Pinia auth store
│   ├── types/            # Shared TS types (round, user, dashboard, analytics)
│   ├── views/            # Page components
│   └── router/           # Vue Router config

docs/archive/             # Stale docs kept for history
```

---

## Test Pyramid (backend)

The backend follows a three-layer test strategy:

1. **Unit tests** — pure business logic in handlers/helpers; no I/O.
2. **In-memory integration tests** — handler-level tests using fake repositories from `repositories/fake_*`.
3. **Gateway / port tests** — repository tests that hit a real MongoDB via `testutil/mongodb.go`.

Run all of them with:

```bash
cd backend && go test ./...
```

## Test Suite (frontend)

Component and store/unit tests run on [Vitest](https://vitest.dev) + `@vue/test-utils` (jsdom environment). Specs live next to the code as `*.spec.ts`; shared setup is in `src/test/setup.ts`.

```bash
cd frontend && npm test          # vitest run (CI uses this)
cd frontend && npm run test:watch # watch mode
```

There is intentionally **no browser/E2E layer** — the Vitest unit + component suite is the frontend test strategy for this repo. Add specs alongside the code you touch.

---

## Quick Start (development)

```bash
# 1. Start MongoDB
cd backend && docker-compose up -d mongodb

# 2. Configure env
cp .env.example .env  # then fill in GEMINI_API_KEY, JWT_SECRET, OAuth creds

# 3. Seed
./reseed-dev.sh

# 4. Run backend
go run main.go        # :8080

# 5. Run frontend
cd ../frontend && npm install && npm run dev   # :5173
```

Open <http://localhost:5173>. Use OAuth, or in dev mode hit
`http://localhost:8080/api/auth/dev-login?email=admin@example.com` (only enabled when `DEV_MODE=true`).

---

## Database Schema

| Collection | Key fields |
|------------|------------|
| `users` | `email`, `name`, `role` (admin/team_admin/member), `team_id`, `photo_url`, `created_at`, `last_login` |
| `teams` | `name`, `description`, `team_admin_id` |
| `team_members` | `team_id`, `user_id` |
| `templates` | `slug` (unique), `name`, `description`, `coaching_persona`, `questions[]` (key, peer_text, self_text, card_title), `competencies[]` (key, name, description). Seeded on every boot by `database.SeedDefaultTemplates` — see `templates_seed.go`. |
| `feedback_rounds` | `subject_id`, `created_by_id`, `template_id`, `status` (draft/active/closed/shared), `deadline`, `reviewers[]` |
| `submissions` | `round_id`, `reviewer_id`, `responses` (JSON string keyed by template question keys), `is_self`, `relationship` (manager/report/peer/cross_functional — required for peer), `interaction_frequency` (daily/weekly/monthly/rarely — required for peer), `ratings[]` (key, score 1–5, justification — required if template has competencies), `private_notes` (peer-only, manager-only audience) |
| `consolidations` | `round_id`, `executive_summary`, `strengths`, `areas_for_improvement`, `actionable_insights`, `question_summaries`, `question_labels` (snapshot from template), `self_vs_others_delta`, `voice_breakdown` (manager / peer / report streams), `competency_ratings[]` (per-axis aggregates with per-voice averages + spread), `manager_only_channel` (synthesis + raw notes; stripped from API/PDF whenever the caller isn't the round creator or a global admin), `admin_notes`, `shared_at` |
| `moderation_logs` | `round_id`, `submission_id`, `model`, `flagged`, `reasons[]` (sanitised — see `sanitiseErr`), `fields_scrubbed[]`, `moderated_at`. Best-effort audit trail of the per-submission Gemini scrub pass. |
| `audit_logs` | `actor_id`, `action`, `round_id`, `description`, `old_value`, `new_value` |

Note: array/object fields on `consolidations` (`strengths`, `areas_for_improvement`, `actionable_insights`, `question_summaries`) are stored as JSON strings in Mongo for legacy reasons. Parse before use. Newer fields (`question_labels`, `self_vs_others_delta`, `voice_breakdown`, `competency_ratings`, `manager_only_channel`) are typed BSON sub-docs.

Indexes are defined in `backend/database/indexes.go` (unique on `templates.slug`; round + timestamp on `moderation_logs`).

---

## Round Lifecycle

`draft` → `active` → `closed` → `shared`

- Admin / team admin creates a round (draft), picking a template (default falls back to the `default`-slug template).
- Starting it activates submission collection. Both peers AND the subject can submit — the subject's submission is flagged `is_self=true` and feeds the self-vs-peers delta.
- Closing freezes submissions and unlocks consolidation.
- Consolidation runs two Gemini passes: a per-submission **moderation scrub** (strips identity-targeted / personality-attack content; persists `moderation_logs`) followed by the **synthesis** (executive summary, strengths, growth areas, top focus areas, question summaries, self-vs-peers delta, voice breakdown, manager-only channel synthesis).
- Server-side computes deterministic aggregates: `competency_ratings` (per-voice averages + spread) and the voice reviewer counts.
- Admin may edit notes. Sharing flips status to `shared` and exposes the consolidation to the subject (in-app + PDF) — but the `manager_only_channel` is stripped from anything the subject sees.

Status transitions are recorded in `audit_logs`. Moderation passes are recorded in `moderation_logs`.

---

## Auth

- Public: `/api/auth/google`, `/api/auth/callback`
- Dev-only (gated by `DEV_MODE=true`): `/api/auth/dev-login?email=…`
- All other `/api/*` routes go through `middleware.AuthMiddleware()`. JWT is read from `Authorization: Bearer …`.
- Some routes additionally require `middleware.AdminOnly()` or `middleware.TeamAdminOrGlobalAdmin()`.

The first user to sign up automatically becomes admin (see `handlers/auth.go`).

---

## Frontend Conventions

- **Composition API** with `<script setup lang="ts">`. Type everything; avoid `any`.
- **API calls** through the `apiClient` from `src/api/client.ts` (it injects the JWT and redirects to `/login` on 401).
- **State** in Pinia stores (`src/stores/`).
- **Styles** scoped SASS (`<style scoped lang="scss">`) using CSS variables (`--color-primary`, `--bg-secondary`, etc.).
- **Charts** are pure SVG (see `RadarChart.vue`) — no chart libraries currently. Keep it that way unless requirements force otherwise.
- **Icons** are Phosphor (`@phosphor-icons/vue`). Import the components you need (`import { PhCheck, PhUsers } from '@phosphor-icons/vue'`) and pass `:size` (px) and `weight` (`regular` / `bold` / `duotone` / `fill`). Convention in this codebase: `regular` for nav/inline text, `duotone` for section titles and empty-state hero icons, `bold` for list bullets and small CTAs, `fill` for warning/sparkle accents. Don't add emoji as a UI element — use a Phosphor component.

---

## API Highlights

| Method & Path | Notes |
|---------------|-------|
| `GET /api/me` | Current user |
| `GET /api/templates` | List available round templates |
| `GET /api/templates/:idOrSlug` | Fetch a single template by ObjectID hex or slug |
| `GET /api/dashboard/stats` | Per-user dashboard counters |
| `GET /api/analytics/me` | Personal analytics (radar + per-round trend) |
| `GET /api/analytics/admin` | Admin analytics (counters, status breakdown, completion trend, team activity, top themes) |
| `GET /api/rounds-for-me` | Rounds the user must review (includes the user's own active rounds for self-assessment) |
| `POST /api/rounds` | Create round (accepts `templateId`; defaults to default slug if absent) |
| `POST /api/submissions` | Submit feedback (peer: requires `relationship`, `interactionFrequency`, and `ratings` if template has competencies; supports `privateNotes`. Self: skips relationship/frequency/private notes) |
| `POST /api/rounds/:id/consolidate` | Generate AI consolidation (admin) |
| `GET /api/consolidations/:roundId` | Fetch consolidation JSON (`manager_only_channel` stripped for the subject) |
| `GET /api/consolidations/:roundId/pdf` | Download PDF (admin / subject after share / round creator — manager-only section omitted for the subject) |
| `POST /api/consolidations/:id/share` | Share with subject (admin) |
| `GET /api/rounds/:id/moderation-logs` | Audit trail of the per-submission scrub pass — admin / round creator only |

See `backend/main.go` for the full route map.

---

## Common Tasks

### Add a new API endpoint

1. Implement the handler in `backend/handlers/<domain>.go`.
2. Register it in `backend/main.go`, with the appropriate middleware.
3. Add tests:
   - Unit tests in the same `handlers/` package
   - Integration tests using fakes if the handler depends on repositories
   - Gateway tests in `repositories/` if it touches new Mongo queries
4. Wire up the frontend call (Axios + Pinia store / view).

### Add a new view

1. Create `frontend/src/views/<Name>View.vue` using `<script setup lang="ts">`.
2. Register the route in `src/router/index.ts` with `meta: { requiresAuth: true }` if needed.
3. Update `NavBar.vue` if the view is reachable from the nav.

### Modify the database schema

1. Update the model in `backend/models/`.
2. Update `mongo-init.js` (validation), `indexes.go`, and `seed_dev.go`.
3. Drop and reseed locally: `docker-compose down -v && docker-compose up -d && ./reseed-dev.sh`.

### Run backend tests

```bash
cd backend
go test ./...                    # full pyramid (unit + integration + gateway)
go test -short ./...             # skip the real-Mongo gateway tests
```

---

## Coding Style

### Go

- Follow `gofmt` / `go vet`. CI should fail on either.
- Keep handlers thin; push business logic into helpers or repositories.
- Always check errors and return early (`http.StatusInternalServerError` with a generic message; log details).
- Convert hex strings to `primitive.ObjectID` at the boundary; don't pass strings deep into the stack.
- Test fakes live next to the interface in `repositories/`.

### TypeScript / Vue

- PascalCase for components, camelCase for functions and variables.
- Prefer named imports; keep imports sorted by source (Vue first, then libraries, then `@/…`).
- Don't introduce `any` to silence the compiler — fix the type instead.
- Keep components small; extract shared visualization primitives (e.g., `RadarChart.vue`) when you find duplication.

---

## Commit Guidelines

- Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`.
- Keep commits focused and atomic.
- Do **not** include Claude as a commit co-author.

---

## Notes for AI Assistants

- The project values simplicity over premature optimization. Prefer the smallest change that solves the problem.
- When in doubt about scope, browse [open issues](https://github.com/mondial7/smart-360/issues) — they track what's *intentionally* deferred.
- Don't fix unrelated TS / lint errors as part of an unrelated change unless explicitly asked.
- The dev-login and `/api/debug/*` endpoints exist for local development only and must remain gated.
- When changing the `consolidations` schema, remember its array/object fields are stored as JSON strings — parse on read, stringify on write.
- The PDF renderer (`backend/handlers/pdf.go`) and the personal analytics endpoint (`backend/handlers/analytics.go`) are intentionally dependency-light. Keep them that way.
