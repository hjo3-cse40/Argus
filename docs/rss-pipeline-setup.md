# RSS Pipeline Setup & Testing Guide

Full end-to-end pipeline: **YouTube → RSSHub → Parse → Normalize → RabbitMQ → Discord**

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed and running
- [Go 1.21+](https://go.dev/dl/) installed
- A Discord server where you can create a webhook

## 1. Create a Discord Webhook

1. Open your Discord server
2. Go to **Server Settings → Integrations → Webhooks**
3. Click **New Webhook**, pick a channel, and click **Copy Webhook URL**

## 2. Create the `.env` File

Create `infra/.env` with the following (this file is gitignored):

```
ENV=dev
PORT=8080

RABBITMQ_URL=amqp://argus:argus@localhost:5672/
API_BASE_URL=http://localhost:8080

DISCORD_WEBHOOK_URL=<paste your webhook URL here>

RSSHUB_BASE_URL=http://localhost:1200
RSSHUB_FEEDS=youtube:youtube/channel/UCWJ2lWNubArHWmf3FIHbfcQ
```

### Finding a YouTube Channel ID

The `RSSHUB_FEEDS` value uses YouTube channel IDs (not names). To find one:
- Go to the channel page on YouTube
- The URL will be `youtube.com/channel/UCxxxxxxx` — that `UCxxxxxxx` part is the ID
- If the URL is `youtube.com/@name`, view the page source and search for `channel_id`

### Adding Multiple Feeds

Comma-separated, in `type:path` format:

```
RSSHUB_FEEDS=youtube:youtube/channel/UCWJ2lWNubArHWmf3FIHbfcQ,youtube:youtube/channel/UC_S45UpAYVuc0fYEcHN9BVQ
```

Supported source types (via RSSHub): `youtube`, `reddit`, `x`, `github`, and
[many more](https://docs.rsshub.app/).

## 3. Start Infrastructure

```bash
cd infra
docker compose up -d
```

This starts three containers:
- **RabbitMQ** (ports 5672, 15672) — message queue
- **Postgres** (port 5432) — database
- **RSSHub** (port 1200) — RSS feed generator

Wait ~15 seconds for RSSHub to be ready, then verify:

```bash
curl http://localhost:1200/youtube/channel/UCWJ2lWNubArHWmf3FIHbfcQ
```

You should see XML with `<item>` entries. If you get a connection error, wait a bit longer.

## 4. Run the Worker

In a terminal:

```bash
cd backend
go run cmd/worker/main.go
```

You should see:

```
Starting worker in dev environment
worker listening on raw_events
```

Leave this running — it consumes events from RabbitMQ and sends them to Discord.

## 5. Run the RSS Poller

In a second terminal:

```bash
cd backend
go run cmd/rss/main.go
```

You should see:

```
Polling 1 feed(s)
[youtube] Fetching: http://localhost:1200/youtube/channel/UCWJ2lWNubArHWmf3FIHbfcQ
[youtube] NBA - YouTube — 30 items
✓ [youtube] NBA - YouTube — <video title>
...
```

The poller fetches the 5 most recent videos, publishes them to RabbitMQ, and then
polls every 5 minutes for new ones.

## 6. Check Discord

Your Discord channel should now have embed messages with:
- Video title (clickable link)
- Platform (youtube)
- Source (channel name)
- Description

## Stopping Everything

```bash
# Ctrl+C in both the worker and poller terminals

# Stop Docker containers
cd infra
docker compose down
```
