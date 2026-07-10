# ADR 0004: Server-side sessions with an HttpOnly cookie

- **Status**: accepted
- **Date**: 2026-07-08
- **Deciders**: Marco Mondini

## Context

The old SPA authenticated with a JWT stored in `localStorage` and sent as an
`Authorization: Bearer` header on every XHR. In a server-rendered app there is
no long-lived JS client to hold a token, and `localStorage` tokens are readable
by any injected script (XSS token theft). We need an auth model that fits SSR +
htmx and has a better browser-security posture.

## Decision

We will use **server-side sessions** stored in a Postgres `sessions` table,
referenced by an opaque session ID carried in a **Secure, HttpOnly,
SameSite=Lax** cookie. The cookie value is `<sessionID>.<HMAC>`, signed with
`SESSION_SECRET`, so a forged or tampered cookie is rejected before any database
lookup. `Secure` is relaxed only when `DEV_MODE=true`.

Google OAuth is unchanged in spirit (state-cookie CSRF on the login redirect);
on callback we upsert the user and create a session. Logout deletes the session
row and clears the cookie.

Because cookies are sent automatically, we add **CSRF protection**: a token
derived from the session (`HMAC("csrf:"+sessionID)`), embedded in forms and sent
by htmx via the `X-CSRF-Token` header, verified on every unsafe method.

## Consequences

- Tokens are never exposed to JavaScript (HttpOnly), removing the localStorage
  XSS-exfiltration class of bug.
- Sessions are revocable (delete the row) and expire server-side — impossible
  with stateless JWTs.
- One extra DB read per request to resolve the session; negligible at this scale
  and cacheable later if needed.
- We must ship CSRF protection (which we did) since we now rely on cookies.

## Alternatives considered

- **Keep stateless JWT, move it to an HttpOnly cookie** — avoids the sessions
  table but keeps non-revocable tokens and still needs CSRF; the table is cheap
  and buys revocation.
- **Encrypted stateless cookie sessions** (e.g. gorilla/securecookie) — viable,
  but a DB-backed session is simpler to reason about and revoke here.
