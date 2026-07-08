# Architecture Decision Records

This log records the significant architecture decisions for Smart 360, one file
per decision. See [ADR-0001](0001-record-architecture-decisions.md) for why we
keep these and [`0000-template.md`](0000-template.md) for the format.

ADRs are immutable once accepted; a reversal gets a new ADR that supersedes the
old one.

| # | Decision | Status |
|---|----------|--------|
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | accepted |
| [0002](0002-rewrite-to-go-htmx-postgres.md) | Rewrite to a server-rendered Go + htmx + Postgres stack | accepted |
| [0003](0003-pgx-hand-written-sql.md) | Data access with pgx and hand-written SQL | accepted |
| [0004](0004-session-cookie-auth.md) | Server-side sessions with an HttpOnly cookie | accepted |
| [0005](0005-chi-router-and-html-template.md) | chi router + html/template rendering | accepted |
| [0006](0006-sse-for-consolidation.md) | Server-Sent Events for consolidation progress | accepted |
| [0007](0007-normalize-schema-reviewers-and-members.md) | Normalize the relational schema | accepted |
