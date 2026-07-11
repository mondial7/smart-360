# Production deployment guide

This guide takes you from "I have a domain and a server" to
"`https://feedback.example.com` is live, OAuth works, and the service
restarts on reboot."

It is intentionally opinionated: **Caddy** for the reverse proxy (automatic
Let's Encrypt) and the **single-binary** install path. An nginx + Certbot
alternative is included at the end.

> If you only need the app on your laptop or LAN, you don't need this guide —
> see the README's "Run it locally" section.

---

## Prerequisites

- A Linux server (or VM) reachable from the internet — Ubuntu 24.04 LTS is
  assumed in the examples.
- A domain name you control, with DNS pointed at the server.
- Ports 80 and 443 open (Let's Encrypt needs 80 for the HTTP-01 challenge; your
  users need 443).
- A **PostgreSQL** instance (local or managed). Examples below assume Postgres
  on the same host on port 5432 with a `smart360` database.
- Google OAuth credentials and (optionally) a Gemini API key.

The app runs its own schema migrations and template seeding on boot — you only
need to create an empty database and grant access.

---

## 1. DNS

Point an `A` (and optionally `AAAA`) record at your server:

```
feedback.example.com.  A    203.0.113.42
feedback.example.com.  AAAA 2001:db8::42
```

Verify: `dig +short feedback.example.com` → `203.0.113.42`.

---

## 2. Install Smart 360

Both paths end up serving on `localhost:8080`.

### Option A — Single binary (recommended for this guide)

```bash
# Replace v1.0.0 with the latest release.
ARCH="$(uname -m)"; case "$ARCH" in x86_64) ARCH=amd64;; aarch64) ARCH=arm64;; esac
curl -L -o smart360.tar.gz \
  "https://github.com/mondial7/smart-360/releases/download/v1.0.0/smart360-v1.0.0-linux-${ARCH}.tar.gz"
tar -xzf smart360.tar.gz
sudo install -o root -g root -m 0755 smart360-v1.0.0-linux-${ARCH}/smart360 /usr/local/bin/smart360
```

### Option B — Docker Compose

Use `docker-compose.prod.yml` (it brings up Postgres + the app). Skip the
systemd step below and use `docker compose -f docker-compose.prod.yml up -d`.

---

## 3. Create the service user, env file, and systemd unit

(Skip this section if you're using Docker Compose.)

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin smart360
sudo mkdir -p /etc/smart360
sudo curl -L -o /etc/smart360/.env \
  https://raw.githubusercontent.com/mondial7/smart-360/main/.env.example
sudo chown -R smart360:smart360 /etc/smart360
sudo chmod 600 /etc/smart360/.env
sudo $EDITOR /etc/smart360/.env
```

Fill in:

```bash
DATABASE_URL=postgres://smart360:<password>@localhost:5432/smart360?sslmode=disable

SESSION_SECRET=<run: openssl rand -hex 32>

# The account that signs in with this email becomes the admin. Set it before
# first sign-in.
ADMIN_EMAIL=you@example.com

APP_URL=https://feedback.example.com

# Structured logs for a shipper (optional): text | json
LOG_FORMAT=json

GOOGLE_CLIENT_ID=<from Google Cloud Console>
GOOGLE_CLIENT_SECRET=<from Google Cloud Console>
GOOGLE_REDIRECT_URL=https://feedback.example.com/auth/callback

GEMINI_API_KEY=<from Google AI Studio — optional>

# DEV_MODE must be false (or unset) in production.
DEV_MODE=false
# Optional: change the listen port (default 8080)
# PORT=8080
```

Create the systemd unit:

```bash
sudo tee /etc/systemd/system/smart360.service >/dev/null <<'UNIT'
[Unit]
Description=Smart 360 Feedback
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
Type=simple
User=smart360
Group=smart360
EnvironmentFile=/etc/smart360/.env
WorkingDirectory=/etc/smart360
ExecStart=/usr/local/bin/smart360
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
UNIT

sudo systemctl daemon-reload
sudo systemctl enable --now smart360
sudo systemctl status smart360
```

Confirm the server is up:

```bash
curl -sf http://127.0.0.1:8080/healthz
# → ok
```

---

## 4. Reverse proxy with HTTPS (Caddy)

Caddy auto-provisions and renews Let's Encrypt certificates, and streams
Server-Sent Events (used by the live consolidation) without extra config.

```bash
# Install Caddy (Ubuntu / Debian)
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install -y caddy
```

Write the Caddyfile:

```bash
sudo tee /etc/caddy/Caddyfile >/dev/null <<'CADDY'
feedback.example.com {
        encode gzip
        reverse_proxy localhost:8080

        header {
                Strict-Transport-Security "max-age=63072000; includeSubDomains; preload"
                X-Content-Type-Options "nosniff"
                X-Frame-Options "SAMEORIGIN"
                Referrer-Policy "strict-origin-when-cross-origin"
        }

        log {
                output file /var/log/caddy/feedback.access.log
                format json
        }
}
CADDY

sudo systemctl reload caddy
```

Visit `https://feedback.example.com` — you should land on the Smart 360 login
page.

---

## 5. Configure the Google OAuth client

1. Open [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials).
2. Edit your OAuth 2.0 client.
3. Under **Authorized redirect URIs**, add:
   `https://feedback.example.com/auth/callback`
4. Under **Authorized JavaScript origins**, add:
   `https://feedback.example.com`
5. Save. Restart the app so it picks up the new `GOOGLE_REDIRECT_URL`:

```bash
sudo systemctl restart smart360
```

Sign in with the address you set as `ADMIN_EMAIL` — that account is the
**Administrator**. Everyone else is a member until you promote them from the
Users page.

---

## 6. Alternative — nginx + Certbot

For readers who already run nginx. Note the SSE-friendly settings
(`proxy_buffering off`, a long read timeout) — without them the live
consolidation progress will appear to stall.

```nginx
# /etc/nginx/sites-available/smart360
server {
    listen 80;
    server_name feedback.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name feedback.example.com;

    ssl_certificate     /etc/letsencrypt/live/feedback.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/feedback.example.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Required for the SSE consolidation stream.
        proxy_buffering off;
        proxy_read_timeout 300s;
    }
}
```

Then:

```bash
sudo ln -s /etc/nginx/sites-available/smart360 /etc/nginx/sites-enabled/
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d feedback.example.com
sudo nginx -t && sudo systemctl reload nginx
```

---

## What this guide does NOT cover

The companion [`SECURITY.md`](../SECURITY.md) spells out the privacy model, the
AI data flow, and the full hardening checklist. Not covered here:

- **Rate limiting** — until a built-in limiter ships, throttle abusive traffic
  at the reverse proxy (`rate_limit` in Caddy 2.7+, `limit_req` in nginx).
- **Database backups** — schedule a nightly `pg_dump` to off-host storage.
- **Observability** — the process logs to stdout (captured by
  `journalctl -u smart360`); there is no Prometheus / OTel integration yet.

(Session-cookie auth and CSRF protection are built in — see
[ADR-0004](adr/0004-session-cookie-auth.md).)

---

## Backups & disaster recovery

All state lives in PostgreSQL — the app itself is stateless (templates and assets
are baked into the binary). So a database backup is a full backup.

The repo ships two helpers that wrap `pg_dump`/`pg_restore`:

```bash
# Back up to ./backups (compressed custom-format dump), pruning dumps > 14 days.
DATABASE_URL=postgres://smart360:<pass>@localhost:5432/smart360 ./scripts/backup.sh

# Restore a dump into a target database (drops & recreates objects).
DATABASE_URL=postgres://smart360:<pass>@localhost:5432/smart360 \
  ./scripts/restore.sh backups/smart360-YYYYMMDDTHHMMSSZ.dump
```

Schedule the backup nightly and **ship the dumps off-host** — they contain all
feedback data. A cron example:

```cron
# /etc/cron.d/smart360-backup  (runs 02:30 daily as the smart360 user)
30 2 * * * smart360 DATABASE_URL='postgres://smart360:<pass>@localhost:5432/smart360' BACKUP_DIR=/var/backups/smart360 RETENTION_DAYS=14 /usr/local/lib/smart360/backup.sh >> /var/log/smart360-backup.log 2>&1
```

Docker Compose users can run the same dump against the `postgres` service:

```bash
docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump --format=custom --no-owner --no-privileges -U smart360 smart360 \
  > "smart360-$(date -u +%Y%m%dT%H%M%SZ).dump"
```

Recovery drill: after restoring into a fresh database and pointing `DATABASE_URL`
at it, start the app — it re-applies any pending migrations on boot. **Test a
restore periodically**; an untested backup is a guess, not a recovery plan.

---

## Operational notes

### Updating to a new version

Single binary:

```bash
curl -L -o smart360.tar.gz https://github.com/mondial7/smart-360/releases/download/v1.1.0/smart360-v1.1.0-linux-amd64.tar.gz
curl -L -o sums.txt        https://github.com/mondial7/smart-360/releases/download/v1.1.0/smart360-v1.1.0-SHA256SUMS.txt
sha256sum -c sums.txt --ignore-missing
tar -xzf smart360.tar.gz
sudo install -o root -g root -m 0755 smart360-v1.1.0-linux-amd64/smart360 /usr/local/bin/smart360
sudo systemctl restart smart360   # runs any new migrations on boot
```

Docker Compose:

```bash
SMART360_VERSION=v1.1.0 docker compose -f docker-compose.prod.yml pull
SMART360_VERSION=v1.1.0 docker compose -f docker-compose.prod.yml up -d
```

### Log tails

```bash
journalctl -u smart360 -f                                  # single binary
docker compose -f docker-compose.prod.yml logs -f app      # Docker Compose
```

### Where logs live

| Component | Location |
|-----------|----------|
| App (systemd) | `journalctl -u smart360` |
| Caddy access | `/var/log/caddy/feedback.access.log` |
| nginx access | `/var/log/nginx/access.log` |
| PostgreSQL | `/var/log/postgresql/postgresql-16-main.log` (default Ubuntu install) |
