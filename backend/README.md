# leoforio backend

Go service that fetches upstream transit data, caches it politely and serves
it to the mobile app. The app never talks to an upstream directly, see
[ADR 0004](../../docs/adr/0004-app-never-calls-upstream.md).

## Run

```bash
go run ./cmd/server
```

Configuration is by environment variable:

| Variable | Default | Meaning |
|---|---|---|
| `LEOFORIO_ADDR` | `:8080` | listen address |
| `LEOFORIO_VERSION` | `dev` | version reported by `/healthz` and `/api/v1/meta` |

## Test

```bash
go vet ./...
go test -race ./...
```

## Layout

| Path | Purpose |
|---|---|
| `cmd/server` | process entry point, graceful shutdown |
| `internal/model` | normalised network types: line, route, stop, trip, arrival, vehicle position |
| `internal/source` | interfaces every upstream source implements |
| `internal/cache` | TTL cache that coalesces concurrent loads of the same key |
| `internal/api` | HTTP handlers, versioned under `/api/v1` |
| `internal/store` | SQLite schema for the static network model |

## Endpoints

| Endpoint | Purpose |
|---|---|
| `GET /healthz` | liveness and version |
| `GET /api/v1/meta` | API name and version |
| `GET /api/v1/stops/{stopID}/arrivals` | expected arrivals for a stop |
| `GET /api/v1/vehicles` | last known vehicle positions |

Endpoints that need an upstream source return `503` until a source is wired
in, which is the current state: upstream discovery is tracked in
[`docs/data-sources/`](../../docs/data-sources/README.md).
