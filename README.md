# Smart 360 Feedback

> AI-powered anonymous 360° peer feedback for professional development.

Smart 360 collects structured anonymous feedback from your team and uses
Google Gemini to consolidate it into clear strengths, growth areas, and
concrete next steps. It runs as a single self-hosted binary, in Docker,
or from source — pick whichever fits your environment.

**Live showcase:** <https://mondial7.github.io/smart-360> (once GitHub Pages is enabled)
**Source:** <https://github.com/mondial7/smart-360>

---

## Table of Contents

- [Highlights](#highlights)
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
- [Contributing & license](#contributing--license)

---

## Highlights

- **Anonymous, structured feedback** — four prompts per reviewer, identities never exposed to the subject.
- **AI consolidation** — Google Gemini turns raw answers into executive summary + strengths + growth areas + action items.
- **PDF export** — branded, print-ready PDF of any shared consolidation.
- **Personal & admin analytics** — radar charts per round, completion trends, theme extraction.
- **Role-based access** — Admin, Team Admin, Member; every status change recorded in an audit log.
- **Single-binary deploy option** — the SPA is embedded into the Go binary, so Homebrew installs the entire app from one file.

See [`PRODUCT-OVERVIEW.md`](PRODUCT-OVERVIEW.md) and [`PRD.md`](PRD.md) for product context, and [`NEXT_STEPS.md`](NEXT_STEPS.md) for the roadmap.

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

See "How it works" in the [showcase page](docs/index.html) or
[`PRODUCT-OVERVIEW.md`](PRODUCT-OVERVIEW.md) for the full round lifecycle.

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
download. The first time you do this, replace the placeholder `version`,
`url`s, and `sha256` values in [`Formula/smart360.rb`](Formula/smart360.rb).

```bash
# 1. Cross-compile + create release archives
make release VERSION=v1.0.0

# 2. Inspect artifacts (./dist/)
ls dist/
#  smart360-v1.0.0-darwin-arm64.tar.gz
#  smart360-v1.0.0-darwin-amd64.tar.gz
#  smart360-v1.0.0-linux-amd64.tar.gz
#  smart360-v1.0.0-linux-arm64.tar.gz
#  smart360-v1.0.0-SHA256SUMS.txt

# 3. Tag and push (this is where GitHub Actions would publish images / pages
#    once you wire that up — see NEXT_STEPS.md #11):
git tag -a v1.0.0 -m "v1.0.0"
git push origin v1.0.0

# 4. Create a GitHub release at https://github.com/mondial7/smart-360/releases/new
#    and attach every file in dist/.

# 5. Update Formula/smart360.rb with the new version + sha256 values
#    from dist/smart360-v1.0.0-SHA256SUMS.txt.

# 6. Test the formula locally:
brew install --build-from-source ./Formula/smart360.rb
smart360 --version

# 7. Publish to the tap so end users get it via `brew install smart360`:
#    Copy the updated Formula/smart360.rb into the tap repo and push.
git -C /path/to/homebrew-tap pull
cp Formula/smart360.rb /path/to/homebrew-tap/Formula/smart360.rb
git -C /path/to/homebrew-tap add Formula/smart360.rb
git -C /path/to/homebrew-tap commit -m "smart360 v1.0.0"
git -C /path/to/homebrew-tap push
```

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

## Contributing & license

Contributions welcome. Read [`CLAUDE.md`](CLAUDE.md) for repo conventions
(commit format, test layout, Phosphor icons, scoped SASS, etc.) and
[`NEXT_STEPS.md`](NEXT_STEPS.md) for the roadmap.

1. Fork and create a feature branch (`feat/your-thing`).
2. Run tests: `cd backend && go test ./...`.
3. Commit using the conventional prefixes (`feat:`, `fix:`, `docs:`, …).
4. Open a PR with a clear description.

Licensed under the [MIT License](LICENSE).
