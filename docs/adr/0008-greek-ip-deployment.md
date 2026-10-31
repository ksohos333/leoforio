# ADR 0008: The backend runs from a Greek IP address

- Status: Accepted
- Date: 2026-10-31
- Deciders: @ksohos333

## Context

The OASA Telematics service answers requests from a Greek address and drops
connections from elsewhere without an answer. The static GTFS feed on
data.gov.gr is not restricted in this way.

leoforio's architecture already puts all upstream access in the backend, so
the restriction applies to one service rather than to every user's phone.
That is a benefit: the phone is never blocked by the operator's geography,
and the backend can be placed where the upstream is reachable.

## Decision

- The backend is deployed on a host with a Greek IP address. A small VPS in
  Greece is the straightforward option; a serverless platform with a
  placement hint for Athens is the alternative.
- The app is never required to reach the upstream itself, so the app works
  for users wherever they are.
- The upstream stays behind the adapters, so if the restriction is ever
  lifted the deployment does not change.

## Consequences

- Hosting has a floor cost that a free tier outside Greece cannot meet.
- The upstream sees one client's traffic from one address rather than many
  clients spread over many addresses, which makes the traffic pattern
  smaller and easier to recognise as legitimate.
- Development from outside Greece needs a relay. The upstream cannot be
  called directly during development, only through the deployed backend.
