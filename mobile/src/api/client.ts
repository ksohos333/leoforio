import { config } from '../config';
import type { Arrival, VehiclePosition } from './types';

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

export const api = {
  arrivals: (stopId: string): Promise<Arrival[]> =>
    getJson<Arrival[]>(`/api/v1/stops/${encodeURIComponent(stopId)}/arrivals`),

  vehicles: (): Promise<VehiclePosition[]> => getJson<VehiclePosition[]>('/api/v1/vehicles'),
};
