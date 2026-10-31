# ADR 0003: React Native with MapLibre Native

- Status: Accepted
- Date: 2026-10-31
- Deciders: @ksohos333

## Context

v1 ships a live bus map, so the map renderer is a core surface, not a
decoration. The project also targets F-Droid and direct APK distribution
before Google Play, and must avoid proprietary map SDKs to keep the build
transparent and the licensing clean.

Candidates considered: Flutter with `flutter_map`, React Native with
`@maplibre/react-native-maplibre`, and a native Kotlin app with MapLibre
Native.

- Flutter + `flutter_map`: low release friction on F-Droid, but raster-first
  rendering and a thinner plugin ecosystem.
- React Native + MapLibre Native: true vector rendering, GPU composited,
  custom styles, and no Google Play Services dependency (important for a
  F-Droid build). Integration roughness and dependency weight are the costs.
- Kotlin + MapLibre Native: best map quality, but a second app language once
  iOS arrives, which is too much for a solo, back-burner project.

The tie was broken with a one-hour spike per candidate (render OSM raster plus
one live marker, produce a release build). MapLibre Native won on map
fidelity, which is the product's signature surface.

## Decision

- React Native community CLI with TypeScript, no Expo runtime dependency.
- Map rendering via `@maplibre/react-native-maplibre`.
- MapLibre styles and tiles are served by our own backend or a configured
  provider (see ADR 0005), never by a hardcoded third-party demo style.

## Consequences

- The map choice is reversible: backend payloads are framework-independent
  GeoJSON, so switching renderers later touches the app layer only.
- Native dependency weight (MapLibre, Hermes) increases Gradle build time and
  F-Droid reproducibility effort.
- Package compatibility with the current React Native version is verified in
  P0 before any map work begins.
