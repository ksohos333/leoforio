/**
 * Runtime configuration for the leoforio app.
 *
 * The API base URL points at the loopback address, which is where the
 * development backend is reached from a device after running
 * `adb reverse tcp:8080 tcp:8080`. On the Android emulator the host is
 * reachable at 10.0.2.2 instead.
 *
 * The tile URL is a development default: the public OpenStreetMap tiles are
 * fine for local work but not for a release, which must use a provider that
 * permits app-scale use (ADR 0005).
 */
export const config = {
  apiBaseUrl: 'http://127.0.0.1:8080',
  vehiclesRefreshMs: 15_000,
  arrivalsRefreshMs: 20_000,
  requestTimeoutMs: 10_000,
  tileUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
  tileAttribution: '© OpenStreetMap contributors',
};

export const athensCenter = {
  longitude: 23.7275,
  latitude: 37.9838,
};
