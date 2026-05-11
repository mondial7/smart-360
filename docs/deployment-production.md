# Production deployment guide

This guide takes you from "I have a domain and a server" to
"`https://feedback.example.com` is live, OAuth works, and the service
restarts on reboot."

It is intentionally opinionated: **Caddy** for the reverse proxy
(automatic Let's Encrypt) and the **single-binary** install path. An
nginx + Certbot alternative is included at the end for readers who
already run nginx.

> If you only need the app on your laptop or LAN, you don't need this
> guide — see the README's "Option 1 — Homebrew" section.

---

## Prerequisites

- A Linux server (or VM) reachable from the internet — Ubuntu 24.04 LTS
  is assumed in the examples.
- A domain name you control, with DNS pointed at the server.
- Ports 80 and 443 open to the internet (Let's Encrypt needs port 80 to
  complete the HTTP-01 challenge, and your users need 443).
- A MongoDB instance (local or managed). Examples below assume MongoDB
  running on the same host on port 27017.
- Google OAuth credentials and a Gemini API key (see the README sections
  on each).

---

## 1. DNS

Point an `A` (and optionally `AAAA`) record at your server:

```
feedback.example.com.  A    203.0.113.42
feedback.example.com.  AAAA 2001:db8::42
```

Wait for propagation (usually a few minutes). Verify:

```bash
dig +short feedback.example.com
# → 203.0.113.42
```

---

## 2. Install Smart 360

Use whichever install path you prefer; all three end up serving on
`localhost:8080`.

### Option A — Single binary (recommended for this guide)

```bash
# Replace v1.0.0 with the latest release.
ARCH="$(uname -m)"; case "$ARCH" in x86_64) ARCH=amd64;; aarch64) ARCH=arm64;; esac
curl -L -o smart360.tar.gz \
  "https://github.com/mondial7/smart-360/releases/download/v1.0.0/smart360-v1.0.0-linux-${ARCH}.tar.gz"
tar -xzf smart360.tar.gz
sudo install -o root -g root -m 0755 smart360-v1.0.0-linux-${ARCH}/smart360 /usr/local/bin/smart360
sudo install -d -o smart360 -g smart360 -m 0750 /etc/smart360 2>/dev/null || true
```

### Option B — Docker Compose

Use `docker-compose.prod.yml`; see the README "Option 2 — Docker Compose
→ Production" section. Skip the systemd step below and use
`docker compose ... up -d` to start.

---

## 3. Create the service user, env file, and systemd unit

(Skip this section if you're using Docker Compose.)

```bash
# Service user
sudo useradd --system --no-create-home --shell /usr/sbin/nologin smart360

# Config directory & env file
sudo mkdir -p /etc/smart360
sudo curl -L -o /etc/smart360/.env \
  https://raw.githubusercontent.com/mondial7/smart-360/main/.env.example
sudo chown -R smart360:smart360 /etc/smart360
sudo chmod 600 /etc/smart360/.env
sudo $EDITOR /etc/smart360/.env
```

Fill in:

```bash
MONGODB_URI=mongodb://<user>:<password>@localhost:27017
MONGODB_DB=smart360

JWT_SECRET=<run: openssl rand -base64 32>

GOOGLE_CLIENT_ID=<from Google Cloud Console>
GOOGLE_CLIENT_SECRET=<from Google Cloud Console>
GOOGLE_REDIRECT_URL=https://feedback.example.com/api/auth/callback

FRONTEND_URL=https://feedback.example.com

GEMINI_API_KEY=<from Google AI Studio>

# Optional: change the listen port (default 8080)
# PORT=8080
```

Create the systemd unit:

```bash
sudo tee /etc/systemd/system/smart360.service >/dev/null <<'UNIT'
[Unit]
Description=Smart 360 Feedback
After=network-online.target
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

You should see the server listening on `:8080`. Confirm:

```bash
curl -sf http://127.0.0.1:8080/api/me
# 401 (no token) — that's fine; means the server is up.
```

---

## 4. Reverse proxy with HTTPS (Caddy)

Caddy auto-provisions and renews Let's Encrypt certificates. It is the
shortest path from "running on localhost" to "publicly reachable on
HTTPS."

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

        # Security headers
        header {
                Strict-Transport-Security "max-age=63072000; includeSubDomains; preload"
                X-Content-Type-Options "nosniff"
                X-Frame-Options "SAMEORIGIN"
                Referrer-Policy "strict-origin-when-cross-origin"
                Permissions-Policy "interest-cohort=()"
        }

        log {
                output file /var/log/caddy/feedback.access.log
                format json
        }
}
CADDY

sudo systemctl reload caddy
```

Caddy will provision a certificate within seconds. Visit
`https://feedback.example.com` — you should land on the Smart 360 login
page.

---

## 5. Configure the Google OAuth client

You almost certainly used `http://localhost:8080/api/auth/callback`
while developing. In production:

1. Open [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials).
2. Edit your OAuth 2.0 client.
3. Under **Authorized redirect URIs**, add:
   `https://feedback.example.com/api/auth/callback`
4. Under **Authorized JavaScript origins**, add:
   `https://feedback.example.com`
5. Save. Changes propagate within a few minutes.

Restart the backend so it picks up the new `GOOGLE_REDIRECT_URL`:

```bash
sudo systemctl restart smart360
```

Sign in. The first user to authenticate is promoted to **Administrator**.

---

## 6. Alternative — nginx + Certbot

For readers who already run nginx, here is the equivalent:

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

    # Cert paths managed by certbot:
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

Certbot installs a renewal timer automatically.

---

## What this guide does NOT cover

These are tracked separately and recommended for any production install,
but they are out of scope here. The companion [`SECURITY.md`](../SECURITY.md)
spells out the privacy model, the AI data flow, and the full hardening
checklist.

- **Rate limiting & CSRF** — tracked in
  [#26](https://github.com/mondial7/smart-360/issues/26) (rate limiting)
  and [#32](https://github.com/mondial7/smart-360/issues/32) (broader
  security hardening). Until these land, throttle abusive traffic at the
  reverse proxy (`rate_limit` directive in Caddy 2.7+, or `limit_req` in
  nginx).
- **Database backups** — tracked in
  [#35](https://github.com/mondial7/smart-360/issues/35). At minimum,
  schedule a nightly `mongodump` and ship the archive to off-host storage.
- **Observability** — tracked in
  [#34](https://github.com/mondial7/smart-360/issues/34). The process
  logs to stdout (captured by `journalctl -u smart360`); there is no
  Prometheus / OTel integration yet.

---

## Operational notes

### Updating to a new version

Single binary:

```bash
# Download, verify, install
curl -L -o smart360.tar.gz https://github.com/mondial7/smart-360/releases/download/v1.1.0/smart360-v1.1.0-linux-amd64.tar.gz
curl -L -o sums.txt        https://github.com/mondial7/smart-360/releases/download/v1.1.0/smart360-v1.1.0-SHA256SUMS.txt
sha256sum -c sums.txt --ignore-missing
tar -xzf smart360.tar.gz
sudo install -o root -g root -m 0755 smart360-v1.1.0-linux-amd64/smart360 /usr/local/bin/smart360
sudo systemctl restart smart360
```

Docker Compose:

```bash
SMART360_VERSION=v1.1.0 docker compose -f docker-compose.prod.yml pull
SMART360_VERSION=v1.1.0 docker compose -f docker-compose.prod.yml up -d
```

### Log tails

```bash
# Single binary
journalctl -u smart360 -f

# Docker Compose
docker compose -f docker-compose.prod.yml logs -f backend
```

### Where logs live

| Component | Location |
|-----------|----------|
| App (systemd) | `journalctl -u smart360` |
| Caddy access | `/var/log/caddy/feedback.access.log` |
| nginx access | `/var/log/nginx/access.log` |
| MongoDB | `/var/log/mongodb/mongod.log` (default Ubuntu install) |
