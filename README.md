# hacelo

Personal app to organize the day's tasks and schedule reminders for the days
ahead. Like an alarm, but with a description and follow-up: when it rings you
pick "Ya lo hago" (it asks again in a few minutes whether you did it) or
"Ahora no puedo" (you reschedule manually).

Built for a single user and installable on a phone without any app store
(Add to Home Screen). The UI is in Spanish; the code is in English.

## How it works

1. Create a task with a description, urgency and time.
2. At that time a notification arrives with two buttons: **Ahora no puedo** and
   **Ya lo hago**.
3. **Ya lo hago** schedules a follow-up after N minutes (configurable per task,
   15 by default) with the buttons **Ya lo hago** / **Ya lo hice**.
   - **Ya lo hice** marks the task as done.
   - **Ya lo hago** again repeats the cycle.
4. After 3 rounds without confirmation the task becomes **overdue** and is
   rescheduled for the next day at the same time.
5. **Ahora no puedo** opens the app so you can pick a new time manually.
6. From the app you can always reschedule any task, whatever its state.

Task states: `pending` -> `verifying` -> `done` | `overdue` -> `pending` (next day).

## Architecture

The backend follows a hexagonal (ports and adapters) architecture. Dependencies
only point inwards: adapters depend on the application, the application depends
on the domain, and the domain depends on nothing.

```
backend/
  cmd/server/                 composition root: wires adapters into services
  cmd/genvapid/               one-off VAPID key generator
  internal/
    domain/                   entities and business rules (task state machine)
    application/
      port/                   inbound ports (use cases) and outbound ports
      service/                use case implementations
    adapters/
      inbound/
        httpapi/              REST API + embedded PWA static files
        scheduler/            periodic trigger for the reminder use case
      outbound/
        sqlite/               task and subscription repositories
        webpush/              Web Push (VAPID) notifier
        lognotifier/          logging-only notifier (no VAPID keys configured)
        clock/                system clock
    config/                   environment configuration
web/                          React + Vite + TypeScript PWA
```

Business rules live in `internal/domain/task.go` and are covered by unit tests
that need no database or network (`go test ./...`).

## Stack

- **Backend**: Go (standard `net/http` router, Go 1.22+), SQLite through
  [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO),
  Web Push with [`webpush-go`](https://github.com/SherClockHolmes/webpush-go).
- **Frontend**: React + Vite + TypeScript, installable PWA with a hand-written
  service worker to control the notification action buttons.
- **Deploy**: a single Go binary with the frontend embedded (`//go:embed`), in a
  Docker container on [Fly.io](https://fly.io).

## Local development

Requires Go 1.22+ and Node 20+.

**Backend** (port 8080):

```bash
cd backend
go mod tidy   # first time only, generates go.sum
go run ./cmd/server
```

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | server port |
| `DB_PATH` | `hacelo.db` | SQLite file path |
| `API_TOKEN` | (empty = no auth) | token required by the API as `Authorization: Bearer` |
| `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` | (empty = log only) | Web Push keys |
| `VAPID_SUBSCRIBER` | `mailto:me@example.com` | contact required by the VAPID standard |

Generate the VAPID keys once:

```bash
cd backend
go run ./cmd/genvapid
```

**Frontend** (port 5173, proxying `/api` to the backend):

```bash
cd web
npm install
npm run dev
```

Installing the PWA and receiving push notifications requires HTTPS (see Deploy).

## Deploy (Fly.io)

```bash
fly launch --no-deploy   # once, if the app does not exist yet
fly volumes create hacelo_data --size 1
fly secrets set \
  API_TOKEN=$(openssl rand -hex 32) \
  VAPID_PUBLIC_KEY=... \
  VAPID_PRIVATE_KEY=... \
  VAPID_SUBSCRIBER=mailto:you@yourdomain.com
fly deploy
```

`.github/workflows/deploy.yml` is a manual workflow that runs `fly deploy` once
the `FLY_API_TOKEN` repository secret exists.

Fly.io has no permanent free tier: one small always-on machine (needed so the
scheduler can check reminders) costs roughly 2-5 USD per month.

## Install on the phone

1. Open the deployed URL in Chrome (Android).
2. Menu -> **Add to Home screen**.
3. Open **Ajustes** inside the app, paste the `API_TOKEN` and tap
   **Activar notificaciones**.

## Roadmap

- Event history per task.
- Edit and delete tasks from the list (the API already supports it).
- Integration tests for the SQLite adapter and HTTP handlers.
