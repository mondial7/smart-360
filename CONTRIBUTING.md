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

Full quick-start lives in [`CLAUDE.md` → Quick Start](CLAUDE.md#quick-start-development).
TL;DR:

```bash
git clone https://github.com/mondial7/smart-360.git
cd smart-360
cp .env.example .env       # fill in JWT_SECRET, GOOGLE_*, GEMINI_API_KEY
make dev-mongo             # start MongoDB
cd backend && ./reseed-dev.sh
cd backend && go run main.go             # :8080
cd frontend && npm install && npm run dev # :5173
```

To run the whole pyramid the way CI does it:

```bash
cd backend && go vet ./... && go test ./...
cd frontend && npm ci && npm run build
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
  logic into helpers / repositories. Convert hex IDs to
  `primitive.ObjectID` at the boundary; don't pass strings deep into
  the stack. Always check errors and return early with a generic
  client message; log the detail.
- **TypeScript / Vue** — Composition API with `<script setup
  lang="ts">`. Type everything; **don't** introduce `any` to silence
  the compiler. Phosphor icons (`@phosphor-icons/vue`), scoped SASS
  with CSS variables, named imports sorted by source.
- **Charts** — pure SVG (see `RadarChart.vue`). Don't add a chart
  library unless requirements force it.

### Tests

The backend follows a three-layer test pyramid:

1. **Unit** — pure handler / helper logic, no I/O.
2. **In-memory integration** — handler-level tests using
   `repositories/fake_*` fakes.
3. **Gateway / port** — repository tests hitting a real MongoDB via
   `testutil/mongodb.go` (uses `memongo`).

`go test ./...` runs all three. `go test -short ./...` skips the
gateway layer.

The frontend has Vitest + `@vue/test-utils` component and store tests
(`*.spec.ts` next to the code). Run them with `npm test`. There's no
browser/E2E layer and none is planned — the Vitest suite is the
frontend test strategy. Add specs next to any code you change.

## Pull request process

1. Fork the repo and create a feature branch
   (`feat/short-description`, `fix/...`, `docs/...`).
2. Make focused commits following the convention above.
3. Run the test pyramid + `npm run build` locally. CI runs the same
   things on every PR.
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
