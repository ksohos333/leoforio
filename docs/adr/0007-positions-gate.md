# ADR 0007: Live positions have an explicit gate and three outcomes

- Status: Accepted
- Date: 2026-10-31
- Deciders: @ksohos333

## Context

The live bus map is the product's signature surface, but live vehicle
positions are exactly the data OASA is least likely to expose openly. It is
possible that no public positions feed exists, in which case the map either
depends on an undocumented endpoint or cannot show live vehicles at all.

## Decision

P1 recon resolves the positions question with exactly three outcomes, recorded
in the decision log:

- **A: public source found.** A positions adapter is built on that source.
- **B: research required.** A timeboxed discovery phase (about three focused
  sessions) inspects the public web app's script bundle and, only if needed,
  the official app's static artefacts. Discovery stops at protection walls;
  bypassing protections is out of scope.
- **C: no usable source.** The v1 map shows route geometry and timetables
  only, with an explicit "no live data" state. Live vehicles move to a later
  release, and the scope change is recorded as an ADR.

Discovery output is documentation and fixtures. No research artefact becomes a
runtime dependency.

## Outcome: branch A (2026-10-31)

Live vehicle positions are published by the operator's own public API.
`getBusLocation&p1=<routeCode>` returns the position, heading and timestamp of
every vehicle on a route, and it was verified to return fresh data during
recon. Arrivals come from `getStopArrivals&p1=<stopCode>`, and route geometry
from `webGetRoutesDetailsAndStops`.

Because a public source exists, the research phase is not needed: no
observation of the official mobile application and no inspection of its
artefacts was required, and the app is never derived from them.

Positions are published per route rather than for the whole fleet, so the
backend queries only the routes a user is looking at, caches each route for a
short time, and never sweeps all 486 routes for a single map view.

## Consequences

- The live map's viability is decided before the map is built, so the map
  work is never speculative.
- Outcome C is a legitimate release, not a failure: honest timetables and
  arrivals still deliver the core promise.
- If upstream terms prohibit republishing positions, the adapter ships
  disabled and the constraint is documented.
