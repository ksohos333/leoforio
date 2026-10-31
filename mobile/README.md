# leoforio mobile

React Native app for Greek public transport, starting with the Athens (OASA)
network. Uses MapLibre Native over OpenStreetMap tiles, and talks only to the
leoforio backend, never to an upstream operator directly (ADR 0004).

## Requirements

- Node 22.11 or newer
- JDK 17
- Android SDK with platform 35 and build tools (for local builds)
- A running leoforio backend (see [`backend/`](../backend/README.md))

## Setup

```bash
npm ci
npm start            # Metro bundler
npm run android      # build and run on a device or emulator
```

The API base URL, the tile URL and the refresh interval live in
[`src/config.ts`](./src/config.ts). The defaults are development values:
`10.0.2.2` reaches the host from an Android emulator, and a physical device
needs the host's LAN address instead. Release builds must point the tile URL
at a provider that permits app-scale use (ADR 0005).

## Checks

```bash
npx tsc --noEmit
npx eslint .
npx jest
```

## Layout

| Path | Purpose |
|---|---|
| `src/api/types.ts` | types mirroring the backend JSON payloads |
| `src/api/client.ts` | typed HTTP client with request timeout |
| `src/map/mapStyle.ts` | raster MapLibre style built from the configured tile URL |
| `src/map/VehicleMap.tsx` | map with route path, stops and vehicle markers |
| `src/screens/SearchScreen.tsx` | stop search and the stops near you |
| `src/screens/StopScreen.tsx` | arrivals at a stop, with the time until each bus |
| `src/screens/LinesScreen.tsx` | the lines the operator publishes |
| `src/screens/RoutesScreen.tsx` | the routes of a line |
| `src/screens/LiveMapScreen.tsx` | live map with polling and honest empty states |
| `src/screens/CreditsScreen.tsx` | disclaimer and attribution to OASA |
| `jest.setup.js` | mocks for the MapLibre native module and `fetch` |

The app asks for the coarse location only, and only when the user taps
"Near me". It is never requested in the background and never required to use
the app.

## Notes

- The map library is `@maplibre/maplibre-react-native`. Earlier packages named
  `react-native-maplibre-gl` and `@maplibre/react-native-maplibre` no longer
  exist on npm; the current package exports `Map`, `Camera`, `GeoJSONSource`
  and `Layer`, and layer paint uses style-spec keys such as `circle-radius`.
- Push notifications are planned for v1.1 and are not wired up yet.
- The location comes from `@react-native-community/geolocation`, which is
  loaded lazily so that it is only needed on a device.
