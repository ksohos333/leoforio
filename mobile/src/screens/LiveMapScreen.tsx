import { useQuery } from '@tanstack/react-query';
import { StyleSheet, Text, View } from 'react-native';

import { api } from '../api/client';
import { config } from '../config';
import { VehicleMap } from '../map/VehicleMap';

export function LiveMapScreen() {
  const { data, isPending, isError, dataUpdatedAt } = useQuery({
    queryKey: ['vehicles'],
    queryFn: api.vehicles,
    refetchInterval: config.vehiclesRefreshMs,
  });

  const vehicles = data ?? [];

  return (
    <View style={styles.container}>
      <VehicleMap vehicles={vehicles} />
      <View style={styles.status}>
        {isError ? (
          <Text style={styles.statusText}>
            Live positions are unavailable right now.
          </Text>
        ) : isPending ? (
          <Text style={styles.statusText}>Loading live positions…</Text>
        ) : vehicles.length === 0 ? (
          <Text style={styles.statusText}>
            No vehicles are reporting a position.
          </Text>
        ) : (
          <Text style={styles.statusText}>
            {vehicles.length} vehicles · updated{' '}
            {new Date(dataUpdatedAt).toLocaleTimeString()}
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
    backgroundColor: 'rgba(255,255,255,0.92)',
    borderRadius: 12,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  statusText: {
    color: '#1a1a1a',
    fontSize: 14,
  },
});
