# Contributing to Smart 360 Feedback

Thanks for considering a contribution. This guide covers the things
that aren't already in [`README.md`](README.md) (which is the install
+ user manual) or [`CLAUDE.md`](CLAUDE.md) (which is the architecture
deep-dive for humans and AI assistants).

---

## Before you start

- **Open an issue first** for any change that's not a one-line bug fix
  or a typo. A 30-second "is this in scope?" conversation saves a lot
  of rebasing later.
- **Pick something from [Issues](https://github.com/mondial7/smart-360/issues)** —
  that's the canonical roadmap.
- Comment on the issue saying you're picking it up so we don't double
  up.

## Development setup

Full quick-start lives in [`CLAUDE.md` → Quick start](CLAUDE.md#quick-start-development).
The app is a single server-rendered Go binary (chi + `html/template` + htmx),
backed by PostgreSQL. There is no separate frontend build. TL;DR:

```bash
git clone https://github.com/mondial7/smart-360.git
cd smart-360
docker compose up -d postgres   # start Postgres (or: make docker-up)
cp .env.example .env            # DEV_MODE=true is fine for local
go run ./cmd/server             # migrates + seeds + serves :8080
```

Sign in via the dev-login on the login page (`DEV_MODE=true`). The first user
to sign in becomes admin.

To run the checks the way CI does:

```bash
go vet ./...
gofmt -l cmd internal web       # must print nothing
go test ./...                   # full pyramid (gateway tests need Docker)
```

## Repository conventions

These are non-negotiable — CI will refuse changes that violate them.

### Commit messages

[Conventional Commits](https://www.conventionalcommits.org/). One of:
`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`, `ci:`,
`perf:`, `style:`, `build:`.

Keep commits focused and atomic; prefer two small commits over one
sprawling one. **Do not include Claude or any AI assistant as a
commit co-author.**

### Code style

- **Go** — `gofmt`, `go vet`. Keep handlers thin and push business
  logic into `internal/ai`, `internal/pdf`, or helpers. IDs are UUID
  strings; queries are parameterized SQL in `internal/repo`. Always
  check errors and return early with a generic client message; log the
  detail. The `ai` and `pdf` packages stay free of HTTP/DB.
- **UI** — server-rendered `html/template` + [htmx](https://htmx.org) +
  SSE. Content templates define `{{define "<name>_content"}}` and render
  inside `base.html`. **Don't** reintroduce a JS framework.
- **Charts** — server-rendered SVG (`internal/view/charts.go`). Don't
  add a client-side chart library.
- **Decisions** — record any significant architecture choice as a new
  ADR in [`docs/adr/`](docs/adr/README.md).

### Tests

One command — `go test ./...` — runs the three-layer pyramid:

1. **Unit** — pure logic in `ai` / `view` / `auth` / `config`.
2. **In-memory integration** — handlers against the `repo` fakes via a
   real `httptest` server (`internal/handlers/*_test.go`). No database.
3. **Gateway** — pgx repositories against a real PostgreSQL via
   `testcontainers-go` (`internal/repo/gateway_*_test.go`).

`go test -short ./...` skips the container-backed gateway layer (no
Docker needed). There is intentionally no browser/E2E layer.

## Pull request process

1. Fork the repo and create a feature branch
   (`feat/short-description`, `fix/...`, `docs/...`).
2. Make focused commits following the convention above.
3. Run `gofmt`, `go vet ./...`, and `go test ./...` locally. CI runs
   the same things (plus a govulncheck-clean dependency tree) on every PR.
4. Open the PR. The template will prompt you for a summary, linked
   issue, and test plan — please fill them in.
5. A maintainer reviews. Address feedback in additional commits
   (don't force-push during review unless asked — it makes the diff
   harder to track).
6. Once approved, a maintainer merges. We use **squash merge** so PRs
   land as a single commit on `main` with the PR title as the
   subject.

## Reporting security issues

**Do not file public issues for security vulnerabilities.** See
[`SECURITY.md`](SECURITY.md) for the private disclosure process.

## License

By contributing you agree your contributions will be licensed under
the [MIT License](LICENSE) — the same license that covers the rest of
the project.
