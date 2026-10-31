import {
  Camera,
  GeoJSONSource,
  Layer,
  Map,
} from '@maplibre/maplibre-react-native';
import { StyleSheet } from 'react-native';

import type { VehiclePosition } from '../api/types';
import { athensCenter, config } from '../config';
import { rasterStyle } from './mapStyle';

interface VehicleMapProps {
  vehicles: VehiclePosition[];
}

export function VehicleMap({ vehicles }: VehicleMapProps) {
  const data = {
    type: 'FeatureCollection' as const,
    features: vehicles.map(vehicle => ({
      type: 'Feature' as const,
      id: vehicle.vehicleId,
      properties: {
        lineId: vehicle.lineId,
        recordedAt: vehicle.recordedAt,
      },
      geometry: {
        type: 'Point' as const,
        coordinates: [vehicle.position.lon, vehicle.position.lat],
      },
    })),
  };

  return (
    <Map
      mapStyle={rasterStyle(config.tileUrl, config.tileAttribution)}
      style={StyleSheet.absoluteFill}>
      <Camera
        zoom={11}
        center={[athensCenter.longitude, athensCenter.latitude]}
      />
      <GeoJSONSource id="vehicles" data={data} />
      <Layer
        id="vehicles"
        type="circle"
        source="vehicles"
        paint={{
          'circle-radius': 6,
          'circle-color': '#2f6fed',
          'circle-stroke-width': 2,
          'circle-stroke-color': '#ffffff',
        }}
      />
    </Map>
  );
}
