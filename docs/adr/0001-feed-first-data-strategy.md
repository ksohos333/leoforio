# ADR 0001: Feed-first data strategy

- Status: Accepted
- Date: 2026-10-31
- Deciders: @ksohos333

## Context

OASA, the Athens public transport authority, exposes no documented, versioned,
SLA-backed API for consumers. Several access paths exist (official GTFS feed,
community-documented JSON endpoints, undocumented endpoints only reachable from
the official mobile app). Each path carries different legal, stability and
reliability properties, and the choice of path shapes the entire product.

## Decision

Sources are consumed through a fixed ladder, and every rung below the first
must be justified in the decision log with a capture date and evidence:

1. Official feed or documented API
2. Community-documented public endpoint
3. Endpoint discovered by observing our own traffic to the public service
   (never by bypassing protections)
4. Feature dropped from the release

The ladder is recorded per source in `docs/data-sources/README.md`. A source
that is not on the ladder is not used.

## Consequences

- The project depends on infrastructure it does not control. Adapter
  isolation, caching and fixture-based contract tests are mandatory, not
  optional.
- Discovery work is scoped and timeboxed (see ADR 0007) instead of becoming
  the default way to obtain data.
- Upstream terms of service are captured alongside every source, including
  the ones we choose not to use.
