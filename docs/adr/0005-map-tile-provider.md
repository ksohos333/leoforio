# ADR 0005: Map tiles come from a configured provider, with attribution

- Status: Accepted
- Date: 2026-10-31
- Deciders: @ksohos333

## Context

The map needs raster or vector tiles. The OpenStreetMap Foundation's public
tile server (`tile.openstreetmap.org`) explicitly disallows app-scale and
bulk use, and a live map with vehicle markers refreshing every 15-30 seconds
generates exactly that kind of load. Serving our own tiles is possible but
carries operational and bandwidth cost out of proportion to a solo project.

## Decision

- Tiles and styles come from a provider with a free tier suitable for
  low-volume app use, or from self-hosted tiles if usage grows past that.
- The provider is configuration, not code: the tile URL and style URL are
  supplied by deployment configuration.
- Required attribution is rendered by the app, in the style and in an
  in-app credits screen.
- The provider's usage terms are captured in `docs/data-sources/`, with the
  same discipline applied to OASA sources.

## Consequences

- There is a per-request cost and a third-party dependency in the map layer.
- Provider quotas can be exceeded by a popular build; the map must degrade
  gracefully (attribution only, no tiles) rather than crash.
- Reusing the official OSM tile server without a provider agreement is out of
  scope for releases.
