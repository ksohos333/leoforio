import { useQueries } from '@tanstack/react-query';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { api } from '../api/client';
import type { Arrival } from '../api/types';
import { config } from '../config';
import type { Favourite } from '../favourites/useFavourites';

interface FavouriteStopsProps {
  favourites: Favourite[];
  onSelect: (favourite: Favourite) => void;
  onRemove: (id: string) => void;
}

/**
 * The stops the user keeps an eye on, each with the next bus.
 *
 * This is the screen the app is really for: open it and see when the bus
 * comes, without searching for anything.
 */
export function FavouriteStops({
  favourites,
  onSelect,
  onRemove,
}: FavouriteStopsProps) {
  const queries = useQueries({
    queries: favourites.map(favourite => ({
      queryKey: ['arrivals', favourite.id],
      queryFn: () => api.arrivals(favourite.id),
      refetchInterval: config.arrivalsRefreshMs,
    })),
  });

  if (!favourites.length) {
    return null;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.heading}>My stops</Text>
      {favourites.map((favourite, index) => {
        const query = queries[index];
        const next = (query?.data ?? [])[0];

        return (
          <Pressable
            key={favourite.id}
            style={styles.row}
            onPress={() => onSelect(favourite)}
            accessibilityRole="button"
            accessibilityLabel={`Next bus at ${favourite.name}`}>
            <View style={styles.text}>
              <Text style={styles.name}>{favourite.name}</Text>
              <Text style={styles.next}>
                {next ? describe(next) : 'no live data'}
              </Text>
            </View>
            <Pressable
              style={styles.remove}
              onPress={() => onRemove(favourite.id)}
              accessibilityRole="button"
              accessibilityLabel={`Remove ${favourite.name}`}>
              <Text style={styles.removeText}>×</Text>
            </Pressable>
          </Pressable>
        );
      })}
    </View>
  );
}

/**
 * Describes the next arrival in one line: the line and how long to wait.
 */
function describe(arrival: Arrival): string {
  const line = arrival.lineId || arrival.routeId;
  return `${line} · ${minutesUntil(arrival.expected)}`;
}

export function minutesUntil(expected: string): string {
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
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#d8dce2',
    paddingBottom: 8,
  },
  heading: {
    paddingHorizontal: 16,
    paddingTop: 4,
    paddingBottom: 6,
    fontSize: 13,
    fontWeight: '700',
    color: '#7b8794',
    letterSpacing: 0.6,
    textTransform: 'uppercase',
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  text: {
    flex: 1,
  },
  name: {
    fontSize: 15,
    color: '#1f2933',
  },
  next: {
    fontSize: 13,
    color: '#2f6fed',
    marginTop: 2,
  },
  remove: {
    paddingHorizontal: 10,
    paddingVertical: 4,
  },
  removeText: {
    fontSize: 22,
    color: '#7b8794',
  },
});
