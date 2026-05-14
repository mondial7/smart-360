# Smart 360 Feedback

[![CI](https://github.com/mondial7/smart-360/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/mondial7/smart-360/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/mondial7/smart-360?display_name=tag&sort=semver)](https://github.com/mondial7/smart-360/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Showcase](https://img.shields.io/badge/site-mondial7.github.io%2Fsmart--360-4f46e5)](https://mondial7.github.io/smart-360)

> AI-powered anonymous 360° peer feedback for professional development.

Smart 360 collects structured anonymous feedback from your team and uses
Google Gemini to consolidate it into clear strengths, growth areas, and
concrete next steps. It runs as a single self-hosted binary, in Docker,
or from source — pick whichever fits your environment.

- **Showcase:** <https://mondial7.github.io/smart-360>
- **Source:** <https://github.com/mondial7/smart-360>
- **Changelog:** [`CHANGELOG.md`](CHANGELOG.md)

---

## Table of Contents

- [What it does](#what-it-does)
- [How a round works](#how-a-round-works)
- [Who can do what](#who-can-do-what)
- [Tech stack](#tech-stack)
- [Requirements](#requirements)
- [Install — choose one path](#install--choose-one-path)
  - [Option 1 — Homebrew (single binary, recommended for local / LAN)](#option-1--homebrew-single-binary-recommended-for-local--lan)
  - [Option 2 — Docker Compose (full stack, recommended for a server)](#option-2--docker-compose-full-stack-recommended-for-a-server)
  - [Option 3 — From source (for development)](#option-3--from-source-for-development)
- [First-time setup (after install)](#first-time-setup-after-install)
- [Configuration reference](#configuration-reference)
- [Google OAuth setup](#google-oauth-setup)
- [Gemini API setup](#gemini-api-setup)
- [Releasing a new version](#releasing-a-new-version)
- [Troubleshooting](#troubleshooting)
- [Architecture](#architecture)
- [Security & privacy](#security--privacy)
- [Contributing & license](#contributing--license)

---

## What it does

Smart 360 runs structured peer-feedback rounds and uses Google Gemini to
turn the raw answers into something a person can actually act on. Designed
for scaleup engineering / product teams where calibrated, growth-oriented
feedback matters more than a pile of platitudes.

- **Growth-framed, behaviourally anchored questions** — the default
  template uses Continue / Block / Amplify / Experiment framing. Question
  text is stored in a `templates` collection, so you can ship one template
  per role family (engineering leadership, IC, PM, design, …) without
  changing code.
- **Self-assessment + delta** — the subject answers the same questions in
  first-person. The AI surfaces a `self_vs_others_delta` block with blind
  spots (peers see, self doesn't), hidden strengths (self underplays), and
  aligned themes — the single highest-signal output of a 360.
- **Voice breakdown** — manager, peer, and direct-report inputs are
  synthesised as distinct streams so a manager's "ready for next level"
  doesn't average into a peer's "easy to work with".
- **Competency rubric (Likert)** — templates can declare competencies;
  reviewers rate 1–5 with a one-line justification. The consolidation
  shows per-voice averages, spread, and a wide-spread warning for
  calibration gaps.
- **Private-to-manager channel** — optional textarea on peer submissions
  feeds a manager-only synthesis. Stripped from the subject's API
  response, in-app view, and PDF.
- **Dedicated moderation pass** — a separate Gemini call scrubs identity-
  targeted / personality-attack / off-topic content before the synthesis
  sees it. Every pass is recorded in `moderation_logs` for the audit
  trail.
- **AI consolidation** — executive summary, top 3 focus areas, growth
  opportunities, per-question summaries.
- **PDF export** — branded, print-ready PDF of any shared consolidation
  (with the manager-only section omitted from the subject's copy).
- **Personal & admin analytics** — per-user radar charts and trend bars,
  plus an org-wide dashboard with completion rates and theme extraction.
- **Role-based access + audit log** — Admin, Team Admin, Member; every
  status transition (`draft → active → closed → shared`) is recorded.
- **Single-binary deploy option** — the Vue SPA is embedded in the Go
  server, so Homebrew installs the whole app as one ~30 MB file.

Reviewer identity is anonymised to the subject (the per-reviewer
breakdown only shows relationship + frequency tags); the admin / round
creator can see whose submission is whose for accountability and audit.

## How a round works

1. **Admin (or Team Admin) creates a round** — picks the subject,
   chooses a template, assigns 3–8 reviewers, sets a deadline. The
   creator is automatically excluded from the reviewer list.
2. **The subject does a self-assessment** — same questions, first-person
   wording. Surfaces as a delta against peer answers in the final report.
3. **Reviewers submit** — each declares their relationship to the
   subject (manager / report / peer / cross-functional) and interaction
   frequency (daily / weekly / monthly / rarely), answers the template's
   questions, rates the subject 1–5 on each competency the template
   defines with a one-line justification, and can optionally leave a
   private note for the manager.
4. **Round closes at the deadline** — no further submissions.
5. **Moderation scrub** — when an admin triggers consolidation, every
   submission goes through a dedicated Gemini call that strips identity-
   targeted / personality-attack / off-topic content. Logged to
   `moderation_logs` for the audit trail.
6. **AI synthesis** — a second Gemini call produces the executive
   summary, top 1–3 focus areas, growth opportunities, per-question
   summaries, the self-vs-peers delta, the manager / peer / report voice
   streams, and a manager-only synthesis of the private notes.
7. **Server-side aggregates** — competency averages per voice, spread,
   and reviewer counts are computed deterministically (not via the AI).
8. **Admin reviews, optionally adds notes, shares** — sharing flips the
   round to `shared` and unlocks the consolidation for the subject.
9. **Subject reads, tracks progress** — sees everything **except** the
   manager-only channel (stripped on both API and PDF).

**Default-template questions (Growth-Framed 360):**

1. What does this person do that has the biggest positive impact on the
   team or product? — *one Situation → Behaviour → Impact example.*
2. Looking at the last 3–6 months, what's currently holding this person
   back from their next level of impact?
3. If this person doubled down on one strength over the next 6 months,
   what should it be — and what would change for the team?
4. What's one concrete experiment or focus area you'd suggest they try
   in the next 30–60 days?

The bundled "Engineering Leadership 360" template ships with
multiplier / scope-and-judgement / technical-depth / coaching framing
for tech leads, EMs, and staff+ engineers.

## Who can do what

| Role | Can do |
|------|--------|
| **Admin** | Everything — manage users + teams, create any round, view all rounds, generate and share any consolidation. First user to sign in is auto-promoted. |
| **Team Admin** | Manage their team's members, create and consolidate rounds **within their team**. |
| **Member** | Submit feedback when assigned as reviewer, read consolidations shared with them (in-app + PDF), see their personal analytics. |

Roadmap and known gaps live in
[GitHub Issues](https://github.com/mondial7/smart-360/issues) — anything
labelled `good first issue` is a friendly entry point.

---

## Tech stack

- **Backend:** Go 1.25 · Gin · MongoDB driver · `generative-ai-go` (Gemini)
- **Frontend:** Vue 3 (Composition API) · TypeScript · Vite · Pinia · SASS
- **Database:** MongoDB 8.0
- **Auth:** Google OAuth 2.0 → JWT
- **PDF:** `github.com/go-pdf/fpdf` (server-side)
- **Distribution:** single embedded Go binary · Docker Compose · Homebrew formula

---

## Requirements

| Path | What you need |
|------|---------------|
| Homebrew | macOS or Linux with Homebrew · a running MongoDB · Google OAuth credentials · Gemini API key |
| Docker | Docker + Docker Compose v2 · Google OAuth credentials · Gemini API key |
| From source | Go 1.25 · Node 20 · running MongoDB · Google OAuth credentials · Gemini API key |

For OAuth and Gemini setup steps see [Google OAuth setup](#google-oauth-setup) and [Gemini API setup](#gemini-api-setup) below.

**Production deployment** — once you've picked an install path, the
[Production deployment guide](docs/deployment-production.md) walks you
from "I have a domain and a server" to a public HTTPS install with
Caddy (or nginx) and a systemd unit.

---

## Install — choose one path

### Option 1 — Homebrew (single binary, recommended for local / LAN)

This is the simplest path. You get one self-contained Go binary with the
Vue SPA embedded, managed by `brew services`.

> **Heads up:** until a tagged release is published on GitHub, the formula
> in this repo carries placeholder URLs. See
> [Releasing a new version](#releasing-a-new-version) to cut your first
> release, or install [from source](#option-3--from-source-for-development) in the meantime.

```bash
# 1. Install and start MongoDB (one option of many)
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community

# 2. Install Smart 360
brew tap mondial7/tap        # repo: github.com/mondial7/homebrew-tap
brew install smart360

# 3. Create your env file from the template
smart360-setup
$EDITOR ~/.config/smart360/.env

# 4. Start as a background service
brew services start smart360
```

Open <http://localhost:8080>. To stop / restart:

```bash
brew services restart smart360
brew services stop smart360
```

**LAN access:** the service binds to `0.0.0.0:8080` by default, so other
machines on your LAN can reach it at `http://<your-ip>:8080`. Update
`GOOGLE_REDIRECT_URL` and `FRONTEND_URL` accordingly (Google OAuth must
accept the redirect URI you publish).

---

### Option 2 — Docker Compose (full stack, recommended for a server)

This brings up MongoDB, the Go API, and an nginx-served frontend in one
command. The repo ships **two compose files** — pick the one that fits:

| File | Image source | When to use |
|------|--------------|-------------|
| `docker-compose.yml` | `build:` from local source | Active development on the stack; you have a clone of the repo and want to iterate on the code. |
| `docker-compose.prod.yml` | `image:` from `ghcr.io/mondial7/smart-360-*` | Production / staging / "pull and run" — no source checkout needed beyond the compose file and an `.env`. |

#### Development (`docker-compose.yml`)

```bash
git clone https://github.com/mondial7/smart-360.git
cd smart-360

cp .env.example .env
$EDITOR .env            # see "Configuration reference" below

docker compose up -d --build
```

#### Production (`docker-compose.prod.yml`)

You only need two files on the host: the compose file and your `.env`.

```bash
mkdir smart360 && cd smart360
curl -O https://raw.githubusercontent.com/mondial7/smart-360/main/docker-compose.prod.yml
curl -O https://raw.githubusercontent.com/mondial7/smart-360/main/.env.example
mv .env.example .env
$EDITOR .env

# Pin to a released version. `latest` works but should be avoided in prod.
SMART360_VERSION=v1.0.0 docker compose -f docker-compose.prod.yml up -d
```

Open <http://localhost> (or whatever you set as `FRONTEND_PORT`).

#### Common commands (same for both files — add `-f docker-compose.prod.yml` for the prod variant)

```bash
docker compose ps                  # service status
docker compose logs -f backend     # tail backend logs
docker compose restart backend     # restart one service
docker compose up -d --build       # rebuild (dev only)
docker compose pull                # pull newer images (prod only)
docker compose down                # stop everything (data preserved)
docker compose down -v             # ⚠️  also drops the MongoDB volume
```

---

### Option 3 — From source (for development)

```bash
git clone https://github.com/mondial7/smart-360.git
cd smart-360

# Build the single binary (frontend + embedded Go server)
make build

cp .env.example .env
$EDITOR .env

# Make sure MongoDB is running locally (or in Docker), then:
./smart360
```

This produces `./smart360` (~30 MB) containing the whole app. Open
<http://localhost:8080>.

For day-to-day development you'll usually prefer running the two halves
separately with hot reload — see [`CLAUDE.md`](CLAUDE.md) "Quick Start (development)".

---

## First-time setup (after install)

1. Open the app and click **Login with Google**.
2. The first user to log in is automatically promoted to **Administrator**.
3. From the **Users** page, promote others to Admin or Team Admin as needed.
4. From the dashboard, create your first feedback round.

See [How a round works](#how-a-round-works) above for the full lifecycle, or the
[showcase page](docs/index.html) for the visual version.

---

## Configuration reference

All configuration is read from the environment. For Homebrew installs the
canonical file is `~/.config/smart360/.env`. For Docker / source installs
it is `.env` in the repo root.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MONGODB_URI` | Yes (non-Docker) | `mongodb://admin:password123@localhost:27017` | Mongo connection string. In Docker Compose this is built from `MONGO_ROOT_USER` + `MONGO_ROOT_PASSWORD`. |
| `MONGODB_DB` | No | `smart360` | Database name. |
| `MONGO_ROOT_USER` | Docker only | `admin` | Mongo admin user (compose). |
| `MONGO_ROOT_PASSWORD` | Docker only | `password123` | **Change before deploying.** |
| `JWT_SECRET` | **Yes** | — | Required. Generate with `openssl rand -base64 32`. The server refuses to start without it. |
| `GOOGLE_CLIENT_ID` | Yes | — | OAuth client ID. |
| `GOOGLE_CLIENT_SECRET` | Yes | — | OAuth client secret. |
| `GOOGLE_REDIRECT_URL` | No | `http://localhost:8080/api/auth/callback` | Must match the URI registered in Google Cloud Console. |
| `FRONTEND_URL` | No | `http://localhost:5173` (dev) / `http://localhost` (Docker) | Used for CORS when the SPA is not served from the same origin. Ignored in single-binary mode. |
| `FRONTEND_PORT` | Docker only | `80` | Host port for the nginx frontend. |
| `GEMINI_API_KEY` | Yes | — | Required for AI consolidation. |
| `PORT` | No | `8080` | Port the Go server listens on. |
| `DEV_MODE` | No | unset | When `true`, enables the dev-login endpoint and seeds demo data. **Never enable in production.** |

### Security notes

- **Generate a real `JWT_SECRET`**: `openssl rand -base64 32` and paste the result.
- **Change `MONGO_ROOT_PASSWORD`** before exposing the stack anywhere.
- **Never commit `.env`** — it's already in `.gitignore`.
- **OAuth in production requires HTTPS** — terminate TLS at a reverse proxy (Caddy, nginx) and update `GOOGLE_REDIRECT_URL`.

---

## Google OAuth setup

1. Open [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials).
2. Create a new project or pick an existing one.
3. **Create credentials → OAuth client ID → Web application**.
4. Add an **Authorized redirect URI**:
   - Local: `http://localhost:8080/api/auth/callback`
   - Production: `https://yourdomain.com/api/auth/callback`
5. Copy the **Client ID** and **Client Secret** into your env file.

For detail, see Google's [OAuth 2.0 docs](https://developers.google.com/identity/protocols/oauth2).

---

## Gemini API setup

1. Open [Google AI Studio](https://makersuite.google.com/app/apikey).
2. Sign in.
3. **Create API Key**.
4. Paste the value into `GEMINI_API_KEY`.

Monitor usage and quota in the [Google Cloud Console](https://console.cloud.google.com/apis/api/generativelanguage.googleapis.com/quotas).

---

## Releasing a new version

Cutting a release builds the cross-platform binaries that Homebrew users
download. The heavy lifting is automated by
[`.github/workflows/release.yml`](.github/workflows/release.yml): push a
`v*` tag and the workflow runs `make release`, attaches the tarballs and
`SHA256SUMS.txt` to a GitHub release, and writes generated release notes.

```bash
# 1. Tag and push — the Release workflow does the cross-compile + publish.
git tag -a v1.0.0 -m "v1.0.0"
git push origin v1.0.0

# 2. Wait for the workflow to finish, then download SHA256SUMS.txt from
#    the release page. Update Formula/smart360.rb with the new `version`
#    and per-arch `sha256` values.

# 3. Test the formula locally:
brew install --build-from-source ./Formula/smart360.rb
smart360 --version

# 4. Publish to the tap so end users get it via `brew install smart360`.
#    Copy the updated Formula/smart360.rb into the tap repo and push.
git -C /path/to/homebrew-tap pull
cp Formula/smart360.rb /path/to/homebrew-tap/Formula/smart360.rb
git -C /path/to/homebrew-tap add Formula/smart360.rb
git -C /path/to/homebrew-tap commit -m "smart360 v1.0.0"
git -C /path/to/homebrew-tap push
```

To dry-run the build locally without tagging, run `make release VERSION=v1.0.0`
and inspect `./dist/`. Automating the tap bump itself is tracked in
[#36](https://github.com/mondial7/smart-360/issues/36).

---

## Troubleshooting

### Server won't start: `JWT_SECRET environment variable is required`

You haven't set `JWT_SECRET`. Generate one with `openssl rand -base64 32`
and put it in your env file.

### OAuth: redirect URI mismatch

Make sure `GOOGLE_REDIRECT_URL` exactly matches one of the redirect URIs
configured for your OAuth client in Google Cloud Console — including the
scheme, host, and `/api/auth/callback` path. For Homebrew installs the
default is `http://localhost:8080/api/auth/callback`.

### AI consolidation fails

Verify `GEMINI_API_KEY` and check the backend log. Most issues are quota
exhaustion or a missing/invalid key. Quota lives at
[Google Cloud Console → Generative Language API](https://console.cloud.google.com/apis/api/generativelanguage.googleapis.com/quotas).

### Port 80 already in use (Docker)

Set `FRONTEND_PORT=3000` (or any free port) in `.env` and rerun
`docker compose up -d`. Open <http://localhost:3000>.

### Single binary returns 404 for `/some-route`

Make sure you built with `make build` (which copies `frontend/dist/` into
`backend/web/dist/` before `go build`). If you ran `go build` directly the
embedded SPA may be empty — the binary still works for the API but won't
serve the UI.

### MongoDB connection errors

- Docker: `docker compose logs mongodb`, ensure healthcheck has passed.
- Homebrew / source: `brew services list`, ensure `mongodb-community` is `started`.
- Validate `MONGODB_URI` reachable with `mongosh "$MONGODB_URI"`.

---

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                Single-binary deployment                  │
│                                                          │
│   smart360 (Go)                                          │
│   ├─ REST API  (/api/*)                                  │
│   └─ Embedded Vue SPA  (everything else)                 │
│                                                          │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
                    MongoDB :27017
```

```
┌──────────────────────────────────────────────────────────┐
│                 Docker Compose deployment                │
│                                                          │
│   nginx (frontend)  ──/api──▶  backend (Go)              │
│                                       │                  │
│                                       ▼                  │
│                                  MongoDB                 │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

Backend lives in `backend/`, frontend in `frontend/`, the GitHub Pages
showcase in `docs/`, and the Homebrew formula in `Formula/`. See
[`CLAUDE.md`](CLAUDE.md) for repository conventions and the test pyramid.

---

## Security & privacy

The product handles sensitive workplace feedback, so the way that data
is stored, who can see it, and what gets sent to Google Gemini are
documented separately in [`SECURITY.md`](SECURITY.md). It also explains
how to report a vulnerability (please don't open a public issue for
exploitable findings — use the private form linked from there).

---

## Contributing & license

Contributions welcome. Read [`CLAUDE.md`](CLAUDE.md) for repo conventions
(commit format, test layout, Phosphor icons, scoped SASS, etc.) and browse
[open issues](https://github.com/mondial7/smart-360/issues) — anything
labeled `good first issue` is a friendly entry point.

1. Fork and create a feature branch (`feat/your-thing`).
2. Run tests: `cd backend && go test ./...`.
3. Commit using the conventional prefixes (`feat:`, `fix:`, `docs:`, …).
4. Open a PR with a clear description.

Licensed under the [MIT License](LICENSE).
