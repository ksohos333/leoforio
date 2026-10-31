# <Source name>

- Base URL: <url>
- Role: timetables | arrivals | vehicle-positions | map-tiles
- Ladder step: 1 | 2 | 3 | 4
- Status: unverified | captured | live | stale | blocked
- Captured: <YYYY-MM-DD> by <who>
- Fixtures: `fixtures/<source>/`

## Terms

- Terms or licence: <link>
- Snapshot: `fixtures/<source>/tos-<YYYY-MM-DD>.<ext>`
- Attribution required: <yes, exact text | no>
- Republishing permitted: <yes | no | partial, explain>
- Personal data in responses: <none observed | describe>
- Our obligations: <rate limits, identification, caching, user agent>

## Capabilities

| Data | Provided | Notes |
|---|---|---|
| Lines | | |
| Routes | | |
| Stops | | |
| Timetables | | |
| Arrivals | | |
| Vehicle positions | | |

## Endpoints

### <purpose>

- Method: <GET | POST>
- URL: <full url or pattern>
- Auth: <none | key | cookie>
- Parameters:

| Name | Type | Example | Notes |
|---|---|---|---|
| | | | |

- Request:

```bash
curl '<url>'
```

- Response shape: <describe the JSON, point at the fixture>

```json
{
}
```

- Field mapping:

| Source path | Meaning | leoforio field |
|---|---|---|
| | | |

## Mapping to the internal model

| Internal type | Populated from | Notes |
|---|---|---|
| `Line` | | |
| `Route` | | |
| `Stop` | | |
| `Trip` | | |
| `Arrival` | | |
| `VehiclePosition` | | |

## Politeness

- Cache TTL: <duration>
- Max request rate: <requests per period>
- User agent: `leoforio/0.1 (+https://github.com/ksohos333/leoforio)`
- Conditional requests: <etag / last-modified supported | not supported>
- Timeouts: <duration>

## Quirks

- <encoding, undocumented enums, empty-result semantics, clock skew, rate
  limit headers, anything that surprised us>

## Change log

| Date | Upstream change | Impact |
|---|---|---|
| | | |
