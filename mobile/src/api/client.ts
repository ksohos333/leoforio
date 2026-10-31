import { config } from '../config';
import type {
  Arrival,
  Line,
  LineString,
  Route,
  Stop,
  VehiclePosition,
} from './types';

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function getJson<T>(path: string): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), config.requestTimeoutMs);

  try {
    const response = await fetch(`${config.apiBaseUrl}${path}`, {
      signal: controller.signal,
    });

    if (!response.ok) {
      const body = await response.text();
      throw new ApiError(response.status, body || response.statusText);
    }

    return (await response.json()) as T;
  } finally {
    clearTimeout(timeout);
  }
}

const route = (routeCode: string) => encodeURIComponent(routeCode);
const stop = (stopCode: string) => encodeURIComponent(stopCode);

export const api = {
  lines: (): Promise<Line[]> => getJson<Line[]>('/api/v1/lines'),

  routes: (lineCode: string): Promise<Route[]> =>
    getJson<Route[]>(`/api/v1/routes?line=${encodeURIComponent(lineCode)}`),

  routeStops: (routeCode: string): Promise<Stop[]> =>
    getJson<Stop[]>(`/api/v1/routes/${route(routeCode)}/stops`),

  routeGeometry: (routeCode: string): Promise<LineString> =>
    getJson<LineString>(`/api/v1/routes/${route(routeCode)}/geometry`),

  vehicles: (routeCode: string): Promise<VehiclePosition[]> =>
    getJson<VehiclePosition[]>(`/api/v1/routes/${route(routeCode)}/vehicles`),

  arrivals: (stopCode: string): Promise<Arrival[]> =>
    getJson<Arrival[]>(`/api/v1/stops/${stop(stopCode)}/arrivals`),

  stop: (stopCode: string): Promise<Stop> =>
    getJson<Stop>(`/api/v1/stops/${stop(stopCode)}`),

  searchStops: (query: string): Promise<Stop[]> =>
    getJson<Stop[]>(`/api/v1/stops/search?q=${encodeURIComponent(query)}`),

  nearbyStops: (
    latitude: number,
    longitude: number,
    radiusKm = 1,
  ): Promise<Stop[]> =>
    getJson<Stop[]>(
      `/api/v1/stops/nearby?lat=${latitude}&lon=${longitude}&radius=${radiusKm}`,
    ),
};
