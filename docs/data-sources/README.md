# Data Sources

Every byte of upstream data leoforio serves is accounted for here. A source
that is not documented in this directory is not used.

## Policy (feed-first ladder)

1. Official feed or documented API
2. Community-documented public endpoint
3. Endpoint discovered by observing our own traffic to the public service
   (never by bypassing protections)
4. Feature dropped from the release

Every rung below rung 1 must be justified in the decision log below with a
capture date and evidence.

## Status

| Source | Role | Ladder step | Status | Doc |
|---|---|---|---|---|
| OASA GTFS | timetables | 1 | unverified | [oasa-gtfs.md](./oasa-gtfs.md) |
| OASA Telematics JSON API | arrivals | 2 | unverified | [oasa-telematics.md](./oasa-telematics.md) |
| OASA web app (observed) | arrivals, maybe positions | 3 | not started | [oasa-web-observed.md](./oasa-web-observed.md) |
| OASA official app (static artefacts) | positions | 3 | not started | [oasa-app-observed.md](./oasa-app-observed.md) |
| Map tile provider | map tiles and styles | 1 | unverified | [map-tiles.md](./map-tiles.md) |

Status values: `unverified` (declared, not confirmed), `captured` (documented
with fixtures), `live` (in production use), `stale` (upstream shape changed,
needs re-capture), `blocked` (terms or access prevent use).

## Decision log

| Date | Decision | Evidence |
|---|---|---|
| 2026-10-31 | App never calls OASA directly; all upstream access is behind backend adapters | [ADR 0004](../adr/0004-app-never-calls-upstream.md) |
| 2026-10-31 | Live positions resolved by explicit three-outcome gate | [ADR 0007](../adr/0007-positions-gate.md) |
| 2026-10-31 | Map tiles come from a configured provider with attribution | [ADR 0005](../adr/0005-map-tile-provider.md) |
| _pending_ | _e.g. arrivals adapter targets the Telematics JSON API_ | _section in oasa-telematics.md_ |

## Conventions

- One document per source, created from `_TEMPLATE.md`.
- Fixtures live in `fixtures/<source>/` and are captured in the same session
  as the document, so documentation and fixtures never disagree.
- Fixtures are the input to adapter contract tests: a fixture that changes
  without the source actually changing upstream is a test failure, not a
  silent doc drift.
- Terms of service are captured as a snapshot with the capture date; links
  alone are not sufficient evidence.
