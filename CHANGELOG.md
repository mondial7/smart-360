# Changelog

All notable changes to Smart 360 Feedback are documented here. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2026-07-11

### Added

- **Self-nomination** (#41) — a member can request a 360 round on themselves
  from **My Feedback → Request feedback**: they pick their reviewers, and the
  round is created as a draft **owned by a manager (their team admin, else a
  global admin) — never the subject**, so the requester can never de-anonymize
  their reviewers. One open request at a time. The owner reviews and starts it.

### Changed

- Non-admin **Rounds** now also lists the rounds you own, so a team admin sees
  and can manage the rounds they created (including self-nominated ones).
- Performance (#33 follow-up): paginated round lists resolve only the users on
  the current page via a batch lookup, instead of loading the whole user table.

## [1.1.0] - 2026-07-11

### Added

- **Cross-round comparison** (#29) — a subject can now see their competency
  scores from peers trending across their shared rounds: a multi-series line
  chart (server-rendered SVG), a round-by-round table, and a summary timeline.
  Linked from **My Feedback** once two or more rounds have been shared.

### Changed

- **Pagination** (#33) — the Rounds (admin), Audit log, and Users lists are now
  paginated (25/page) with Prev/Next controls. Paged queries carry a unique
  `id` tiebreaker so page boundaries are deterministic even when timestamps
  collide.

## [1.0.0] - 2026-07-11

First tagged release. Delivers the app as a single server-rendered Go binary
backed by PostgreSQL. (The project was previously used only locally on a
Go/Gin + Vue + MongoDB stack, which was never published; this is the first
public release.)

### Changed — full rewrite to a server-rendered Go stack

Smart 360 is now a **single server-rendered Go application** (`html/template` +
htmx + SSE) backed by **PostgreSQL**, replacing the Go/Gin JSON API + Vue 3 SPA +
MongoDB split. Product behaviour and feature set are preserved.

- **UI**: Vue/Vite/Pinia/TypeScript SPA → Go `html/template` + htmx, with charts
  drawn as server-rendered SVG (no JS chart library) and consolidation progress
  streamed over Server-Sent Events.
- **Database**: MongoDB → PostgreSQL via pgx with hand-written SQL. Embedded
  arrays are normalized into `team_members` / `round_reviewers` join tables;
  legacy JSON-string consolidation fields become typed `jsonb`.
- **Auth**: JWT-in-`localStorage` → server-side sessions in an HttpOnly, signed
  cookie, with CSRF protection on all unsafe requests. The OAuth token no longer
  travels in a URL.
- **Deploy**: one embedded binary (templates + assets via `//go:embed`);
  `docker-compose.yml` now runs Postgres + the app.
- **Docs**: added an [Architecture Decision Record log](docs/adr/README.md)
  (ADR-0001 through ADR-0007) capturing the rewrite decisions.

The reusable Go layers — Gemini moderation + synthesis, the fpdf renderer, and
the competency aggregation math — were carried over intact.

### Prior unreleased work (pre-rewrite)

Major reshape of the 360 round itself — questions, persona, and
consolidation output are now growth-framed and substantially richer.
Everything below ships behind the existing `GEMINI_API_KEY` requirement
for the AI features; the non-AI fallback still works without a key.

### Added

- **Growth-framed peer questions** — replaced the four generic prompts
  (key strengths / improvements / observed behaviours / advice) with
  Continue / Block / Amplify / Experiment framing, anchored on the last
  3–6 months with a Situation → Behaviour → Impact scaffold on the first
  question.
- **Self-assessment + delta** — subjects now answer the same questions
  in first-person; the consolidation surfaces `self_vs_others_delta`
  with blind spots, hidden strengths, aligned themes, and a coaching
  summary.
- **Reviewer-relationship metadata** — every peer submission declares
  `relationship` (manager / report / peer / cross-functional) and
  `interactionFrequency` (daily / weekly / monthly / rarely). Required
  for peer submissions; ignored for self.
- **Manager / peer / report voice breakdown** — the consolidation emits
  three distinct synthesis streams gated by server-side reviewer counts,
  so a manager's signal isn't averaged with peer signal.
- **Configurable round templates** — new `templates` collection with
  per-template question wording, coaching persona, card titles, and
  competencies. Two templates seeded on every boot: "Growth-Framed 360"
  (default) and "Engineering Leadership 360" (tech leads, EMs, staff+).
  Both create-round wizards now include a template picker.
- **Likert competency rubric** — templates can declare competencies;
  reviewers rate 1–5 with a one-line justification. The consolidation
  shows per-voice averages, the spread, and a wide-spread warning to
  surface calibration gaps for the 1:1.
- **Private-to-manager channel** — optional textarea on peer submissions
  feeds into a `manager_only_channel` block (synthesis + themes + raw
  relationship-tagged notes). Stripped from the subject's API response,
  in-app view, and PDF; visible only to round creator / global admin.
  Audit endpoint: `GET /api/rounds/:id/moderation-logs`.
- **Dedicated moderation pass** — a separate, audited Gemini call scrubs
  identity-targeted / personality-attack / off-topic content per
  submission before the synthesis prompt sees it. Persists a
  `moderation_logs` row per submission with `fields_scrubbed`, `reasons`,
  and the model used. Parallelised across submissions with a 25s
  per-call timeout; synthesis call has a 60s timeout. Error strings are
  scrubbed of any embedded API keys before being persisted.
- **Recency anchor surfaced in the UI** — submit form has a dashed-
  border aside explaining the 3–6 month window; the create-round review
  step warns admins that reviewers will see the prompt.
- **New endpoints** — `GET /api/templates`, `GET /api/templates/:idOrSlug`,
  `GET /api/rounds/:id/moderation-logs`.

### Changed

- AI prompt persona is now driven by the template's `coaching_persona`
  field (was hard-coded as "Engineering Team Lead"). The default
  template's persona is "a thoughtful coach helping someone grow over
  the next 6 months". `actionable_insights` is now explicitly capped at
  3 in the prompt — these are the manager's "top focus areas".
- Submit-form question text is fetched from the round's template
  (was hard-coded in the Vue component). Same for the per-question
  display labels in `ConsolidationView` / `MyFeedbackView` / PDF —
  they're snapshotted onto the consolidation as `question_labels` at
  generation time so the UI doesn't need an extra fetch.
- Subjects can now submit their own self-assessment via the standard
  `/api/submissions` endpoint (previously this was forbidden).
- "Rounds for me" now includes the user's own active rounds so the
  self-assessment surfaces alongside peer reviews they owe.

### Security

- `moderation_logs.reasons` is sanitised before persistence — the genai
  SDK embeds the live Gemini API key in URLs inside its error strings,
  and the audit-log surface would otherwise leak the key to anyone with
  admin / round-creator read access. Regex-redacts `?key=` / `&key=`
  query parameters before any error reaches the log or stdout.
- Manager-only synthesis is stripped on three independent surfaces (API,
  dashboard listing, PDF) when the caller is not the round creator or a
  global admin. Subjects never see it.

### Internal

- Submission model gained `is_self`, `relationship`, `interaction_frequency`,
  `ratings[]`, and `private_notes` fields.
- Consolidation model gained `question_labels`, `self_vs_others_delta`,
  `voice_breakdown`, `competency_ratings[]`, and `manager_only_channel`
  fields. Newer fields are typed BSON sub-docs; legacy string-encoded
  fields (`strengths`, `areas_for_improvement`, etc.) are left as-is.
- Unique index on `templates.slug`; round + timestamp index on
  `moderation_logs` for the audit-trail query path.
- Backend unit/integration tests extended to cover every new code path:
  `validateRatings`, `aggregateCompetencyRatings`, `countByVoice`,
  `buildVoiceBreakdown`, `collectPrivateNotes`, `buildManagerOnlyChannel`,
  `canSeeManagerOnlyChannel`, `applyModerationCleaned`, `sanitiseErr`,
  plus the self-assessment integration tests against the parallel
  `testSubmissionHandler` used by the integration suite.

### Known gaps post-Phase-2

- Browser UI smoke for the new sections (rubric, voices, manager-only,
  template picker) has not been done — tracked in [#50](https://github.com/mondial7/smart-360/issues/50).
- The moderation response schema declares `cleaned` as a free-form
  `TypeObject`, which Gemini occasionally rejects, causing all
  per-submission calls to hit the 25s timeout. When this happens the
  synthesis runs on un-moderated content and the audit log records the
  timeout. Tracked in [#48](https://github.com/mondial7/smart-360/issues/48).
- `POST /rounds/:id/consolidate` returns the consolidation with a zero
  `ObjectID`; the frontend masks this by re-fetching. Tracked in
  [#47](https://github.com/mondial7/smart-360/issues/47).
- `dashboard` / `admin_analytics` handlers discard Mongo errors silently
  on counter queries. Tracked in [#49](https://github.com/mondial7/smart-360/issues/49).

## [1.0.0] — TBD

First publicly distributed release. The Go rewrite is feature-complete,
distributable as a single binary via Homebrew, as a Docker Compose
stack, or from source.

### Distribution

- **Single-binary install via Homebrew** — the Vue SPA is now embedded
  into the Go server with `//go:embed`. `brew tap mondial7/tap && brew
  install smart360` produces one ~30 MB binary that serves both the API
  and the UI on a single port.
- **Pre-built Docker images on GHCR** — `ghcr.io/mondial7/smart-360-backend`
  and `…-frontend`, multi-arch (`linux/amd64`, `linux/arm64`).
- **Pull-and-run Docker Compose file** (`docker-compose.prod.yml`) that
  references the GHCR images, so production users don't need a source
  checkout.
- **Cross-platform release binaries** for macOS and Linux × amd64 / arm64,
  attached to every GitHub Release.
- **Automated release pipeline** (GitHub Actions) that cuts tarballs,
  publishes images, creates the Release, and updates the Homebrew tap
  from a single `v*.*.*` tag push.

### Features

- AI consolidation of anonymous feedback via Google Gemini, with
  executive summary, strengths, growth areas, and concrete next steps.
- Branded PDF export of any shared consolidation.
- Personal analytics dashboard (radar chart per round + trend) and admin
  analytics (counters, completion trends, theme extraction, team activity).
- Role-based access control (Admin, Team Admin, Member) with audit logs
  on every status transition.
- Google OAuth 2.0 authentication with JWT-backed sessions.
- Phosphor-based icon system across the SPA.

### Documentation

- README rewritten around three install paths (Homebrew, Docker, source).
- New [production deployment guide](docs/deployment-production.md)
  covering DNS, systemd, Caddy + Let's Encrypt, OAuth configuration, and
  an nginx + Certbot alternative.
- Static showcase site at `docs/` (GitHub Pages, no build step).

### Quality

- Backend test pyramid (unit + in-memory integration + real-MongoDB
  gateway), ~80 packages tests in CI on every PR.
- Frontend type-check (`vue-tsc`) now passes cleanly and runs in CI.
- Security audit closed 8 access-control and auth findings prior to release.

### Known limitations (deferred to post-1.0)

See [open issues](https://github.com/mondial7/smart-360/issues) for the
full list. The most notable gaps for self-hosters:

- **No built-in rate limiting or CSRF protection** — mitigate at the
  reverse proxy (Caddy `rate_limit`, nginx `limit_req`) until [#26](https://github.com/mondial7/smart-360/issues/26)
  (rate-limiting) and [#32](https://github.com/mondial7/smart-360/issues/32)
  (broader security hardening) land.
- **No automated MongoDB backups** — schedule a nightly `mongodump`
  yourself. Tracked in [#35](https://github.com/mondial7/smart-360/issues/35).
- **No structured logging / metrics / tracing** — the process logs to
  stdout only. Tracked in [#34](https://github.com/mondial7/smart-360/issues/34).

[Unreleased]: https://github.com/mondial7/smart-360/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/mondial7/smart-360/releases/tag/v1.0.0
