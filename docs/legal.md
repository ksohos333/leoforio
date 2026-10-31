# Licensing and compliance

This document records what leoforio is allowed to do with the data it uses,
what it deliberately does not do, and which statements are verified against a
primary source rather than repeated from community lore.

## Summary

| Source | Licence status | How leoforio uses it |
|---|---|---|
| OASA GTFS on data.gov.gr | Published by OASA, a public body, on the national open data portal. The dataset page carries **no explicit Creative Commons licence statement**; it says the data is provided *as is* and that the transport services bear no responsibility for third-party use. Reuse rests on the Greek public-sector open data framework (law 4727/2020). | Imported into the backend as the static network model: stops, routes, geometry. Attribution given. Refreshed by an importer job. |
| OASA Telematics API | **Undocumented and unlicensed.** The service's own terms of use say the material is available for **personal use only**, and that reproduction or use for commercial purposes without written permission is criminally prosecuted. | Queried live for arrivals, vehicle positions, schedules and geometry. Never persisted beyond a short in-memory cache, never redistributed. |

The code is AGPL-3.0. The **data is not**: the live feed must only ever be
used for personal, non-commercial purposes, so leoforio must not become a
commercial product.

## Verified statements

Each of these was checked against a primary source on 2026-10-31.

- **The Telematics terms restrict use to personal purposes.** The terms page
  states that the rights to the informational material belong exclusively to
  OASA, that it is made available to the public only for personal use, and
  that any reproduction or use for commercial purposes without prior written
  permission is criminally prosecuted.
  Snapshot: `data-sources/fixtures/oasa/terms-2026-10-31.html`.
- **The GTFS dataset page states no Creative Commons licence.** The page for
  the dataset carries no CC BY-NC 4.0 or any other licence identifier. The
  only licensing language is that the data are provided *as is* and that the
  Athens transport services bear no responsibility for the results of any
  third-party use.
  Snapshot: `data-sources/fixtures/oasa/datagovgr-dataset-2026-10-31.html`.
  Community projects sometimes describe this feed as CC BY-NC 4.0; that
  description was not found on the dataset page and is not relied on here.
- **The dataset lives on data.gov.gr**, the national open data portal, with
  the portal's own terms of use and data dissemination decisions attached.
  This is the basis for treating the static feed as reusable public-sector
  data.

## What leoforio does to stay within these terms

1. **Unofficial and unaffiliated.** The app and this repository state plainly
   that leoforio is not affiliated with, endorsed by or supported by OASA.
2. **Attribution.** OASA is credited as the data source in the repository, in
   the app's credits screen and in the map attribution.
3. **Transient use of the live feed.** Realtime responses are held in an
   in-memory cache for seconds (arrivals 25 s, vehicle positions 15 s per
   route) and are never written to disk, never aggregated into a stored
   dataset and never redistributed to third parties. The app shows the
   timestamp of what it displays, so a user can see how old it is.
4. **Non-commercial.** No advertising, no analytics, no tracking, no paid
   tier, no data sold to anyone. The project stays free and open source.
5. **Polite access.** A descriptive User-Agent, request timeouts, rate
   limiting, and request coalescing so that upstream traffic tracks distinct
   stops and routes rather than user count. The official web application's own
   polling interval (30 s for one stop) is the reference.
6. **Honest failure.** When the live feed is unavailable or its data is stale,
   the app says so. It never invents an arrival time.
7. **Acknowledged instability.** The documentation states that the Telematics
   API is undocumented and may change or disappear without notice, so no
   availability is promised for it.

## Known open questions

- Whether OASA would grant written permission, and on what terms. No request
  has been made. Asking would reduce the ambiguity for every project in this
  space, and is the single most useful legal step available.
- Whether the static GTFS may be used in a commercial fork. The dataset page
  does not say, so the answer is currently "do not".
- The dataset also publishes a `.rar` copy of the feed; the `.zip` copy is
  used.

## Forks

A fork that adds advertising, a paid tier, or any other commercial use is
outside what these terms permit for the live feed. The licence of the code
does not change that: the terms restrict the data, and the data is what makes
the app useful.
