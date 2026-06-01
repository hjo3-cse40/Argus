# Argus

Argus monitors RSS feeds from platforms such as YouTube, Reddit, X, and any other Atom or RSS feed, filters content by keywords, and delivers focused notifications to the in-app dashboard and an optional Discord webhook.

**Try it live:** [https://argus.masondrake.dev](https://argus.masondrake.dev) - Deployed on Kubernetes. Accounts are separate from local setup

## Required software

Before you start, install and verify:

| Requirement | Notes |
|-------------|--------|
| **Docker Desktop** | Must be installed and **running** |
| **Docker Compose** | Included with Docker Desktop (`docker compose version`) |
| **Go 1.24+** | Only needed if you run the API/worker locally outside Docker |

For the quick-start path below, **Docker is enough** — you do not need Go installed unless you choose local development.

---

## Quick start (recommended)

### 0. Clone the Repository

```bash
git clone https://github.com/hjo3-cse40/Argus.git
cd Argus
```

### 1. Start the stack

From the project root:

```bash
cd infra
docker compose up --build -d
```

This starts Postgres, RabbitMQ, RSSHub, the API, frontend, worker, and RSS poller.

To check that containers are healthy:

From `infra/` directory:

```bash
docker compose ps
```

### 2. Open the app

| Service | URL |
|---------|-----|
| **Web UI** | http://localhost:3000 |
| **API** | http://localhost:8080 |
| **RabbitMQ UI** | http://localhost:15672 (login: `argus` / `argus`) |

Verify the API:

```bash
curl http://localhost:8080/health
```

Expected: `{"ok":true}`

### 3. Create an account

1. Open http://localhost:3000
2. Click **Register** and create an account
3. Log in — you'll land on the dashboard

Auth uses **session cookies** (not JWT). The browser sends the cookie automatically on later requests.

**Note:** Local accounts are separate from the deployed (Kubernetes) environment. Register again on the production URL if you use both.

### 4. Add a source to monitor

A **platform** is the destination type (YouTube, Reddit, X, or Other) plus optional Discord webhook and filters. A **sub-channel** is the specific feed to watch within a platform.

1. Go to **Platforms**
2. Create a platform (e.g. YouTube, Reddit, X, or Other) and optionally set a Discord webhook URL (see Discord Webhook section)
3. Add a **sub-channel** (subsource):
   - **YouTube:** channel URL or handle (e.g. `https://youtube.com/@handle`)
   - **Reddit:** subreddit name (e.g. `r/golang` or `https://reddit.com/r/golang`)
   - **X:** profile URL or username (e.g. `https://x.com/username` or `username`)
   - **Other:** direct feed URL (e.g. `https://example.com/feed.xml`)
   ```markdown 
   > You can paste a URL for any platform; Argus derives the identifier automatically.
4. Optionally add **keyword filters** (include/exclude) under **Filters**. With no filters, all new items are delivered. Excludes are checked first; multiple keywords can **match any** or **match all**.

The RSS poller checks feeds **every 5 minutes** and only ingests items it hasn't seen before. Allow up to one cycle (5 minutes) after adding a source. The first poll may include a few recent existing posts.

### 5. View notifications

Open **Notifications** (full delivery history) or **Dashboard** (by Platform) in the UI to see delivered items. If you configured a Discord webhook on the platform, matching events are also posted to Discord.

---

## Stop the stack

```bash
cd infra
docker compose down
```

To remove database data as well:

```bash
cd infra
docker compose down -v
```

## Discord webhook (optional)

1. In your Discord server: **Server Settings → Integrations → Webhooks**
2. Click **New Webhook** and copy the URL
3. Paste it when creating or editing a platform on the **Platforms** page

---

## Local development (optional)

Use this if you're changing backend or frontend code and want hot reload instead of rebuilding Docker images.

### Prerequisites

- Go 1.24+
- Node.js 22+ (for the Next.js frontend)
- Docker Desktop running (for Postgres, RabbitMQ, RSSHub)

### 1. Start infrastructure only

If the full stack is already running, stop it first (`cd infra && docker compose down`), then start only the backing services:

```bash
cd infra
docker compose up -d db rabbitmq rsshub
```

### 2. Run the API

```bash
cd backend
go run ./cmd/api
```

API: http://localhost:8080

### 3. Run the worker

In a new terminal:

```bash
cd backend
go run ./cmd/worker
```

Optional — Discord delivery:

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
go run ./cmd/worker
```

### 4. Run the RSS poller

In another terminal:

```bash
cd backend
go run ./cmd/rss
```

### 5. Run the frontend

In another terminal:

```bash
cd frontend
npm install
npm run dev
```

Frontend: http://localhost:3000 (proxies `/api` to the Go API)

---

## Configuration

Configuration is via **environment variables**. Defaults work for local Docker Compose.

Common variables:

| Variable | Default (Compose) | Purpose |
|----------|-------------------|---------|
| `DB_HOST` | `db` / `localhost` | Postgres host |
| `DB_PORT` | `5432` / `5433` | Postgres port (5433 on host) |
| `RABBITMQ_URL` | `amqp://argus:argus@rabbitmq:5672/` | Message queue |
| `RSSHUB_BASE_URL` | `http://rsshub:1200` | RSSHub for some feeds |
| `DISCORD_WEBHOOK_URL` | — | Fallback Discord webhook when platform has none set |

---

## How it works

```
RSS Poller  →  RabbitMQ (raw_events)  →  Worker  →  Discord + in-app notifications
     ↑                                        ↑
  Subsources                             Keyword filters
  (per channel/subreddit)                (per platform)
```

1. **Subsources** define what to watch (a YouTube channel, Reddit subreddit, etc.).
2. The **RSS poller** fetches feeds every 5 minutes and deduplicates by URL.
3. The **worker** applies keyword filters and delivers to Discord and the dashboard.
4. **Platforms** group subsources and hold Discord webhook + filter settings.

---

## Useful commands

**Publish a test event (API must be running):**

```bash
curl -X POST http://localhost:8080/debug/publish
```

**List deliveries (requires login cookie):**

```bash
curl -b cookies.txt http://localhost:8080/api/deliveries
```

**Connect to Postgres:**

```bash
cd infra
docker exec -it $(docker compose ps -q db) psql -U argus -d argus
```

---

## Deployment

Production runs on Kubernetes (k3s + ArgoCD). See `k3s/` for manifests. Deployment details are maintained separately from this onboarding guide.

---

## Project layout

```
Argus/
├── backend/          Go API, worker, RSS poller
├── frontend/         Next.js web UI
├── infra/            Docker Compose for local dev
├── k3s/              Kubernetes manifests
└── static/           Legacy static assets
```
