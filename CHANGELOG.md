# Changelog

All notable changes to Smart 360 Feedback are documented here. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Nothing yet — the next release will be `1.0.0` once the items in
[issue #11](https://github.com/mondial7/smart-360/issues/11) are
checked off.

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

See [`NEXT_STEPS.md`](NEXT_STEPS.md) for the full list. The most notable
gaps for self-hosters:

- **No built-in rate limiting or CSRF protection** — mitigate at the
  reverse proxy (Caddy `rate_limit`, nginx `limit_req`) until the
  enhancements in `NEXT_STEPS.md` §7 land.
- **No automated MongoDB backups** — schedule a nightly `mongodump`
  yourself. Tracked in `NEXT_STEPS.md` §10.
- **No structured logging / metrics / tracing** — the process logs to
  stdout only. Tracked in `NEXT_STEPS.md` §9.

[Unreleased]: https://github.com/mondial7/smart-360/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/mondial7/smart-360/releases/tag/v1.0.0
