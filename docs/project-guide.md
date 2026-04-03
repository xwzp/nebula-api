# Nebula API — Project Guide

## Quick Reference

```
mise run help            # Show this quick reference
mise run help:detail     # Show full project guide
mise run status          # Git branch, worktrees, recent commits
```

### Development

| Command | Description |
|---------|-------------|
| `mise run dev-up` | Build and start all services (Docker) |
| `mise run dev-rebuild` | Rebuild nebula-api only (keeps DB/Redis) |
| `mise run dev-logs` | Follow nebula-api container logs |
| `mise run dev-down` | Stop all services |
| `mise run dev-reset` | Stop all + delete volumes (DB data lost) |
| `mise run run` | Run Go backend locally (port 3000) |
| `mise run build` | Build Go binary |
| `mise run fe-dev` | Start frontend dev server (port 5173) |
| `mise run fe-build` | Build frontend for production |
| `mise run fe-lint` | Lint frontend code |

### Testing

| Command | Description |
|---------|-------------|
| `mise run test` | Run all Go tests |
| `mise run test:controller` | Run controller tests (subscription, topup, token) |
| `mise run e2e` | Run Playwright E2E tests (headless) |
| `mise run e2e:headed` | Run E2E tests with visible browser |
| `mise run e2e:codegen` | Open Playwright codegen to record test specs |

### Production

| Command | Description |
|---------|-------------|
| `mise run prod-up` | Build and start production |
| `mise run prod-rebuild` | Rebuild nebula-api in production |
| `mise run prod-logs` | Follow production logs |
| `mise run prod-status` | Show production container status |
| `mise run prod-pull` | Git pull + rebuild production |
| `mise run prod-backup` | Backup PostgreSQL to timestamped SQL file |
| `mise run prod-restore -- path/to/backup.sql` | Restore from backup |

### Other

| Command | Description |
|---------|-------------|
| `mise run docs-dev` | Start docs dev server |
| `mise run docs-build` | Build docs for production |

---

## Architecture

Layered architecture: Router -> Middleware -> Controller -> Service -> Model

```
router/        — HTTP routing (API, relay, dashboard, web)
controller/    — Request handlers
service/       — Business logic (billing, quota, error handling)
model/         — Data models and DB access (GORM)
relay/         — AI API relay/proxy core
  relay/channel/   — Provider adapters (openai/, claude/, gemini/, aws/, etc.)
  relay/common/    — RelayInfo, shared relay types
  relay/constant/  — Relay modes (chat, embedding, image, audio, rerank, responses)
middleware/    — Auth, rate limiting, CORS, logging, distribution
setting/       — Configuration subsystems (ratio, model, operation, system, performance)
common/        — Shared utilities (JSON, crypto, Redis, env, rate-limit, etc.)
dto/           — Data transfer objects (request/response structs)
constant/      — Constants (API types, channel types, context keys)
types/         — Type definitions (relay formats, file sources, errors)
i18n/          — Backend internationalization (go-i18n, en/zh)
oauth/         — OAuth provider implementations
pkg/           — Internal packages (cachex, ionet)
web/           — React frontend
```

## Tech Stack

- **Backend**: Go 1.22+, Gin web framework, GORM v2 ORM
- **Frontend**: React 18, Vite, Semi Design UI (@douyinfe/semi-ui)
- **Databases**: SQLite, MySQL, PostgreSQL (all three must be supported)
- **Cache**: Redis (go-redis) + in-memory LRU (pkg/cachex HybridCache)
- **Auth**: Session cookies + access token fallback, WebAuthn/Passkeys, OAuth
- **Frontend package manager**: Bun
- **E2E Testing**: Playwright (Chromium, headless, locale zh-CN)
- **Backend Testing**: Go standard testing + httptest + SQLite in-memory

## Relay System — Request Flow

The relay system proxies client requests to 40+ upstream AI providers. Each provider implements the `Adaptor` interface (`relay/channel/adapter.go`):

```
Client Request
  -> Router (relay-router.go)
  -> Middleware: TokenAuth -> Distribute (selects channel) -> RateLimit
  -> Controller.Relay(c, relayFormat)
  -> Helper (TextHelper / ImageHelper / AudioHelper / etc.)
      -> GetAdaptor(apiType) — factory returns provider-specific adapter
      -> adaptor.ConvertOpenAIRequest() — convert to provider format
      -> service.PreConsumeBilling() — pre-charge quota
      -> adaptor.DoRequest() — HTTP call to upstream
      -> adaptor.DoResponse() — parse response, extract usage
      -> service.SettleBilling() — adjust quota based on actual usage
  -> Response to client
```

**Key interfaces** (`relay/channel/adapter.go`):
- `Adaptor` — synchronous requests (chat, embedding, image, audio, rerank)
- `TaskAdaptor` — async tasks (video/music generation) with polling and three-phase billing

**Channel distribution** (`middleware/distributor.go`): Selects the best channel based on model availability, channel priority, user group, and affinity.

## Billing System — Two-Phase Pattern

1. **Pre-consume** (`service.PreConsumeBilling`): Estimate quota cost and deduct upfront, creating a `BillingSession`
2. **Settle** (`service.SettleBilling`): After response, calculate actual cost from token usage. Charge delta or refund excess.

Sources: `BillingSourceWallet` (user quota) or `BillingSourceSubscription` (recurring plan).

## Subscription & Topup Tier System

### Subscription Plan Groups

Subscriptions use a **plan group** model:
- `subscription_plan_group` — display-level entity (title, subtitle, tag, structured feature list)
- `subscription_plan` — billing variant under a group (monthly, yearly, etc.)

The month/year toggle on the frontend switches between variants within the same group.

**Admin APIs** (require admin auth):
- `GET/POST /api/subscription/admin/groups` — list/create groups
- `PUT/PATCH/DELETE /api/subscription/admin/groups/:id` — update/status/delete group
- `POST /api/subscription/admin/groups/:id/plans` — create plan under group
- `PUT/PATCH /api/subscription/admin/plans/:id` — update/status plan

**Public API** (no auth):
- `GET /api/subscription/public-plans` — returns groups with nested plan variants

### Topup Tiers

Pay-as-you-go tiers stored in `topup_tier` table (dedicated DB table, replaces old JSON settings):
- Each tier has: title, subtitle, tag, amount, discount, bonus_quota, structured features

**Admin APIs**:
- `GET/POST /api/topup/admin/tiers` — list/create
- `PUT/PATCH/DELETE /api/topup/admin/tiers/:id` — update/status/delete

**Public API**:
- `GET /api/topup/tiers` — returns enabled tiers

### Structured Feature List

Both subscription groups and topup tiers support a JSON feature list:
```json
[{"text": "Feature description", "icon": "check|x|info", "style": "default|highlight|disabled"}]
```

Stored as TEXT column (not JSONB) for cross-DB compatibility.

## Settings System

Runtime settings stored in the `options` DB table, loaded at startup and synced periodically (`SyncOptions`). Organized into subsystems under `setting/`: ratio (pricing), model, operation, system, performance, console, reasoning.

## Context Keys

Gin context (`*gin.Context`) carries request-scoped state through the middleware/handler chain. Key constants are in `constant/context_key.go`. Access via `common.SetContextKey()` / `common.GetContextKey()`.

## Internationalization (i18n)

### Backend (`i18n/`)
- Library: `nicksnyder/go-i18n/v2`
- Languages: en, zh-CN, zh-TW
- Message keys defined in `i18n/keys.go`
- Usage: `i18n.T(c, i18n.MsgSomeKey, map[string]any{"Field": value})`

### Frontend (`web/src/i18n/`)
- Library: `i18next` + `react-i18next` + `i18next-browser-languagedetector`
- Languages: zh (fallback), en, fr, ru, ja, vi
- Translation files: `web/src/i18n/locales/{lang}.json` — flat JSON, keys are Chinese source strings

## Testing

### Go Backend Tests

Located alongside source files (e.g., `controller/subscription_test.go`).

Pattern:
- SQLite in-memory DB per test (`file:TestName?mode=memory&cache=shared`)
- `gin.TestMode` + `common.RedisEnabled = false`
- `httptest.NewRecorder()` + `gin.CreateTestContext()`
- Seed data with helper functions, call handler directly, decode response

### Playwright E2E Tests

Located in `web/e2e/`. Config: `web/playwright.config.js`.

- Dedicated Go server on port 3099 with `e2e-test.db`
- `web/e2e/auth.js` — handles setup wizard + admin login
- Locale: `zh-CN` (headless Chromium defaults to en-US)
- Use `npx playwright codegen http://localhost:3099 --locale zh-CN` to record specs

## Key Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Server listen port |
| `SQL_DSN` | (none) | MySQL/PostgreSQL connection string |
| `SQLITE_PATH` | `./sqlite.db` | SQLite file path (used when no SQL_DSN) |
| `REDIS_CONN_STRING` | (none) | Redis connection string |
| `GIN_MODE` | `debug` | Set to `release` for production |
| `MEMORY_CACHE_ENABLED` | (false) | Enable in-memory cache |
| `SESSION_SECRET` | (random) | Required for multi-node deployments |

See `.env.example` for the full list.
