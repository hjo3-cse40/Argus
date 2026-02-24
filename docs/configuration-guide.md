# Configuration Guide - Collector Team

## Overview

The Argus backend uses **environment variables** as its configuration. The **config file** is a `.env` file: the app loads variables from `.env`, `infra/.env`, or `../infra/.env` (relative to the working directory or the binary) automatically via `godotenv`, so you can put secrets and overrides in a `.env` file instead of exporting them in the shell.

The app also supports environment-based configuration (dev/stage/prod) with a centralized config loading system.

## Environment Variables

The application uses the `ENV` environment variable to determine which environment it's running in. Valid values are:
- `dev` - Development environment (default)
- `stage` - Staging environment
- `prod` - Production environment

### Configuration Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ENV` | Environment name (dev/stage/prod) | `dev` | No |
| `PORT` | API server port | `8080` | No |
| `API_BASE_URL` | Base URL for the API | `http://localhost:8080` | No |
| `RABBITMQ_URL` | Full RabbitMQ connection URL | Auto-built from components | No* |
| `RABBITMQ_HOST` | RabbitMQ host | `localhost` | No* |
| `RABBITMQ_PORT` | RabbitMQ port | `5672` | No* |
| `RABBITMQ_USER` | RabbitMQ username | `argus` | No* |
| `RABBITMQ_PASS` | RabbitMQ password | `argus` | No* |
| `DB_HOST` | Database host | `localhost` | No |
| `DB_PORT` | Database port | `5432` | No |
| `DB_USER` | Database user | `argus` | No |
| `DB_PASSWORD` | Database password | `argus` | No |
| `DB_NAME` | Database name | `argus` | No |
| `DISCORD_WEBHOOK_URL` | Discord webhook URL for notifications (worker) | *(none)* | Yes for worker |

*Either provide `RABBITMQ_URL` or the individual components (HOST, PORT, USER, PASS)

**DISCORD_WEBHOOK_URL** must be a valid Discord webhook URL starting with `https://discord.com/api/webhooks/` when set. The worker fails to start if it is missing or malformed.

## Usage

### Setting Environment Variables

#### Option 1: Export in shell
```bash
export ENV=dev
export PORT=8080
export RABBITMQ_URL=amqp://argus:argus@localhost:5672/
```

#### Option 2: Use a .env file (recommended)

The app **loads a `.env` file automatically** (no need to `source` it). Place a `.env` in the project root, in `backend/`, or in `infra/` with your variables, for example:

```bash
# infra/.env or backend/.env
ENV=dev
PORT=8080
RABBITMQ_URL=amqp://argus:argus@localhost:5672/
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/your-id/your-token
```

Then run the API or worker from a directory where one of those paths exists; config is loaded at startup.

#### Option 3: Inline with command
```bash
ENV=dev PORT=8080 go run ./cmd/api
```

### Example Configurations

#### Development
```bash
ENV=dev PORT=8080 RABBITMQ_URL=amqp://argus:argus@localhost:5672/ go run ./cmd/api
```

#### Staging
```bash
ENV=stage \
  PORT=8080 \
  RABBITMQ_HOST=rabbitmq-staging.example.com \
  RABBITMQ_USER=argus_staging \
  RABBITMQ_PASS=secure_password \
  go run ./cmd/api
```

#### Production
```bash
ENV=prod \
  PORT=8080 \
  RABBITMQ_URL=amqp://user:pass@rabbitmq-prod.example.com:5672/ \
  DB_HOST=db-prod.example.com \
  DB_USER=argus_prod \
  DB_PASSWORD=secure_password \
  go run ./cmd/api
```

## Example Environment Files

See `docs/env.example`, `docs/env.dev.example`, `docs/env.stage.example`, and `docs/env.prod.example` for template configurations.

**Note:** Actual `.env` files are gitignored for security. Copy the example files and customize as needed.

## Code Usage

The configuration is loaded in `main.go`:

```go
import "argus-backend/internal/config"

cfg, err := config.Load()
if err != nil {
    log.Fatalf("failed to load config: %v", err)
}

// Use config values
mqClient, err := mq.Connect(cfg.RabBITMQ.URL)
port := cfg.Port
```

## Validation

The config loader validates:
- `ENV` must be one of: `dev`, `stage`, or `prod`
- `DISCORD_WEBHOOK_URL` (when set) must start with `https://discord.com/api/webhooks/`; otherwise `config.Load()` returns an error
- All other values have sensible defaults

## Best Practices

1. **Never commit secrets** - Use environment variables or secret management
2. **Use different configs per environment** - Don't use production configs in dev
3. **Document required variables** - Update this guide when adding new config
4. **Test config loading** - Verify config works in all environments
