/**
 * Types mirroring the backend JSON payloads.
 *
 * These mirror the Go types in backend/internal/model. Changing a field
 * there means changing it here in the same commit.
 */

export interface Position {
  lat: number;
  lon: number;
}

export interface Line {
  id: string;
  code: string;
  name: string;
}

export interface Route {
  id: string;
  lineId: string;
  direction: string;
  name: string;
  stopIds: string[] | null;
}

export interface Stop {
  id: string;
  code: string;
  name: string;
  lat: number;
  lon: number;
  distanceKm?: number;
}

export interface Arrival {
  stopId: string;
  lineId: string;
  routeId: string;
  vehicleId?: string;
  scheduled?: string;
  expected: string;
}

export interface VehiclePosition {
  vehicleId: string;
  lineId: string;
  routeId: string;
  position: Position;
  heading?: number;
  recordedAt: string;
}

export interface LineString {
  type: 'Feature';
  geometry: {
    type: 'LineString';
    coordinates: [number, number][];
  };
  properties: Record<string, unknown>;
}
