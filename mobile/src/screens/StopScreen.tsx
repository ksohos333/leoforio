import { useQuery } from '@tanstack/react-query';
import { StyleSheet, Text, View } from 'react-native';

import { api } from '../api/client';
import type { Stop } from '../api/types';
import { config } from '../config';

interface StopScreenProps {
  stop: Stop;
}

/**
 * The arrivals at a stop.
 *
 * The upstream publishes the remaining minutes rather than a time, and it
 * sometimes has nothing to report for a stop. Both cases are shown as they
 * are: the screen never invents an arrival.
 */
export function StopScreen({ stop }: StopScreenProps) {
  const arrivals = useQuery({
    queryKey: ['arrivals', stop.id],
    queryFn: () => api.arrivals(stop.id),
    refetchInterval: config.arrivalsRefreshMs,
  });

  const expected = arrivals.data ?? [];

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.name}>{stop.name}</Text>
        <Text style={styles.code}>stop {stop.code}</Text>
      </View>

      {arrivals.isError ? (
        <Text style={styles.message}>
          Arrivals are unavailable right now. The operator may be having
          trouble, or the network is.
        </Text>
      ) : arrivals.isPending ? (
        <Text style={styles.message}>Loading arrivals…</Text>
      ) : expected.length === 0 ? (
        <Text style={styles.message}>
          No buses are reported for this stop right now. That happens late at
          night and on lines where the operator has no live data.
        </Text>
      ) : (
        <View style={styles.list}>
          {expected.map((arrival, index) => (
            <View key={`${arrival.routeId}-${arrival.vehicleId}-${index}`} style={styles.row}>
              <Text style={styles.route}>{arrival.routeId}</Text>
              <Text style={styles.minutes}>
                {minutesUntil(arrival.expected)}
              </Text>
            </View>
          ))}
          <Text style={styles.updated}>
            updated {new Date(arrivals.dataUpdatedAt).toLocaleTimeString()}
          </Text>
        </View>
      )}
    </View>
  );
}

/**
 * Formats the remaining time as a whole number of minutes, rounded up, so
 * that a bus that is due now reads as "now" rather than "0 min ago".
 */
function minutesUntil(expected: string): string {
  const milliseconds = new Date(expected).getTime() - Date.now();
  const minutes = Math.ceil(milliseconds / 60000);

  if (minutes <= 0) {
    return 'due now';
  }
  if (minutes === 1) {
    return '1 min';
  }
  return `${minutes} min`;
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  header: {
    paddingHorizontal: 16,
    paddingTop: 16,
    paddingBottom: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#d8dce2',
  },
  name: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1f2933',
  },
  code: {
    fontSize: 12,
    color: '#7b8794',
    marginTop: 2,
  },
  list: {
    paddingHorizontal: 16,
    paddingTop: 12,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 14,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#eceff2',
  },
  route: {
    fontSize: 16,
    color: '#1f2933',
  },
  minutes: {
    fontSize: 18,
    fontWeight: '600',
    color: '#2f6fed',
  },
  updated: {
    marginTop: 14,
    fontSize: 12,
    color: '#7b8794',
  },
  message: {
    padding: 20,
    fontSize: 14,
    lineHeight: 21,
    color: '#3e4c59',
  },
});
