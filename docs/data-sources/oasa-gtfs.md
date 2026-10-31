# OASA GTFS

- Base URL:
  `https://data.gov.gr/dataset/fb049bb1-aea6-4443-95fa-8b941dd6a057/resource/119db488-16ea-4c76-b560-41c472872390/download/osy_gtfs.zip`
- Role: static network model, stop coordinates, route geometry, timetables
- Ladder step: 1 (official open data portal, published by OASA)
- Status: captured, **service calendar expired**
- Captured: 2026-10-31 by @ksohos333
- Dataset page:
  `https://data.gov.gr/dataset/fb049bb1-aea6-4443-95fa-8b941dd6a057`

## Terms

- Published by OASA on the Greek national open data portal (data.gov.gr).
  Reuse rests on the Greek public-sector open data framework (law 4727/2020),
  which is the basis other projects in this space rely on.
- The dataset page carries **no explicit Creative Commons licence
  statement**. Its only licensing language is that the data are provided
  *as is* and that the Athens transport services bear no responsibility for
  the results of any third-party use.
- Snapshot: `fixtures/oasa/datagovgr-dataset-2026-10-31.html`
- Community projects sometimes describe this feed as CC BY-NC 4.0; that
  description was not found on the dataset page and is not relied on here.
- Attribution required: the agency is `OASA / Athens Urban Transport
  Organization`, timezone `Europe/Athens`, and the feed must be credited to
  OASA.
- Personal data: none.
- The same personal-use restriction that applies to the Telematics service
  applies to OASA's informational material, so this feed is used for personal,
  non-commercial purposes as well.

## Contents

30 MB zip, eight files:

| File | Rows | Purpose |
|---|---|---|
| `agency.txt` | 1 | the operator |
| `stops.txt` | 8 041 | stops with coordinates |
| `routes.txt` | 486 | routes with short name, long name, colour |
| `trips.txt` | 97 218 | trips per route and service day |
| `stop_times.txt` | 4 237 955 | scheduled times |
| `shapes.txt` | 65 710 | route geometry points |
| `calendar.txt` | 7 | service days per weekday |
| `calendar_dates.txt` | 2 | exceptions |

## Findings

**The service calendar is expired.** `calendar.txt` covers
`20260706`–`20261006`. At capture time (2026-10-31) the feed's timetable
validity ended 25 days earlier, so its `stop_times.txt` describes a period
that is already in the past. The feed is still useful for the parts that do
not expire quickly: stop coordinates, route names and colours, and shape
geometry.

**Timetables therefore come from the Telematics API**, not from this feed:
`getScheduleDaysMasterline` and `getSchedLines` give the current schedule for
a line. The GTFS remains the source for the static network model.

**Stop identifiers do not match the Telematics API.** The GTFS uses identifiers
such as `100001`, while the Telematics API uses `StopCode` values such as
`10175` and `StopID` values such as `18` or `060134`. These are separate
numbering systems, so a join table is required before arrivals can be
attributed to a GTFS stop. The `stop_url` column in `stops.txt` points at
`http://telematics.oasa.gr/#stationInfo_<stop_id>`, which suggests the GTFS
stop identifier is the value the Telematics web application expects, and that
the `StopCode` seen in the API is a different, older code.

**No realtime counterpart.** There is no GTFS-RT feed. Realtime data exists
only through the Telematics API, see
[oasa-telematics.md](./oasa-telematics.md).

## Politeness

- The feed is a 30 MB download, so it is fetched by an importer job, not per
  request, and cached on disk.
- Refresh: daily check of the download timestamp, imported only when the
  content changed.

## Quirks

- File timestamps inside the archive do not match the calendar validity
  period, so freshness has to be judged from `calendar.txt`.
- `stop_times.txt` is 225 MB uncompressed, which is why the import is a
  streaming job rather than a load into memory.
- `routes.txt` has no `route_id` to line mapping table; the relationship to
  the Telematics `LineCode` has to be derived, because the line code that a
  vehicle runs under live is not always the route identifier its trips use in
  the GTFS.

## Change log

| Date | Upstream change | Impact |
|---|---|---|
| 2026-10-31 | Initial capture; service calendar expired 2026-10-06 | Timetables must come from the Telematics API |
