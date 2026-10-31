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
| OASA GTFS (data.gov.gr) | static model, geometry | 1 | captured, calendar expired | [oasa-gtfs.md](./oasa-gtfs.md) |
| OASA Telematics API | arrivals, positions, schedules | 2 | captured | [oasa-telematics.md](./oasa-telematics.md) |
| OASA web app (observed) | arrival polling behaviour | 3 | not needed | [oasa-web-observed.md](./oasa-web-observed.md) |
| OASA official app (static artefacts) | positions | 3 | not needed | [oasa-app-observed.md](./oasa-app-observed.md) |
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
| 2026-10-31 | Live positions come from the public `getBusLocation` endpoint, so the research phase is not needed | [ADR 0007](../adr/0007-positions-gate.md), [oasa-telematics.md](./oasa-telematics.md) |
| 2026-10-31 | Timetables come from `getSchedLines`, because the GTFS service calendar expired on 2026-10-06 | [oasa-gtfs.md](./oasa-gtfs.md) |
| 2026-10-31 | The project stays non-commercial: OASA's terms allow personal use only | [oasa-telematics.md](./oasa-telematics.md) |
| 2026-10-31 | Written permission is requested from OASA, in parallel with development | [permission-request.md](../permission-request.md) |
| 2026-10-31 | The backend is deployed on a Greek IP address, because the upstream is only reachable from Greece | [ADR 0008](../adr/0008-greek-ip-deployment.md) |
| _pending_ | _e.g. arrivals adapter targets the Telematics API_ | _section in oasa-telematics.md_ |

## Conventions

- One document per source, created from `_TEMPLATE.md`.
- Fixtures live in `fixtures/<source>/` and are captured in the same session
  as the document, so documentation and fixtures never disagree.
- Fixtures are the input to adapter contract tests: a fixture that changes
  without the source actually changing upstream is a test failure, not a
  silent doc drift.
- Terms of service are captured as a snapshot with the capture date; links
  alone are not sufficient evidence.
