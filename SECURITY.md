# Security & Privacy

Smart 360 collects sensitive workplace feedback, so the way we handle that data —
and how we use AI on top of it — is a first-class concern. This document is the
single source of truth for:

- [Reporting a vulnerability](#reporting-a-vulnerability)
- [Authentication & sessions](#authentication--sessions)
- [Anonymity & privacy by design](#anonymity--privacy-by-design)
- [How Google Gemini is used](#how-google-gemini-is-used)
- [Operational hardening checklist](#operational-hardening-checklist)

---

## Reporting a vulnerability

If you've found something exploitable, **please do not open a public GitHub
issue**. Use the GitHub
[private vulnerability reporting](https://github.com/mondial7/smart-360/security/advisories/new)
form for this repo — it routes the report straight to the maintainers and keeps
the details private until a fix ships.

What helps triage move quickly:

- Affected commit / branch (e.g. `main` at SHA `…`).
- Reproduction steps — a minimal `curl` or screen recording is gold.
- Impact you observed and what you think the worst case is.

---

## Authentication & sessions

- Login is **Google OAuth 2.0**, scoped to `profile` + `email` only (no drive,
  calendar, or contacts). A login-CSRF state cookie protects the OAuth redirect.
- On success the server creates a **server-side session** (a row in the
  `sessions` table) and sets an **HttpOnly, Secure, SameSite=Lax** cookie holding
  an opaque, HMAC-signed session ID. Tokens are never exposed to JavaScript and
  never travel in a URL. Sessions are revocable and expire server-side.
- Every state-changing request (POST/PUT/DELETE) is **CSRF-protected** with a
  per-session token (`X-CSRF-Token` header for htmx, hidden `csrf_token` field
  for forms). The one state-changing GET (the consolidation/log SSE streams) is
  guarded by a separate session-derived stream token.
- **Admin bootstrap** is deterministic: the user whose email matches
  `ADMIN_EMAIL` is (and stays) the global admin. There is no "first user wins"
  race. All other role changes go through the admin Users page, which refuses to
  demote the last admin.
- `dev-login` (which bypasses OAuth) is only mounted when `DEV_MODE=true` and is
  the single most important thing to keep disabled in production.

See [ADR-0004](docs/adr/0004-session-cookie-auth.md) for the rationale.

## Transport & browser hardening

- **Security headers** on every response: a strict `Content-Security-Policy`
  (`script-src 'self'` — all JS is self-hosted, no inline scripts),
  `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
  `Referrer-Policy: strict-origin-when-cross-origin`, and a restrictive
  `Permissions-Policy`.
- **Rate limiting** (per-IP, in-memory): tight on auth endpoints, a cap on
  feedback submissions, and a backstop on all authenticated routes. A reverse
  proxy should still throttle in front for production.
- Static analysis: `govulncheck` and `gosec` run clean (reviewed false positives
  are annotated with `#nosec` + justification).

---

## Anonymity & privacy by design

The product promise is "reviewers can be honest because their identity never
reaches the subject." That promise is enforced in code, not only by policy.

### What gets stored

| Table | Contains | Identifies a reviewer? |
|-------|----------|------------------------|
| `users` | Google profile (name, email, photo), role, timestamps | Only the logged-in identity |
| `sessions` | Session ID, user ID, expiry | No feedback content |
| `feedback_rounds` / `round_reviewers` | Subject, creator, template, status; reviewer assignments | Assignment only, no content |
| `submissions` | `round_id`, `reviewer_id`, `responses`/`ratings` (jsonb), `private_notes` | **Yes — see below** |
| `consolidations` | AI/aggregate summary, admin notes, manager-only channel | No reviewer identity |
| `moderation_logs` | Per-submission scrub audit (reasons, fields scrubbed) | No content, no reviewer name |
| `audit_logs` | Actor, action, round, before/after values | Actor only (status transitions, not reviewers) |

`submissions.reviewer_id` exists to enforce "one submission per reviewer per
round" and to show reviewers their own draft. It is **never** surfaced to the
round subject; handler-level checks gate every read path.

### What each role can see

| Role | Own submissions | Others' submissions | Consolidated feedback | Manager-only channel |
|------|-----------------|---------------------|-----------------------|----------------------|
| **Admin** | Yes | Yes (for audit) | All rounds | Yes |
| **Team Admin** | Yes | Rounds they created | Rounds they created | Rounds they created |
| **Member (reviewer)** | Yes | No | Only shared with them as subject | No |
| **Member (subject)** | n/a | No | Yes, after sharing | **Never** |

Two invariants are enforced in `internal/handlers` and covered by tests: the
**manager-only channel is stripped from anything the subject sees** (in-app and
PDF), and reviewer identities are never returned to the subject.

### Data minimisation

- The PDF export contains only the consolidated summary — no raw submissions, no
  reviewer identifiers (and no manager-only section in the subject's copy).
- Audit logs record status transitions, not submission contents.

---

## How Google Gemini is used

AI is **triggered by an admin** after a round closes — there is no automatic /
background processing. Two passes run (see [ADR-0006](docs/adr/0006-sse-for-consolidation.md)),
both built in `internal/ai`:

1. **Moderation scrub** (`moderation.go`) — each submission is sent individually
   and Gemini returns a cleaned version with identity-targeted / personality-
   attack / off-topic content removed. Recorded in `moderation_logs`.
2. **Synthesis** (`synthesis.go`) — the scrubbed submissions are combined into
   the consolidation.

### What is sent to Gemini

Per submission: the free-text question answers, competency rating justifications,
the reviewer's **relationship and interaction-frequency labels** (needed so the
model can weight signal), and any private manager-only note. What is **not**
sent:

- The subject's or reviewer's name, email, photo, or any user ID.
- The round ID, team name, or organization name.
- Any audit-log or timestamp metadata.

Submissions are presented as unordered blocks; the response is constrained to a
typed JSON schema so the model cannot reflect metadata into a structured field.

### Key protection

The Gemini SDK embeds the API key in request URLs, which appear in its error
strings. `ai.sanitiseErr` scrubs `key=…` to `key=REDACTED` before any error is
logged or persisted to `moderation_logs`, so the live key can't leak through the
audit trail.

### When Gemini is bypassed

If `GEMINI_API_KEY` is unset, moderation is a no-op and synthesis falls back to a
non-AI combine of the raw answers. No feedback content leaves the process unless
the operator enables Gemini.

### Third-party terms

Once data is sent to Gemini, Google's API terms and privacy policy apply:

- [Google AI Studio terms](https://ai.google.dev/gemini-api/terms)
- [Google API Services User Data Policy](https://developers.google.com/terms/api-services-user-data-policy)
- [Generative AI prohibited use policy](https://policies.google.com/terms/generative-ai/use-policy)

As of the `gemini-flash-latest` model used here, the free tier excludes API
content from training; **verify this is still true at the time you deploy.**

---

## Operational hardening checklist

Things the application can't enforce on your behalf:

- [ ] Generate a strong `SESSION_SECRET` (`openssl rand -hex 32`) and rotate it on
      any suspected compromise.
- [ ] Use a strong Postgres password and restrict network access to the database.
- [ ] Terminate TLS at a reverse proxy (Caddy / nginx + Let's Encrypt). OAuth
      requires HTTPS in production; set `APP_URL` / `GOOGLE_REDIRECT_URL` to the
      public HTTPS URLs.
- [ ] Set `ADMIN_EMAIL` to the owner's address before first sign-in.
- [ ] The app rate-limits per IP already; for extra depth, also throttle at the
      reverse proxy (`rate_limit` in Caddy, `limit_req` in nginx).
- [ ] Schedule an off-host backup with `scripts/backup.sh` (see the deployment
      guide's "Backups & disaster recovery") and test a restore periodically.
- [ ] **Never** set `DEV_MODE=true` in production — it unlocks `dev-login` and
      relaxes the Secure cookie flag.
