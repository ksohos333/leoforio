# OASA Telematics API

- Base URL: `https://telematics.oasa.gr/api/`
- Role: live arrivals, live vehicle positions, stop and route lookups, schedules
- Ladder step: 2 (community-documented public endpoint, operated by OASA)
- Status: captured
- Captured: 2026-10-31 by @ksohos333
- Fixtures: `fixtures/oasa/`
- Documentation: community docs at
  [oasa-telematics-api.readthedocs.io](https://oasa-telematics-api.readthedocs.io/en/latest/)
  (source: `giannis-papaioannou/OASA-Telematics-API-Documentation`). This is not
  official OASA documentation, so every endpoint below was verified against the
  live service.

## Terms

- Terms or licence: `https://telematics.oasa.gr/terms.php`
- Snapshot: `fixtures/oasa/terms-2026-10-31.html`
- The terms state that the rights to the informational material (maps, diagrams,
  lines, routes) belong exclusively to OASA and that the material is made
  available to the public **for personal use only**; reproduction or use for
  **commercial purposes** without OASA's prior written permission is criminally
  prosecuted.
- leoforio is free, open source, non-commercial and carries no advertising or
  tracking, which is consistent with personal use. Anyone forking it for
  commercial use must obtain permission from OASA first. This constrains the
  project: it stays non-commercial, and it is the reason the data is proxied
  rather than redistributed.
- Personal data in responses: none observed. Vehicle identifiers (`veh_code`,
  `VEH_NO`) are equipment numbers, not personal data, and are not persisted.

## Request shape

- Method: `POST` works for every endpoint; `GET` also works. The official web
  application uses `POST` with an empty body.
- Parameters are **positional**: `p1`, `p2`, `p3`. Named parameters are
  ignored and the service answers `{"error":"nothing requested"}`.
- An unknown action gives the same error, and a known action without its
  required parameters answers `null`.
- Auth: none. No API key, no token, no signature.

```bash
curl -X POST 'https://telematics.oasa.gr/api/?act=getStopArrivals&p1=10175'
```

## Endpoints

| Action | Parameters | Returns | Fixture |
|---|---|---|---|
| `webGetLangs` | – | UI language labels | – |
| `webGetLines` | – | all lines: `LineCode`, `LineID`, `LineDescr`, `LineDescrEng` (471 entries) | `webGetLines.json` |
| `webGetLinesWithMLInfo` | – | lines plus `ml_code`, `sdc_code`, `mld_master` | `webGetLinesWithMLInfo.json` |
| `webGetRoutes` | `p1=LineCode` | routes of a line: `RouteCode`, `RouteDescr`, `RouteType`, `RouteDistance` | `webGetRoutes.json` |
| `webGetStops` | `p1=RouteCode` | stops of a route with coordinates and order | `webGetStops.json` |
| `webRoutesForStop` | `p1=StopCode` | routes serving a stop, with the line number riders see | `webRoutesForStop.json` |
| `getStopNameAndXY` | `p1=StopCode` | stop name, coordinates, heading | `getStopNameAndXY.json` |
| `getStopArrivals` | `p1=StopCode` | **live arrivals** for a stop | `getStopArrivals.json` |
| `getBusLocation` | `p1=RouteCode` | **live vehicle positions** for a route | `getBusLocation.json` |
| `getClosestStops` | `p1=lat`, `p2=lng` | nearest stops with distance | `getClosestStops.json` |
| `webGetRoutesDetailsAndStops` | `p1=RouteCode` | route geometry plus stops | `webGetRoutesDetailsAndStops.json` |
| `getSchedLines` | `p1=mlCode`, `p2=sdcCode`, `p3=lineCode` | timetable entries (`come`, `go`) | `getSchedLines.json` |
| `getScheduleDaysMasterline` | `p1=lineCode` | `sdc_code` identifiers per line | `getScheduleDaysMasterline.json` |

### Arrivals

```json
[
  { "route_code": "2484", "veh_code": "12237", "btime2": "11" }
]
```

`btime2` is the number of minutes until the vehicle arrives. The service also
returns vehicle code and route code, but no absolute arrival time and no
scheduled time, so the arrival is reconstructed as `now + btime2`.

### Vehicle positions

```json
[
  {
    "VEH_NO": "12237",
    "CS_DATE": "Oct 31 2026 10:19:23:000PM",
    "CS_LAT": "37.988269",
    "CS_LNG": "23.746624",
    "ROUTE_CODE": "2484",
    "VEH_HEADING": "247"
  }
]
```

Positions are exposed **per route**, not fleet-wide: showing every vehicle in
Athens means one request per route code, which is 486 requests for the whole
network. The app therefore queries the routes it is actually showing, and the
backend caches per route.

### Route geometry

`webGetRoutesDetailsAndStops` returns the ordered path of the route as
`routed_x` / `routed_y` (longitude, latitude) pairs plus `routed_order`, which
is what the map draws.

## Mapping to the internal model

| Internal type | Source fields | Notes |
|---|---|---|
| `Line` | `LineCode`, `LineID`, `LineDescr`, `LineDescrEng` | `LineCode` is the id used by every other endpoint |
| `Route` | `RouteCode`, `LineCode`, `RouteDescr`, `RouteDistance` | `RouteCode` is the id used by stops, geometry and positions |
| `Stop` | `StopCode`, `StopDescr`, `StopDescrEng`, `StopLat`, `StopLng`, `StopHeading`, `RouteStopOrder` | `StopCode` is the id used by arrivals |
| `Arrival` | `route_code` → `RouteID`, `veh_code` → `VehicleID`, `btime2` → `Expected` | `Expected = now + btime2 minutes` |
| `VehiclePosition` | `VEH_NO` → `VehicleID`, `ROUTE_CODE` → `RouteID`, `CS_LAT`/`CS_LNG` → `Position`, `CS_DATE` → `RecordedAt`, `VEH_HEADING` | heading is kept as an extra field |

## Politeness

- Cache TTL: arrivals 25 s, positions 15 s per route, static data 24 h.
- The backend coalesces concurrent requests for the same key into a single
  upstream call, so upstream traffic tracks distinct stops and routes rather
  than user count.
- User agent: `leoforio/0.1 (+https://github.com/ksohos333/leoforio)`.
- No rate limit is published. The official web application polls
  `getStopArrivals` for one stop every 30 s, which is the reference for how
  fast a single stop may be polled.
- Positions must be requested per route. Fetching all 486 routes for one map
  view is not acceptable; only the routes the user is looking at are fetched.

## Quirks

- Positional parameters only. `stopID=10175` is silently ignored and the
  service answers `{"error":"nothing requested"}`.
- An error may arrive where a list was expected: `{"error":"nothing requested"}`
  is a valid answer to a request with an unknown action or missing
  parameters. The client detects that shape and reports it as an upstream
  error rather than a decoding failure.
- **Field types drift without notice.** The service is undocumented and has
  changed identifiers such as `route_code`, `btime2` and `StopCode` from
  strings to numbers and back. Every field is therefore read leniently, as a
  string or a number, and identifiers are never coerced to numbers: `StopID`
  keeps a significant leading zero (`060134`) where `StopCode` is `60134`.
- `CS_DATE` uses the format `Oct 31 2026 10:19:23:000PM`: a millisecond field
  glued to the AM/PM marker, and no timezone. Values are Europe/Athens local
  time and need a custom parser rather than a standard layout. Athens is on
  UTC+2 in winter and UTC+3 in summer.
- `btime2` is a string, and the count is minutes, not seconds.
- Real-time information is temporarily unavailable on some lines; the official
  site carries a notice about it. Empty arrays are a normal response, not an
  error.
- **A query with no results can answer with an empty string** rather than an
  empty array: `getBusLocation` for a route with no vehicles returns `""`. The
  client reads that as an empty result, not as a failure.
- A single digit day is written with two spaces, as in `Nov  1 2026`, so the
  timestamp parser has to accept both `Nov 1 2026` and `Nov  1 2026`.
- There is no GTFS-RT equivalent in this API: realtime data exists only as
  arrivals per stop and positions per route.
- Stop identifiers differ between this API and the official GTFS feed, see
  [oasa-gtfs.md](./oasa-gtfs.md).
- The service is reachable over HTTPS only; plain HTTP requests time out.

## Operational notes

These are the failure modes that actually bite, and what leoforio does about
each one.

- **The service is only reachable from Greece.** Requests from outside the
  country are dropped without an answer. The backend therefore has to run on
  a Greek IP address, which is a hosting constraint rather than a coding
  one, see [ADR 0008](../adr/0008-greek-ip-deployment.md).
- **The operator rate limits.** A `403` or `429` answer is treated as a
  rejection: the client stops talking to the upstream for ten minutes instead
  of retrying into a longer ban, and the cache keeps serving the last known
  values meanwhile. Other failures are retried twice with a doubling delay.
- **The whole service can disappear.** In late August 2026 the telematics
  system was taken down for several days by a cyberattack, taking the
  official app, the website and the station displays with it. The app must
  degrade to "no live data" rather than fail, and the cache holds the last
  known state for that reason.
- **The DNS answer is worth watching.** `telematics.oasa.gr` currently
  resolves to `195.46.22.91`, which works from Greece; community reports
  describe that address dropping packets at times, and the operator's
  addresses sit in a small range. A future maintainer who sees unexplained
  timeouts should try the other addresses in that range, keeping the
  hostname for TLS.
- **There is no published changelog and no schema.** A CI job that runs the
  live test against the upstream and fails when the shape changes is the
  only reliable way to notice a schema shift. The contract tests in
  `backend/internal/source/oasa` run against the captured fixtures and fail
  the moment the adapter stops matching them.

## Change log

| Date | Upstream change | Impact |
|---|---|---|
| 2026-10-31 | Initial capture of all thirteen endpoints | — |
