# ADR 0004: The app never talks to OASA directly

- Status: Accepted
- Date: 2026-10-31
- Deciders: @ksohos333

## Context

OASA's endpoints are a free public service with no SLA and unknown rate
limits. If every user's phone polled OASA on its own, N users would mean N
upstream requests per stop, which is effectively a denial-of-service against a
public utility, and would get the project's users blocked.

## Decision

All upstream access lives in the backend, behind source adapters
(`internal/source/...`). The mobile app calls only the leoforio API.

- Upstream responses are cached with a TTL (arrivals 20-30s, static data
  daily).
- Concurrent requests for the same key are coalesced into one upstream call.
- The backend identifies itself with a descriptive User-Agent and applies
  per-upstream rate limiting and timeouts.

## Consequences

- Upstream traffic is proportional to distinct stops requested, not to user
  count.
- The app keeps working from cache when OASA is degraded or unreachable, and
  can show a last-known timestamp instead of inventing an ETA.
- The backend is a required component: the app is not useful without it,
  which makes its availability a project responsibility.
