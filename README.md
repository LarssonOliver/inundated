# 🌊 inundated

> A personal time and task management system.

<p align="center">
  <img alt="GitHub License" src="https://img.shields.io/github/license/LarssonOliver/inundated">
  <a href="https://woodpecker.larssonoliver.com/repos/3" target="_blank">
    <img src="https://woodpecker.larssonoliver.com/api/badges/3/status.svg" alt="status-badge" />
  </a>
  <img alt="GitHub Tag" src="https://img.shields.io/github/v/tag/LarssonOliver/inundated">
  <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/LarssonOliver/inundated">
</p>

**inundated** is a lightweight, self-hosted app that helps me track time across 
projects and tasks. It is developed by me, for me. 

---

## 🛠 Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go |
| Frontend | Vue 3 + Pinia + TypeScript + Vite |
| Database | PostgreSQL (or in-memory for dev) |
| API spec | OpenAPI 3 |

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/) 
- [Node.js](https://nodejs.org/) 
- [PostgreSQL](https://www.postgresql.org/) (or use the built-in in-memory store for quick local testing)

### Running in development

1. **Clone the repo**

   ```bash
   git clone https://github.com/LarssonOliver/inundated.git
   cd inundated
   ```

2. **Start the backend** (defaults to the in-memory store — no database required for local tinkering)

   ```bash
   make dev
   ```

3. **Start the frontend** (in a separate terminal)

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   The frontend dev server will open at `http://localhost:5173` and proxy API requests to the backend.

### Connecting to PostgreSQL

Set the `DATABASE_URL` environment variable before starting the server:

```bash
export DATABASE_URL="postgresql://user:password@localhost/inundated"
make dev
```

### Enabling authentication (OIDC)

By default inundated runs in **userless mode**: no login, and every resource is
shared. To put it behind an OpenID Connect provider, set the issuer URL, client
credentials, and the app's public origin:

```bash
export OIDC_ISSUER_URL="https://accounts.example.com"
export OIDC_CLIENT_ID="inundated"
export OIDC_CLIENT_SECRET="…"
export PUBLIC_BASE_URL="https://inundated.example.com"
# optional:
export OIDC_SCOPES="openid,profile,email"   # default
export OIDC_HTTP_TIMEOUT="10s"              # default
```

The redirect URI is derived as `$PUBLIC_BASE_URL/api/auth/callback` — register
exactly that with your provider.

`PUBLIC_BASE_URL` also controls cookie security: when it is an `https://` origin
the session cookie is marked `Secure`. Set it to your real `https://` origin in
any deployment reachable over HTTPS, including userless ones — otherwise the
browser accepts the cookie over plain HTTP too.

The first user to log in adopts all pre-existing resources. Once any user
exists the server refuses to start without OIDC configured, so authentication
can't be silently switched off.

#### Closing registration

By default any identity your OIDC provider authenticates gets an account on
first login. Once every expected user has logged in at least once, set:

```bash
export DISABLE_USER_REGISTRATION=true   # or --disable-user-registration
```

With this on, a login from an identity that has no account is rejected (the
account is not created and no resources are adopted); existing users are
unaffected. Leave it off for the very first login so you can enrol yourself.

### Deploying behind a reverse proxy

inundated does not terminate TLS or emit HSTS itself. Run it behind a proxy
(Caddy, nginx, Traefik, …) that:

- terminates HTTPS and sets `Strict-Transport-Security`;
- sets `X-Forwarded-For` to the real client address **and strips any
  client-supplied copies** — the built-in per-IP rate limits (100 req/min
  general, 10 req/min on `/api/auth/*`) key off that value;
- forwards to the app's `HOST:PORT`.

**Client-IP resolution.** Set `TRUSTED_PROXIES` to the proxy's address(es) — a
comma-separated list of CIDRs or bare IPs (e.g. `10.0.0.0/8,192.168.1.10`).
inundated reads the forwarded header only when the request's direct TCP peer is
in that list; otherwise it rate-limits by the raw connection address, so a
client reachable directly (or over a pod network) can't spoof its IP with a
forged header. Leave `TRUSTED_PROXIES` unset and the header is ignored
entirely — **including on an existing deployment, where every request then
shares one rate-limit bucket until you set it.**

By default only `X-Forwarded-For` is trusted (walked right-to-left, skipping
trusted hops). If your proxy publishes the client address under a different
header, set `TRUSTED_PROXY_HEADERS` (priority-ordered; e.g.
`X-Real-IP,X-Forwarded-For`). Only `X-Forwarded-For`, `X-Real-IP` and
`True-Client-IP` are accepted; whichever you list, the proxy must set it and
strip inbound copies.

Also set `PUBLIC_BASE_URL` to the external `https://` origin: it is registered
as a trusted origin so the cross-origin request check keeps working for older
browsers when the proxy rewrites the `Host` header. API request bodies are
capped at 1 MiB.

### Trying the auth flow locally with Docker Compose

`docker-compose.yml` runs PostgreSQL and a mock OIDC provider
([mock-oauth2-server](https://github.com/navikt/mock-oauth2-server)). Run the app
itself on the host so it shares `localhost` with your browser and the issuer URL.
`make dev-auth` starts the backend with the matching flags:

```bash
docker compose up -d

# terminal 1 — backend
make dev-auth

# terminal 2 — frontend
cd frontend && npm run dev
```

Open <http://localhost:5173> and follow a login link. The mock provider shows a
form where you enter any username and, optionally, a `claims` JSON blob such as
`{"email":"alice@example.com","name":"Alice"}` to mint the ID token. The first
account to log in adopts the existing resources.

```bash
docker compose down -v   # stop and wipe the database
```

---

## 🏗 Building for Production

Build a single self-contained binary (frontend assets embedded):

```bash
make build
./bin/inundated
```

Or build and push a multi-arch Docker image using buildx:

```bash
make image-push
```

---

## 🧪 Testing

```bash
# Run all tests (backend + frontend)
make test

# Backend only
make test-backend

# Frontend only
make test-frontend
```

---

## 🔧 Code Generation

The REST API is driven by the OpenAPI spec in `openapi/inundated.yaml`. After editing the spec, regenerate the server stubs and TypeScript client:

```bash
make generate
```

---

## 📄 License

MIT © [Oliver Larsson](https://github.com/LarssonOliver) — see [LICENSE](LICENSE) for details.
