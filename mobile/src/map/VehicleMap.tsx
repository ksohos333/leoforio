import {
  Camera,
  GeoJSONSource,
  Layer,
  Map,
} from '@maplibre/maplibre-react-native';
import type { Feature, FeatureCollection, Point } from 'geojson';
import { useMemo } from 'react';
import { StyleSheet } from 'react-native';

import type { LineString, Stop, VehiclePosition } from '../api/types';
import { athensCenter, config } from '../config';
import { rasterStyle } from './mapStyle';

interface VehicleMapProps {
  geometry?: LineString;
  vehicles: VehiclePosition[];
  stops: Stop[];
}

export function VehicleMap({ geometry, vehicles, stops }: VehicleMapProps) {
  const vehicleData = useMemo<FeatureCollection<Point>>(
    () => ({
      type: 'FeatureCollection',
      features: vehicles.map(
        (vehicle): Feature<Point> => ({
          type: 'Feature',
          id: vehicle.vehicleId,
          properties: {
            lineId: vehicle.lineId,
            recordedAt: vehicle.recordedAt,
          },
          geometry: {
            type: 'Point',
            coordinates: [vehicle.position.lon, vehicle.position.lat],
          },
        }),
      ),
    }),
    [vehicles],
  );

  const stopData = useMemo<FeatureCollection<Point>>(
    () => ({
      type: 'FeatureCollection',
      features: stops.map(
        (busStop): Feature<Point> => ({
          type: 'Feature',
          id: busStop.id,
          properties: { name: busStop.name },
          geometry: {
            type: 'Point',
            coordinates: [busStop.lon, busStop.lat],
          },
        }),
      ),
    }),
    [stops],
  );

  const camera = useMemo(() => {
    const coordinates = geometry?.geometry.coordinates ?? [];
    if (!coordinates.length) {
      return {
        zoom: 11,
        center: [athensCenter.longitude, athensCenter.latitude] as [number, number],
      };
    }

    const longitudes = coordinates.map(coordinate => coordinate[0]);
    const latitudes = coordinates.map(coordinate => coordinate[1]);
    const minLongitude = Math.min(...longitudes);
    const maxLongitude = Math.max(...longitudes);
    const minLatitude = Math.min(...latitudes);
    const maxLatitude = Math.max(...latitudes);

    const span = Math.max(maxLongitude - minLongitude, maxLatitude - minLatitude);
    const zoom = Math.max(10, Math.min(16, Math.round(Math.log2(360 / Math.max(span, 0.002)))));

    return {
      zoom,
      center: [
        (minLongitude + maxLongitude) / 2,
        (minLatitude + maxLatitude) / 2,
      ] as [number, number],
    };
  }, [geometry]);

  return (
    <Map
      mapStyle={rasterStyle(config.tileUrl, config.tileAttribution)}
      style={StyleSheet.absoluteFill}>
      <Camera zoom={camera.zoom} center={camera.center} />
      {geometry ? (
        <>
          <GeoJSONSource id="route" data={geometry} />
          <Layer
            id="route"
            type="line"
            source="route"
            paint={{
              'line-color': '#2f6fed',
              'line-width': 4,
              'line-opacity': 0.85,
            }}
          />
        </>
      ) : null}
      <GeoJSONSource id="stops" data={stopData} />
      <Layer
        id="stops"
        type="circle"
        source="stops"
        paint={{
          'circle-radius': 3,
          'circle-color': '#1f2933',
          'circle-stroke-width': 1,
          'circle-stroke-color': '#ffffff',
        }}
      />
      <GeoJSONSource id="vehicles" data={vehicleData} />
      <Layer
        id="vehicles"
        type="circle"
        source="vehicles"
        paint={{
          'circle-radius': 7,
          'circle-color': '#e5484d',
          'circle-stroke-width': 2,
          'circle-stroke-color': '#ffffff',
        }}
      />
    </Map>
  );
}
