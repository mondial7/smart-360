# Security & Privacy

Smart 360 collects sensitive workplace feedback, so the way we handle
that data — and how we use AI on top of it — is a first-class concern.
This document is the single source of truth for:

- [Reporting a vulnerability](#reporting-a-vulnerability)
- [Anonymity & privacy by design](#anonymity--privacy-by-design)
- [How Google Gemini is used](#how-google-gemini-is-used)
- [Operational hardening checklist](#operational-hardening-checklist)
- [Known security work in progress](#known-security-work-in-progress)

---

## Reporting a vulnerability

If you've found something exploitable, **please do not open a public
GitHub issue**. Use the GitHub
[private vulnerability reporting](https://github.com/mondial7/smart-360/security/advisories/new)
form for this repo — it routes the report straight to the maintainers
and keeps the details private until a fix ships.

What helps a triage move quickly:

- Affected commit / version (e.g. `v1.0.0`, or `main` at SHA `…`).
- Reproduction steps — a minimal `curl` or screen recording is gold.
- Impact you observed and what you think the worst case is.

We aim to acknowledge within **3 business days** and to publish a fix
(or a documented workaround) within **30 days** for high-severity
findings. Lower-severity issues are scheduled like any other backlog
item and tracked in [open issues](https://github.com/mondial7/smart-360/issues)
labelled `security`.

---

## Anonymity & privacy by design

The product promise is "reviewers can be honest because their identity
never reaches the subject." That promise is enforced in code, not only
by policy.

### What gets stored

| Collection | Contains | Identifies a reviewer? |
|------------|----------|------------------------|
| `users` | Google profile (name, email, photo), role, timestamps | Yes — but only for the logged-in identity |
| `feedback_rounds` | Subject ID, creator ID, deadline, status | No reviewer identity |
| `submissions` | `round_id`, `reviewer_id`, free-text `responses` JSON | **Yes — see below** |
| `consolidations` | AI-generated summary + admin notes | No |
| `audit_logs` | Actor ID, action, round ID, before/after values | Actor only (status transitions, not reviewers) |

`submissions.reviewer_id` exists so the app can enforce "one submission
per reviewer per round" and surface the reviewer's own draft back to
them. It is **never** returned to the round subject and is gated even
from admins behind explicit handler-level checks — eight separate
leak paths were closed pre-1.0 (see commit
[`c9097b9`](https://github.com/mondial7/smart-360/commit/c9097b9)).

### What each role can see

| Role | Their own submissions | Others' submissions | Consolidated feedback | Reviewer identities |
|------|----------------------|---------------------|-----------------------|---------------------|
| **Admin** | Yes | Yes (raw, for audit) — gated by handler check | Yes for all rounds | Yes (a deliberate trade-off for moderation) |
| **Team Admin** | Yes | Within own team only | Within own team only | Yes within own team |
| **Member (reviewer)** | Yes (own only) | No | Only consolidations shared with them as a subject | No, ever |
| **Member (subject)** | n/a | No | Yes, after admin shares | No, ever |

The "subject never sees reviewer identities" guarantee is what keeps the
360° process honest. Every code path that could leak it is covered by a
test in `backend/handlers/*_test.go`.

### Data minimisation

- Google OAuth scopes are limited to `profile` + `email`. No drive, no
  calendar, no contacts.
- The PDF export omits raw submissions and any reviewer identifiers — it
  only contains the consolidated summary.
- Audit logs intentionally do **not** record submission contents; only
  status transitions.

---

## How Google Gemini is used

AI is **opt-in per round** and triggered by an admin after the round
closes. There is no automatic / background processing.

### What is sent to Gemini

The Gemini prompt is built in `backend/handlers/consolidation.go`
(`generateGeminiConsolidation`). For each submission in the round, we
send only the four free-text answers, labelled by question letter:

```
Feedback from reviewer:
Strengths: <text>
Areas for improvement: <text>
Observed behaviors: <text>
Growth advice: <text>
```

What is **not** sent:

- ❌ The subject's name, email, photo, or any user ID.
- ❌ The reviewer's name, email, photo, or any user ID.
- ❌ The round ID, the team name, or the organization name.
- ❌ Any reviewer-discriminating ordering ("Reviewer 1", "Reviewer 2", …).
- ❌ Any audit-log or submission-timestamp metadata.

Each submission is concatenated with a blank line between them — Gemini
sees N unordered chunks of free text and a generic prompt asking for a
JSON-structured consolidation. The response is constrained to a typed
JSON schema (`model.ResponseSchema`) so the model cannot reflect
metadata back into a structured field.

### What is stored from Gemini's response

Only the five JSON fields documented in the prompt — executive summary,
strengths, areas for improvement, actionable insights, and the four
per-question summaries. The raw HTTP response is not persisted.

### When Gemini is bypassed

If `GEMINI_API_KEY` is unset, the same endpoint returns a non-AI
fallback consolidation (a concatenation of the raw answers). This is
useful for development / offline use but also means the only third
party that ever sees feedback content is the Gemini API when the
operator chooses to enable it.

### Third-party terms

Once data is sent to Gemini, Google's API terms and privacy policy
apply. Operators who self-host should review them and decide whether
they want to enable Gemini at all:

- [Google AI Studio terms](https://ai.google.dev/gemini-api/terms)
- [Google API Services User Data Policy](https://developers.google.com/terms/api-services-user-data-policy)
- [Generative AI prohibited use policy](https://policies.google.com/terms/generative-ai/use-policy)

As of the [`gemini-flash-latest`](https://ai.google.dev/gemini-api/docs/models)
model used here, the free tier explicitly excludes API content from
training; paid tiers default to the same exclusion. **Verify this is
still true at the time you deploy.**

---

## Operational hardening checklist

Things the application can't enforce on your behalf — the
[Production deployment guide](docs/deployment-production.md) walks
through each in detail.

- [ ] Generate a strong `JWT_SECRET` (`openssl rand -base64 32`) and
      rotate it on any suspected compromise.
- [ ] Change the default `MONGO_ROOT_PASSWORD` before exposing the stack.
- [ ] Terminate TLS at a reverse proxy (Caddy / nginx + Let's Encrypt).
      OAuth requires HTTPS in production.
- [ ] Throttle abusive traffic at the reverse proxy (`rate_limit`
      directive in Caddy, `limit_req` in nginx) until the built-in
      rate limiter ships ([#26](https://github.com/mondial7/smart-360/issues/26)).
- [ ] Schedule a nightly `mongodump` to off-host storage
      ([#35](https://github.com/mondial7/smart-360/issues/35) will
      automate this).
- [ ] Never enable `DEV_MODE=true` in production — it unlocks the
      `dev-login` endpoint and seed data.

---

## Known security work in progress

A pre-1.0 audit produced eight access-control / auth findings — all
closed before release in commit
[`c9097b9`](https://github.com/mondial7/smart-360/commit/c9097b9).
New work is tracked as GitHub issues. Currently open:

- [#22](https://github.com/mondial7/smart-360/issues/22) — Stop delivering JWT via URL query parameter on OAuth callback.
- [#23](https://github.com/mondial7/smart-360/issues/23) — Audit-log every AI consolidation edit.
- [#24](https://github.com/mondial7/smart-360/issues/24) — Reduce PII exposure in `/api/users` for non-admin callers.
- [#25](https://github.com/mondial7/smart-360/issues/25) — Guardrail admin role escalation/demotion (avoid bricked install).
- [#26](https://github.com/mondial7/smart-360/issues/26) — Rate-limit auth + submission endpoints.
- [#27](https://github.com/mondial7/smart-360/issues/27) — First-user-becomes-admin race condition on empty DB.
- [#32](https://github.com/mondial7/smart-360/issues/32) — Broader hardening: CSRF, CSP, input validation, secrets management.
- [#20](https://github.com/mondial7/smart-360/issues/20) — Vite ecosystem upgrade + transitive dependabot alerts.

If something belongs on this list but isn't here yet, the private
reporting form linked above is the right path.
