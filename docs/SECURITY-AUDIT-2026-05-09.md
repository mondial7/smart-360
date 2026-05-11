# Security Audit — 2026-05-09

Branch: `smart-360-in-go`. Eight confirmed findings after false-positive
filtering. Listed in priority order; status starts as `open` and moves to
`fixed` as we work through them.

---

## 1. `GetRoundSubmissions` exposes all reviewers + responses

- **Severity:** HIGH
- **Status:** fixed
- **File:** `backend/handlers/submissions.go` (`GetRoundSubmissions`, ~L17)
- **Route:** `GET /api/submissions/round/:roundId` (`backend/main.go:103`,
  only `AuthMiddleware()`)

Handler returns the raw `[]models.Submission` — including `reviewerId` and
free-text `responses` — for any round, with no authorization check beyond
JWT.

**Exploit:** Any authenticated user iterates round IDs (easily harvested
from `/api/rounds-for-me`, `/api/rounds/:id`, etc.) and obtains each
reviewer's identity paired with their feedback content. Defeats the 360
anonymity guarantee.

**Fix:** Restrict to admin / round creator. If returned to non-admins,
strip `reviewer_id` from the response.

---

## 2. `GetSubmissionDetails` returns any submission to any user

- **Severity:** HIGH
- **Status:** fixed
- **File:** `backend/handlers/submissions.go` (`GetSubmissionDetails`, ~L52)
- **Route:** `GET /api/submissions/:submissionId` (`backend/main.go:105`,
  only `AuthMiddleware()`)

Looks up a submission by ObjectID and returns the full document
(`reviewerId`, `roundId`, `responses`) with no ownership check. By
contrast, `UpdateSubmission` in the same file enforces
`existingSubmission.ReviewerID == currentUser.ID`.

**Exploit:** Submission IDs leak from #1 and other list endpoints.
Attacker downloads each anonymous review tied to its real author.

**Fix:** Allow only the submission's reviewer (read-back), the round
creator, or admin. Strip `reviewer_id` for everyone except author and
admin.

---

## 3. `GetConsolidation` exposes AI consolidation to any user

- **Severity:** HIGH
- **Status:** fixed
- **File:** `backend/handlers/consolidation.go` (`GetConsolidation`, ~L98)
- **Route:** `GET /api/consolidations/:roundId` (`backend/main.go:109`,
  only `AuthMiddleware()`)

Returns the full consolidation (`executive_summary`, `strengths`,
`areas_for_improvement`, `actionable_insights`, `question_summaries`,
`admin_notes`, `shared_at`) with no caller/relationship/status check. The
sibling `DownloadConsolidationPDF` (`backend/handlers/pdf.go:18`) properly
enforces "admin OR creator OR (subject AND `SharedAt != nil`)".

**Exploit:** Any authenticated user reads AI-consolidated feedback for
any subject, including drafts and internal `admin_notes`.

**Fix:** Mirror the PDF handler's check. Return 403 otherwise.

---

## 4. `GetRoundDetails` leaks reviewer roster + PII

- **Severity:** HIGH
- **Status:** fixed
- **File:** `backend/handlers/rounds.go` (`GetRoundDetails`, ~L348) and
  `backend/handlers/round_helpers.go` (`getPopulatedRound`)
- **Route:** `GET /api/rounds/:id` (`backend/main.go:91`, only
  `AuthMiddleware()`)

Returns the populated round with embedded `*models.User` for subject,
creator, and every reviewer. `models.User` has no `json:"-"` redaction, so
email, photo URL, role, team_id, last_login leak for everyone attached.

**Exploit:** Any authenticated user with any round ID dumps the reviewer
roster + PII. Combined with #1, enables systematic deanonymization.

**Fix:** Restrict to admin / round creator / subject (after share) /
listed reviewers. Or strip user PII for callers outside that set.

---

## 5. `SubmitFeedback` accepts phantom submissions

- **Severity:** HIGH
- **Status:** fixed
- **File:** `backend/handlers/submissions.go` (`SubmitFeedback`, ~L102)
- **Route:** `POST /api/submissions` (`backend/main.go:101`, only
  `AuthMiddleware()`)

Inserts a submission for `(round_id, current user)` without checking the
round exists, that its status is `active`, that the caller is in
`round.Reviewers`, or that the caller is not the subject. The only gate
is duplicate-prevention on `(round_id, reviewer_id)`. `UpdateSubmission`
in the same file does enforce `RoundActive` + ownership.

**Exploit:** Any authenticated user POSTs an arbitrary `roundId` with
free-text `responses`, becoming a phantom reviewer on any round in any
status. Pollutes AI consolidation, analytics, and PDFs. Self-injection
lets the subject game their own consolidation. Phantom-submitting first
also blocks legitimate reviewers via the duplicate check.

**Fix:** Load the round, ensure `status == active`, verify caller is in
`round.Reviewers[].reviewer_id` and not equal to `subject_id`. 403
otherwise.

---

## 6. OAuth login CSRF — hardcoded `state`

- **Severity:** HIGH
- **Status:** fixed
- **File:** `backend/handlers/auth.go:46` (initiate),
  `backend/handlers/auth.go:50-157` (callback)

`AuthCodeURL("state")` uses the literal string `"state"` as the OAuth
state for everyone. `GoogleCallback` reads only `code` and never
validates `state`. No PKCE, no cookie binding.

**Exploit:** Attacker initiates OAuth, intercepts their own `code`, and
tricks a victim's browser into hitting `/api/auth/callback?code=…`. The
backend mints a JWT for the attacker's account and the victim is silently
logged into it. Victim then submits feedback into attacker-owned account
which the attacker reads (compounded by #1 / #2).

**Fix:** Generate a per-request crypto-random nonce, store it in a
short-lived HttpOnly+Secure+SameSite=Strict cookie, pass it as `state`,
reject on mismatch. Optional: PKCE.

---

## 7. JWT secret falls back to a public constant

- **Severity:** HIGH
- **Status:** fixed
- **File:** `backend/middleware/auth.go:28`,
  `backend/handlers/auth.go:160`

Both verifier and issuer fall back to the literal
`"your-secret-key-change-in-production"` when `JWT_SECRET` is unset. No
startup check fails the process when the env var is empty.

**Exploit:** Any deployment that boots without `JWT_SECRET` set signs and
verifies tokens with a public constant. Attacker mints HS256 JWTs for any
user whose ObjectID they have (IDs leak from many endpoints).
Impersonates anyone, including the global admin.

**Fix:** `log.Fatal` on startup if `JWT_SECRET` is empty (or < 32 bytes).
Remove the literal fallback from both call sites. Centralize the lookup.

---

## 8. Round creation + reviewer assignment open to any member

- **Severity:** MEDIUM
- **Status:** fixed
- **File:** `backend/handlers/rounds.go` (`CreateFeedbackRound`,
  `AddReviewersToRound`)
- **Routes:** `POST /api/rounds` (`backend/main.go:89`),
  `POST /api/rounds/:id/reviewers` (`backend/main.go:93`), both only
  `AuthMiddleware()`

`CreateFeedbackRound` only blocks self-targeting and only applies a
team-boundary check when the caller's role is `team_admin`. Plain members
fall through. `AddReviewersToRound` gates only on
`round.CreatedByID == currentUser.ID`. The parallel team-admin route
`/api/teams/:id/rounds/create-batch` is correctly gated by
`TeamAdminOrGlobalAdmin`.

**Exploit:** A regular member creates a draft round targeting any
employee, then enlists arbitrary employees as reviewers.

**Fix:** Add `TeamAdminOrGlobalAdmin()` on both routes. In
`AddReviewersToRound`, validate added reviewers are within scope (same
team for team admins; anywhere for global admin).
