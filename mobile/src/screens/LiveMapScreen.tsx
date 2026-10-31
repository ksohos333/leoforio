import { useQuery } from '@tanstack/react-query';
import { StyleSheet, Text, View } from 'react-native';

import { api } from '../api/client';
import type { Route } from '../api/types';
import { config } from '../config';
import { VehicleMap } from '../map/VehicleMap';

interface LiveMapScreenProps {
  route: Route;
}

export function LiveMapScreen({ route }: LiveMapScreenProps) {
  const vehicles = useQuery({
    queryKey: ['vehicles', route.id],
    queryFn: () => api.vehicles(route.id),
    refetchInterval: config.vehiclesRefreshMs,
  });

  const geometry = useQuery({
    queryKey: ['geometry', route.id],
    queryFn: () => api.routeGeometry(route.id),
  });

  const stops = useQuery({
    queryKey: ['stops', route.id],
    queryFn: () => api.routeStops(route.id),
  });

  const liveVehicles = vehicles.data ?? [];

  return (
    <View style={styles.container}>
      <VehicleMap
        geometry={geometry.data}
        vehicles={liveVehicles}
        stops={stops.data ?? []}
      />
      <View style={styles.status}>
        {vehicles.isError ? (
          <Text style={styles.statusText}>
            Live positions are unavailable right now.
          </Text>
        ) : vehicles.isPending ? (
          <Text style={styles.statusText}>Loading live positions…</Text>
        ) : liveVehicles.length === 0 ? (
          <Text style={styles.statusText}>
            No vehicles are reporting a position on this route.
          </Text>
        ) : (
          <Text style={styles.statusText}>
            {liveVehicles.length} vehicle
            {liveVehicles.length === 1 ? '' : 's'} · updated{' '}
            {new Date(vehicles.dataUpdatedAt).toLocaleTimeString()}
          </Text>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  status: {
    position: 'absolute',
    left: 16,
    right: 16,
    bottom: 32,
    backgroundColor: 'rgba(255,255,255,0.94)',
    borderRadius: 12,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  statusText: {
    color: '#1f2933',
    fontSize: 14,
  },
});
