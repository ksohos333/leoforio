# ADR 0002: Go backend with embedded SQLite

- Status: Accepted
- Date: 2026-10-31
- Deciders: @ksohos333

## Context

The app needs a backend that can fetch upstream OASA data, cache it politely,
serve it to mobile clients, and persist a static network model (lines, routes,
stops, timetables). It is operated by one person on a back burner, so
deployment cost and operational surface matter more than raw throughput.

## Decision

- Language: Go (1.26+).
- Storage: SQLite via a pure-Go driver (`modernc.org/sqlite`, no cgo), so the
  service cross-compiles to a single static binary.
- Transport: JSON over HTTPS, versioned under `/api/v1`.
- Caching: in-process TTL cache with request coalescing (`singleflight`), so
  concurrent clients for the same key produce exactly one upstream call.
- Deployment: one binary plus one SQLite file, sized for a free-tier host.

## Consequences

- No separate database to operate, and no cgo in the build.
- SQLite write concurrency is bounded; the static model is written by an
  importer job, not by request handlers.
- The cache is per-process. A second instance multiplies upstream traffic, so
  the deployment runs a single instance until that changes.
