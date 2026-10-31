/**
 * Runtime configuration for the leoforio app.
 *
 * The values below are development defaults: 10.0.2.2 reaches the host
 * machine from an Android emulator, and the tile URL points at the public
 * OpenStreetMap tiles, which is fine for local development but not for a
 * release. Release builds must point the tile URL at a provider that permits
 * app-scale use and the API base URL at the deployed backend, per ADR 0005.
 */
export const config = {
  apiBaseUrl: 'http://10.0.2.2:8080',
  vehiclesRefreshMs: 15_000,
  requestTimeoutMs: 10_000,
  tileUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
  tileAttribution: '© OpenStreetMap contributors',
};

export const athensCenter = {
  longitude: 23.7275,
  latitude: 37.9838,
};
